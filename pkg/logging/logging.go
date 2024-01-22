package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var Log *Logger

type Logger struct {
	*zerolog.Logger
}

func init() {
	Log = New()
}

func New() *Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	return &Logger{
		&logger,
	}
}

func (l *Logger) ErrorStack(err error) {
	l.Error().Caller(1).Stack().Err(err).Send()
}
