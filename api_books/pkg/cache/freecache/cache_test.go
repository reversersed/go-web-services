package freecache

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntryCount(t *testing.T) {
	cache := NewCache(0)

	cache.Set([]byte("1"), []byte("1"), 0)
	cache.Set([]byte("2"), []byte("1"), 0)
	cache.Set([]byte("3"), []byte("1"), 0)

	assert.EqualValues(t, cache.EntryCount(), 3)

	cache.Set([]byte("0"), []byte("1"), 0)

	assert.EqualValues(t, cache.EntryCount(), 4)

	cache.Delete([]byte("1"))
	cache.Delete([]byte("2"))

	assert.EqualValues(t, cache.EntryCount(), 2)
}
func TestGetCache(t *testing.T) {
	cache := NewCache(0)

	cache.Set([]byte("1"), []byte("512"), 0)

	got, err := cache.Get([]byte("1"))

	if assert.NoError(t, err) {
		assert.Equal(t, string(got), "512")
		_, err = cache.Get([]byte("23123"))
		assert.Error(t, err)
	}
}

func TestSetCache(t *testing.T) {
	cache := NewCache(0) // min size 512 kb

	var body []byte
	for i := 0; i < 600; i++ { // creating byte slice of 600 bytes (if trying to insert 1/1024 value of cache size, error will be thrown)
		body = slices.Insert(body, 0, byte(2))
	}

	err := cache.Set([]byte("2"), body, 0)
	assert.Error(t, err)
}

func BenchmarkCacheSet(b *testing.B) {
	cache := NewCache(0)

	for i := 0; i < b.N; i++ {
		cache.Set([]byte{byte(i)}, []byte{byte(5)}, 0)
	}
}
func BenchmarkCacheGet(b *testing.B) {
	cache := NewCache(0)
	cache.Set([]byte{byte(5)}, []byte("hello"), 0)

	for i := 0; i < b.N; i++ {
		_, _ = cache.Get([]byte{byte(5)})
	}
}
func BenchmarkCacheSetAndDelete(b *testing.B) {
	cache := NewCache(0)

	for i := 0; i < b.N; i++ {
		cache.Set([]byte{byte(5)}, []byte{byte(5)}, 0)
		cache.Delete([]byte{byte(5)})
	}
}
func BenchmarkCacheEntryCount(b *testing.B) {
	cache := NewCache(0)
	cache.Set([]byte{byte(1)}, []byte{byte(1)}, 0)
	cache.Set([]byte{byte(2)}, []byte{byte(1)}, 0)
	cache.Set([]byte{byte(3)}, []byte{byte(1)}, 0)
	cache.Set([]byte{byte(4)}, []byte{byte(1)}, 0)

	for i := 0; i < b.N; i++ {
		_ = cache.EntryCount()
	}
}
func BenchmarkCacheErrorSet(b *testing.B) {
	cache := NewCache(0) // min size 512 kb

	var body []byte
	for i := 0; i < 600; i++ { // creating byte slice of 600 bytes (if trying to insert 1/1024 value of cache size, error will be thrown)
		body = slices.Insert(body, 0, byte(2))
	}

	for i := 0; i < b.N; i++ {
		_ = cache.Set([]byte("2"), body, 0)
	}
}
