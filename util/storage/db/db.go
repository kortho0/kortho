package db

import (
	"fmt"
	"kortho/util/mixed"
	"kortho/util/storage"

	"github.com/dgraph-io/badger"
)

func New(name string) storage.DB {
	opts := badger.DefaultOptions(name)
	if db, err := badger.Open(opts); err != nil {
		fmt.Println("===", err)
		return nil
	} else {
		return &bgStore{db}
	}
}

func (db *bgStore) Close() error {
	return db.db.Close()
}

func (db *bgStore) Del(k []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := del(tx, k); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Set(k, v []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := set(tx, k, v); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Get(k []byte) ([]byte, error) {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return get(tx, k)
}

func (db *bgStore) Mclear(m []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := mclear(tx, m); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Mdel(m, k []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := mdel(tx, m, k); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Mset(m, k, v []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := mset(tx, m, k, v); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Mget(m, k []byte) ([]byte, error) {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return mget(tx, m, k)
}

func (db *bgStore) Mkeys(m []byte) ([][]byte, error) {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return mkeys(tx, m)
}

func (db *bgStore) Mvals(m []byte) ([][]byte, error) {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return mvals(tx, m)
}

func (db *bgStore) Mkvs(m []byte) ([][]byte, [][]byte, error) {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return mkvs(tx, m)
}

func (db *bgStore) Llen(k []byte) int64 {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return llen(tx, k)
}

func (db *bgStore) Lclear(k []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := llclear(tx, k); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Llpush(k, v []byte) (int64, error) {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if n, err := llpush(tx, k, v); err != nil {
		return -1, err
	} else {
		if err = tx.Commit(); err != nil {
			return -1, err
		}
		return n, nil
	}
}

func (db *bgStore) Llpop(k []byte) ([]byte, error) {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	v, err := llpop(tx, k)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return v, nil
}

func (db *bgStore) Lrpush(k, v []byte) (int64, error) {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if n, err := lrpush(tx, k, v); err != nil {
		return -1, err
	} else {
		if err = tx.Commit(); err != nil {
			return -1, err
		}
		return n, nil
	}
}

func (db *bgStore) Lrpop(k []byte) ([]byte, error) {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	v, err := lrpop(tx, k)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return v, nil
}

func (db *bgStore) Lrange(k []byte, start, end int64) ([][]byte, error) {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return lrange(tx, k, start, end)
}

func (db *bgStore) Lset(k []byte, idx int64, v []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := lset(tx, k, idx, v); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Lindex(k []byte, idx int64) ([]byte, error) {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return lindex(tx, k, idx)
}

func (db *bgStore) Sclear(k []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := sclear(tx, k); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Sdel(k, v []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := sdel(tx, k, v); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Sadd(k, v []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := sadd(tx, k, v); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Selem(k, v []byte) (bool, error) {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return selem(tx, k, v)
}

func (db *bgStore) Smembers(k []byte) ([][]byte, error) {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return smembers(tx, k)
}

func (db *bgStore) Zclear(k []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := zclear(tx, k); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Zdel(k, v []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := zdel(tx, k, v); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Zadd(k []byte, score int32, v []byte) error {
	tx := db.db.NewTransaction(true)
	defer tx.Discard()
	if err := zadd(tx, k, score, v); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *bgStore) Zscore(k, v []byte) (int32, error) {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return zscore(tx, k, v)
}

func (db *bgStore) Zrange(k []byte, start, end int32) ([][]byte, error) {
	tx := db.db.NewTransaction(false)
	defer tx.Discard()
	return zrange(tx, k, start, end)
}

func (db *bgStore) NewTransaction() storage.Transaction {
	tx := db.db.NewTransaction(true)
	return &bgTransaction{tx}
}

func (tx *bgTransaction) Cancel() error {
	tx.tx.Discard()
	return nil
}

func (tx *bgTransaction) Commit() error {
	return tx.tx.Commit()
}

func (tx *bgTransaction) Del(k []byte) error {
	return del(tx.tx, k)
}

func (tx *bgTransaction) Set(k, v []byte) error {
	return set(tx.tx, k, v)
}

func (tx *bgTransaction) Get(k []byte) ([]byte, error) {
	return get(tx.tx, k)
}

func (tx *bgTransaction) Mclear(m []byte) error {
	return mclear(tx.tx, m)
}

func (tx *bgTransaction) Mdel(m, k []byte) error {
	return mdel(tx.tx, m, k)
}

func (tx *bgTransaction) Mset(m, k, v []byte) error {
	return mset(tx.tx, m, k, v)
}

func (tx *bgTransaction) Mget(m, k []byte) ([]byte, error) {
	return mget(tx.tx, m, k)
}

func (tx *bgTransaction) Mkeys(m []byte) ([][]byte, error) {
	return mkeys(tx.tx, m)
}

func (tx *bgTransaction) Mvals(m []byte) ([][]byte, error) {
	return mvals(tx.tx, m)
}

func (tx *bgTransaction) Mkvs(m []byte) ([][]byte, [][]byte, error) {
	return mkvs(tx.tx, m)
}

func (tx *bgTransaction) Llen(k []byte) int64 {
	return llen(tx.tx, k)
}

func (tx *bgTransaction) Lclear(k []byte) error {
	return llclear(tx.tx, k)
}

func (tx *bgTransaction) Llpush(k, v []byte) (int64, error) {
	return llpush(tx.tx, k, v)
}

func (tx *bgTransaction) Llpop(k []byte) ([]byte, error) {
	return llpop(tx.tx, k)
}

func (tx *bgTransaction) Lrpush(k, v []byte) (int64, error) {
	return lrpush(tx.tx, k, v)
}

func (tx *bgTransaction) Lrpop(k []byte) ([]byte, error) {
	return lrpop(tx.tx, k)
}

func (tx *bgTransaction) Lrange(k []byte, start, end int64) ([][]byte, error) {
	return lrange(tx.tx, k, start, end)
}

func (tx *bgTransaction) Lset(k []byte, idx int64, v []byte) error {
	return lset(tx.tx, k, idx, v)
}

func (tx *bgTransaction) Lindex(k []byte, idx int64) ([]byte, error) {
	return lindex(tx.tx, k, idx)
}

func (tx *bgTransaction) Sclear(k []byte) error {
	return sclear(tx.tx, k)
}

func (tx *bgTransaction) Sdel(k, v []byte) error {
	return sdel(tx.tx, k, v)
}

func (tx *bgTransaction) Sadd(k, v []byte) error {
	return sadd(tx.tx, k, v)
}

func (tx *bgTransaction) Selem(k, v []byte) (bool, error) {
	return selem(tx.tx, k, v)
}

func (tx *bgTransaction) Smembers(k []byte) ([][]byte, error) {
	return smembers(tx.tx, k)
}

func (tx *bgTransaction) Zclear(k []byte) error {
	return zclear(tx.tx, k)
}

func (tx *bgTransaction) Zdel(k, v []byte) error {
	return zdel(tx.tx, k, v)
}

func (tx *bgTransaction) Zadd(k []byte, score int32, v []byte) error {
	return zadd(tx.tx, k, score, v)
}

func (tx *bgTransaction) Zscore(k, v []byte) (int32, error) {
	return zscore(tx.tx, k, v)
}

func (tx *bgTransaction) Zrange(k []byte, start, end int32) ([][]byte, error) {
	return zrange(tx.tx, k, start, end)
}

func del(tx *badger.Txn, k []byte) error {
	return tx.Delete(k)
}

func set(tx *badger.Txn, k, v []byte) error {
	return tx.Set(k, v)
}

func get(tx *badger.Txn, k []byte) ([]byte, error) {
	it, err := tx.Get(k)
	if err == badger.ErrKeyNotFound {
		err = storage.NotExist
	}
	if err != nil {
		return nil, err
	}
	return it.ValueCopy(nil)
}

func mclear(tx *badger.Txn, m []byte) error {
	k := eMapKey(m, []byte{})
	opt := badger.DefaultIteratorOptions
	opt.Prefix = k
	opt.PrefetchValues = false
	itr := tx.NewIterator(opt)
	defer itr.Close()
	for itr.Seek(k); itr.ValidForPrefix(k); itr.Next() {
		if err := del(tx, itr.Item().Key()); err != nil {
			return err
		}
	}
	return nil
}

func mdel(tx *badger.Txn, m, k []byte) error {
	return del(tx, eMapKey(m, k))
}

func mset(tx *badger.Txn, m, k, v []byte) error {
	return set(tx, eMapKey(m, k), v)
}

func mget(tx *badger.Txn, m, k []byte) ([]byte, error) {
	return get(tx, eMapKey(m, k))
}

func mkeys(tx *badger.Txn, m []byte) ([][]byte, error) {
	var ks [][]byte

	k := eMapKey(m, []byte{})
	opt := badger.DefaultIteratorOptions
	opt.Prefix = k
	opt.PrefetchValues = false
	itr := tx.NewIterator(opt)
	defer itr.Close()
	for itr.Seek(k); itr.ValidForPrefix(k); itr.Next() {
		ks = append(ks, dMapKey(itr.Item().KeyCopy(nil)))
	}
	return ks, nil
}

func mvals(tx *badger.Txn, m []byte) ([][]byte, error) {
	var vs [][]byte

	k := eMapKey(m, []byte{})
	opt := badger.DefaultIteratorOptions
	opt.Prefix = k
	opt.PrefetchValues = false
	itr := tx.NewIterator(opt)
	defer itr.Close()
	for itr.Seek(k); itr.ValidForPrefix(k); itr.Next() {
		if v, err := itr.Item().ValueCopy(nil); err != nil {
			return nil, err
		} else {
			vs = append(vs, v)
		}
	}
	return vs, nil
}

func mkvs(tx *badger.Txn, m []byte) ([][]byte, [][]byte, error) {
	var ks, vs [][]byte

	k := eMapKey(m, []byte{})
	opt := badger.DefaultIteratorOptions
	opt.Prefix = k
	opt.PrefetchValues = false
	itr := tx.NewIterator(opt)
	defer itr.Close()
	for itr.Seek(k); itr.ValidForPrefix(k); itr.Next() {
		ks = append(ks, dMapKey(itr.Item().KeyCopy(nil)))
		if v, err := itr.Item().ValueCopy(nil); err != nil {
			return nil, nil, err
		} else {
			vs = append(vs, v)
		}
	}
	return ks, vs, nil
}

func lnew(tx *badger.Txn, k []byte) error {
	return set(tx, eListMetaKey(k), eListMetaValue(0, 0))
}

func llen(tx *badger.Txn, k []byte) int64 {
	if start, end, err := listStartEnd(tx, k); err != nil {
		return 0
	} else {
		return end - start
	}
}

func llclear(tx *badger.Txn, k []byte) error {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		return err
	}
	for ; start < end; start++ {
		if err := del(tx, eListKey(k, start)); err != nil {
			return err
		}
	}
	return del(tx, eListMetaKey(k))
}

func llpush(tx *badger.Txn, k, v []byte) (int64, error) {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		if err = lnew(tx, k); err != nil {
			return -1, err
		}
	}
	if start-1 == end {
		return -1, storage.OutOfSize
	}
	if err := set(tx, eListKey(k, start-1), v); err != nil {
		return -1, err
	}
	if err := set(tx, eListMetaKey(k), eListMetaValue(start-1, end)); err != nil {
		return -1, err
	}
	return end - start + 1, nil
}

func llpop(tx *badger.Txn, k []byte) ([]byte, error) {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		return nil, err
	}
	if start == end {
		return nil, storage.OutOfSize
	}
	v, err := get(tx, eListKey(k, start))
	if err != nil {
		return nil, err
	}
	if err := set(tx, eListMetaKey(k), eListMetaValue(start+1, end)); err != nil {
		return nil, err
	}
	return v, nil
}

func lrpush(tx *badger.Txn, k, v []byte) (int64, error) {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		if err = lnew(tx, k); err != nil {
			return -1, err
		}
	}
	if start == end+1 {
		return -1, storage.OutOfSize
	}
	if err := set(tx, eListKey(k, end), v); err != nil {
		return -1, err
	}
	if err := set(tx, eListMetaKey(k), eListMetaValue(start, end+1)); err != nil {
		return -1, err
	}
	return end - start + 1, nil
}

func lrpop(tx *badger.Txn, k []byte) ([]byte, error) {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		return nil, err
	}
	if start == end {
		return nil, storage.OutOfSize
	}
	v, err := get(tx, eListKey(k, end-1))
	if err != nil {
		return nil, err
	}
	if err := set(tx, eListMetaKey(k), eListMetaValue(start, end-1)); err != nil {
		return nil, err
	}
	return v, nil
}

func lset(tx *badger.Txn, k []byte, idx int64, v []byte) error {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		return err
	}
	switch {
	case idx >= 0:
		idx += start
	default:
		idx += end
	}
	if idx < start || idx >= end {
		return storage.OutOfSize
	}
	return set(tx, eListKey(k, idx+start), v)
}

func lindex(tx *badger.Txn, k []byte, idx int64) ([]byte, error) {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		return nil, err
	}
	switch {
	case idx >= 0:
		idx += start
	default:
		idx += end
	}
	if idx < start || idx >= end {
		return []byte{}, nil
	}
	return get(tx, eListKey(k, idx))
}

func lrange(tx *badger.Txn, k []byte, start, end int64) ([][]byte, error) {
	var vs [][]byte

	x, y, err := listStartEnd(tx, k)
	if err != nil {
		return nil, err
	}
	if end < 0 {
		end += y
	} else {
		end += x
	}
	if start < 0 {
		start += y
	} else {
		start += x
	}
	if start < x {
		start = x
	}
	if end >= y {
		end = y
	}
	for ; start <= end; start++ {
		if v, err := get(tx, eListKey(k, start)); err != nil {
			continue
		} else {
			vs = append(vs, v)
		}
	}
	return vs, nil
}

func sclear(tx *badger.Txn, k []byte) error {
	k = eSetKey(k, []byte{})
	opt := badger.DefaultIteratorOptions
	opt.Prefix = k
	opt.PrefetchValues = false
	itr := tx.NewIterator(opt)
	defer itr.Close()
	for itr.Seek(k); itr.ValidForPrefix(k); itr.Next() {
		if err := del(tx, itr.Item().Key()); err != nil {
			return err
		}
	}
	return nil
}

func sdel(tx *badger.Txn, k, v []byte) error {
	return del(tx, eSetKey(k, v))
}

func sadd(tx *badger.Txn, k, v []byte) error {
	return set(tx, eSetKey(k, v), []byte{})
}

func selem(tx *badger.Txn, k, v []byte) (bool, error) {
	_, err := get(tx, eSetKey(k, v))
	switch {
	case err == nil:
		return true, nil
	case err == storage.NotExist:
		return false, nil
	default:
		return false, err
	}
}

func smembers(tx *badger.Txn, k []byte) ([][]byte, error) {
	var vs [][]byte

	k = eSetKey(k, []byte{})
	opt := badger.DefaultIteratorOptions
	opt.Prefix = k
	opt.PrefetchValues = false
	itr := tx.NewIterator(opt)
	defer itr.Close()
	for itr.Seek(k); itr.ValidForPrefix(k); itr.Next() {
		vs = append(vs, dSetKey(itr.Item().KeyCopy(nil)))
	}
	return vs, nil
}

func zclear(tx *badger.Txn, k []byte) error {
	key := []byte{}
	key = append([]byte("sz"), mixed.E32func(uint32(len(k)))...)
	key = append(key, k...)
	key = append(key, byte('+'))
	opt := badger.DefaultIteratorOptions
	opt.Prefix = key
	opt.PrefetchValues = false
	itr := tx.NewIterator(opt)
	defer itr.Close()
	for itr.Seek(key); itr.ValidForPrefix(key); itr.Next() {
		score, v := dZetScore(itr.Item().Key())
		if err := del(tx, eZetKey(k, v)); err != nil {
			return err
		}
		if err := del(tx, eZetScore(k, v, score)); err != nil {
			return err
		}
	}
	return nil
}

func zdel(tx *badger.Txn, k, v []byte) error {
	key := eZetKey(k, v)
	buf, err := get(tx, key)
	if err != nil {
		return err
	}
	if err := del(tx, key); err != nil {
		return err
	}
	score, _ := mixed.D32func(buf)
	if err := del(tx, eZetScore(k, v, int32(score))); err != nil {
		return err
	}
	return nil
}

func zscore(tx *badger.Txn, k, v []byte) (int32, error) {
	if buf, err := get(tx, eZetKey(k, v)); err != nil {
		return -1, err
	} else {
		score, _ := mixed.D32func(buf)
		return int32(score), nil
	}
}

func zadd(tx *badger.Txn, k []byte, score int32, v []byte) error {
	if err := set(tx, eZetKey(k, v), mixed.E32func(uint32(score))); err != nil {
		return err
	}
	if err := set(tx, eZetScore(k, v, score), []byte{}); err != nil {
		return err
	}
	return nil
}

func zrange(tx *badger.Txn, k []byte, start, end int32) ([][]byte, error) {
	var vs [][]byte

	key := []byte{}
	key = append([]byte("sz"), mixed.E32func(uint32(len(k)))...)
	key = append(key, k...)
	key = append(key, byte('+'))
	opt := badger.DefaultIteratorOptions
	opt.Prefix = key
	opt.PrefetchValues = false
	itr := tx.NewIterator(opt)
	defer itr.Close()
	for itr.Seek(eZetScore(k, []byte{}, start)); itr.ValidForPrefix(key); itr.Next() {
		if score, v := dZetScore(itr.Item().Key()); score > end {
			break
		} else {
			vs = append(vs, mixed.Dup(v))
		}
	}
	return vs, nil
}

func listStartEnd(tx *badger.Txn, k []byte) (int64, int64, error) {
	if v, err := get(tx, eListMetaKey(k)); err != nil {
		return 0, 0, err
	} else {
		start, end := dListMetaValue(v)
		return start, end, nil
	}
}

func eListMetaKey(k []byte) []byte {
	return append([]byte{'l'}, k...)
}

func dListMetaKey(buf []byte) []byte {
	return buf[1:]
}

func eListMetaValue(start, end int64) []byte {
	return append(mixed.E64func(uint64(start)), mixed.E64func(uint64(end))...)
}

func dListMetaValue(buf []byte) (int64, int64) {
	start, _ := mixed.D64func(buf[:8])
	end, _ := mixed.D64func(buf[8:16])
	return int64(start), int64(end)
}

func eListKey(k []byte, idx int64) []byte {
	buf := []byte{}
	buf = append([]byte{'l'}, k...)
	buf = append(buf, mixed.E64func(uint64(idx))...)
	return buf
}

func dListKey(buf []byte) []byte {
	n := len(buf)
	return buf[1 : n-8]
}

func eMapKey(m, k []byte) []byte {
	buf := []byte{}
	buf = append([]byte{'m'}, mixed.E32func(uint32(len(m)))...)
	buf = append(buf, m...)
	buf = append(buf, byte('+'))
	buf = append(buf, k...)
	return buf
}

func dMapKey(buf []byte) []byte {
	buf = buf[1:]
	n, _ := mixed.D32func(buf[:4])
	return buf[5+n:]
}

func eSetKey(k, v []byte) []byte {
	buf := []byte{}
	buf = append([]byte{'s'}, mixed.E32func(uint32(len(k)))...)
	buf = append(buf, k...)
	buf = append(buf, byte('+'))
	buf = append(buf, v...)
	return buf
}

func dSetKey(buf []byte) []byte {
	buf = buf[1:]
	n, _ := mixed.D32func(buf[:4])
	return buf[5+n:]
}

func eZetKey(k, v []byte) []byte {
	buf := []byte{}
	buf = append([]byte{'z'}, mixed.E32func(uint32(len(k)))...)
	buf = append(buf, k...)
	buf = append(buf, byte('+'))
	buf = append(buf, v...)
	return buf
}

func dZetKey(buf []byte) []byte {
	buf = buf[1:]
	n, _ := mixed.D32func(buf[:4])
	return buf[5+n:]
}

func eZetScore(k, v []byte, score int32) []byte {
	buf := []byte{}
	buf = append([]byte("sz"), mixed.E32func(uint32(len(k)))...)
	buf = append(buf, k...)
	buf = append(buf, byte('+'))
	buf = append(buf, mixed.EB32func(uint32(score))...)
	buf = append(buf, v...)
	return buf
}

func dZetScore(buf []byte) (int32, []byte) {
	buf = buf[2:]
	n, _ := mixed.D32func(buf[:4])
	score, _ := mixed.DB32func(buf[5+n : 9+n])
	return int32(score), buf[9+n:]
}
