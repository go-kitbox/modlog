package modlog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExist_NewExistAndSetAndExist(t *testing.T) {
	target := NewExist(2)

	require.False(t, target.Exist("k1"))
	target.Set("k1")
	require.True(t, target.Exist("k1"))
	require.False(t, target.Exist("k2"))
}

func TestExist_CopyIsIndependent(t *testing.T) {
	target := NewExist(2)
	target.Set("a")

	cloned := target.Copy()
	require.True(t, cloned.Exist("a"))

	cloned.Set("b")
	require.True(t, cloned.Exist("b"))
	require.False(t, target.Exist("b"))
}
