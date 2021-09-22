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
	FindOne(id int) (*Commit, error)
	Write(commit *Commit) error
}

type DB struct {
	db          *bolt.DB
	projectName string
}

func NewStore(pn string) Store {
	return &DB{
		projectName: pn,
	}
}

func (db *DB) Open() error {
	bolt, err := bolt.Open("assets.db", 0600, nil)
	if err != nil {
		return err
	}

	db.db = bolt
	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) FindOne(id int) (*Commit, error) {
	res := &Commit{}
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.projectName))
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
	err := db.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(db.projectName))
		if err != nil {
			return err
		}
		v := b.Get(itob(commit.Id))

		if v != nil {
			return nil
		}

		buf, err := json.Marshal(commit)
		if err != nil {
			return err
		}

		return b.Put(itob(commit.Id), buf)
	})
	if err != nil {
		return err
	}

	return nil
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
