package freecache

import (
	"slices"
	"testing"
)

func TestEntryCount(t *testing.T) {
	cache := NewCache(0)

	cache.Set([]byte("1"), []byte("1"), 0)
	cache.Set([]byte("2"), []byte("1"), 0)
	cache.Set([]byte("3"), []byte("1"), 0)

	if cache.EntryCount() != 3 {
		t.Errorf("excepted entry count 3 but got %d", cache.EntryCount())
	}

	cache.Set([]byte("0"), []byte("1"), 0)

	if cache.EntryCount() != 4 {
		t.Errorf("excepted entry count 4 but got %d", cache.EntryCount())
	}

	cache.Delete([]byte("1"))
	cache.Delete([]byte("2"))

	if cache.EntryCount() != 2 {
		t.Errorf("excepted entry count 2 but got %d", cache.EntryCount())
	}
}
func TestGetCache(t *testing.T) {
	cache := NewCache(0)

	cache.Set([]byte("1"), []byte("512"), 0)

	got, err := cache.Get([]byte("1"))
	if err != nil {
		t.Errorf("excepted value but got error %v", err)
	}
	if string(got) != "512" {
		t.Errorf("excepted value 512 but got %s", string(got))
	}

	_, err = cache.Get([]byte("23123"))
	if err == nil {
		t.Error("excepted error but got nil")
	}
}

func TestSetCache(t *testing.T) {
	cache := NewCache(0) // min size 512 kb

	var body []byte
	for i := 0; i < 600; i++ { // creating byte slice of 600 bytes (if trying to insert 1/1024 value of cache size, error will be thrown)
		body = slices.Insert(body, 0, byte(2))
	}

	err := cache.Set([]byte("2"), body, 0)
	if err == nil {
		t.Error("excepted error but got nil")
	}
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
