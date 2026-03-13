package modlog

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	microlog "go-micro.dev/v4/logger"
	"go.uber.org/zap/zapcore"
)

func TestMicroLogger_NewInitOptionsString(t *testing.T) {
	base, _ := newObservedLogger(t, zapcore.DebugLevel)
	target := NewMicroLogger(base)
	require.Equal(t, "zap-micro", target.String())

	err := target.Init(func(o *microlog.Options) {
		o.CallerSkipCount = 1
	})
	require.NoError(t, err)
	require.Equal(t, 1, target.Options().CallerSkipCount)
}

func TestMicroLogger_Fields(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	target := NewMicroLogger(base)
	target.Init(func(o *microlog.Options) {
		o.CallerSkipCount = -1
	})

	withFields := target.Fields(map[string]interface{}{"k1": "v1", "k2": 2})
	withFields.Log(microlog.InfoLevel, "ok")

	require.Equal(t, 1, logs.Len())
	ctx := logs.All()[0].ContextMap()
	require.Equal(t, "v1", ctx["k1"])
	require.Equal(t, int64(2), ctx["k2"])
}

func TestMicroLogger_LogAllLevels(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	target := NewMicroLogger(base)

	target.Log(microlog.InfoLevel, "info")
	target.Log(microlog.DebugLevel, "debug")
	target.Log(microlog.TraceLevel, "trace")
	target.Log(microlog.WarnLevel, "warn")
	target.Log(microlog.ErrorLevel, "error")
	target.Log(microlog.Level(99), "other")

	require.Equal(t, 6, logs.Len())
}

func TestMicroLogger_LogfAllLevels(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	target := NewMicroLogger(base)

	target.Logf(microlog.InfoLevel, "i-%d", 1)
	target.Logf(microlog.DebugLevel, "d-%d", 2)
	target.Logf(microlog.TraceLevel, "t-%d", 3)
	target.Logf(microlog.WarnLevel, "w-%d", 4)
	target.Logf(microlog.ErrorLevel, "e-%d", 5)
	target.Logf(microlog.Level(99), "o-%d", 6)

	require.Equal(t, 6, logs.Len())
	require.Equal(t, "i-1", logs.All()[0].Message)
	require.Equal(t, "o-6", logs.All()[5].Message)
}

func TestMicroLogger_LogFatalLevel(t *testing.T) {
	if os.Getenv("MODLOG_MICRO_FATAL_CHILD") == "1" {
		base, _ := newObservedLogger(t, zapcore.DebugLevel)
		target := NewMicroLogger(base)
		target.Log(microlog.FatalLevel, "fatal")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMicroLogger_LogFatalLevel", "-test.v")
	cmd.Env = append(os.Environ(), "MODLOG_MICRO_FATAL_CHILD=1")
	err := cmd.Run()
	require.Error(t, err)
}

func TestMicroLogger_LogfFatalLevel(t *testing.T) {
	if os.Getenv("MODLOG_MICRO_FATALF_CHILD") == "1" {
		base, _ := newObservedLogger(t, zapcore.DebugLevel)
		target := NewMicroLogger(base)
		target.Logf(microlog.FatalLevel, "f-%d", 1)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMicroLogger_LogfFatalLevel", "-test.v")
	cmd.Env = append(os.Environ(), "MODLOG_MICRO_FATALF_CHILD=1")
	err := cmd.Run()
	require.Error(t, err)
}
