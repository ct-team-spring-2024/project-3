package internal

type db interface {
	Delete(key string) (bool, error)
	Set(Key string, value []byte) error
	Get(key string) ([]byte, error)
}
