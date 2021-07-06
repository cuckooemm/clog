package cloghttp

import (
	"github.com/cuckooemm/clog"
	"net/http"
)

type handler struct{}

func Handler() http.Handler {
	return handler{}
}

func (handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPut {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	levelStr := req.URL.Query().Get("level")
	level, err := clog.ParseLevel(levelStr)
	if err != nil {
		_, _ = w.Write([]byte("support level [trace,debug,info,warn,error,fatal,panic,nil]\n eg: curl -X PUT http://host:port/path?level=info"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	clog.SetGlobalLevel(level)
	w.WriteHeader(http.StatusOK)
	return
}
