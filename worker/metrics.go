package worker

import (
	"github.com/prometheus/client_golang/prometheus"
)

func RemoveCounter(c *Counter) {
	prometheus.Unregister(c.c)
}

func CreateCounter(opts CounterOpts, labels []string) *Counter {
	m := prometheus.NewCounterVec(
		opts, labels,
	)
	prometheus.MustRegister(m)
	c := Counter{c: m}
	return &c
}

type (
	MetricLabels = prometheus.Labels
	CounterOpts  = prometheus.CounterOpts
	GaugeOpts    = prometheus.GaugeOpts
)

type Counter struct {
	c *prometheus.CounterVec
}

func (c *Counter) Inc(ml MetricLabels) {
	c.c.With(ml).Inc()
}

func RemoveGauge(g *Gauge) {
	prometheus.Unregister(g.g)
}

func CreateGauge(opts GaugeOpts, labels []string) *Gauge {
	m := prometheus.NewGaugeVec(
		opts, labels,
	)
	prometheus.MustRegister(m)
	g := Gauge{g: m}
	return &g
}

type Gauge struct {
	g *prometheus.GaugeVec
}

func (g *Gauge) Inc(ml MetricLabels) {
	g.g.With(ml).Inc()
}

func (g *Gauge) Dec(ml MetricLabels) {
	g.g.With(ml).Dec()
}

func (g *Gauge) Set(val float64, ml MetricLabels) {
	g.g.With(ml).Set(val)
}

func (g *Gauge) Add(val float64, ml MetricLabels) {
	g.g.With(ml).Add(val)
}

func (g *Gauge) Sub(val float64, ml MetricLabels) {
	g.g.With(ml).Sub(val)
}
