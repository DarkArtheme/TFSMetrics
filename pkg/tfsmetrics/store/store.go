package store

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"go-marathon-team-3/pkg/tfsmetrics/repointerface"

	bolt "go.etcd.io/bbolt"
)

type Store interface {
	Open() error
	Close() error
	FindOne(id int) (*repointerface.Commit, error)
	Write(commit *repointerface.Commit) error
}

type DB struct {
	DB          *bolt.DB
	ProjectName string
}

func NewStore(pn string) Store {
	return &DB{
		ProjectName: pn,
	}
}

func (db *DB) Open() error {
	bolt, err := bolt.Open("assets.db", 0600, nil)
	if err != nil {
		return err
	}

	db.DB = bolt
	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) FindOne(id int) (*repointerface.Commit, error) {
	res := &repointerface.Commit{}
	err := db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.ProjectName))
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

func (db *DB) Write(commit *repointerface.Commit) error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(db.ProjectName))
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
