package tfsmetrics

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_FindOne(t *testing.T) {
	store, err := TestStore()
	require.NoError(t, err)
	defer store.Close()
	defer func() {
		os.Remove(store.db.Path())
	}()
	commit := Commit{
		Id:          1,
		Author:      "ivan",
		Email:       "example@example.com",
		AddedRows:   1,
		DeletedRows: 2,
		Date:        time.Time{},
		Message:     "hello world",
		Hash:        "",
	}

	err = store.Write(&commit)
	require.NoError(t, err)

	tests := []struct {
		name string

		id      int
		want    *Commit
		wantErr bool
	}{
		{
			name:    "ok",
			id:      commit.Id,
			want:    &commit,
			wantErr: false,
		},
		{
			name:    "no item",
			id:      2,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := store.FindOne(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, c)
		})
	}
}

func TestDB_Write(t *testing.T) {
	store, err := TestStore()
	require.NoError(t, err)
	defer store.Close()
	defer func() {
		os.Remove(store.db.Path())
	}()
	commit := Commit{
		Id:          1,
		Author:      "ivan",
		Email:       "example@example.com",
		AddedRows:   1,
		DeletedRows: 2,
		Date:        time.Time{},
		Message:     "hello world",
		Hash:        "",
	}

	tests := []struct {
		name string

		commit *Commit
	}{
		{
			name:   "ok",
			commit: &commit,
		},
		{
			name:   "equal data",
			commit: &commit,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Write(tt.commit)
			assert.NoError(t, err)

			c, err := store.FindOne(tt.commit.Id)
			assert.NoError(t, err)
			assert.Equal(t, tt.commit, c)
		})
	}
}
