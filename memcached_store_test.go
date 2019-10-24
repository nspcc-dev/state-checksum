package xorkv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemCachedStorePersist(t *testing.T) {
	// persistent Store
	ps := NewMemoryStore()
	// cached Store
	ts := NewMemCachedStore(ps)
	// persisting nothing should do nothing
	c, err := ts.Persist()
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, c)
	// persisting one key should result in one key in ps and nothing in ts
	assert.NoError(t, ts.Put([]byte("key"), []byte("value")))
	c, err = ts.Persist()
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, c)
	v, err := ps.Get([]byte("key"))
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte("value"), v)
	v, err = ts.MemoryStore.Get([]byte("key"))
	assert.Equal(t, ErrKeyNotFound, err)
	assert.Equal(t, []byte(nil), v)
	// now we overwrite the previous `key` contents and also add `key2`,
	assert.NoError(t, ts.Put([]byte("key"), []byte("newvalue")))
	assert.NoError(t, ts.Put([]byte("key2"), []byte("value2")))
	// this is to check that now key is written into the ps before we do
	// persist
	v, err = ps.Get([]byte("key2"))
	assert.Equal(t, ErrKeyNotFound, err)
	assert.Equal(t, []byte(nil), v)
	// two keys should be persisted (one overwritten and one new) and
	// available in the ps
	c, err = ts.Persist()
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, c)
	v, err = ts.MemoryStore.Get([]byte("key"))
	assert.Equal(t, ErrKeyNotFound, err)
	assert.Equal(t, []byte(nil), v)
	v, err = ts.MemoryStore.Get([]byte("key2"))
	assert.Equal(t, ErrKeyNotFound, err)
	assert.Equal(t, []byte(nil), v)
	v, err = ps.Get([]byte("key"))
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte("newvalue"), v)
	v, err = ps.Get([]byte("key2"))
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte("value2"), v)
	// we've persisted some values, make sure successive persist is a no-op
	c, err = ts.Persist()
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, c)
	// test persisting deletions
	err = ts.Delete([]byte("key"))
	assert.Equal(t, nil, err)
	c, err = ts.Persist()
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, c)
	v, err = ps.Get([]byte("key"))
	assert.Equal(t, ErrKeyNotFound, err)
	assert.Equal(t, []byte(nil), v)
	v, err = ps.Get([]byte("key2"))
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte("value2"), v)
}

func TestCachedGetFromPersistent(t *testing.T) {
	key := []byte("key")
	value := []byte("value")
	ps := NewMemoryStore()
	ts := NewMemCachedStore(ps)

	assert.NoError(t, ps.Put(key, value))
	val, err := ts.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, value, val)
	assert.NoError(t, ts.Delete(key))
	val, err = ts.Get(key)
	assert.Equal(t, err, ErrKeyNotFound)
	assert.Nil(t, val)
}

func TestCachedSeek(t *testing.T) {
	var (
		// Given this prefix...
		goodPrefix = []byte{'f'}
		// these pairs should be found...
		lowerKVs = []kvSeen{
			{[]byte("foo"), []byte("bar"), false},
			{[]byte("faa"), []byte("bra"), false},
		}
		// and these should be not.
		deletedKVs = []kvSeen{
			{[]byte("fee"), []byte("pow"), false},
			{[]byte("fii"), []byte("qaz"), false},
		}
		// and these should be not.
		updatedKVs = []kvSeen{
			{[]byte("fuu"), []byte("wop"), false},
			{[]byte("fyy"), []byte("zaq"), false},
		}
		ps = NewMemoryStore()
		ts = NewMemCachedStore(ps)
	)
	for _, v := range lowerKVs {
		require.NoError(t, ps.Put(v.key, v.val))
	}
	for _, v := range deletedKVs {
		require.NoError(t, ps.Put(v.key, v.val))
		require.NoError(t, ts.Delete(v.key))
	}
	for _, v := range updatedKVs {
		require.NoError(t, ps.Put(v.key, []byte("stub")))
		require.NoError(t, ts.Put(v.key, v.val))
	}
	foundKVs := make(map[string][]byte)
	ts.Seek(goodPrefix, func(k, v []byte) {
		foundKVs[string(k)] = v
	})
	assert.Equal(t, len(foundKVs), len(lowerKVs)+len(updatedKVs))
	for _, kv := range lowerKVs {
		assert.Equal(t, kv.val, foundKVs[string(kv.key)])
	}
	for _, kv := range deletedKVs {
		_, ok := foundKVs[string(kv.key)]
		assert.Equal(t, false, ok)
	}
	for _, kv := range updatedKVs {
		assert.Equal(t, kv.val, foundKVs[string(kv.key)])
	}
}

func TestCachedStateSimple(t *testing.T) {
	ps := NewMemoryStore()
	s := NewMemCachedStore(ps)
	h0 := Uint256{}
	kv1 := [][]byte{[]byte("key"), []byte("value")}
	kv2 := [][]byte{[]byte("foo"), []byte("bar")}
	kv3 := [][]byte{[]byte("bar"), []byte("baz")}
	kv3s := [][]byte{[]byte("bar"), []byte("zab")}

	// Put three KV pairs into the store
	require.NoError(t, s.Put(kv1[0], kv1[1]))
	require.NoError(t, s.Put(kv2[0], kv2[1]))
	require.NoError(t, s.Put(kv3[0], kv3[1]))

	// Change checksum and state checksum should match.
	require.Equal(t, s.ChangeChecksum(), s.Checksum())

	// After persisting state checksums in s and ps should match, but
	// change checksum should be zero now.
	_, err := s.Persist()
	require.Nil(t, err)
	require.Equal(t, ps.Checksum(), s.Checksum())
	require.Equal(t, h0, s.ChangeChecksum())
	kv123Sum := s.Checksum()

	// Delete kv3.
	require.NoError(t, s.Delete(kv3[0]))
	changeSumAfterDelete := s.ChangeChecksum()

	// Change checksum shouldn't change after the second delete.
	require.NoError(t, s.Delete(kv3[0]))
	require.Equal(t, changeSumAfterDelete, s.ChangeChecksum())

	// After persisting ps only has two key-value pairs, checksums should match.
	_, err = s.Persist()
	require.Nil(t, err)
	require.Equal(t, ps.Checksum(), s.Checksum())

	// Add and delete kv3 again, should be a noop.
	require.NoError(t, s.Put(kv3[0], kv3[1]))
	require.NoError(t, s.Delete(kv3[0]))
	require.Equal(t, ps.Checksum(), s.Checksum())
	require.Equal(t, h0, s.ChangeChecksum())

	// Put kv3 and update it while not persisting, change checksum should match
	// an updated KV pair.
	require.NoError(t, s.Put(kv3[0], kv3[1]))
	require.NoError(t, s.Put(kv3s[0], kv3s[1]))
	require.Equal(t, HashKV(string(kv3s[0]), kv3s[1]), s.ChangeChecksum())

	// Put old kv3 value and persist it, we should end up with the same sum as
	// in the first part of the test.
	require.NoError(t, s.Put(kv3[0], kv3[1]))
	_, err = s.Persist()
	require.Nil(t, err)
	require.Equal(t, ps.Checksum(), s.Checksum())
	require.Equal(t, kv123Sum, s.Checksum())

	// Update persisted kv3 and persist that change.
	require.NoError(t, s.Put(kv3s[0], kv3s[1]))
	_, err = s.Persist()
	require.Nil(t, err)
	require.Equal(t, ps.Checksum(), s.Checksum())
}

func newMemCachedStoreForTesting(t *testing.T) Store {
	return NewMemCachedStore(NewMemoryStore())
}
