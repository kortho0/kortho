package db

const PAGE_SIZE = 4 * 1024 * 1024

type db struct {
	name string
}

type DB interface {
	Close() error
	Del([]byte) error
	SetExecute([]byte) error
	Set([]byte, []byte) error
	Get([]byte) ([]byte, error)
	GetExecute() ([]byte, error)
}
