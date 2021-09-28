package exporter

import (
	"go-marathon-team-3/pkg/tfsmetrics/repointerface"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter interface {
	GetDataByProject(iterator repointerface.CommitIterator, project string) map[string]ByProject
	GetDataByAuthor(iterator repointerface.CommitIterator, author string, project string) map[string]ByAuthor
	// Принимает КОПИЮ итератора и создает по нему метрики для проекта
	GetProjectMetrics(iterator repointerface.CommitIterator, project string)
}

type metrics struct {
	commits     prometheus.Counter
	addedRows   prometheus.Counter
	deletedRows prometheus.Counter
}

func newMetrics(author string, email string, project string) *metrics {
	m := &metrics{
		commits: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "commits",
			Help:        "commits counter",
			ConstLabels: map[string]string{"author": author, "email": email, "project": project},
		}),
		addedRows: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "added_rows",
			Help:        "added_rows counter",
			ConstLabels: map[string]string{"author": author, "email": email, "project": project},
		}),
		deletedRows: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "deleted_rows",
			Help:        "deleted_rows counter",
			ConstLabels: map[string]string{"author": author, "email": email, "project": project},
		}),
	}
	prometheus.MustRegister(m.commits, m.addedRows, m.deletedRows)
	return m
}

type exporter struct {
	authors      map[string]*metrics
	dataByAuthor map[string]ByAuthor
}

func NewExporter() Exporter {
	return &exporter{
		authors:      make(map[string]*metrics),
		dataByAuthor: make(map[string]ByAuthor),
	}
}

func (e *exporter) GetProjectMetrics(iterator repointerface.CommitIterator, project string) {
	for commit, err := iterator.Next(); err == nil; commit, err = iterator.Next() {
		if m, ok := e.authors[commit.Author]; ok {
			m.commits.Inc()
			m.addedRows.Add(float64(commit.AddedRows))
			m.deletedRows.Add(float64(commit.DeletedRows))
		} else {
			e.authors[commit.Author] = newMetrics(commit.Author, commit.Email, project)
			e.authors[commit.Author].commits.Inc()
			e.authors[commit.Author].addedRows.Add(float64(commit.AddedRows))
			e.authors[commit.Author].deletedRows.Add(float64(commit.DeletedRows))
		}
	}
}

type ByAuthor struct {
	Projects    []string
	Commits     int
	AddedRows   int
	DeletedRows int
}

type ByProject struct {
	Author      string
	Commits     int
	AddedRows   int
	DeletedRows int
}

func (e *exporter) GetDataByProject(iterator repointerface.CommitIterator, project string) map[string]ByProject {
	res := make(map[string]ByProject)
	for commit, err := iterator.Next(); err == nil; commit, err = iterator.Next() {
		if proj, ok := res[project]; ok {
			proj.Commits += 1
			proj.AddedRows += commit.AddedRows
			proj.DeletedRows += commit.DeletedRows
		} else {
			res[project] = ByProject{
				Author:      commit.Author,
				Commits:     1,
				AddedRows:   commit.AddedRows,
				DeletedRows: commit.DeletedRows,
			}
		}
	}
	return res
}

func (e *exporter) GetDataByAuthor(iterator repointerface.CommitIterator, author string, project string) map[string]ByAuthor {
	for commit, err := iterator.Next(); err == nil; commit, err = iterator.Next() {
		if auth, ok := e.dataByAuthor[author]; ok {
			is := false
			for _, v := range auth.Projects {
				if v == project {
					is = true
					break
				}
			}
			if !is {
				auth.Projects = append(auth.Projects, project)
			}
			auth.Commits += 1
			auth.AddedRows += commit.AddedRows
			auth.DeletedRows += commit.DeletedRows
		} else {
			e.dataByAuthor[author] = ByAuthor{
				Projects:    []string{project},
				Commits:     1,
				AddedRows:   commit.AddedRows,
				DeletedRows: commit.DeletedRows,
			}
		}
	}
	return nil
}
