package kvstore

type Client[T any] interface {
	Set(key string, value T)
	Get(key string) (T, error)
	Delete(key string)
}
