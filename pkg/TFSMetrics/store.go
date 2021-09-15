package tfsmetrics

import (
	bolt "go.etcd.io/bbolt"
)

type Store interface {
	Open() error
	Close()
	FindOne(id int) (*Commit, error)
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

func (db *DB) Close() {
	db.Close()
}

func (db *DB) FindOne(id int) (*Commit, error) {
	return nil, nil
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
