package test

import (
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func Logger() log.Logger {
	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stdout)

	logger = level.NewFilter(logger, level.AllowDebug())
	logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	return logger
}
