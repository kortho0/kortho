package db

import "github.com/dgraph-io/badger"

type bgStore struct {
	db *badger.DB
}

type bgTransaction struct {
	tx *badger.Txn
}
