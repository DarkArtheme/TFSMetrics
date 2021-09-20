package tfsmetrics

import (
	"go-marathon-team-3/pkg/tfsmetrics/azure"
	"go-marathon-team-3/pkg/tfsmetrics/repointerface"
)

type commitsCollection struct {
	nameOfProject string
	azure         azure.AzureInterface
}

func (c *commitsCollection) Open() error {
	return c.azure.TfvcClientConnection()
}

func (c *commitsCollection) GetCommitIterator() (repointerface.CommitIterator, error) {
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

func (i *iterator) Next() (*repointerface.Commit, error) {
	if i.index < len(i.commits) {
		i.index++
		changeSet, err := i.azure.GetChangesetChanges(i.commits[i.index-1], i.nameOfProject)
		if err != nil {
			return nil, err
		}
		return &repointerface.Commit{
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
	return nil, repointerface.ErrNoMoreItems
}
