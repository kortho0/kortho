package storage

import "errors"

var (
	NotExist  = errors.New("NotExist")
	OutOfSize = errors.New("Out of Size")
)

type DB interface {
	Close() error
	Del([]byte) error
	Set([]byte, []byte) error
	Get([]byte) ([]byte, error)

	Mclear([]byte) error
	Mdel([]byte, []byte) error
	Mkeys([]byte) ([][]byte, error)
	Mvals([]byte) ([][]byte, error)
	Mset([]byte, []byte, []byte) error
	Mget([]byte, []byte) ([]byte, error)
	Mkvs([]byte) ([][]byte, [][]byte, error)

	Llen([]byte) int64
	Lclear([]byte) error
	Llpop([]byte) ([]byte, error)
	Lrpop([]byte) ([]byte, error)
	Lset([]byte, int64, []byte) error
	Lrpush([]byte, []byte) (int64, error)
	Llpush([]byte, []byte) (int64, error)
	Lindex([]byte, int64) ([]byte, error)
	Lrange([]byte, int64, int64) ([][]byte, error)

	Sclear([]byte) error
	Sadd([]byte, []byte) error
	Sdel([]byte, []byte) error
	Smembers([]byte) ([][]byte, error)
	Selem([]byte, []byte) (bool, error)

	Zclear([]byte) error
	Zdel([]byte, []byte) error
	Zadd([]byte, int32, []byte) error
	Zscore([]byte, []byte) (int32, error)
	Zrange([]byte, int32, int32) ([][]byte, error)

	NewTransaction() Transaction
}

type Transaction interface {
	Commit() error
	Cancel() error

	Del([]byte) error
	Set([]byte, []byte) error
	Get([]byte) ([]byte, error)

	Mclear([]byte) error
	Mdel([]byte, []byte) error
	Mkeys([]byte) ([][]byte, error)
	Mvals([]byte) ([][]byte, error)
	Mset([]byte, []byte, []byte) error
	Mget([]byte, []byte) ([]byte, error)
	Mkvs([]byte) ([][]byte, [][]byte, error)

	Llen([]byte) int64
	Lclear([]byte) error
	Llpop([]byte) ([]byte, error)
	Lrpop([]byte) ([]byte, error)
	Lset([]byte, int64, []byte) error
	Lrpush([]byte, []byte) (int64, error)
	Llpush([]byte, []byte) (int64, error)
	Lindex([]byte, int64) ([]byte, error)
	Lrange([]byte, int64, int64) ([][]byte, error)

	Sclear([]byte) error
	Sadd([]byte, []byte) error
	Sdel([]byte, []byte) error
	Smembers([]byte) ([][]byte, error)
	Selem([]byte, []byte) (bool, error)

	Zclear([]byte) error
	Zdel([]byte, []byte) error
	Zadd([]byte, int32, []byte) error
	Zscore([]byte, []byte) (int32, error)
	Zrange([]byte, int32, int32) ([][]byte, error)
}
