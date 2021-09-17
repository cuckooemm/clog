package cloghttp

import (
	"fmt"
	"github.com/cuckooemm/clog"
	"net/http"
)

type handler struct{}

func Handler() http.Handler {
	return handler{}
}

func (handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPut {
		_, _ = w.Write([]byte("Please use the PUT method to try again.\r\n"))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	levelStr := req.URL.Query().Get("level")
	if len(levelStr) == 0 {
		_, _ = w.Write([]byte("Parameter `level` is required.\r\n"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newLevel, err := clog.ParseLevel(levelStr)
	if err != nil {
		_, _ = w.Write([]byte("Support level [trace,debug,info,warn,error,nil].\r\n \t eg: curl -X PUT http://host:port/path?level=info\r\n"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	preLevel := clog.GlobalLevel()
	if newLevel == preLevel {
		_, _ = w.Write([]byte(fmt.Sprintf("The log level has not changed.\r\ncurrent global log level is [%s].\r\n", preLevel.String())))
		w.WriteHeader(http.StatusOK)
		return
	}
	clog.SetGlobalLevel(newLevel)
	_, _ = w.Write([]byte(fmt.Sprintf("Successfully changed the log level to [%s] (old:%s).\r\n", newLevel.String(), preLevel.String())))
	w.WriteHeader(http.StatusOK)
	return
}
