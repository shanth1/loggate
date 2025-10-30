package common

import (
	"github.com/rs/zerolog"
	"github.com/shanth1/gotools/log"
)

func GetLogger() zerolog.Logger {
	return log.New(AppName, "trace")
}

func GetGenLogger() zerolog.Logger {
	return log.New("loggen", "trace")
}
