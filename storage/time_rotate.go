package storage

import (
	"errors"
	"fmt"
	"github.com/cuckooemm/clog"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type TimeRotate struct {
	path          string
	dirPath       string
	name          string
	maxDay        int    // 备份文件保存时间
	maxBackups    int    // 备份文件数量
	compress      bool   // 备份文件是否压缩
	compressAfter int    // 几天后的日志进行压缩
	timeFormat    string // 文件时间格式
	interval      int64  // 时间间隔
	fd            *os.File
	mu            sync.Mutex
	startMill     sync.Once
}

func (r *TimeRotate) WriteLevel(level clog.Level, p []byte) (int, error) {
	return r.Write(p)
}

func (r *TimeRotate) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.fd.Write(p)
}

func newTimeFileWrite(path string, interval time.Duration, day, backups, compressAfter int, compress bool) *TimeRotate {
	fw := new(TimeRotate)
	fw.path = path
	fw.dirPath = filepath.Dir(path)
	if err := os.MkdirAll(fw.dirPath, 0755); err != nil {
		panic(err)
	}
	fw.name = filepath.Base(path)
	fw.maxDay = day
	fw.maxBackups = backups
	fw.compressAfter = compressAfter
	fw.compress = compress
	fw.interval = int64(interval.Seconds())
	if interval >= time.Minute {
		if interval >= time.Hour {
			if interval >= time.Hour*24 {
				fw.timeFormat = "20060102"
			} else {
				fw.timeFormat = "2006010215"
			}
		} else {
			fw.timeFormat = "200601021504"
		}
	} else {
		panic("file split time is too short")
	}

	if err := fw.firstOpenExistOrNew(); err != nil {
		panic(err)
	}

	go fw.whileRun()

	return fw
}

func (r *TimeRotate) whileRun() {
	for {
		splitTIme := time.Unix(time.Now().Unix()/r.interval*r.interval+r.interval, 0)
		select {
		case <-time.After(splitTIme.Sub(time.Now())):
			r.mu.Lock()
			_ = r.openNew(splitTIme)
			r.mu.Unlock()
			r.processCompress()
		}
	}
}

func (r *TimeRotate) processCompress() {
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
	if r.maxDay > 0 {
		cutoff := currentTime().AddDate(0, 0, -r.maxDay)
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

func (r *TimeRotate) oldLogFiles() ([]logInfo, error) {
	var (
		files    []os.FileInfo
		logFiles []logInfo
		err      error
	)
	if files, err = ioutil.ReadDir(r.dirPath); err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		if t, err := r.timeFromName(f.Name(), r.name, ""); err == nil {
			logFiles = append(logFiles, logInfo{t, f})
			continue
		}
		if t, err := r.timeFromName(f.Name(), r.name, compressSuffix); err == nil {
			logFiles = append(logFiles, logInfo{t, f})
			continue
		}
		// 如果有错误  则不是生成的文件
	}

	sort.Sort(byFormatTime(logFiles))
	return logFiles, nil
}

func (r *TimeRotate) timeFromName(filename, prefix, compressSuffix string) (time.Time, error) {
	if !strings.HasPrefix(filename, prefix) {
		return time.Time{}, errors.New("mismatched prefix")
	}
	s := len(filename) - len(prefix) - len(r.timeFormat) - 1
	if s < 0 {
		return time.Time{}, errors.New("mismatched log file")
	}
	if s > 0 {
		if len(compressSuffix) != s {
			return time.Time{}, errors.New("mismatched log file")
		}
		if !strings.HasSuffix(filename, compressSuffix) {
			return time.Time{}, errors.New("mismatched log file")
		}
		filename = filename[:len(filename)-len(compressSuffix)]
	}
	return time.Parse(r.timeFormat, filename[len(prefix)+1:])
}

func (r *TimeRotate) backupName(tm time.Time) string {
	//prefix, ext := r.prefixAndExt()
	timestamp := time.Unix(tm.Unix()/r.interval*r.interval, 0)
	return filepath.Join(r.dirPath, fmt.Sprintf("%s.%s", r.name, timestamp.Format(r.timeFormat)))
}

func (r *TimeRotate) prefixAndExt() (prefix, ext string) {
	ext = filepath.Ext(r.name)
	prefix = r.name[:len(r.name)-len(ext)]
	return
}

func (r *TimeRotate) firstOpenExistOrNew() error {
	var (
		info os.FileInfo
		err  error
	)

	if info, err = os.Stat(r.path); err != nil {
		if os.IsNotExist(err) {
			return r.openNew(time.Now())
		}
		return fmt.Errorf("error getting log file info: %s", err.Error())
	}
	if time.Now().Sub(info.ModTime()) > time.Duration(r.interval)*time.Second {
		return r.openNew(info.ModTime().Add(time.Duration(r.interval) * time.Second))
	}
	if r.fd, err = os.OpenFile(r.path, os.O_APPEND|os.O_WRONLY, 0644); err != nil {
		// open old log to failed - ignore
		// open a new log file.
		return r.openNew(time.Now())
	}
	return nil
}

// openNew opens a new log file for writing, moving any old log file out of the way.
// This methods assumes the file has already been closed.
func (r *TimeRotate) openNew(tm time.Time) error {
	mode := os.FileMode(0644)
	if r.fd != nil {
		if err := r.fd.Close(); err != nil {
			return err
		}
	}
	info, err := os.Stat(r.path)
	// backup file if file exist
	if err == nil {
		// Copy the mode off the old logfile.
		mode = info.Mode()
		// move the existing file
		if err = os.Rename(r.path, r.backupName(tm)); err != nil {
			return fmt.Errorf("can't rename log file: %s", err)
		}
	}
	if r.fd, err = os.OpenFile(r.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode); err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}

	return nil
}
