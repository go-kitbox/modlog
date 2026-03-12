package modlog

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap/zapcore"
)

func TestNewZeroLogger(t *testing.T) {
	logger, err := NewEasyLogger(true, false, "", "test")
	require.NoError(t, err, "创建 logger 失败")

	zeroLog := NewZeroLogger(logger)
	require.NotNil(t, zeroLog, "NewZeroLogger 返回 nil")
}

func TestZeroLogger_Debug(t *testing.T) {
	logger, _ := NewEasyLogger(true, false, "", "test")
	zeroLog := NewZeroLogger(logger)

	zeroLog.Debug("debug message")
	zeroLog.Debug("debug with fields", logx.Field("key1", "value1"), logx.Field("key2", 123))
}

func TestZeroLogger_Info(t *testing.T) {
	logger, _ := NewEasyLogger(true, false, "", "test")
	zeroLog := NewZeroLogger(logger)

	zeroLog.Info("info message")
	zeroLog.Info("info with fields", logx.Field("module", "auth"), logx.Field("user_id", 1001))
}

func TestZeroLogger_Error(t *testing.T) {
	logger, _ := NewEasyLogger(true, false, "", "test")
	zeroLog := NewZeroLogger(logger)

	zeroLog.Error("error message")
	zeroLog.Error("error with fields", logx.Field("error", "connection failed"), logx.Field("retry", 3))
}

func TestZeroLogger_Slow(t *testing.T) {
	logger, _ := NewEasyLogger(true, false, "", "test")
	zeroLog := NewZeroLogger(logger)

	zeroLog.Slow("slow query warning")
	zeroLog.Slow("slow query with duration", logx.Field("duration", "2.5s"), logx.Field("sql", "SELECT * FROM users"))
}

func TestZeroLogger_Stat(t *testing.T) {
	logger, _ := NewEasyLogger(true, false, "", "test")
	zeroLog := NewZeroLogger(logger)

	zeroLog.Stat("statistics message")
	zeroLog.Stat("cache hit rate", logx.Field("hit_rate", "95.5%"), logx.Field("total", 10000))
}

func TestZeroLogger_Alert(t *testing.T) {
	logger, _ := NewEasyLogger(true, false, "", "test")
	zeroLog := NewZeroLogger(logger)

	zeroLog.Alert("alert message")
	zeroLog.Alert("system overload")
}

func TestZeroLogger_Stack(t *testing.T) {
	logger, _ := NewEasyLogger(true, false, "", "test")
	zeroLog := NewZeroLogger(logger)

	zeroLog.Stack("goroutine stack trace")
}

func TestZeroLogger_Close(t *testing.T) {
	logger, _ := NewEasyLogger(true, false, "", "test")
	zeroLog := NewZeroLogger(logger)

	err := zeroLog.Close()
	require.NoError(t, err, "Close 返回错误")
}

func TestSetZeroLogger(t *testing.T) {
	cfg := &Config{
		Service: "test",
		Level:   zapcore.DebugLevel,
		Debug:   true,
	}

	logger, err := cfg.Build()
	require.NoError(t, err, "构建 logger 失败")

	SetZeroLogger(logger)

	logx.Info("通过 go-zero logx 输出")
	logx.Debug("调试信息", logx.Field("module", "auth"))
	logx.Error("错误信息", logx.Field("code", 500))
}

func TestZeroLogger_WithDerive(t *testing.T) {
	cfg := &Config{
		Service: "parent",
		Level:   zapcore.DebugLevel,
		Debug:   true,
	}

	parentLogger, err := cfg.Build()
	require.NoError(t, err, "构建 parent logger 失败")

	authLogger := parentLogger.Derive("auth")
	zeroLog := NewZeroLogger(authLogger)

	zeroLog.Info("auth module log", logx.Field("user_id", 1001))
	zeroLog.Debug("auth debug", logx.Field("action", "login"))
}

func TestConvertLogFields(t *testing.T) {
	fields := []logx.LogField{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: 123},
		{Key: "key3", Value: true},
	}

	zapFields := convertLogFields(fields)

	require.Equal(t, len(fields), len(zapFields), "字段数量不匹配")
}
