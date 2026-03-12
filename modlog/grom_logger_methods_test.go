package modlog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	gormlogger "gorm.io/gorm/logger"
)

func TestGormLogger_NewAndLogMode(t *testing.T) {
	base, _ := newObservedLogger(t, zapcore.DebugLevel)
	target := NewGormLogger(base, time.Second, map[string]zapcore.Level{
		"order": zapcore.InfoLevel,
	})
	require.NotNil(t, target)

	require.NotNil(t, target.LogMode(gormlogger.Info))
	require.NotNil(t, target.LogMode(gormlogger.Warn))
	require.NotNil(t, target.LogMode(gormlogger.Error))
	require.NotNil(t, target.LogMode(gormlogger.Silent))
}

func TestGormLogger_InfoWarnError(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	target := NewGormLogger(base, time.Second, nil).(*GormLogger)

	target.Info(context.Background(), "hello-%d", 1)
	target.Warn(context.Background(), "warn-msg", "k", "v")
	target.Error(context.Background(), "err-msg", "k", "v")

	require.Equal(t, 3, logs.Len())
	require.Equal(t, "hello-1", logs.All()[0].Message)
	require.Equal(t, "warn-msg", logs.All()[1].Message)
	require.Equal(t, "err-msg", logs.All()[2].Message)
}

func TestGormLogger_TraceBranches(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	target := NewGormLogger(base, 50*time.Millisecond, map[string]zapcore.Level{
		"payment": zapcore.InfoLevel,
	}).(*GormLogger)

	fc := func() (string, int64) {
		return "select 1", 2
	}

	ctxIgnoreErr := context.WithValue(context.Background(), IgnoreErrorKey, errors.New("ignore"))
	target.Trace(ctxIgnoreErr, time.Now(), fc, errors.New("ignore"))

	ctxIgnoreMsg := context.WithValue(context.Background(), IgnoreErrorMsgKey, "skip-me")
	target.Trace(ctxIgnoreMsg, time.Now(), fc, errors.New("skip-me detail"))

	target.Trace(context.Background(), time.Now(), fc, errors.New("db-error"))

	target.Trace(context.Background(), time.Now().Add(-time.Second), fc, nil)

	ctxNoModule := context.WithValue(context.Background(), ModuleKey, "unknown")
	target.Trace(ctxNoModule, time.Now(), fc, nil)

	ctxModule := context.WithValue(context.Background(), ModuleKey, "payment")
	target.Trace(ctxModule, time.Now(), fc, nil)

	require.GreaterOrEqual(t, logs.Len(), 3)
	require.GreaterOrEqual(t, logs.FilterMessage("执行错误").Len(), 1)
	require.GreaterOrEqual(t, logs.FilterMessage("慢查询").Len(), 1)
	require.GreaterOrEqual(t, logs.FilterMessage("执行成功").Len(), 1)
}

func TestGormLogger_AutoSkip(t *testing.T) {
	base, _ := newObservedLogger(t, zapcore.DebugLevel)
	target := NewGormLogger(base, time.Second, nil).(*GormLogger)

	require.NotPanics(t, func() {
		target.AutoSkip()
	})
}
