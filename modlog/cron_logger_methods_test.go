package modlog

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestCronLogger_PrintfAndError(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	target := NewCronLogger(base)

	target.Printf("job=%s", "cleanup")
	target.Error(errors.New("boom"), "task=%s", "sync")

	require.Equal(t, 2, logs.Len())
	entries := logs.All()
	require.Equal(t, "job=cleanup", entries[0].Message)
	require.Equal(t, "task=sync", entries[1].Message)
	require.Equal(t, "boom", entries[1].ContextMap()["error"])
}

func TestDeriveCronLogger(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	derived := DeriveCronLogger(base, "topicA", "methodB")

	derived.Info("run")

	require.Equal(t, 1, logs.Len())
	ctx := logs.All()[0].ContextMap()
	require.Equal(t, []interface{}{"topicA", "methodB"}, ctx["topic/method"])
}

func TestCronFormatTimes(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	out := cronFormatTimes([]interface{}{"a", now, "b", 1})

	require.Equal(t, "a", out[0])
	require.Equal(t, now.Format(time.RFC3339), out[1])
	require.Equal(t, "b", out[2])
	require.Equal(t, 1, out[3])
}

func TestCronFormatString(t *testing.T) {
	require.Equal(t, "%s", cronFormatString(0))
	require.Equal(t, "%s, %v=%v", cronFormatString(2))
	require.Equal(t, "%s, %v=%v, %v=%v", cronFormatString(4))
}
