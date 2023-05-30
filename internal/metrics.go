package internal

import (
	"fmt"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
)

type Metrics struct {
	counters   map[string]proxywasm.MetricCounter
	histograms map[string]proxywasm.MetricHistogram
}

const MetricPrefix = "envoy_wasm_auth_plugin"

func NewMetrics() *Metrics {
	return &Metrics{
		counters:   make(map[string]proxywasm.MetricCounter),
		histograms: make(map[string]proxywasm.MetricHistogram),
	}
}

func (m *Metrics) Increment(name string, tags [][2]string) {
	fullName := metricName(name, tags)
	if _, exists := m.counters[fullName]; !exists {
		m.counters[fullName] = proxywasm.DefineCounterMetric(fullName)
	}
	m.counters[fullName].Increment(1)
}

func (m *Metrics) Histogram(name string, tags [][2]string, value uint64) {
	fullName := metricName(name, tags)
	if _, exists := m.histograms[fullName]; !exists {
		m.histograms[fullName] = proxywasm.DefineHistogramMetric(fullName)
	}
	m.histograms[fullName].Record(value)
}

func metricName(name string, tags [][2]string) string {
	fullName := fmt.Sprintf("%s_%s", MetricPrefix, name)

	for _, t := range tags {
		fullName += fmt.Sprintf("_%s=.=%s;.;", t[0], t[1])
	}
	return fullName
}
