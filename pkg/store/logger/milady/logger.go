package onex

import (
	"github.com/moweilong/milady/pkg/log"
)

// miladyLogger is a logger that implements the Logger interface.
// It uses the log package to log error messages with additional context.
type miladyLogger struct{}

// NewLogger creates and returns a new instance of miladyLogger.
func NewLogger() *miladyLogger {
	return &miladyLogger{}
}

// Error logs an error message with the provided context using the log package.
func (l *miladyLogger) Error(err error, msg string, kvs ...any) {
	log.Errorw(err, msg, kvs...)
}
