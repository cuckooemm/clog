package main

import (
	"github.com/cuckooemm/clog"
	"github.com/cuckooemm/clog/cloghttp"
	"github.com/cuckooemm/clog/storage"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	changeLogLevel()
}

func writeLogFile() {
	path := "./log/api.log"
	s := storage.Opt.WithFile(path).Backups(100).Compress().SaveTime(10).MaxSize(10).Done()
	clog.NewOption().WithLogLevel(clog.InfoLevel).WithTimestamp().WithWriter(s).Default()
	clog.Set.SetBaseTimeDurationInteger()

	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		gid := i
		go func() {
			for j := 0; j < 1000; j++ {
				clog.Info().Int("goroutine id", gid).Int("idx", j).Str("msg", "suibian").Msg("dhad")
				clog.Warn().TimeDur("timedur", time.Second*3+time.Minute*9).Cease()
				clog.Info().Stack().Cease()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func changeLogLevel() {
	var mux = http.NewServeMux()
	mux.Handle("/changelog", cloghttp.Handler())
	srv := http.Server{
		Addr:    ":9999",
		Handler: mux,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				clog.Info().Msg("service exit...")
				return
			}
			panic(err)
		}
	}()
	clog.NewOption().WithLogLevel(clog.TraceLevel).WithTimestamp().WithWriter(os.Stdout).Default()
	clog.Set.SetBaseTimeDurationInteger()
	for i := 0; i < 1000; i++ {
		clog.Trace().Msg("trace")
		clog.Debug().Msg("debug")
		clog.Info().Msg("info")
		clog.Warn().Msg("warn")
		clog.Error().Msg("error")
		time.Sleep(time.Millisecond * 300)
	}
}
