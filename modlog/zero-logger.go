package modlog

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type zeroLogger struct {
	Logger
}

func (z *zeroLogger) Alert(v interface{}) {
	z.Logger.Error(fmt.Sprintf("[ALERT] %v", v))
}

func (z *zeroLogger) Close() error {
	if closer, ok := z.Logger.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

func (z *zeroLogger) Debug(v interface{}, fields ...logx.LogField) {
	zapFields := convertLogFields(fields)
	z.Logger.Debug(fmt.Sprintf("%v", v), zapFields...)
}

func (z *zeroLogger) Error(v interface{}, fields ...logx.LogField) {
	zapFields := convertLogFields(fields)
	z.Logger.Error(fmt.Sprintf("%v", v), zapFields...)
}

func (z *zeroLogger) Info(v interface{}, fields ...logx.LogField) {
	zapFields := convertLogFields(fields)
	z.Logger.Info(fmt.Sprintf("%v", v), zapFields...)
}

func (z *zeroLogger) Severe(v interface{}) {
	z.Logger.Fatal(fmt.Sprintf("[SEVERE] %v", v))
}

func (z *zeroLogger) Slow(v interface{}, fields ...logx.LogField) {
	zapFields := convertLogFields(fields)
	zapFields = append(zapFields, zap.String("type", "slow"))
	z.Logger.Warn(fmt.Sprintf("%v", v), zapFields...)
}

func (z *zeroLogger) Stack(v interface{}) {
	z.Logger.Error(fmt.Sprintf("[STACK] %v", v), zap.String("stack", fmt.Sprintf("%v", v)))
}

func (z *zeroLogger) Stat(v interface{}, fields ...logx.LogField) {
	zapFields := convertLogFields(fields)
	zapFields = append(zapFields, zap.String("type", "stat"))
	z.Logger.Info(fmt.Sprintf("%v", v), zapFields...)
}

func convertLogFields(fields []logx.LogField) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	return zapFields
}

func NewZeroLogger(logger Logger) logx.Writer {
	return &zeroLogger{
		Logger: logger,
	}
}

func SetZeroLogger(logger Logger) {
	logx.SetWriter(NewZeroLogger(logger))
}
