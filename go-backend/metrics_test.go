package main

import (
	"testing"
	"time"
)

func TestMetrics_Record(t *testing.T) {
	m := NewMetrics()
	m.Record("GET", "/api/users", 200, 10*time.Millisecond)
	m.Record("POST", "/api/users", 201, 20*time.Millisecond)
	m.Record("GET", "/api/tasks", 500, 5*time.Millisecond)

	snap := m.Snapshot()

	if snap.TotalRequests != 3 {
		t.Errorf("expected 3 total requests, got %d", snap.TotalRequests)
	}
	if snap.TotalErrors != 1 {
		t.Errorf("expected 1 error (500), got %d", snap.TotalErrors)
	}
	expectedRate := 1.0 / 3.0
	if snap.ErrorRate < expectedRate-0.01 || snap.ErrorRate > expectedRate+0.01 {
		t.Errorf("expected error rate ~%.2f, got %.2f", expectedRate, snap.ErrorRate)
	}
}

func TestMetrics_ByStatus(t *testing.T) {
	m := NewMetrics()
	m.Record("GET", "/a", 200, time.Millisecond)
	m.Record("GET", "/b", 200, time.Millisecond)
	m.Record("POST", "/a", 201, time.Millisecond)
	m.Record("GET", "/a", 404, time.Millisecond)
	m.Record("GET", "/a", 500, time.Millisecond)

	snap := m.Snapshot()

	if snap.ByStatus["2xx"] != 3 {
		t.Errorf("expected 3 2xx, got %d", snap.ByStatus["2xx"])
	}
	if snap.ByStatus["4xx"] != 1 {
		t.Errorf("expected 1 4xx, got %d", snap.ByStatus["4xx"])
	}
	if snap.ByStatus["5xx"] != 1 {
		t.Errorf("expected 1 5xx, got %d", snap.ByStatus["5xx"])
	}
}

func TestMetrics_ByMethod(t *testing.T) {
	m := NewMetrics()
	m.Record("GET", "/a", 200, time.Millisecond)
	m.Record("GET", "/b", 200, time.Millisecond)
	m.Record("POST", "/a", 201, time.Millisecond)

	snap := m.Snapshot()

	if snap.ByMethod["GET"] != 2 {
		t.Errorf("expected 2 GET, got %d", snap.ByMethod["GET"])
	}
	if snap.ByMethod["POST"] != 1 {
		t.Errorf("expected 1 POST, got %d", snap.ByMethod["POST"])
	}
}

func TestMetrics_ByPath(t *testing.T) {
	m := NewMetrics()
	m.Record("GET", "/api/users", 200, time.Millisecond)
	m.Record("GET", "/api/users", 200, time.Millisecond)
	m.Record("GET", "/api/tasks", 200, time.Millisecond)

	snap := m.Snapshot()

	if snap.ByPath["/api/users"] != 2 {
		t.Errorf("expected 2 for /api/users, got %d", snap.ByPath["/api/users"])
	}
	if snap.ByPath["/api/tasks"] != 1 {
		t.Errorf("expected 1 for /api/tasks, got %d", snap.ByPath["/api/tasks"])
	}
}

func TestMetrics_AvgDuration(t *testing.T) {
	m := NewMetrics()
	m.Record("GET", "/a", 200, 10*time.Millisecond)
	m.Record("GET", "/b", 200, 30*time.Millisecond)

	snap := m.Snapshot()

	// Average should be ~20ms
	if snap.AvgDuration == "" {
		t.Error("expected non-empty AvgDuration")
	}
}

func TestMetrics_Uptime(t *testing.T) {
	m := NewMetrics()
	snap := m.Snapshot()
	if snap.Uptime == "" {
		t.Error("expected non-empty uptime")
	}
}

func TestMetrics_ZeroRequests(t *testing.T) {
	m := NewMetrics()
	snap := m.Snapshot()
	if snap.TotalRequests != 0 {
		t.Errorf("expected 0 total, got %d", snap.TotalRequests)
	}
	if snap.ErrorRate != 0 {
		t.Errorf("expected 0 error rate, got %f", snap.ErrorRate)
	}
}
