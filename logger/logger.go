package logger

import (
	"errors"
	"io"
	"os"

	"github.com/ftqo/gothor/config"
	"github.com/rs/zerolog"
)

func New(c config.Logger) (zerolog.Logger, error) {
	var level zerolog.Level
	var err error

	switch c.Level {
	case "trace":
		level = zerolog.TraceLevel
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "fatal":
		level = zerolog.FatalLevel
	case "panic":
		level = zerolog.PanicLevel
	default:
		err = errors.New("invalid log level")
		return zerolog.Logger{}, err
	}

	zerolog.SetGlobalLevel(level)

	var output io.Writer
	switch c.Format {
	case "json":
		output = os.Stderr
	case "console":
		output = zerolog.ConsoleWriter{Out: os.Stderr}
	default:
		err = errors.New("invalid log format")
		return zerolog.Logger{}, err
	}

	logger := zerolog.New(output).With().Timestamp().Logger()

	return logger, nil
}
