package tfsmetrics

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_repositoryCache_Cache(t *testing.T) {
	iter := testIterator{
		index: 0,
		commits: []Commit{
			{Id: 1,
				Author: "Ivan"},
			{Id: 2,
				Author: "Peter"},
			{Id: 3,
				Author: "Vity"},
		},
	}

	store, err := TestStore()
	require.NoError(t, err)
	defer store.Close()
	defer func() {
		os.Remove(store.db.Path())
	}()

	cacher := repositoryCache{
		store: store,
	}

	iterator, err := cacher.Cache(&iter)
	require.NoError(t, err)

	time.Sleep(time.Second)
	commit, err := iterator.Next()
	assert.NoError(t, err)
	assert.Equal(t, &Commit{Id: 1,
		Author: "Ivan"}, commit)

	commit, err = iterator.Next()
	assert.NoError(t, err)
	assert.Equal(t, &Commit{Id: 2,
		Author: "Peter"}, commit)

	commit, err = iterator.Next()
	assert.NoError(t, err)
	assert.Equal(t, &Commit{Id: 3,
		Author: "Vity"}, commit)

	commit, err = iterator.Next()
	assert.Error(t, err)
	assert.Nil(t, commit)
}
