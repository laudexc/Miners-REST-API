package myHttp

import (
	"net/http"
	"prj2/internal"
	"prj2/logic"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry        *prometheus.Registry
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewMetrics(enterprise *logic.Enterprise) *Metrics {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		registry: registry,
		requestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "miners_http_requests_total",
			Help: "Total number of HTTP requests handled by the Miners REST API.",
		}, []string{"method", "path", "status"}),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "miners_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path", "status"}),
	}

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		m.requestsTotal,
		m.requestDuration,
		newEnterpriseCollector(enterprise),
	)

	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *Metrics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		rw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		status := strconv.Itoa(rw.statusCode)
		path := routePath(r)

		m.requestsTotal.WithLabelValues(r.Method, path, status).Inc()
		m.requestDuration.WithLabelValues(r.Method, path, status).Observe(time.Since(startedAt).Seconds())
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func routePath(r *http.Request) string {
	route := mux.CurrentRoute(r)
	if route == nil {
		return "unknown"
	}

	template, err := route.GetPathTemplate()
	if err != nil || template == "" {
		return "unknown"
	}

	return template
}

type enterpriseCollector struct {
	enterprise *logic.Enterprise

	balance      *prometheus.Desc
	activeMiners *prometheus.Desc
	hiredMiners  *prometheus.Desc
	equipment    *prometheus.Desc
}

func newEnterpriseCollector(enterprise *logic.Enterprise) prometheus.Collector {
	return &enterpriseCollector{
		enterprise: enterprise,
		balance: prometheus.NewDesc(
			"miners_enterprise_balance",
			"Current enterprise coal balance.",
			nil,
			nil,
		),
		activeMiners: prometheus.NewDesc(
			"miners_active_total",
			"Current number of active miners.",
			nil,
			nil,
		),
		hiredMiners: prometheus.NewDesc(
			"miners_hired_total",
			"Total number of hired miners by class.",
			[]string{"class"},
			nil,
		),
		equipment: prometheus.NewDesc(
			"miners_equipment_owned",
			"Whether equipment is owned by type. 1 means owned, 0 means not owned.",
			[]string{"type"},
			nil,
		),
	}
}

func (c *enterpriseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.balance
	ch <- c.activeMiners
	ch <- c.hiredMiners
	ch <- c.equipment
}

func (c *enterpriseCollector) Collect(ch chan<- prometheus.Metric) {
	snapshot := c.enterprise.Summary()

	ch <- prometheus.MustNewConstMetric(c.balance, prometheus.GaugeValue, float64(snapshot.Balance))
	ch <- prometheus.MustNewConstMetric(c.activeMiners, prometheus.GaugeValue, float64(snapshot.ActiveCount))

	for _, class := range []internal.MinerClass{internal.WeakClass, internal.NormalClass, internal.StrongClass} {
		ch <- prometheus.MustNewConstMetric(
			c.hiredMiners,
			prometheus.GaugeValue,
			float64(snapshot.HiredStats[class]),
			string(class),
		)
	}

	for _, equipmentType := range internal.EquipmentTypes() {
		value := 0.0
		if snapshot.Equipment[equipmentType] {
			value = 1
		}

		ch <- prometheus.MustNewConstMetric(
			c.equipment,
			prometheus.GaugeValue,
			value,
			string(equipmentType),
		)
	}
}
