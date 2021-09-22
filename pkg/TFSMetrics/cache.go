package tfsmetrics

import (
	"log"
	"sync"
	"time"
)

var idChan chan int = make(chan int, 10)
var wg sync.WaitGroup = sync.WaitGroup{}

type Cacher interface {
	Cache(iterator CommitIterator) (CommitIterator, error)
}

type repositoryCache struct {
	store Store
}

func NewCacher(projectName string, store Store) Cacher {
	return &repositoryCache{
		store: store,
	}
}

func (rc *repositoryCache) Cache(iterator CommitIterator) (CommitIterator, error) {
	commit, err := iterator.Next()
	if err != nil {
		return nil, err
	}
	rc.store.Write(commit)
	storeIterator := NewStoreIterator(commit, rc.store)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for commit, err := iterator.Next(); err == nil; commit, err = iterator.Next() {
			err := rc.store.Write(commit)
			if err != nil {
				log.Panic(err)
				return
			}
			idChan <- commit.Id
		}
	}()
	return storeIterator, nil
}

type storeIterator struct {
	index int
	ids   []int

	store Store
}

func NewStoreIterator(commit *Commit, store Store) CommitIterator {
	si := &storeIterator{
		index: 0,
		ids:   []int{commit.Id},
		store: store,
	}
	go func(iter storeIterator) {
		for {
			id := <-idChan
			si.ids = append(si.ids, id)
		}
	}(*si)
	return si
}

func (si *storeIterator) Next() (*Commit, error) {
	for i := 0; i < 3; i++ {
		if si.index < len(si.ids) {
			si.index++
			commit, err := si.store.FindOne(si.ids[si.index-1])
			if err != nil {
				return nil, err
			}
			return commit, nil
		} else {
			time.Sleep(time.Millisecond * 500)
		}
	}
	wg.Wait()
	if si.index < len(si.ids) {
		si.index++
		commit, err := si.store.FindOne(si.ids[si.index-1])
		if err != nil {
			return nil, err
		}
		return commit, nil
	}
	return nil, errNoMoreItems
}
