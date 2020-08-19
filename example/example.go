package main

import (
	"github.com/cuckooemm/clog"
	"os"
)

func main() {
	//clog.SetGlobalLevel()
	clog.NewOption().WithLogLevel(clog.DebugLevel).WithTimestamp().WithWriter(os.Stdout).Default()
	clog.Set.SetTimeFormat(clog.TimeFormatUnixMicro).SetCallerSkipFrameCount(2)
	clog.Info().Caller().Int8("hah", 2).Done()
}
