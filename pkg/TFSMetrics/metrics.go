package tfsmetrics

import (
	"errors"
	"go-marathon-team-3/pkg/TFSMetrics/azure"
	"time"
)

var errNoMoreItems error = errors.New("no more items")

type Repository interface {
	Open() error // вызываем azure.TfvcClientConnection()
	GetCommitIterator() (CommitIterator, error)
}

type CommitIterator interface {
	Next() (*Commit, error)
}

type Commit struct {
	Id          int
	Author      string // обязательное поле
	Email       string
	AddedRows   int       // обязательное поле
	DeletedRows int       // обязательное поле
	Date        time.Time // обязательное поле
	Message     string
	Hash        string
}

type commitsCollection struct {
	nameOfProject string
	azure         azure.AzureInterface
}

func NewCommitCollection(nameOfProject string, azure *azure.Azure) Repository {
	return &commitsCollection{
		nameOfProject: nameOfProject,
		azure:         azure,
	}
}

func (c *commitsCollection) Open() error {
	return c.azure.TfvcClientConnection()
}

func (c *commitsCollection) GetCommitIterator() (CommitIterator, error) {
	changeSets, err := c.azure.GetChangesets(c.nameOfProject)
	if err != nil {
		return nil, err
	}
	return &iterator{
		nameOfProject: c.nameOfProject,
		azure:         c.azure.Azure(),
		commits:       changeSets,
	}, nil
}

type iterator struct {
	index         int
	nameOfProject string
	azure         azure.AzureInterface
	commits       []*int
}

func (i *iterator) Next() (*Commit, error) {
	if i.index < len(i.commits) {
		i.index++
		changeSet, err := i.azure.GetChangesetChanges(i.commits[i.index-1], i.nameOfProject)
		if err != nil {
			return nil, err
		}
		return &Commit{
			Id:          changeSet.Id,
			Author:      changeSet.Author,
			Email:       changeSet.Email,
			AddedRows:   changeSet.AddedRows,
			DeletedRows: changeSet.DeletedRows,
			Date:        changeSet.Date,
			Message:     changeSet.Message,
			Hash:        changeSet.Hash,
		}, nil
	}
	return nil, errNoMoreItems
}
