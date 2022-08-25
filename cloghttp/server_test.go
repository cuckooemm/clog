package cloghttp

import (
	"github.com/ethreads/clog"
	"net/http"
	"testing"
)

func TestHandler_ServeHTTP(t *testing.T) {
	var mux = http.NewServeMux()
	mux.Handle("/changelog", Handler())
	srv := http.Server{
		Addr:    ":9999",
		Handler: mux,
	}
	if err := srv.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			clog.Info().Msg("service exit...")
			return
		}
		panic(err)
	}
}
