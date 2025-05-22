package util

import "github.com/prometheus/client_golang/prometheus"

const Namespace = "tungsten_dns"

func MakeSubsystemOptsFactory(Subsystem string) func(Name string, Help string) prometheus.GaugeOpts {
	return func(Name string, Help string) prometheus.GaugeOpts {
		return prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      Name,
			Help:      Help,
		}
	}
}

type DNSMetrics struct {
	MetricsEnabled bool

	totalQueriesCounter           *prometheus.CounterVec
	queriesByRecordTypeCounter    *prometheus.CounterVec
	queriesByResponderTypeCounter *prometheus.CounterVec
}

func NewDNSMetrics() *DNSMetrics {
	return new(DNSMetrics)
}

const PrometheusNamespace = "tungsten_dns"

func (dm *DNSMetrics) SetupAndRegisterCollectors(registry *prometheus.Registry) {
	dm.totalQueriesCounter = prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: PrometheusNamespace, Name: "total_queries"}, []string{"zone"})
	dm.queriesByRecordTypeCounter = prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: PrometheusNamespace, Name: "total_queries"}, []string{"zone", "type"})
	dm.queriesByResponderTypeCounter = prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: PrometheusNamespace, Name: "total_queries", Help: "Total number of queries"}, []string{"zone", "responder"})

	registry.MustRegister(dm.totalQueriesCounter, dm.queriesByRecordTypeCounter, dm.queriesByResponderTypeCounter)
	dm.MetricsEnabled = true
}

func (dm *DNSMetrics) CountQuery(zone string, qType string, responder string) {
	if !dm.MetricsEnabled {
		return
	}
	dm.totalQueriesCounter.WithLabelValues(zone).Inc()
	dm.queriesByRecordTypeCounter.WithLabelValues(zone, qType).Inc()
	dm.queriesByResponderTypeCounter.WithLabelValues(zone, responder).Inc()
}
