package writer

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/cuckooemm/clog"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	backupTimeFormat = "2006-01-02T15-04-05.000"
	compressSuffix   = ".gz"
)

var (
	currentTime = time.Now
)

type FileWrite struct {
	path       string
	dirPath    string
	name       string
	maxSize    int64 // 文件最大大小
	lastSize   int64 // 剩余可写空间
	maxLine    int64 // 文件最大可写行
	lastLine   int64 // 文件剩余可写行
	maxAge     int64 // 备份文件保存时间
	maxBackups int64 // 备份文件数量
	compress   bool  // 备份文件是否压缩
	// 多久后的文件开启压缩   文件压缩后缀
	file      *os.File
	mu        sync.Mutex
	millCh    chan struct{}
	startMill sync.Once
}

func newFileWrite(path string, size, line, age, backups int64, compress bool) *FileWrite {
	var fw = new(FileWrite)
	fw.path = path
	fw.dirPath = filepath.Dir(path)
	if err := os.MkdirAll(fw.dirPath, os.ModeDir); err != nil {
		panic(err)
	}
	fw.name = filepath.Base(path)
	fw.maxAge = age
	fw.maxBackups = backups
	fw.maxSize = size
	if fw.maxSize <= 0 {
		fw.maxSize = math.MaxInt64
	}
	fw.maxLine = line
	if fw.maxLine <= 0 {
		fw.maxLine = math.MaxInt64
	}
	if fw.maxAge > 0 || fw.compress || fw.maxBackups > 0 {
		fw.millCh = make(chan struct{}, 1)
		go fw.millRun()
		fw.mill()
	}
	fw.compress = compress

	if err := fw.firstOpenExistOrNew(); err != nil {
		panic(err)
	}

	return fw
}

func (fw *FileWrite) WriteLevel(level clog.Level, p []byte) (n int, err error) {
	return fw.Write(p)
}

func (fw *FileWrite) Write(p []byte) (n int, err error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	var writeLen = int64(len(p))
	if writeLen > fw.lastSize {
		if err := fw.rotate(); err != nil {
			return 0, err
		}
	}
	if fw.lastLine <= 0 {
		if err := fw.rotate(); err != nil {
			return 0, err
		}
	}
	fw.lastLine--
	n, err = fw.file.Write(p)
	fw.lastSize -= int64(n)
	return n, err
}

func (fw *FileWrite) firstOpenExistOrNew() error {
	var (
		info os.FileInfo
		err  error
	)
	if info, err = os.Stat(fw.path); err != nil {
		if os.IsNotExist(err) {
			return fw.openNew()
		}
		return fmt.Errorf("error getting log file info: %s", err.Error())
	}

	if fw.file, err = os.OpenFile(fw.path, os.O_APPEND|os.O_RDWR, 0644); err != nil {

		// open old log to failed - ignore
		// open a new log file.
		return fw.openNew()
	}

	if info.Size() >= fw.maxSize {
		return fw.rotate()
	}
	fw.lastSize = fw.maxSize - info.Size()
	// 得到文件当前行数
	if fw.lastLine, err = lineCounter(fw.file); err != nil {
		return err
	}

	if fw.lastLine >= fw.maxLine {
		return fw.rotate()
	}
	// 计算剩余行数
	fw.lastLine = fw.maxLine - fw.lastLine
	return nil
}

func (fw *FileWrite) rotate() error {
	if err := fw.openNew(); err != nil {
		return err
	}
	fw.mill()
	return nil
}

func (fw *FileWrite) mill() {
	select {
	case fw.millCh <- struct{}{}:
	default:
	}
}

func (fw *FileWrite) millRun() {
	for range fw.millCh {
		fw.millRunOnce()
	}
}

func (fw *FileWrite) millRunOnce() {
	files, err := fw.oldLogFiles()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "clog: get old log file err %s\n", err.Error())
		return
	}
	var compress, remove []logInfo

	if fw.maxBackups > 0 && fw.maxBackups < int64(len(files)) {
		preserved := make(map[string]struct{})
		var remaining []logInfo
		for _, f := range files {
			// Only count the uncompressed log file or the
			// compressed log file, not both.
			fn := f.Name()
			if strings.HasSuffix(fn, compressSuffix) {
				fn = fn[:len(fn)-len(compressSuffix)]
			}
			preserved[fn] = struct{}{}
			if int64(len(preserved)) > fw.maxBackups {
				remove = append(remove, f)
			} else {
				remaining = append(remaining, f)
			}
		}
		files = remaining
	}
	if fw.maxAge > 0 {
		cutoff := currentTime().AddDate(0, 0, int(-fw.maxAge))
		var remaining []logInfo
		for _, f := range files {
			if f.timestamp.Before(cutoff) {
				remove = append(remove, f)
			} else {
				remaining = append(remaining, f)
			}
		}
		files = remaining
	}

	// 压缩7天后的文件
	if fw.compress {
		compressTime := currentTime().AddDate(0, 0, -7)
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), compressSuffix) && f.timestamp.Before(compressTime) {
				compress = append(compress, f)
			}
		}
	}

	for _, f := range remove {
		if err := os.Remove(filepath.Join(fw.dirPath, f.Name())); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "clog: remove old log file %s err %s\n", f.Name(), err.Error())
		}
	}
	for _, f := range compress {
		fn := filepath.Join(fw.dirPath, f.Name())
		if err := compressLogFile(fn, fn+compressSuffix); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "clog: compress log file %s err %s\n", f.Name(), err.Error())
		}
	}
	return
}

func (fw *FileWrite) oldLogFiles() ([]logInfo, error) {
	files, err := ioutil.ReadDir(fw.dirPath)

	if err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	var logFiles []logInfo
	prefix, ext := fw.prefixAndExt()
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if t, err := fw.timeFromName(f.Name(), prefix, ext); err == nil {
			logFiles = append(logFiles, logInfo{t, f})
			continue
		}
		if t, err := fw.timeFromName(f.Name(), prefix, ext+compressSuffix); err == nil {
			logFiles = append(logFiles, logInfo{t, f})
			continue
		}
		// 如果有错误  则不是生成的文件
	}

	sort.Sort(byFormatTime(logFiles))
	return logFiles, nil
}

// openNew opens a new log file for writing, moving any old log file out of the
// way.  This methods assumes the file has already been closed.
func (fw *FileWrite) openNew() error {
	mode := os.FileMode(0600)
	info, err := os.Stat(fw.path)
	if fw.file != nil {
		if err := fw.file.Close(); err != nil {
			return err
		}
	}
	// 文件存在 备份此文件
	if err == nil {
		// Copy the mode off the old logfile.
		mode = info.Mode()
		// move the existing file
		if err := os.Rename(fw.path, fw.backupName()); err != nil {
			return fmt.Errorf("can't rename log file: %s", err)
		}
	}
	f, err := os.OpenFile(fw.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}
	fw.file = f
	fw.lastSize = fw.maxSize
	fw.lastLine = fw.maxLine
	return nil
}

func (fw *FileWrite) prefixAndExt() (prefix, ext string) {
	ext = filepath.Ext(fw.name)
	prefix = fw.name[:len(fw.name)-len(ext)]
	return prefix + "-", ext
}

func (fw *FileWrite) timeFromName(filename, prefix, ext string) (time.Time, error) {
	if !strings.HasPrefix(filename, prefix) {
		return time.Time{}, errors.New("mismatched prefix")
	}
	if !strings.HasSuffix(filename, ext) {
		return time.Time{}, errors.New("mismatched extension")
	}
	ts := filename[len(prefix) : len(filename)-len(ext)]
	return time.Parse(backupTimeFormat, ts)
}

// backupName creates a new filename from the given name, inserting a timestamp
// between the filename and the extension, using the local time if requested
// (otherwise UTC).
func (fw *FileWrite) backupName() string {
	prefix, ext := fw.prefixAndExt()
	timestamp := currentTime().Format(backupTimeFormat)
	return filepath.Join(fw.dirPath, fmt.Sprintf("%s%s%s", prefix, timestamp, ext))
}

// compressLogFile compresses the given log file, removing the
// uncompressed log file if successful.
func compressLogFile(src, dst string) (err error) {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer f.Close()

	fi, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat log file: %v", err)
	}

	//if err := chown(dst, fi); err != nil {
	//	return fmt.Errorf("failed to chown compressed log file: %v", err)
	//}

	// If this file already exists, we presume it was created by
	// a previous attempt to compress the log file.
	gzf, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
	if err != nil {
		return fmt.Errorf("failed to open compressed log file: %v", err)
	}
	defer gzf.Close()

	gz := gzip.NewWriter(gzf)

	defer func() {
		if err != nil {
			os.Remove(dst)
			err = fmt.Errorf("failed to compress log file: %v", err)
		}
	}()

	if _, err := io.Copy(gz, f); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	if err := gzf.Close(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Remove(src); err != nil {
		return err
	}

	return nil
}

func lineCounter(r io.Reader) (int64, error) {
	var (
		buf           = make([]byte, 32*1024)
		count   int64 = 0
		lineSep       = []byte{'\n'}
	)

	for {
		c, err := r.Read(buf)
		count += int64(bytes.Count(buf[:c], lineSep))
		switch {
		case err == io.EOF:
			return count, nil
		case err != nil:
			return count, err
		}
	}
}
