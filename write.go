package clog

import (
	"io"
	"sync"
)

// LevelWriter 定义接口 接收带有 Level 信息
type LevelWriter interface {
	io.Writer
	WriteLevel(level Level, p []byte) (n int, err error)
}

type levelWriterAdapter struct {
	io.Writer
}

func (lw levelWriterAdapter) WriteLevel(l Level, p []byte) (n int, err error) {
	return lw.Write(p)
}

type syncWriter struct {
	mu sync.Mutex
	lw LevelWriter
}

func SyncWriter(w io.Writer) io.Writer {
	if lw, ok := w.(LevelWriter); ok {
		return &syncWriter{lw: lw}
	}
	return &syncWriter{lw: levelWriterAdapter{w}}
}

// Write implements the io.Writer interface.
func (s *syncWriter) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lw.Write(p)
}

// WriteLevel implements the LevelWriter interface.
func (s *syncWriter) WriteLevel(l Level, p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lw.WriteLevel(l, p)
}

type multiLevelWriter struct {
	writers []LevelWriter
}

func (t multiLevelWriter) Write(p []byte) (n int, err error) {
	for _, w := range t.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	return len(p), nil
}

func (t multiLevelWriter) WriteLevel(l Level, p []byte) (n int, err error) {
	for _, w := range t.writers {
		n, err = w.WriteLevel(l, p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	return len(p), nil
}

// MultiLevelWriter 将写入复制到所提供的 Writer.
// 类似于 UNIX tee(1) 命令.
// 如果实现了 LevelWriter 接口, 则优先使用 WriteLevel 方法.
func MultiLevelWriter(writers ...io.Writer) LevelWriter {
	lwriters := make([]LevelWriter, 0, len(writers))
	for _, w := range writers {
		if lw, ok := w.(LevelWriter); ok {
			lwriters = append(lwriters, lw)
		} else {
			lwriters = append(lwriters, levelWriterAdapter{w})
		}
	}
	return multiLevelWriter{lwriters}
}
