package write

import (
	"github.com/cuckooemm/clog"
)

func NewFileWrite(path string) clog.LevelWriter {
	return &FileWrite{f: path}
}

type FileWrite struct {
	f string
}

func (f *FileWrite) WriteLevel(level clog.Level, p []byte) (n int, err error) {

}
func (f *FileWrite) Write(p []byte) (n int, err error) {

}
