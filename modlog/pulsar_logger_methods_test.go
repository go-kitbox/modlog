package modlog

import (
	"errors"
	"testing"

	pulsarlog "github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestPulsarLogger_NewAndFieldMethods(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	target := NewPulsarLogger(base)

	sub := target.SubLogger(pulsarlog.Fields{"k1": "v1"})
	require.NotNil(t, sub)
	_, ok := sub.(*pulsarLogger)
	require.True(t, ok)

	entry := target.WithFields(pulsarlog.Fields{"k2": 2})
	require.NotNil(t, entry)
	_, ok = entry.(*pulsarLogger)
	require.True(t, ok)

	_ = target.WithField("k3", true)
	_ = target.WithError(errors.New("boom"))

	target.Infof("hello-%d", 1)
	require.Equal(t, "hello-1", logs.All()[0].Message)
}

func TestPulsarLogger_LevelsAndFormat(t *testing.T) {
	base, logs := newObservedLogger(t, zapcore.DebugLevel)
	target := NewPulsarLogger(base)

	target.Debug("d")
	target.Info("i")
	target.Warn("w")
	target.Error("e")
	target.Debugf("d-%d", 1)
	target.Infof("i-%d", 2)
	target.Warnf("w-%d", 3)
	target.Errorf("e-%d", 4)

	require.Equal(t, 8, logs.Len())
	require.Equal(t, "d-1", logs.All()[4].Message)
	require.Equal(t, "e-4", logs.All()[7].Message)
}

