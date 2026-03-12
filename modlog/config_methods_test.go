package modlog

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestConfig_NewConfigAndNewEasyLogger(t *testing.T) {
	cfg := NewConfig()
	require.NotNil(t, cfg)

	logger, err := NewEasyLogger(true, true, "", "svc")
	require.NoError(t, err)
	require.NotNil(t, logger)
}

func TestConfig_Tidy(t *testing.T) {
	cfg := &Config{LevelToPath: map[string]string{"info": "info.log"}}
	require.NoError(t, cfg.tidy())
	require.Equal(t, "info.log", cfg.levelToPath[zapcore.InfoLevel])
}

func TestConfig_TidyWithInvalidLevel(t *testing.T) {
	cfg := &Config{LevelToPath: map[string]string{"bad-level": "x.log"}}
	require.Error(t, cfg.tidy())
}

func TestConfig_NewFromYamlAndTomlError(t *testing.T) {
	_, err := NewConfigFromYamlData(strings.NewReader(":::bad"))
	require.Error(t, err)

	_, err = NewConfigFromToml([]byte("==bad"))
	require.Error(t, err)
}

func TestConfig_BuildInvalidTimeZone(t *testing.T) {
	cfg := &Config{
		Service:  "svc",
		Level:    zapcore.InfoLevel,
		TimeZone: "Bad/Timezone",
	}
	_, err := cfg.Build()
	require.Error(t, err)
}

func TestConfig_BuildDefaults(t *testing.T) {
	cfg := &Config{
		Service: "svc",
		Level:   zapcore.InfoLevel,
	}
	logger, err := cfg.Build()
	require.NoError(t, err)
	require.NotNil(t, logger)
	require.Equal(t, defaultTimeZone, cfg.TimeZone)
	require.Equal(t, defaultTimeLayout, cfg.TimeLayout)
}

func TestConfig_NewEncoderConfig(t *testing.T) {
	cfg := &Config{
		TimeLayout: defaultTimeLayout,
		location:   time.UTC,
	}

	enc := cfg.newEncoderConfig()
	require.Equal(t, "T", enc.TimeKey)
	require.NotNil(t, enc.EncodeTime)

	cfg.Dev = true
	devEnc := cfg.newEncoderConfig()
	require.NotNil(t, devEnc.EncodeCaller)
}

func TestFillLumberjack(t *testing.T) {
	target := &lumberjack.Logger{}
	fillLumberjack(target)
	require.Equal(t, defaultRotateMaxSize, target.MaxSize)
	require.Equal(t, defaultRotateMaxBackups, target.MaxBackups)
	require.Equal(t, defaultRotateMaxAge, target.MaxAge)

	custom := &lumberjack.Logger{MaxSize: 1, MaxBackups: 2, MaxAge: 3}
	fillLumberjack(custom)
	require.Equal(t, 1, custom.MaxSize)
	require.Equal(t, 2, custom.MaxBackups)
	require.Equal(t, 3, custom.MaxAge)
}

func TestLevelEnableWithExcept(t *testing.T) {
	enabler := newLevelEnablerWithExcept(zapcore.DebugLevel, map[zapcore.Level]string{
		zapcore.InfoLevel: "info.log",
		zapcore.WarnLevel: "warn.log",
	}, zapcore.WarnLevel)

	require.True(t, enabler.Enabled(zapcore.DebugLevel))
	require.False(t, enabler.Enabled(zapcore.InfoLevel))
	require.True(t, enabler.Enabled(zapcore.WarnLevel))
}

