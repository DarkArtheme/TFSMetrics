package exporter

import (
	"go-marathon-team-3/pkg/tfsmetrics/repointerface"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter interface {
	// Принимает КОПИЮ итератора и создает по нему метрики по проекту
	GetProjectMetrics(iterator repointerface.CommitIterator)
}

type metrics struct {
	commits     prometheus.Counter
	addedRows   prometheus.Counter
	deletedRows prometheus.Counter
}

func newMetrics(auther string) *metrics {
	m := &metrics{
		commits: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "commits",
			Help:        "commits counter",
			ConstLabels: map[string]string{"auther": auther},
		}),
		addedRows: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "added_rows",
			Help:        "added_rows counter",
			ConstLabels: map[string]string{"auther": auther},
		}),
		deletedRows: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "deleted_rows",
			Help:        "deleted_rows counter",
			ConstLabels: map[string]string{"auther": auther},
		}),
	}
	prometheus.MustRegister(m.commits, m.addedRows, m.deletedRows)
	return m
}

type exporter struct {
	authers map[string]*metrics
}

func NewExporter() Exporter {
	return &exporter{
		authers: make(map[string]*metrics),
	}
}

func (e *exporter) GetProjectMetrics(iterator repointerface.CommitIterator) {
	for commit, err := iterator.Next(); err == nil; commit, err = iterator.Next() {
		if m, ok := e.authers[commit.Author]; ok {
			m.commits.Inc()
			m.addedRows.Add(float64(commit.AddedRows))
			m.deletedRows.Add(float64(commit.DeletedRows))
		} else {
			e.authers[commit.Author] = newMetrics(commit.Author)
			e.authers[commit.Author].commits.Inc()
			e.authers[commit.Author].addedRows.Add(float64(commit.AddedRows))
			e.authers[commit.Author].deletedRows.Add(float64(commit.DeletedRows))
		}
	}
}
