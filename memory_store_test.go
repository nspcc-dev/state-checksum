package xorkv

import (
	"testing"
)

func newMemoryStoreForTesting(t *testing.T) Store {
	return NewMemoryStore()
}
