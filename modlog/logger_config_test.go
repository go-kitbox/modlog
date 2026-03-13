package modlog

import (
	"strings"
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestConfig_Unmarshal(t *testing.T) {
	yamlCfg := `
service: test   # 服务名称
level: debug    # 日志级别，分别为debug,info,warn,error,fatal,panic
filePath: "../logs/"   # 日志路径, 本地文件路径,如果为空，表示不输出到文件
timeZone: ""   # 时区，默认defaultTimeZone,可以从https://www.zeitverschiebung.net/en/ 查询时区信息
timeLayout: "" # 输出时间格式,默认为defaultTimeLayout,任何Go支持的格式都是合法的
debug: true     # 是否调试，调试模式会输出完整的代码行信息,其他模式只会输出项目内部的
levelToPath:
  debug: ../logs/debug.log
  info: ../logs/info.log
rotate:
  maxSize: 200
`

	var (
		cfg     *Config
		wantCfg = &Config{
			Service:    "test",
			Level:      zapcore.DebugLevel,
			FilePath:   "../logs/",
			TimeZone:   "",
			TimeLayout: "",
			Debug:      true,
			Rotate: &RotateConfig{
				MaxSize:    200,
				MaxBackups: 0,
				MaxAge:     0,
			},
			LevelToPath: map[string]string{
				zapcore.DebugLevel.String(): `../logs/debug.log`,
				zapcore.InfoLevel.String():  `../logs/info.log`,
			},
		}
		err error
	)

	// data, _ := yaml.Marshal(wantCfg)
	// t.Log(string(data))
	cfg, err = NewConfigFromYamlData(strings.NewReader(yamlCfg))
	require.NoError(t, err, `从yaml数据解析配置`)
	require.EqualValues(t, wantCfg, cfg, `结果一致`)

	infoLogger, err := cfg.Build()
	require.NoError(t, err, `构建错误`)

	infoLogger.Info(`info`)
	infoLogger.Debug(`debug`)
}

func TestConfig_Build(t *testing.T) {
	// debug = true
	var (
		cfg = &Config{
			Service: "test",
			Level:   zapcore.DebugLevel,
			Debug:   true,
			// FilePath: `a`,
		}
	)

	originLogger, err := cfg.Build()
	require.NoError(t, err, `构建错误`)

	// With 添加字段
	originLogger = originLogger.With(zap.String(`a`, `x`), zap.String(`b`, "y"))
	// Debug输出可见
	originLogger.Debug(`a`)

	// 验证日志器
	infoLogger := originLogger.Derive(`提现`)
	// Debug 可见
	infoLogger.Debug(`提现可见`)
	// 设置为Info
	infoLogger = infoLogger.SetLevel(zapcore.InfoLevel)
	// Debug 不可见
	infoLogger.Debug(`提现不可见`)
	// Debug 可见
	originLogger.Debug(`origin Debug`)
	// Info 可见
	infoLogger.Info(`提现可见.Info`)
	// Warn 可见
	infoLogger.Warn(`提现可见.Warn`)

	infoLogger = infoLogger.With(zap.String(`info`, `info`))

	infoLogger.Info(`infoLogger.Info`)
	// 再次衍生
	debugLogger := infoLogger.Derive(`汇总`)
	// Debug不可见
	debugLogger.Debug(`不可见`)
	// 设置为Debug
	debugLogger = debugLogger.SetLevel(zapcore.DebugLevel)
	// Debug可见
	debugLogger.Debug(`汇总可见`)

	taskLogger := debugLogger.Start()
	taskLogger.Info(`开始`)

	// 三次衍生

	warnLogger := debugLogger.Derive(`三次`)
	warnLogger.Info(`三次`)
}

func TestJSON(t *testing.T) {
	var (
		jsonCfg = &Config{
			Service: "test",
			Level:   zapcore.DebugLevel,
			Debug:   true,
			JSON:    true,
		}
	)

	originJSONLogger, err := jsonCfg.Build()

	require.NoError(t, err, `构建JSON输出`)

	// With 添加字段
	originJSONLogger = originJSONLogger.With(zap.String(`a`, `x`), zap.String(`b`, "y"))
	// Debug输出可见
	originJSONLogger.Debug(`a`)

	// 验证日志器
	infoJSONLogger := originJSONLogger.Derive(`提现`)
	// Debug 可见
	infoJSONLogger.Debug(`debug1`)
	// 设置为Info
	infoJSONLogger = infoJSONLogger.SetLevel(zapcore.InfoLevel)
	// Debug 不可见
	infoJSONLogger.Debug(`debug2`)
	// Debug 可见
	originJSONLogger.Debug(`origin Debug`)
	// Info 可见
	infoJSONLogger.Info(`infoJSONLogger.Info`)
	// Warn 可见
	infoJSONLogger.Warn(`infoJSONLogger.Warn`)

	infoJSONLogger = infoJSONLogger.With(zap.String(`info`, `info`))

	infoJSONLogger.Info(`infoJSONLogger.Info`)
	// 再次衍生
	debugJSONLogger := infoJSONLogger.Derive(`汇总`)
	// Debug不可见
	debugJSONLogger.Debug(`debug1`)
	// 设置为Debug
	debugJSONLogger = debugJSONLogger.SetLevel(zapcore.DebugLevel)
	// Debug可见
	debugJSONLogger.Debug(`debug2`)

	taskJSONLogger := debugJSONLogger.Start()
	taskJSONLogger.Info(`开始`)
}

func TestLogger_SetLevel(t *testing.T) {
	testLogger, _ := NewEasyLogger(true, false, ``, `test`)
	testLogger = testLogger.SetLevel(zapcore.WarnLevel)
	testLogger.Info(`info`)
	testLogger.Warn(`warn`)

	testLogger = testLogger.Derive(`a`)
	testLogger.Info(`info`)
	testLogger.Warn(`warn`)
}

func TestLogger_AddCallerSkip(t *testing.T) {
	var (
		cfg = &Config{
			Service: "test",
			Level:   zapcore.DebugLevel,
			Debug:   true,
		}
	)

	originLogger, err := cfg.Build()
	require.NoError(t, err, `构建错误`)

	logger := originLogger.Derive(`a`)

	logger.Info(`a`)
	logger = logger.AddCallerSkip(0)
	logger.Info(`a`)
}

func TestLogger_Start(t *testing.T) {
	testLogger, _ := NewEasyLogger(true, false, ``, `test`)

	testLogger = testLogger.With(zap.String(`key`, `b`))

	logger := testLogger.Start()

	logger.Info(`a`)

	logger = logger.With(zap.String(`key`, `a`))

	logger.Info(`a`)
}

func TestViper(t *testing.T) {
	data := `[conf]
service =  "test"   # 服务名称
level =  "debug"    # 日志级别，分别为debug,info,warn,error,fatal,panic
filePath =  "../logs/"   # 日志路径, 本地文件路径,如果为空，表示不输出到文件
timeZone =  ""   # 时区，默认defaultTimeZone,可以从https = //www.zeitverschiebung.net/en/ 查询时区信息
timeLayout =  "" # 输出时间格式,默认为defaultTimeLayout,任何Go支持的格式都是合法的
debug =  true     # 是否调试，调试模式会输出完整的代码行信息,其他模式只会输出项目内部的
rotate.maxSize =  200
levelToPath.debug = "../logs/debug.log"
levelToPath.info = "../logs/info.log"
levelToPath.warn ="../logs/warn.log"
`
	viperCfg := viper.New()
	viperCfg.SetConfigType(`toml`)
	require.NoError(t, viperCfg.ReadConfig(strings.NewReader(data)), `读取`)

	require.EqualValues(t, `test`, viperCfg.GetString(`conf.service`))

	marshaledData, err := toml.Marshal(viperCfg.Get(`conf`))

	require.NoError(t, err)
	t.Log(string(marshaledData))

	var cfg = &Config{}

	require.NoError(t, toml.Unmarshal(marshaledData, cfg))

	t.Log(cfg.LevelToPath)

	var (
		logger Logger
	)

	logger, err = cfg.Build()
	require.NoError(t, err)

	logger = logger.Derive(`mysql`)
	for i := 0; i < 3; i++ {
		logger.Info(`info-toml`)
		logger.Debug(`debug-toml`)
		logger.Warn(`warn`)
	}

	require.EqualValues(t, "test", cfg.Service)
	require.EqualValues(t, zapcore.DebugLevel, cfg.Level)
	require.EqualValues(t, "../logs/", cfg.FilePath)
	require.EqualValues(t, "../logs/debug.log", cfg.LevelToPath["debug"])
	require.EqualValues(t, "../logs/info.log", cfg.LevelToPath["info"])
	require.EqualValues(t, "../logs/warn.log", cfg.LevelToPath["warn"])
	require.EqualValues(t, defaultTimeZone, cfg.TimeZone)
	require.EqualValues(t, defaultTimeLayout, cfg.TimeLayout)
}

func TestNewConfigFromToml(t *testing.T) {
	data := `
		Service = 'test'
        Level = 'debug'
        FilePath = 'a'
        TimeZone = 'b'
        TimeLayout = 'c'
        Debug = true
        Dev = false
        JSON = false
        HideConsole = false
        
        [Rotate]
        MaxSize = 200
        MaxBackups = 0
        MaxAge = 0
`

	var (
		cfg     = &Config{}
		wantCfg = &Config{
			Service:    "test",
			Level:      zapcore.DebugLevel,
			FilePath:   "a",
			TimeZone:   "b",
			TimeLayout: "c",
			Debug:      true,
			Rotate: &RotateConfig{
				MaxSize:    200,
				MaxBackups: 0,
				MaxAge:     0,
			},
		}
		err error
	)

	tstData, _ := toml.Marshal(wantCfg)

	t.Log(string(tstData))

	cfg, err = NewConfigFromToml([]byte(data))
	require.NoError(t, err, `读取`)
	require.EqualValues(t, wantCfg, cfg)
}

func TestLogger_AddLogrus(t *testing.T) {
	var (
		cfg = &Config{
			Service: "test",
			Level:   zapcore.DebugLevel,
			Debug:   true,
			LevelToPath: map[string]string{
				zapcore.DebugLevel.String(): `../logs/debug.log`,
				zapcore.InfoLevel.String():  `../logs/info.log`,
			},
			Rotate: &RotateConfig{},
			// FilePath: `a`,
		}
	)

	originLogger, err := cfg.Build()

	require.NoError(t, err, `构建基础`)

	originLogger.Info(`info`)

	originLogger.Debug(`debug`)

	derived := originLogger.Derive(`derive`)

	derived.Info(`info`)
	derived.Derive(`debug`)

	// 再次derive
	derived = derived.Derive(`derive`)

	derived.Info(`info`)
	derived.Derive(`debug`)
}

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
