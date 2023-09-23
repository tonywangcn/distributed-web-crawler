package log

import (
	"fmt"

	"github.com/tonywangcn/distributed-web-crawler/pkg/logging"
)

// Debug debug
func Debug(format string, v ...interface{}) {

	logging.SubLog.Debug().Msg(fmt.Sprintf(format, v...))
}

// Info info
func Info(format string, v ...interface{}) {

	logging.SubLog.Info().Msg(fmt.Sprintf(format, v...))
}

// Warn warning
func Warn(format string, v ...interface{}) {

	logging.SubLog.Warn().Msg(fmt.Sprintf(format, v...))
}

// Error error
func Error(format string, v ...interface{}) {

	logging.SubLog.Error().Msg(fmt.Sprintf(format, v...))
}
