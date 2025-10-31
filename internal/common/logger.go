package common

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// TODO: changed to gotools pkg

func GetLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"

	return zerolog.New(os.Stdout).With().Timestamp().
		Str("app", AppName).
		Str("service", "loggate-internal").
		Logger()
}

func GetGenLogger() zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().
		Timestamp().
		Str("app", "loggen").
		Logger()
}
