package utils

import (
	"runtime"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	Registry = prometheus.NewRegistry()

	RSS = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "process_resident_set_size_bytes",
		Help: "Resident Set Size — total physical RAM used by the process",
	})
	EventloopLag = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "eventloop_lag_seconds",
		Help: "Event loop lag — delay of the event loop for synchronous blocking detection",
	})
	CPUTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "process_cpu_seconds_total",
		Help: "Total user + system CPU time spent (seconds)",
	})
	HeapBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "process_heap_bytes",
		Help: "Heap memory usage (used / total)",
	}, []string{"state"})
	Uptime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "process_start_time_seconds",
		Help: "Start time of the process since unix epoch (seconds)",
	})
	Inflight = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Number of HTTP requests currently being processed (active)",
	}, []string{"method"})

	RequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "route", "status", "ok"})

	RequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10},
	}, []string{"method", "route", "status", "ok"})
)

func init() {
	// Register standard Go runtime and Garbage Collection metrics
	Registry.MustRegister(prometheus.NewGoCollector())

	Registry.MustRegister(RSS)
	Registry.MustRegister(EventloopLag)
	Registry.MustRegister(CPUTotal)
	Registry.MustRegister(HeapBytes)
	Registry.MustRegister(Uptime)
	Registry.MustRegister(Inflight)
	Registry.MustRegister(RequestsTotal)
	Registry.MustRegister(RequestDuration)
}

func InflightMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := c.Method()
		Inflight.WithLabelValues(method).Inc()
		defer Inflight.WithLabelValues(method).Dec()

		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Seconds()

		route := c.Path()
		if r := c.Route(); r != nil {
			route = r.Path
		}

		status := c.Response().StatusCode()
		statusStr := strconv.Itoa(status)
		okStr := "false"
		if status >= 200 && status < 300 {
			okStr = "true"
		}

		RequestsTotal.WithLabelValues(method, route, statusStr, okStr).Inc()
		RequestDuration.WithLabelValues(method, route, statusStr, okStr).Observe(duration)

		return err
	}
}

var startTime = time.Now()

func StartSystemMetrics() {
	Uptime.Set(float64(startTime.Unix()))

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		var prevCPUSec float64
		if initialCPU, err := GetCPUSec(); err == nil {
			prevCPUSec = initialCPU
		}

		for range ticker.C {
			if rssVal, err := GetRSSBytes(); err == nil {
				RSS.Set(float64(rssVal))
			}

			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			HeapBytes.WithLabelValues("used").Set(float64(m.HeapAlloc))
			HeapBytes.WithLabelValues("total").Set(float64(m.HeapSys))

			lag := measureLag()
			EventloopLag.Set(lag)

			if currCPU, err := GetCPUSec(); err == nil {
				delta := currCPU - prevCPUSec
				if delta > 0 {
					CPUTotal.Add(delta)
				}
				prevCPUSec = currCPU
			}
		}
	}()
}

func measureLag() float64 {
	start := time.Now()
	time.Sleep(1 * time.Millisecond)
	lag := time.Since(start) - 1*time.Millisecond
	if lag < 0 {
		return 0
	}
	return lag.Seconds()
}
