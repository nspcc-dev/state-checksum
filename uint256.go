package xorkv

import (
	"crypto/sha256"
)

type Uint256 [sha256.Size]byte

// Xor xores input value into the Uint256 element.
func (u *Uint256) Xor(o Uint256) {
	for i := range u {
		u[i] ^= o[i]
	}
}

// Equals compares two Uint256 values.
func (u *Uint256) Equals(o Uint256) bool {
	for i := range u {
		if u[i] != o[i] {
			return false
		}
	}
	return true
}

// HashKV returns Uint256 with a hash of given key and value.
func HashKV(k string, v []byte) Uint256 {
	return sha256.Sum256(append([]byte(k), v...))
}
