package tfsmetrics

var idChan chan int = make(chan int, 10)

type Cacher interface {
	Cache(iterator CommitIterator) (CommitIterator, error)
}

type repositoryCache struct {
	store Store
}

func (rc *repositoryCache) Cache(iterator CommitIterator) (CommitIterator, error) {
	commit, err := iterator.Next()
	if err != nil {
		return nil, err
	}
	rc.store.Write(commit)
	storeIterator := NewStoreIterator(commit, rc.store)
	go func() {
		for commit, err := iterator.Next(); err == nil; commit, err = iterator.Next() {
			idChan <- commit.Id
			rc.store.Write(commit)
		}
	}()
	return storeIterator, nil
}

type storeIterator struct {
	index int
	ids   []int

	projectName string
	store       Store
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
	if si.index < len(si.ids) {
		si.index++
		commit, err := si.store.FindOne(si.ids[si.index-1], si.projectName)
		if err != nil {
			return nil, err
		}
		return commit, nil
	}
	return nil, errNoMoreItems
}
