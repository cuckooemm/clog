package writer

type option struct {
	path     string
	size     int
	line     int
	day      int
	backups  int
	compress bool
}
type Capacity int64

const (
	B Capacity = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
)

// NewWrite 返回一个实现 LevelWrite 接口的实例
func NewWrite(path string) *option {
	var opt = new(option)
	opt.path = path
	return opt
}

// Mb
func (o *option) WithMaxSize(size Capacity) *option {
	o.size = int(size)
	return o
}

func (o *option) WithMaxLine(line int) *option {
	o.line = line
	return o
}

func (o *option) WithMaxAge(day int) *option {
	o.day = day
	return o
}

func (o *option) WithCompress() *option {
	o.compress = true
	return o
}
func (o *option) WithBackups(backups int) *option {
	o.backups = backups
	return o
}

func (o *option) Done() *FileWrite {
	return newFileWrite(o.path, o.size, o.line, o.day, o.backups, o.compress)
}
