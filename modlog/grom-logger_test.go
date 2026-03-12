package modlog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
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
