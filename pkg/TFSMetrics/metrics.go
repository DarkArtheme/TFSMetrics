package tfsmetrics

import "time"

type Repository interface {
	Open() error
	GetCommitIterator() (CommitIterator, error)
}

type CommitIterator interface {
	Next() (*Commit, error)
}

type Commit struct {
	Author      string
	Email       string
	AddedRows   int
	DeletedRows int
	Date        time.Time
	Message     string
	Hash        string
}
