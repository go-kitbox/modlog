package modlog

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/utils"

	gormlogger "gorm.io/gorm/logger"
)

var (
	todo = context.TODO()
)

type GormLogger struct {
	Logger
	slowThreshold time.Duration
	minLevels     map[string]zapcore.Level
}

/*
NewGormLogger 创建 GORM 日志适配器
参数:
*	logger        	Logger         	基础日志器
*	slowThreshold 	time.Duration  	慢查询阈值
*	minLevel      	map[string]zapcore.Level	模块最小日志级别
返回值:
*	gormlogger.Interface	gormlogger.Interface	GORM 日志接口
*/
func NewGormLogger(logger Logger, slowThreshold time.Duration, minLevel map[string]zapcore.Level) gormlogger.Interface {
	targetLogger := &GormLogger{Logger: logger, slowThreshold: slowThreshold, minLevels: minLevel}
	targetLogger.AutoSkip()
	return targetLogger
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	var targetLevel zapcore.Level

	switch level {
	case gormlogger.Info:
		targetLevel = zapcore.InfoLevel
	case gormlogger.Warn:
		targetLevel = zapcore.WarnLevel
	case gormlogger.Error:
		targetLevel = zapcore.ErrorLevel
	case gormlogger.Silent:
		targetLevel = zapcore.PanicLevel
	}

	l.Logger = l.Logger.SetLevel(targetLevel)

	return l
}

func (l GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Info(fmt.Sprintf(msg, data...))
}

func (l GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	l.Logger.Warn(s, zap.Any(`值`, append([]interface{}{utils.FileWithLineNum()}, i...)))
}

func (l GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	l.Logger.Error(s, zap.Any(`值`, append([]interface{}{utils.FileWithLineNum()}, i...)))
}

func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil:
		value := ctx.Value(IgnoreErrorKey)
		if value != nil {
			if ignoreError, ok := value.(error); ok {
				if err == ignoreError {
					return
				}
			}
		}

		value = ctx.Value(IgnoreErrorMsgKey)
		if value != nil {
			if ignoreMsg, ok := value.(string); ok {
				if strings.Contains(err.Error(), ignoreMsg) {
					return
				}
			}
		}

		l.Logger.Error(`执行错误`, zap.String(`错误`, err.Error()), zap.Int64(`影响行数`, rows), zap.Duration(`耗时`, elapsed), zap.String(sqlField, sql))
	case elapsed > l.slowThreshold && l.slowThreshold != 0:
		l.Logger.Warn(`慢查询`, zap.Duration(`阈值`, l.slowThreshold), zap.Int64(`影响行数`, rows), zap.Duration(`耗时`, elapsed), zap.String(sqlField, sql))
	default:
		value := ctx.Value(ModuleKey)
		if value != nil {
			if module, ok := value.(string); ok {
				if _, exist := l.minLevels[module]; !exist {
					return
				}

				switch l.minLevels[module] {
				case zapcore.DebugLevel, zapcore.InfoLevel:
					l.Logger.Info(`执行成功`, zap.Int64(`影响行数`, rows), zap.Duration(`耗时`, elapsed), zap.String(sqlField, sql))
				}
			}
		}
	}
}

var (
	sqlField = `SQL`
	lineKey  = `gormLine`
	fileKey  = `gormFile`
)

type CtxKey string

const (
	IgnoreErrorKey    = CtxKey(`ignoreError`)
	IgnoreErrorMsgKey = CtxKey(`ignoreErrMsg`)
	ModuleKey         = CtxKey(`module`)
)

var (
	gormPackage    = filepath.Join("gorm.io", "gorm")
	zapgormPackage = filepath.Join("fighterlyt", "gormlogger")
)

func (l *GormLogger) AutoSkip() {
	for i := 2; i < 15; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.HasSuffix(file, "_test.go"):
		case strings.Contains(file, gormPackage):
		case strings.Contains(file, zapgormPackage):
		default:
			l.Logger = l.Logger.AddCallerSkip(i)
		}
	}
}
