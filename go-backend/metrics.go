package main

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks request counts, durations, and error counts for observability.
type Metrics struct {
	mu              sync.RWMutex
	totalRequests   atomic.Int64
	totalErrors     atomic.Int64
	statusCounts    map[int]*atomic.Int64
	methodCounts    map[string]*atomic.Int64
	pathCounts      map[string]*atomic.Int64
	totalDurationNs atomic.Int64
	startTime       time.Time
}

// MetricsResponse represents the JSON response for the metrics endpoint.
type MetricsResponse struct {
	Uptime        string           `json:"uptime"`
	TotalRequests int64            `json:"totalRequests"`
	TotalErrors   int64            `json:"totalErrors"`
	ErrorRate     float64          `json:"errorRate"`
	AvgDuration   string           `json:"avgDuration"`
	ByStatus      map[string]int64 `json:"byStatus"`
	ByMethod      map[string]int64 `json:"byMethod"`
	ByPath        map[string]int64 `json:"byPath"`
}

// NewMetrics creates a new Metrics tracker.
func NewMetrics() *Metrics {
	return &Metrics{
		statusCounts: make(map[int]*atomic.Int64),
		methodCounts: make(map[string]*atomic.Int64),
		pathCounts:   make(map[string]*atomic.Int64),
		startTime:    time.Now(),
	}
}

// Record records a single request's method, path, status code, and duration.
func (m *Metrics) Record(method, path string, status int, duration time.Duration) {
	m.totalRequests.Add(1)
	m.totalDurationNs.Add(int64(duration))

	if status >= 400 {
		m.totalErrors.Add(1)
	}

	// Status counts
	m.mu.Lock()
	sc, ok := m.statusCounts[status]
	if !ok {
		sc = &atomic.Int64{}
		m.statusCounts[status] = sc
	}

	mc, ok := m.methodCounts[method]
	if !ok {
		mc = &atomic.Int64{}
		m.methodCounts[method] = mc
	}

	pc, ok := m.pathCounts[path]
	if !ok {
		pc = &atomic.Int64{}
		m.pathCounts[path] = pc
	}
	m.mu.Unlock()

	sc.Add(1)
	mc.Add(1)
	pc.Add(1)
}

// Snapshot returns a point-in-time snapshot of all metrics.
func (m *Metrics) Snapshot() MetricsResponse {
	total := m.totalRequests.Load()
	errors := m.totalErrors.Load()
	durationNs := m.totalDurationNs.Load()

	var errorRate float64
	if total > 0 {
		errorRate = float64(errors) / float64(total)
	}

	var avgDuration time.Duration
	if total > 0 {
		avgDuration = time.Duration(durationNs / total)
	}

	m.mu.RLock()
	byStatus := make(map[string]int64, len(m.statusCounts))
	for code, counter := range m.statusCounts {
		key := statusCodeToString(code)
		byStatus[key] += counter.Load()
	}

	byMethod := make(map[string]int64, len(m.methodCounts))
	for method, counter := range m.methodCounts {
		byMethod[method] = counter.Load()
	}

	byPath := make(map[string]int64, len(m.pathCounts))
	for path, counter := range m.pathCounts {
		byPath[path] = counter.Load()
	}
	m.mu.RUnlock()

	return MetricsResponse{
		Uptime:        time.Since(m.startTime).Round(time.Second).String(),
		TotalRequests: total,
		TotalErrors:   errors,
		ErrorRate:     errorRate,
		AvgDuration:   avgDuration.String(),
		ByStatus:      byStatus,
		ByMethod:      byMethod,
		ByPath:        byPath,
	}
}

func statusCodeToString(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "other"
	}
}
