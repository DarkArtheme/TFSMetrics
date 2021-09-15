package store

import (
	tfsmetrics "go-marathon-team-3/pkg/TFSMetrics"

	bolt "go.etcd.io/bbolt"
)

type Store interface {
	Open() error
	Close()
	FindOne(id int) (*tfsmetrics.Commit, error)
	Write(commit tfsmetrics.Commit) error
	WriteBatch(commits []tfsmetrics.Commit) error
}

type DB struct {
	db        *bolt.DB
	batch     []*tfsmetrics.Commit
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

func (db *DB) FindOne(id int) (*tfsmetrics.Commit, error) {
	return nil, nil
}

func (db *DB) Write(commit tfsmetrics.Commit) error {

	return nil
}

func (db *DB) WriteBatch(commits []tfsmetrics.Commit) error {
	return nil
}
