package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	Setup()
}

var SubLog zerolog.Logger

// Setup initialize the log instance
func Setup() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	SubLog = log.Output(
		zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: false,
		},
	)
	SubLog.Info().Msg("Logging is initialized!")
}
