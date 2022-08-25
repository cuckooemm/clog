package cloghttp

import (
	"fmt"
	"github.com/ethreads/clog"
	"net/http"
)

type handler struct{}

func Handler() http.Handler {
	return handler{}
}

func (handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPut {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Please use the PUT method to try again.\r\n"))
		return
	}
	levelStr := req.URL.Query().Get("level")
	if len(levelStr) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Parameter `level` is required.\r\n"))
		return
	}
	newLevel, err := clog.ParseLevel(levelStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Support level [trace,debug,random,info,warn,error,nil].\r\n \t eg: curl -X PUT http://host:port/path?level=info\r\n"))
		return
	}
	preLevel := clog.GlobalLevel()
	if newLevel == preLevel {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf("The log level has not changed.\r\ncurrent global log level is [%s].\r\n", preLevel.String())))
		return
	}
	clog.SetGlobalLevel(newLevel)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("Successfully changed the log level to [%s] (old:%s).\r\n", newLevel.String(), preLevel.String())))
	return
}
