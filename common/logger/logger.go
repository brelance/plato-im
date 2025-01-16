package logger

import (
	"os"
	"time"

	"github.com/brelance/plato/common/config"
	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func Init() {
	level := config.GetLogLevelForLogger()
	logLevel, err := zerolog.ParseLevel(level)

	if err != nil {
		panic(err)
	}

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFieldName = "timestamp"
	Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(logLevel)
}
