package main

import (
	"fmt"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const program string = "spark_nanny"

var (
	commonLabels = []string{"spark_app"}

	buildInfo = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: program,
		Name:      "build_info",
		Help: fmt.Sprintf(
			"A metric with a constant '1' value labeled by version, build_date, commit, and goversion from which %s was built.",
			program,
		),
		ConstLabels: prometheus.Labels{
			"version":    version,
			"build_date": buildDate,
			"commit":     commit,
			"goversion":  runtime.Version(),
		},
	})

	killCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: program,
		Name:      "kill_count",
		Help:      "The number of times a Spark job was killed",
	}, commonLabels)

	pokeDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: program,
		Name:      "poke_duration_seconds",
		Help:      "The duration of the poke operation",
		// we might need more fine grained buckets in the future, but this should be a good start
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2},
	}, commonLabels)
)
