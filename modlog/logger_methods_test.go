package modlog

import (
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestEnsureDuplicateKeys(t *testing.T) {
	target := ensureDuplicateKeys(nil)
	require.NotNil(t, target)

	existing := NewExist(1)
	require.Same(t, existing, ensureDuplicateKeys(existing))
}

func TestLogger_DeriveWithAndWithWhenNotExist(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)

	child := base.Derive("module")
	child = child.With(zap.String("k", "v"))
	child.Info("msg")

	once := child.WithWhenNotExist("trace", zap.String("trace", "id1"))
	once.Info("once")

	twice := once.WithWhenNotExist("trace", zap.String("trace", "id2"))
	twice.Info("twice")

	entries := logs.All()
	require.Equal(t, "module", entries[0].LoggerName)
	require.Equal(t, "v", entries[0].ContextMap()["k"])
	require.Equal(t, "id1", entries[1].ContextMap()["trace"])
	require.Equal(t, "id1", entries[2].ContextMap()["trace"])
}

func TestLogger_BasicLevelsAndStart(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	base.Debug("d")
	base.Info("i")
	base.Warn("w")
	base.Error("e")

	started := base.Start()
	started.Info("started")

	require.Equal(t, 5, logs.Len())
	require.NotEmpty(t, logs.All()[4].ContextMap()["任务ID"])
}

func TestLogger_Panic(t *testing.T) {
	base, _ := newObservedLogger(t, zapcore.DebugLevel)
	require.Panics(t, func() {
		base.Panic("panic")
	})
}

func TestLogger_Fatal(t *testing.T) {
	if os.Getenv("MODLOG_FATAL_CHILD") == "1" {
		base, _ := newObservedLogger(t, zapcore.DebugLevel)
		base.Fatal("fatal")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestLogger_Fatal", "-test.v")
	cmd.Env = append(os.Environ(), "MODLOG_FATAL_CHILD=1")
	err := cmd.Run()
	require.Error(t, err)
}

func TestLogger_SetLevelAndAddCallerSkip(t *testing.T) {
	encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writeSyncer = zapcore.AddSync(io.Discard)
	inputCores = nil
	HiddenConsole = true

	base, _ := newObservedLogger(t, zapcore.DebugLevel)
	leveled := base.SetLevel(zapcore.WarnLevel)
	leveled.Info("hidden")
	leveled.Warn("shown")

	skipped := leveled.AddCallerSkip(1)
	require.NotNil(t, skipped)
}
