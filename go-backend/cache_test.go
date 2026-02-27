package main

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := NewCache(5 * time.Minute)
	c.Set("key1", "value1")

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestCache_Miss(t *testing.T) {
	c := NewCache(5 * time.Minute)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestCache_Expiration(t *testing.T) {
	c := NewCache(50 * time.Millisecond)
	c.Set("key1", "value1")

	// Should be present immediately
	_, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit before expiry")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	_, ok = c.Get("key1")
	if ok {
		t.Fatal("expected cache miss after expiry")
	}
}

func TestCache_Invalidate(t *testing.T) {
	c := NewCache(5 * time.Minute)
	c.Set("key1", "val1")
	c.Set("key2", "val2")

	c.Invalidate()

	_, ok1 := c.Get("key1")
	_, ok2 := c.Get("key2")
	if ok1 || ok2 {
		t.Fatal("expected all entries cleared after invalidation")
	}
}

func TestCache_Stats(t *testing.T) {
	c := NewCache(5 * time.Minute)
	c.Set("key1", "val1")

	// 1 hit
	c.Get("key1")
	// 2 misses
	c.Get("miss1")
	c.Get("miss2")

	stats := c.Stats()
	if stats.Entries != 1 {
		t.Errorf("expected 1 entry, got %d", stats.Entries)
	}
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 2 {
		t.Errorf("expected 2 misses, got %d", stats.Misses)
	}
	expectedRate := 1.0 / 3.0
	if stats.HitRate < expectedRate-0.01 || stats.HitRate > expectedRate+0.01 {
		t.Errorf("expected hit rate ~%.2f, got %.2f", expectedRate, stats.HitRate)
	}
}

func TestCache_Overwrite(t *testing.T) {
	c := NewCache(5 * time.Minute)
	c.Set("key1", "original")
	c.Set("key1", "updated")

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if val != "updated" {
		t.Errorf("expected updated, got %v", val)
	}
}
