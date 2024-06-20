package cache

//go:generate mockgen -source=cache.go -destination=mocks/cache.go

type Cache interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte, expiration int) error
	Delete(key []byte) (affected bool)

	EntryCount() int64
}
