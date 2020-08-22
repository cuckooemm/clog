package writer

type option struct {
	path     string
	size     int64
	line     int64
	age      int64
	backups  int64
	compress bool
}

// NewWrite 返回一个实现 LevelWrite 接口的实例
func NewWrite(path string) *option {
	var opt = new(option)
	opt.path = path
	return opt
}

func (o *option) WithMaxSize(size int64) *option {
	o.size = size
	return o
}

func (o *option) WithMaxLine(line int64) *option {
	o.line = line
	return o
}

func (o *option) WithMaxAge(age int64) *option {
	o.age = age
	return o
}

func (o *option) WithCompress() *option {
	o.compress = true
	return o
}
func (o *option) WithBackups(backups int64) *option {
	o.backups = backups
	return o
}

func (o *option) Done() *FileWrite {
	return newFileWrite(o.path, o.size, o.line, o.age, o.backups, o.compress)
}
