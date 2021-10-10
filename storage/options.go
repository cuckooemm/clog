package storage

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type timeRotateOption struct {
	timeRotate *TimeRotate
}
type sizeRotateOption struct {
	sizeRotate *SizeRotate
}

// NewTimeSplitFile 根据时间切分文件
func NewTimeSplitFile(path string, interval time.Duration) *timeRotateOption {
	f := new(timeRotateOption)
	f.timeRotate = &TimeRotate{
		path:     path,
		dirPath:  filepath.Dir(path),
		name:     filepath.Base(path),
		interval: int64(interval.Seconds()),
	}
	f.timeRotate.mu = &sync.Mutex{}
	if err := os.MkdirAll(f.timeRotate.dirPath, 0755); err != nil {
		panic(err)
	}
	if interval >= time.Minute {
		if interval >= time.Hour {
			if interval >= time.Hour*24 {
				f.timeRotate.timeFormat = "20060102"
			} else {
				f.timeRotate.timeFormat = "2006010215"
			}
		} else {
			f.timeRotate.timeFormat = "200601021504"
		}
	} else {
		panic("file split time is too short")
	}
	return f
}

// Backups 设置最大保存数量,达到阈值根据时间戳删除旧文件
func (t *timeRotateOption) Backups(backups int) *timeRotateOption {
	t.timeRotate.maxBackups = backups
	return t
}

// SaveTime 设置文件保存时间,单位天
func (t *timeRotateOption) SaveTime(day int) *timeRotateOption {
	t.timeRotate.saveDay = day
	return t
}

// Compress 开启并压缩n天前的日志
func (t *timeRotateOption) Compress(day int) *timeRotateOption {
	if day > 0 {
		if t.timeRotate.saveDay > 0 && t.timeRotate.saveDay < day {
			return t
		}
		t.timeRotate.compress = true
		t.timeRotate.compressAfter = day
	}
	return t
}

// Finish 返回io.Writer实例
func (t *timeRotateOption) Finish() *TimeRotate {
	if err := t.timeRotate.firstOpenExistOrNew(); err != nil {
		panic(err)
	}
	t.timeRotate.ch = make(chan struct{}, 1)
	go t.timeRotate.whileRun()
	return t.timeRotate
}

// NewSizeSplitFile 按文件大小分隔
func NewSizeSplitFile(path string) *sizeRotateOption {
	var o = new(sizeRotateOption)
	o.sizeRotate = &SizeRotate{
		path:    path,
		dirPath: filepath.Dir(path),
		name:    filepath.Base(path),
	}
	if err := os.MkdirAll(o.sizeRotate.dirPath, 0755); err != nil {
		panic(err)
	}
	o.sizeRotate.mu = &sync.Mutex{}
	return o
}

// MaxSize 设置文件大小上限,单位Mb
func (o *sizeRotateOption) MaxSize(m int) *sizeRotateOption {
	o.sizeRotate.maxSize = m * (1 << 20)
	return o
}

// MaxLine 设置文件行数上限
func (o *sizeRotateOption) MaxLine(line int) *sizeRotateOption {
	o.sizeRotate.maxLine = line
	return o
}

// SaveTime 设置日志保存天数,单位天
func (o *sizeRotateOption) SaveTime(day int) *sizeRotateOption {
	o.sizeRotate.saveDay = day
	return o
}

// Compress 开始并设置压缩n天前的文件
func (o *sizeRotateOption) Compress(day int) *sizeRotateOption {
	if day > 0 {
		if o.sizeRotate.saveDay > 0 && o.sizeRotate.saveDay < day {
			return o
		}
		o.sizeRotate.compress = true
		o.sizeRotate.compressAfter = day
	}
	return o
}

// Backups 设置文件保存数量上限,达到阈值后根据时间删除旧文件
func (o *sizeRotateOption) Backups(total int) *sizeRotateOption {
	o.sizeRotate.maxBackups = total
	return o
}

// Finish 返回is.Writer实例
func (o *sizeRotateOption) Finish() *SizeRotate {
	if o.sizeRotate.saveDay > 0 || o.sizeRotate.compress || o.sizeRotate.maxBackups > 0 {
		o.sizeRotate.millCh = make(chan struct{}, 1)
		go o.sizeRotate.millRun()
		o.sizeRotate.mill()
	}

	if err := o.sizeRotate.firstOpenExistOrNew(); err != nil {
		panic(err)
	}
	return o.sizeRotate
}
