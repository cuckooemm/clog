package main

import (
	"github.com/cuckooemm/clog"
	"github.com/cuckooemm/clog/writer"
	"sync"
)

func main() {
	//clog.SetGlobalLevel()
	path := "./log/api.log"
	write := writer.NewWrite(path).WithBackups(100).WithCompress().WithMaxAge(0).WithMaxSize(writer.MB * 10).Done()
	clog.NewOption().WithLogLevel(clog.DebugLevel).WithTimestamp().WithWriter(write).Default()
	clog.Set.SetTimeFormat(clog.TimeFormatUnixMicro).SetCallerSkipFrameCount(2)
	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		gid := i
		go func() {
			for j := 0; j < 1000; j++ {
				clog.Info().Caller().Int("goroutine id", gid).Int("idx", j).Str("mes", "suibian").Cease()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
