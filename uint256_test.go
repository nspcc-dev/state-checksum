package xorkv

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUint256Equals(t *testing.T) {
	v0 := Uint256{}
	v1 := Uint256{}
	v2 := Uint256{}
	for i := range v2 {
		v2[i] = 0x5a
	}
	require.Equal(t, true, v1.Equals(v0))
	require.Equal(t, false, v1.Equals(v2))
}

func TestUint256Xor(t *testing.T) {
	v0 := Uint256{}
	v1 := Uint256{}
	v2 := Uint256{}
	for i := range v2 {
		v2[i] = 0x5a
	}
	v1.Xor(v2)
	require.Equal(t, v2, v1)
	v1.Xor(v2)
	require.Equal(t, v0, v1)
}

func TestHashKV(t *testing.T) {
	ref := Uint256(sha256.Sum256([]byte("kv")))
	require.Equal(t, ref, HashKV("k", []byte("v")))
}
