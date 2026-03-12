package modlog

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap/zapcore"
)

func TestMongoLogger_NewAndOptions(t *testing.T) {
	base, _ := newObservedLogger(t, zapcore.DebugLevel)
	target := NewMongoLogger(base, 1024)
	require.NotNil(t, target)
	require.NotNil(t, target.Options())
}

func TestMongoLogger_InfoAndError(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	target := NewMongoLogger(base, 1024)

	target.Info(int(options.LogLevelDebug), "debug-msg", "k1", "v1")
	target.Info(int(options.LogLevelInfo), "info-msg", "k2", 2)
	target.Info(99, "fallback-msg", "k3", true)
	target.Error(errors.New("boom"), "error-msg", "k4", "v4")

	require.Equal(t, 4, logs.Len())
	require.Equal(t, "debug-msg", logs.All()[0].Message)
	require.Equal(t, "v1", logs.All()[0].ContextMap()["k1"])
	require.Equal(t, "info-msg", logs.All()[1].Message)
	require.Equal(t, int64(2), logs.All()[1].ContextMap()["k2"])
	require.Equal(t, "fallback-msg", logs.All()[2].Message)
	require.Equal(t, true, logs.All()[2].ContextMap()["k3"])
	require.Equal(t, "boom", logs.All()[3].ContextMap()["error"])
}

func TestAnyToZapFieldMongo(t *testing.T) {
	fields := anyToZapFieldMongo("a", 1, "b")
	require.Len(t, fields, 2)
	require.Equal(t, "a", fields[0].Key)
	require.Equal(t, "数据3", fields[1].Key)

	require.Panics(t, func() {
		_ = anyToZapFieldMongo(1, "v")
	})
}

func TestMongoLogger_CommandMonitor(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	target := NewMongoLogger(base, 1024)
	monitor := target.CommandMonitor()
	require.NotNil(t, monitor)

	cmdRaw, err := bson.Marshal(bson.D{{Key: "find", Value: "users"}})
	require.NoError(t, err)
	replyRaw, err := bson.Marshal(bson.D{{Key: "ok", Value: 1}})
	require.NoError(t, err)

	monitor.Started(context.Background(), &event.CommandStartedEvent{
		RequestID: 1,
		Command:   cmdRaw,
	})
	monitor.Succeeded(context.Background(), &event.CommandSucceededEvent{
		CommandFinishedEvent: event.CommandFinishedEvent{
			RequestID: 1,
		},
		Reply: replyRaw,
	})
	monitor.Failed(context.Background(), &event.CommandFailedEvent{
		CommandFinishedEvent: event.CommandFinishedEvent{
			RequestID: 1,
		},
		Failure: "timeout",
	})

	require.Equal(t, 3, logs.Len())
	require.Equal(t, "开始执行", logs.All()[0].Message)
	require.Equal(t, "执行成功", logs.All()[1].Message)
	require.Equal(t, "执行失败", logs.All()[2].Message)
}
