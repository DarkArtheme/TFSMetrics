package tfsmetrics

type Cacher interface {
	Cache(iterator CommitIterator) (CommitIterator, error)
}

type repositoryCache struct{}

func (rc *repositoryCache) Cache(CommitIterator) (CommitIterator, error) {
	return nil, nil
}
