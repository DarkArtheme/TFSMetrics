package exporter

import (
	"go-marathon-team-3/pkg/tfsmetrics/repointerface"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter interface {
	// Принимает КОПИЮ итератора и создает по нему метрики для проекта
	GetProjectMetrics(iterator repointerface.CommitIterator, project string)
}

type metrics struct {
	commits     prometheus.Counter
	addedRows   prometheus.Counter
	deletedRows prometheus.Counter
}

func newMetrics(author string, project string) *metrics {
	m := &metrics{
		commits: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "commits",
			Help:        "commits counter",
			ConstLabels: map[string]string{"author": author, "project": project},
		}),
		addedRows: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "added_rows",
			Help:        "added_rows counter",
			ConstLabels: map[string]string{"author": author, "project": project},
		}),
		deletedRows: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "deleted_rows",
			Help:        "deleted_rows counter",
			ConstLabels: map[string]string{"author": author, "project": project},
		}),
	}
	prometheus.MustRegister(m.commits, m.addedRows, m.deletedRows)
	return m
}

type exporter struct {
	authors map[string]*metrics
}

func NewExporter() Exporter {
	return &exporter{
		authors: make(map[string]*metrics),
	}
}

func (e *exporter) GetProjectMetrics(iterator repointerface.CommitIterator, project string) {
	for commit, err := iterator.Next(); err == nil; commit, err = iterator.Next() {
		if m, ok := e.authors[commit.Author]; ok {
			m.commits.Inc()
			m.addedRows.Add(float64(commit.AddedRows))
			m.deletedRows.Add(float64(commit.DeletedRows))
		} else {
			e.authors[commit.Author] = newMetrics(commit.Author, project)
			e.authors[commit.Author].commits.Inc()
			e.authors[commit.Author].addedRows.Add(float64(commit.AddedRows))
			e.authors[commit.Author].deletedRows.Add(float64(commit.DeletedRows))
		}
	}
}
