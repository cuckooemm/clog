# clog

## 结构化日志库


### Installation
`go get -u https://github.com/cuckooemm/clog`

### Getting Started

#### Example
```
    // init
    clog.NewOption().WithLogLevel(clog.InfoLevel).WithTimestamp().WithWriter(os.Stdout).Default()
```
- `withTimestamp`
  - 通过hook方式为每条日志添加时间戳，默认添加至行首
  - {"time":"2021-07-05T23:23:15+08:00","level":"warn","foo":"bar"}
    
- `WithWriter`
  - 添加输出(需实现`io.Writer`接口)，默认输出至`os.stderr`标准输出
    
- `Default`
  - 设置默认log实例
    
- `WithLogLevel`
  - 设置日志等级
    
- `WithPreHook` `WithHook`
  - 设置前置hook与后置hook
    
- `Logger`
  - 返回log实例
  ```go
    import (
	    "github.com/cuckooemm/clog"
    )
    log := clog.NewOption().WithLogLevel(clog.InfoLevel).WithTimestamp().WithWriter(s).Logger()
    log.Info().Int("foo", 123).Int("bar", 123).Msg("")
  
    // 未初始化的默认log不能使用此形式
    // panic
    clog.Info().Int("foo", 123).Int("bar", 123).Msg("")

    // 默认Log初始化后可通过`CopyDefault`获取到默认logger副本
    // 配置数据继承全局log
    log := clog.CopyDefault()
    // 设置此log实例前缀 
    log.AppendStrPrefix("trackId","foobar")
    log.Info().Int("foo", 123).Int("bar", 123).Msg("")
  ```

#### Output
```go
  // 按时间切割
  s := storage.NewTimeSplitFile(path, time.Minute).Backups(3).SaveTime(3).Compres(2).Finish()
  // 按行切割
  s := storage.NewSizeSplitFile(path).Backups(10).MaxSize(50).SaveTime(4).Compress(3).Finish()
  clog.NewOption().WithWriter(s).Logger()

```

- `Backups`
  - 设置最大保留日志个数

- `SaveTime`
  - 设置最大保留天数

- `Compres`
  - 压缩N天前的日志

- `MaxSize`
  - 设置日志文件最大大小(仅支持`NewSizeSplitFile`)

- `MaxLine`
  - 设置日志文件最大行数(仅支持`NewSizeSplitFile`)
  
#### ChangeLogLevel
```go
	var mux = http.NewServeMux()
	// 注册handler
	mux.Handle("/changelog", cloghttp.Handler())
	srv := http.Server{
		Addr:    ":9999",
		Handler: mux,
	}
```
即可通过`curl -X PUT http://url/path?level=info`方式动态修改程序日志等级