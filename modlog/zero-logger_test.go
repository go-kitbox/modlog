package modlog

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap/zapcore"
)

type closableLogger struct {
	Logger
	retErr error
}

func (c closableLogger) Close() error {
	return c.retErr
}

func TestNewZeroLogger(t *testing.T) {
	base, _ := newObservedLogger(t, zapcore.DebugLevel)
	zeroLog := NewZeroLogger(base)
	require.NotNil(t, zeroLog)
}

func TestZeroLogger_DebugInfoError(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	zeroLog := NewZeroLogger(base)

	zeroLog.Debug("debug", logx.Field("k1", "v1"))
	zeroLog.Info("info", logx.Field("k2", 2))
	zeroLog.Error("error", logx.Field("k3", true))

	require.Equal(t, 3, logs.Len())
	require.Equal(t, "debug", logs.All()[0].Message)
	require.Equal(t, "v1", logs.All()[0].ContextMap()["k1"])
	require.Equal(t, "info", logs.All()[1].Message)
	require.Equal(t, int64(2), logs.All()[1].ContextMap()["k2"])
	require.Equal(t, "error", logs.All()[2].Message)
	require.Equal(t, true, logs.All()[2].ContextMap()["k3"])
}

func TestZeroLogger_SlowStatAlertStack(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	zeroLog := NewZeroLogger(base)

	zeroLog.Slow("slow", logx.Field("cost", "1s"))
	zeroLog.Stat("stat", logx.Field("rate", "99%"))
	zeroLog.Alert("alert")
	zeroLog.Stack("stack")

	require.Equal(t, 4, logs.Len())
	require.Equal(t, "slow", logs.All()[0].Message)
	require.Equal(t, "slow", logs.All()[0].ContextMap()["type"])
	require.Equal(t, "stat", logs.All()[1].Message)
	require.Equal(t, "stat", logs.All()[1].ContextMap()["type"])
	require.Equal(t, "[ALERT] alert", logs.All()[2].Message)
	require.Equal(t, "[STACK] stack", logs.All()[3].Message)
}

func TestZeroLogger_CloseAndSevere(t *testing.T) {
	base, _ := newObservedLogger(t, zapcore.DebugLevel)
	zeroLog := NewZeroLogger(base)

	require.NoError(t, zeroLog.Close())

	closeErr := errors.New("close-err")
	withCloser := &zeroLogger{Logger: closableLogger{Logger: base, retErr: closeErr}}
	require.Equal(t, closeErr, withCloser.Close())

	if os.Getenv("MODLOG_SEVERE_CHILD") == "1" {
		internal, ok := zeroLog.(*zeroLogger)
		require.True(t, ok)
		internal.Severe("down")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestZeroLogger_CloseAndSevere", "-test.v")
	cmd.Env = append(os.Environ(), "MODLOG_SEVERE_CHILD=1")
	err := cmd.Run()
	require.Error(t, err)
}

func TestSetZeroLogger(t *testing.T) {
	base, _ := newObservedLogger(t, zapcore.DebugLevel)
	SetZeroLogger(base)

	logx.Info("go-zero-info")
	logx.Debug("go-zero-debug", logx.Field("module", "auth"))
	logx.Error("go-zero-error", logx.Field("code", 500))
}

func TestZeroLogger_WithDerive(t *testing.T) {
	parentLogger, logs := newObservedLogger(t, zapcore.DebugLevel)

	authLogger := parentLogger.Derive("auth")
	zeroLog := NewZeroLogger(authLogger)

	zeroLog.Info("auth module log", logx.Field("user_id", 1001))
	zeroLog.Debug("auth debug", logx.Field("action", "login"))

	require.Equal(t, 2, logs.Len())
	require.Equal(t, "auth", logs.All()[0].LoggerName)
}

func TestConvertLogFields(t *testing.T) {
	fields := []logx.LogField{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: 123},
		{Key: "key3", Value: true},
	}

	zapFields := convertLogFields(fields)

	require.Equal(t, len(fields), len(zapFields))
	require.Equal(t, "key1", zapFields[0].Key)
	require.Equal(t, "key2", zapFields[1].Key)
	require.Equal(t, "key3", zapFields[2].Key)
}
