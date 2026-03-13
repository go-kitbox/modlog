package modlog

import (
	"testing"

	"github.com/stretchr/testify/require"
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

func TestExist_NewExistAndSetAndExist(t *testing.T) {
	target := NewExist(2)

	require.False(t, target.Exist("k1"))
	target.Set("k1")
	require.True(t, target.Exist("k1"))
	require.False(t, target.Exist("k2"))
}

func TestExist_CopyIsIndependent(t *testing.T) {
	target := NewExist(2)
	target.Set("a")

	cloned := target.Copy()
	require.True(t, cloned.Exist("a"))

	cloned.Set("b")
	require.True(t, cloned.Exist("b"))
	require.False(t, target.Exist("b"))
}
