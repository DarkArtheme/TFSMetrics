package tfsmetrics

import (
	"encoding/binary"
	"encoding/json"
	"errors"

	bolt "go.etcd.io/bbolt"
)

type Store interface {
	Open() error
	Close() error
	FindOne(id int, prName string) (*Commit, error)
	Write(commit *Commit) error
	WriteBatch() error
}

type DB struct {
	db        *bolt.DB
	batch     []*Commit
	batchSize int
}

func (db *DB) Open() error {
	bolt, err := bolt.Open("assets/statistics.db", 0600, nil)
	if err != nil {
		return err
	}

	db.db = bolt
	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) FindOne(id int, prName string) (*Commit, error) {
	if len(db.batch) != 0 {
		for _, v := range db.batch {
			if v.Id == id {
				return v, nil
			}
		}
	}
	res := &Commit{}
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(prName))
		v := b.Get(itob(id))

		if v == nil {
			return errors.New("no item")
		}
		if err := json.Unmarshal(v, res); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (db *DB) Write(commit *Commit) error {
	db.batch = append(db.batch, commit)
	if len(db.batch) == db.batchSize {
		if err := db.WriteBatch(); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) WriteBatch() error {
	return nil
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
