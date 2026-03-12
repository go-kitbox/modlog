
## 目标

模块日志器 modlog 是基于 zap 的结构化日志库，提供以下核心特性：

### 功能特性
- [X] 不同的日志级别
- [X] 不同的模块可以使用不同的日志级别（通过 Derive 实现）
- [X] 级别接受内嵌式定义，层级不受影响
- [X]  运行时调整日志级别
- [X]  支持输出到 stdout 和文件，并且文件提供 rotate
- [X]  使用结构化的输出，而不是 fmt 格式
- [X]  携带代码行数，方便定位
  - [X]  开发环境展示完整路径，可点击
  - [X]  线上环境展示相对路径

### 适配器支持
- **GORM** - 提供 `NewGormLogger` 将 GORM SQL 日志按模块输出
- **MongoDB** - 提供 `NewMongoLogger` 将 MongoDB 操作日志按模块输出
- **Cron** - 提供 `NewCronLogger` 将定时任务日志按模块输出
- **Pulsar** - 提供 `NewPulsarLogger` 将消息队列日志按模块输出
- **go-micro** - 提供 `NewMicroLogger` 将微服务日志按模块输出
- **go-zero** - 提供 `NewZeroLogger` 将 go-zero 框架日志按模块输出

## 概念

### 日志器的层级和领域
模块通过 `Derive` 方法衍生子日志器，每个子日志器可以有独立的日志级别。

```
projectLogger (项目日志器)
    ├── moduleLoggerA (模块A日志器)
    │   ├── operationA (功能A日志器)
    │   └── operationB (功能B日志器)
    └── moduleLoggerB (模块B日志器)
```

* **Derive** - 衍生新的子日志器，每个实例有单独的级别，互不影响
* **With** - 添加字段信息，可以添加各自不同信息

### 结构化输出和fmt输出
结构化输出是指输出信息以多个字段出现，每个字段 **key:value**

```json
{"pri": "6", "host": "192.168.0.1", "ident": "fluentd", "pid": "11111", "message": "[error] Syslog test"}
```

fmt 输出格式：
```
<4>Nov 21 2:53:17 192.168.0.1 fluentd[11111]: [error] Syslog test
```

结构化输出强制信息添加描述，更加明确，而且可以被第三方工具所使用。

### 日志级别和结构化字段
都由 zapcore 包 (**go.uber.org/zap/zapcore**) 或者 zap 包 (**go.uber.org/zap**) 提供

#### 日志级别
**zapcore.Level**

```go
const (
    DebugLevel Level = iota - 1
    InfoLevel
    WarnLevel
    ErrorLevel
    DPanicLevel
    PanicLevel
    FatalLevel
)
```

#### 结构化字段
**zap.Field** - 常见的类型可以直接操作

```go
func String(key string, val string) Field
func Int(key string, val int) Field
func Int64(key string, val int64) Field
func Bool(key string, val bool) Field
func Duration(key string, val time.Duration) Field
func Any(key string, value interface{}) Field
// 更多字段类型请参考 zap 包文档
```

## 使用
### 安装
```bash
go get github.com/go-kitbox/modlog
```

### 快速开始
```go
package main

import (
    "github.com/go-kitbox/modlog/modlog"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func main() {
    // 创建配置
    cfg := &modlog.Config{
        Service: "my-service",
        Level:   zapcore.DebugLevel,
        Debug:   true,
    }

    // 构建日志器
    logger, err := cfg.Build()
    if err != nil {
        panic(err)
    }

    // 基础日志
    logger.Info("服务启动")

    // 衍生模块日志器
    authLogger := logger.Derive("auth")
    authLogger.Info("auth模块日志")

    // 衍生数据库日志器
    dbLogger := logger.Derive("db")
    dbLogger.Info("db模块日志")
}
```

### 配置
#### yaml 配置
```yaml
service: test          # 服务名称
level: debug           # 日志级别：debug, info, warn, error, fatal, panic
filePath: "logs/app"  # 日志路径，本地文件路径，如果为空，表示不输出到文件
timeZone: "Asia/Shanghai"  # 时区，默认 Asia/Shanghai
timeLayout: "2006-01-02 15:04:05.000"  # 时间格式
debug: true            # 是否调试模式
json: true             # 是否作为完整JSON输出
rotate:
  maxSize: 200        # 单个日志文件最大大小(MB)
  maxBackups: 10    # 最大备份数量
  maxAge: 7          # 最大保留天数
levelToPath:
  info: "logs/info.log"
  error: "logs/error.log"
```

#### toml 配置
```toml
service = "test"
level = "debug"
filePath = "logs/app"
timeZone = "Asia/Shanghai"
timeLayout = "2006-01-02 15:04:05.000"
debug = true
json = true

[rotate]
maxSize = 200
maxBackups = 10
maxAge = 7

[levelToPath]
info = "logs/info.log"
error = "logs/error.log"
```

#### 使用 viper
对于 viper 用户，需要这样处理：
1. `Get(x)` x 是整个配置的章节名称
2. 使用 `toml.Marshal` 序列化
3. 使用 `NewConfigFromToml`

参考 `config_test.go` 中的 `TestViper`

### 模块化日志示例
```go
// 创建基础日志器
cfg := &modlog.Config{
    Service: "my-app",
    Level:   zapcore.DebugLevel,
    Debug:   true,
}
baseLogger, _ := cfg.Build()

// 衍生 auth 模块日志器
authLogger := baseLogger.Derive("auth")
authLogger.Info("auth模块启动")

// 衍生 db 模块日志器
dbLogger := baseLogger.Derive("db")
dbLogger.Info("db模块启动")

// auth 模块可以设置不同的日志级别
authLogger = authLogger.SetLevel(zapcore.InfoLevel)
authLogger.Debug("这条不会输出") // 因为级别是 Info
authLogger.Info("这条会输出")
```

### GORM 日志适配器
GORM 日志可以通过 `NewGormLogger` 创建，会自动携带模块信息：

```go
// 创建 auth 模块的 GORM 日志器
authLogger := baseLogger.Derive("auth")
gormLogger := modlog.NewGormLogger(
    authLogger.AddCallerSkip(4),
    time.Second,  // 慢查询阈值
    map[string]zapcore.Level{
        "auth": zapcore.InfoLevel,  // auth 模块的 SQL 日志级别
    },
)

// 配合 GORM 使用
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: gormLogger,
})

// SQL 日志会自动携带 auth 模块标识
// 输出示例： INFO    auth    /path/to/file.go:123    执行成功    {"SQL": "SELECT * FROM users", "影响行数": 10}
```

### MongoDB 日志适配器
```go
// 创建 MongoDB 模块的日志器
mongoLogger := baseLogger.Derive("mongo")
mLogger := modlog.NewMongoLogger(mongoLogger, 10240)

// 配合 MongoDB 使用
clientOptions := options.Client().
    ApplyURI("mongodb://localhost:27017").
    SetLoggerOptions(mLogger.Options())
```

### Cron 日志适配器
```go
// 创建 Cron 模块的日志器
cronLogger := baseLogger.Derive("cron")
cLogger := modlog.NewCronLogger(cronLogger)
```

### Pulsar 日志适配器
```go
// 创建 Pulsar 模块的日志器
pulsarLogger := baseLogger.Derive("pulsar")
pLogger := modlog.NewPulsarLogger(pulsarLogger)
```

### 微服务日志适配器 (go-micro)
```go
// 创建微服务模块的日志器
microLogger := baseLogger.Derive("micro")
mLogger := modlog.NewMicroLogger(microLogger)
```

### go-zero 日志适配器
go-zero 框架的日志可以通过 `NewZeroLogger` 适配，实现日志统一输出：

```go
import (
    "github.com/go-kitbox/modlog/modlog"
    "github.com/zeromicro/go-zero/core/logx"
)

// 方式一：创建适配器手动使用
zeroLogger := baseLogger.Derive("zero")
writer := modlog.NewZeroLogger(zeroLogger)
writer.Info("info message", logx.Field("key", "value"))
writer.Error("error message", logx.Field("code", 500))
writer.Slow("slow query", logx.Field("duration", "2.5s"))

// 方式二：全局替换 go-zero 默认日志器
modlog.SetZeroLogger(baseLogger)
// 之后所有 logx 调用都会使用 modlog 输出
logx.Info("通过 go-zero logx 输出")
logx.Debug("调试信息", logx.Field("module", "auth"))
```

go-zero 日志适配器支持的方法：
- `Debug(v, fields...)` - Debug 级别日志
- `Info(v, fields...)` - Info 级别日志
- `Error(v, fields...)` - Error 级别日志
- `Slow(v, fields...)` - 慢查询警告（Warn 级别 + type=slow 标记）
- `Stat(v, fields...)` - 统计日志（Info 级别 + type=stat 标记）
- `Alert(v)` - 告警日志（Error 级别 + [ALERT] 前缀）
- `Severe(v)` - 严重错误（Fatal 级别）
- `Stack(v)` - 堆栈日志（Error 级别）

## 日志输出示例
```
2021-04-03 19:16:54.839	INFO    my-app    /path/to/file.go:123    服务启动    {"系统": "my-app"}
2021-04-03 19:16:54.839	INFO    auth     /path/to/auth.go:45     auth模块启动    {"系统": "my-app"}
2021-04-03 19:16:54.839	INFO    db      /path/to/db.go:67      db模块启动    {"系统": "my-app"}
2021-04-03 19:16:54.839	INFO    auth    /path/to/auth.go:89     执行成功    {"SQL": "SELECT * FROM users", "影响行数": 10}
```

## API 参考
### Config
```go
type Config struct {
    Service     string            // 服务名称
    Level       zapcore.Level     // 日志级别
    FilePath    string            // 日志文件路径
    TimeZone    string            // 时区
    TimeLayout  string            // 时间格式
    Debug       bool              // 是否调试模式
    Dev         bool              // 是否开发环境
    JSON        bool              // 是否JSON输出
    HideConsole bool              // 是否隐藏控制台输出
    Rotate      *RotateConfig     // 日志轮转配置
    LevelToPath map[string]string // 不同级别的日志路径
}
```

### Logger 接口
```go
type Logger interface {
    Derive(name string) Logger                              // 衍生子日志器
    With(fields ...zap.Field) Logger                // 添加字段
    WithWhenNotExist(key string, field zap.Field) Logger  // 条件添加字段
    Debug(msg string, fields ...zap.Field)          // Debug 日志
    Info(msg string, fields ...zap.Field)           // Info 日志
    Warn(msg string, fields ...zap.Field)           // Warn 日志
    Error(msg string, fields ...zap.Field)         // Error 日志
    Fatal(msg string, fields ...zap.Field)        // Fatal 日志
    Panic(msg string, fields ...zap.Field)        // Panic 日志
    Start() Logger                                // 携带任务ID
    SetLevel(level zapcore.Level) Logger          // 设置日志级别
    AddCallerSkip(skip int) Logger               // 添加调用栈跳过
}
```

## 测试
运行测试：
```bash
go test ./modlog -v
```

运行特定测试:
```bash
go test ./modlog -v -run TestConfig_Build
go test ./modlog -v -run TestGormLogger_Module
```
