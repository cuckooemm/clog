package storage

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/cuckooemm/clog"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	backupTimeFormat = "20060102150405"
	compressSuffix   = ".gz"
)

var currentTime = time.Now

type SizeRotate struct {
	path          string
	dirPath       string
	name          string
	maxSize       int  // 文件最大大小
	lastSize      int  // 剩余可写空间
	maxLine       int  // 文件最大可写行
	lastLine      int  // 文件剩余可写行
	saveDay       int  // 备份文件保存时间
	maxBackups    int  // 备份文件数量
	compress      bool // 备份文件是否压缩
	compressAfter int  // 几天后的日志进行压缩
	fd            *os.File
	mu            *sync.Mutex
	millCh        chan struct{}
}

func (r *SizeRotate) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.millCh != nil {
		close(r.millCh)
	}
	_ = r.fd.Sync()
	_ = r.fd.Close()
}
func (r *SizeRotate) WriteLevel(level clog.Level, p []byte) (n int, err error) {
	return r.Write(p)
}

func (r *SizeRotate) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.maxSize > 0 {
		if len(p) > r.lastSize {
			if err = r.rotate(); err != nil {
				return 0, err
			}
		}
		r.lastSize -= len(p)
	}
	if r.maxLine > 0 {
		if r.lastLine <= 0 {
			if err = r.rotate(); err != nil {
				return 0, err
			}
		}
		r.lastLine--
	}
	n, err = r.fd.Write(p)
	return n, err
}

func (r *SizeRotate) firstOpenExistOrNew() error {
	var (
		info os.FileInfo
		err  error
	)
	if info, err = os.Stat(r.path); err != nil {
		if os.IsNotExist(err) {
			return r.openNew()
		}
		return fmt.Errorf("error getting log file info: %s", err.Error())
	}

	if r.fd, err = os.OpenFile(r.path, os.O_APPEND|os.O_WRONLY, 0644); err != nil {
		// open old log to failed - ignore
		// open a new log file.
		return r.openNew()
	}
	if r.maxSize > 0 {
		if info.Size() >= int64(r.maxSize) {
			return r.rotate()
		} else {
			r.lastSize = int(int64(r.maxSize) - info.Size())
		}
	}
	if r.maxLine > 0 {
		// 得到文件当前行数
		var curLine int
		if curLine, err = lineCounter(r.fd); err != nil {
			return err
		}
		if curLine >= r.maxLine {
			return r.rotate()
		} else {
			r.lastLine = r.maxLine - curLine
		}
	}
	return nil
}

func (r *SizeRotate) rotate() error {
	if err := r.openNew(); err != nil {
		return err
	}
	r.mill()
	return nil
}

func (r *SizeRotate) mill() {
	select {
	case r.millCh <- struct{}{}:
	default:
	}
}

func (r *SizeRotate) millRun() {
	for range r.millCh {
		r.millRunOnce()
	}
}

func (r *SizeRotate) millRunOnce() {
	var (
		compress, remove, files []logInfo
		err                     error
	)
	if files, err = r.oldLogFiles(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "clog: get old log file err %s\n", err.Error())
		return
	}
	if r.maxBackups > 0 && r.maxBackups < len(files) {
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
			if len(preserved) > r.maxBackups {
				remove = append(remove, f)
			} else {
				remaining = append(remaining, f)
			}
		}
		files = remaining
	}
	if r.saveDay > 0 {
		cutoff := currentTime().AddDate(0, 0, -r.saveDay)
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

	// 压缩n天后的文件
	if r.compress {
		compressTime := currentTime().AddDate(0, 0, -r.compressAfter)
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), compressSuffix) && f.timestamp.Before(compressTime) {
				compress = append(compress, f)
			}
		}
	}

	for _, f := range remove {
		if err = os.Remove(filepath.Join(r.dirPath, f.Name())); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "clog: remove old log file %s err %s\n", f.Name(), err.Error())
		}
	}
	for _, f := range compress {
		fn := filepath.Join(r.dirPath, f.Name())
		if err = compressLogFile(fn, fn+compressSuffix); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "clog: compress log file %s err %s\n", f.Name(), err.Error())
		}
	}
	return
}

func (r *SizeRotate) oldLogFiles() ([]logInfo, error) {
	var (
		files    []os.FileInfo
		logFiles []logInfo
		err      error
	)
	if files, err = ioutil.ReadDir(r.dirPath); err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	prefix, ext := r.prefixAndExt()
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if t, err := r.timeFromName(f.Name(), prefix, ext); err == nil {
			logFiles = append(logFiles, logInfo{t, f})
			continue
		}
		if t, err := r.timeFromName(f.Name(), prefix, ext+compressSuffix); err == nil {
			logFiles = append(logFiles, logInfo{t, f})
			continue
		}
		// 如果有错误  则不是生成的文件
	}

	sort.Sort(byFormatTime(logFiles))
	return logFiles, nil
}

// openNew opens a new log file for writing, moving any old log file out of the way.
// This methods assumes the file has already been closed.
func (r *SizeRotate) openNew() error {
	mode := os.FileMode(0644)
	if r.fd != nil {
		if err := r.fd.Close(); err != nil {
			return err
		}
	}
	info, err := os.Stat(r.path)
	// 文件存在 备份此文件
	if err == nil {
		// Copy the mode off the old logfile.
		mode = info.Mode()
		// move the existing file
		if err = os.Rename(r.path, r.backupName()); err != nil {
			return fmt.Errorf("can't rename log file: %s. err: %s", r.path, err.Error())
		}
	}
	if r.fd, err = os.OpenFile(r.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode); err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}
	r.lastSize = r.maxSize
	r.lastLine = r.maxLine
	return nil
}

func (r *SizeRotate) prefixAndExt() (prefix, ext string) {
	ext = filepath.Ext(r.name)
	prefix = r.name[:len(r.name)-len(ext)] + "-"
	return
}

func (r *SizeRotate) timeFromName(filename, prefix, ext string) (time.Time, error) {
	if !strings.HasPrefix(filename, prefix) {
		return time.Time{}, errors.New("mismatched prefix")
	}
	if !strings.HasSuffix(filename, ext) {
		return time.Time{}, errors.New("mismatched extension")
	}
	return time.ParseInLocation(backupTimeFormat, filename[len(prefix):len(filename)-len(ext)], time.Local)
}

// backupName creates a new filename from the given name, inserting a timestamp
// between the filename and the extension, using the local time if requested
// (otherwise UTC).
func (r *SizeRotate) backupName() string {
	prefix, ext := r.prefixAndExt()
	timestamp := currentTime().Format(backupTimeFormat)
	return filepath.Join(r.dirPath, fmt.Sprintf("%s%s%s", prefix, timestamp, ext))
}

// compressLogFile compresses the given log file, removing the
// uncompressed log file if successful.
func compressLogFile(src, dst string) (err error) {
	var (
		f, gzf *os.File
		fi     os.FileInfo
	)
	if f, err = os.Open(src); err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer f.Close()

	if fi, err = os.Stat(src); err != nil {
		return fmt.Errorf("failed to stat log file: %v", err)
	}
	// If this file already exists, we presume it was created by
	// a previous attempt to compress the log file.
	if gzf, err = os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode()); err != nil {
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

	if _, err = io.Copy(gz, f); err != nil {
		return err
	}
	if err = gz.Close(); err != nil {
		return err
	}
	if err = gzf.Close(); err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Remove(src); err != nil {
		return err
	}

	return nil
}

func lineCounter(r io.Reader) (int, error) {
	var (
		buf     = make([]byte, 32*1024)
		count   = 0
		lineSep = []byte{'\n'}
	)

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)
		switch {
		case err == io.EOF:
			return count, nil
		case err != nil:
			return count, err
		}
	}
}
