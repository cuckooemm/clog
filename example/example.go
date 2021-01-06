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
	//changeLogLevel()
	//newSearchExample()
	writeLogFile()
}

func writeLogFile() {

	path := "./log/api.log"
	//s := storage.NewSizeSplitFile(path).Backups(10).MaxSize(50).SaveTime(4).Compress(3).Finish()
	s := storage.NewTimeSplitFile(path, time.Minute).Backups(3).SaveTime(3).Compress(2).Finish()
	defer s.Close()
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

type SearchExample struct {
	count int
	log   clog.Logger
}

func newSearchExample() {
	example := SearchExample{
		log: clog.NewOption().WithLogLevel(clog.InfoLevel).WithWriter(os.Stdout).Logger(),
	}
	example.log.ResetStrPrefix("searchId", time.Now().UnixNano())

	example.log.Info().Interface("data", 1).Msg("")
	example.log.Error().Interface("data", 1).Msg("")
	example.log.Info().Interface("data", 1).Msg("")
	example.log.AppendStrPrefix("sec", time.Now().UnixNano())
	example.log.Info().Interface("data", 1).Msg("")
	example.log.Info().Interface("data", 1).Msg("")
	example.log.Info().Interface("data", 1).Msg("")
}
