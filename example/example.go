package main

import (
	"github.com/cuckooemm/clog"
	"github.com/cuckooemm/clog/storage"
	"sync"
)

func main() {
	//clog.SetGlobalLevel()
	path := "./log/api.log"
	s := storage.Opt.WithFile(path).Backups(100).Compress().SaveTime(10).MaxSize(10).Done()
	clog.NewOption().WithLogLevel(clog.DebugLevel).WithTimestamp().WithWriter(s).Default()
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
