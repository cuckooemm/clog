package main

import (
	"github.com/cuckooemm/clog"
	"github.com/cuckooemm/clog/storage"
	"sync"
	"time"
)

func main() {
	//clog.SetGlobalLevel()
	path := "./log/api.log"
	s := storage.Opt.WithFile(path).Backups(100).Compress().SaveTime(10).MaxSize(10).Done()
	clog.NewOption().WithLogLevel(clog.DebugLevel).WithTimestamp().WithWriter(s).WithPrefix("app", "node1").Default()
	clog.Set.SetBaseTimeDurationInteger()
	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		gid := i
		go func() {
			for j := 0; j < 1000; j++ {
				clog.Info().Int("goroutine id", gid).Int("idx", j).Str("msg", "suibian").Msg("dhad")
				clog.Warn().TimeDur("timedur", time.Second*3+time.Minute*9).Cease()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
