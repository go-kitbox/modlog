package modlog

import (
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar/log"
	"go.uber.org/zap"
)

type PulsarLogger struct {
	Logger
}

func NewPulsarLogger(logger Logger) *PulsarLogger {
	return &PulsarLogger{Logger: logger}
}

func (p PulsarLogger) SubLogger(fields log.Fields) log.Logger {
	zapFields := make([]zap.Field, 0, len(fields))

	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	return NewPulsarLogger(p.Logger.With(zapFields...))
}

func (p PulsarLogger) WithFields(fields log.Fields) log.Entry {
	return p.SubLogger(fields)
}

func (p PulsarLogger) WithField(name string, value interface{}) log.Entry {
	return NewPulsarLogger(p.Logger.With(zap.Any(name, value)))
}

func (p PulsarLogger) WithError(err error) log.Entry {
	return NewPulsarLogger(p.Logger.With(zap.Error(err)))
}

func (p PulsarLogger) Debug(args ...interface{}) {
	p.Logger.Debug(``, zap.Any(`参数`, args))
}

func (p PulsarLogger) Info(args ...interface{}) {
	p.Logger.Info(``, zap.Any(`参数`, args))
}

func (p PulsarLogger) Warn(args ...interface{}) {
	p.Logger.Warn(``, zap.Any(`参数`, args))
}

func (p PulsarLogger) Error(args ...interface{}) {
	p.Logger.Error(``, zap.Any(`参数`, args))
}

func (p PulsarLogger) Debugf(format string, args ...interface{}) {
	p.Logger.Debug(fmt.Sprintf(format, args...))
}

func (p PulsarLogger) Infof(format string, args ...interface{}) {
	p.Logger.Info(fmt.Sprintf(format, args...))
}

func (p PulsarLogger) Warnf(format string, args ...interface{}) {
	p.Logger.Warn(fmt.Sprintf(format, args...))
}

func (p PulsarLogger) Errorf(format string, args ...interface{}) {
	p.Logger.Error(fmt.Sprintf(format, args...))
}
