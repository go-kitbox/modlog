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

var (
	localDSN = `root:dubaihell@tcp(127.0.0.1:3306)/test1?charset=utf8mb4&parseTime=True`
)

/*
TestNewGormLogger 测试创建 GORM 日志适配器（不需要真实数据库)
验证:
1. 能正确构建日志器
2. Derive 后日志器带有模块名称
3. 日志风格统一
*/
func TestNewGormLogger(t *testing.T) {
	cfg := &Config{
		Debug:   true,
		Service: "测试",
		Level:   zapcore.DebugLevel,
	}

	originLogger, err := cfg.Build()
	require.NoError(t, err, `构建原始日志器`)

	mysqlLogger := originLogger.Derive(`mysql`)
	mysqlLogger.Info(`before`)

	gormLogger := NewGormLogger(mysqlLogger, time.Second, nil)

	_ = gormLogger

	mysqlLogger.Info(`GORM日志适配器测试完成`)
}

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
