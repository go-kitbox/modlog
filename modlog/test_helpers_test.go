package modlog

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func newObservedLogger(t *testing.T, level zapcore.Level) (Logger, *observer.ObservedLogs) {
	t.Helper()

	core, logs := observer.New(level)
	base := zap.New(core, zap.WithCaller(true), zap.WithFatalHook(zapcore.WriteThenNoop))

	return NewLogger(base, "", -1, false, false, nil, nil), logs
}
