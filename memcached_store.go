package xorkv

import (
	"crypto/sha256"
)

// MemCachedStore is a wrapper around persistent store that caches all changes
// being made for them to be later flushed in one batch.
type MemCachedStore struct {
	MemoryStore

	// Persistent Store.
	ps Store

	stateSum Uint256
}

// NewMemCachedStore creates a new MemCachedStore object.
func NewMemCachedStore(lower Store) *MemCachedStore {
	return &MemCachedStore{
		MemoryStore: *NewMemoryStore(),
		ps:          lower,
	}
}

// Delete implements the Store interface.
func (s *MemCachedStore) Delete(key []byte) error {
	strKey := string(key)
	// Double Delete is a noop.
	if s.del[strKey] {
		return nil
	}
	if val, ok := s.mem[strKey]; ok {
		// The value was added, but now we're deleting it.
		s.stateSum.Xor(HashKV(strKey, val))
	} else if val, err := s.ps.Get(key); err == nil {
		// The value is present in the lower store, but now we're deleting it.
		s.stateSum.Xor(HashKV(strKey, val))
	}
	return s.MemoryStore.Delete(key)
}

// Put implements the Store interface.
func (s *MemCachedStore) Put(key, value []byte) error {
	strKey := string(key)
	if oldVal, ok := s.mem[strKey]; ok {
		// We've already updated the value and now are doing it again.
		s.stateSum.Xor(HashKV(strKey, oldVal))
	} else if oldVal, err := s.ps.Get(key); err == nil {
		// The first update to already existing value.
		s.stateSum.Xor(HashKV(strKey, oldVal))
	}
	s.stateSum.Xor(HashKV(strKey, value))
	return s.MemoryStore.Put(key, value)
}

// Get implements the Store interface.
func (s *MemCachedStore) Get(key []byte) ([]byte, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	k := string(key)
	if val, ok := s.mem[k]; ok {
		return val, nil
	}
	if _, ok := s.del[k]; ok {
		return nil, ErrKeyNotFound
	}
	return s.ps.Get(key)
}

// Seek implements the Store interface.
func (s *MemCachedStore) Seek(key []byte, f func(k, v []byte)) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	s.MemoryStore.Seek(key, f)
	s.ps.Seek(key, func(k, v []byte) {
		elem := string(k)
		// If it's in mem, we already called f() for it in MemoryStore.Seek().
		_, present := s.mem[elem]
		if !present {
			// If it's in del, we shouldn't be calling f() anyway.
			_, present = s.del[elem]
		}
		if !present {
			f(k, v)
		}
	})
}

// Persist flushes all the MemoryStore contents into the (supposedly) persistent
// store ps.
func (s *MemCachedStore) Persist() (int, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	batch := s.ps.Batch()
	keys, dkeys := 0, 0
	for k, v := range s.mem {
		batch.Put([]byte(k), v)
		keys++
	}
	for k := range s.del {
		batch.Delete([]byte(k))
		dkeys++
	}
	var err error
	if keys != 0 || dkeys != 0 {
		err = s.ps.PutBatch(batch)
	}
	if err == nil {
		s.mem = make(map[string][]byte)
		s.del = make(map[string]bool)
	}
	return keys, err
}

// Checksum returns current storage contents checksum incrementally calculated
// by the storage change operations.
func (s *MemCachedStore) Checksum() Uint256 {
	return s.stateSum
}

// ChangeChecksum returns checksum for the current storage changeset relative
// to the persistent store.
func (s *MemCachedStore) ChangeChecksum() Uint256 {
	var calcChangeSum = Uint256{}

	for k, v := range s.mem {
		calcChangeSum.Xor(HashKV(k, v))
	}
	for k := range s.del {
		// Don't checksum if key is absent in the lower store, as it's
		// a no-op effectively.
		if _, err := s.ps.Get([]byte(k)); err == nil {
			calcChangeSum.Xor(sha256.Sum256([]byte(k)))
		}
	}
	return calcChangeSum
}

// Close implements Store interface, clears up memory and closes the lower layer
// Store.
func (s *MemCachedStore) Close() error {
	// It's always successful.
	_ = s.MemoryStore.Close()
	return s.ps.Close()
}
