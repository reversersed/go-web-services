package freecache

import (
	"sync"

	"github.com/coocood/freecache"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/cache"
)

type cacherepo struct {
	sync.Mutex
	cache *freecache.Cache
}

func (c *cacherepo) EntryCount() int64 {
	c.Lock()
	defer c.Unlock()

	return c.cache.EntryCount()
}
func NewCache(size int) cache.Cache {
	return &cacherepo{cache: freecache.NewCache(size)}
}

func (r *cacherepo) GetIterator() cache.Iterator {
	return &iterator{r.cache.NewIterator()}
}

func (r *cacherepo) Get(uuid []byte) ([]byte, error) {
	r.Lock()
	defer r.Unlock()
	return r.cache.Get(uuid)
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
