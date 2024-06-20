package freecache

import (
	"sync"

	"github.com/coocood/freecache"
)

type cacherepo struct {
	sync.Mutex
	cache *freecache.Cache
}

func NewCache(size int) *cacherepo {
	return &cacherepo{cache: freecache.NewCache(size)}
}
func (c *cacherepo) EntryCount() int64 {
	c.Lock()
	defer c.Unlock()

	return c.cache.EntryCount()
}
func (r *cacherepo) Get(uuid []byte) ([]byte, error) {
	r.Lock()
	defer r.Unlock()
	got, err := r.cache.Get(uuid)
	return got, err
}

func (r *cacherepo) Set(key, val []byte, expireIn int) error {
	r.Lock()
	defer r.Unlock()

	err := r.cache.Set(key, val, expireIn)
	if err != nil {
		return err
	}
	return nil
}

func (r *cacherepo) Delete(key []byte) (affected bool) {
	r.Lock()
	defer r.Unlock()

	return r.cache.Del(key)
}
