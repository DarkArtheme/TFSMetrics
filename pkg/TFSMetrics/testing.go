package tfsmetrics

import (
	bolt "go.etcd.io/bbolt"
)

type testIterator struct {
	index   int
	commits []Commit
}

func (ti *testIterator) Next() (*Commit, error) {
	if ti.index < len(ti.commits) {
		ti.index++
		return &ti.commits[ti.index-1], nil
	}
	return nil, errNoMoreItems
}

func TestStore() (*DB, error) {
	bolt, err := bolt.Open("pkg", 0600, nil)
	if err != nil {
		return nil, err
	}
	store := DB{
		db:          bolt,
		projectName: "project1",
	}
	return &store, nil
}
