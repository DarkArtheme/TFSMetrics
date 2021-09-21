package cache

import (
	"go-marathon-team-3/pkg/tfsmetrics/repointerface"
	"go-marathon-team-3/pkg/tfsmetrics/store"
	"log"
	"sync"
	"time"
)

var wg sync.WaitGroup = sync.WaitGroup{}

type Cacher interface {
	Cache(iterator repointerface.CommitIterator) (repointerface.CommitIterator, error)
}

type repositoryCache struct {
	store  store.Store
	idChan chan int
}

func NewCacher(projectName string, store store.Store) Cacher {
	return &repositoryCache{
		store:  store,
		idChan: make(chan int, 10),
	}
}

func (rc *repositoryCache) Cache(iterator repointerface.CommitIterator) (repointerface.CommitIterator, error) {
	commit, err := iterator.Next()
	if err != nil {
		return nil, err
	}
	rc.store.Write(commit)
	storeIterator := NewStoreIterator(commit, rc.store, rc.idChan)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for commit, err := iterator.Next(); err == nil; commit, err = iterator.Next() {
			err := rc.store.Write(commit)
			if err != nil {
				log.Panic(err)
				return
			}
			rc.idChan <- commit.Id
		}
	}()
	return storeIterator, nil
}

type storeIterator struct {
	index int
	ids   []int

	store store.Store
}

func NewStoreIterator(commit *repointerface.Commit, store store.Store, idChan chan int) repointerface.CommitIterator {
	si := &storeIterator{
		index: 0,
		ids:   []int{commit.Id},
		store: store,
	}
	go func() {
		for {
			id := <-idChan
			si.ids = append(si.ids, id)
		}
	}()
	return si
}

func (si *storeIterator) Next() (*repointerface.Commit, error) {
	for i := 0; i < 3; i++ {
		if si.index < len(si.ids) {
			commit, err := si.store.FindOne(si.ids[si.index])
			si.index++
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
	return nil, repointerface.ErrNoMoreItems
}
