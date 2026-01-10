package storage

import (
	"testing"
	"time"
)

func TestSetGetDeleteExists(t *testing.T) {
	s := NewStorage()
	s.Set("k1", "v1", 0)
	if v, ok := s.Get("k1"); !ok || v != "v1" {
		t.Fatalf("expected k1=v1, got %v,%v", v, ok)
	}
	if !s.Exists("k1") {
		t.Fatalf("expected k1 to exist")
	}
	if !s.Delete("k1") {
		t.Fatalf("expected delete k1 true")
	}
	if s.Exists("k1") {
		t.Fatalf("expected k1 not exist after delete")
	}
}

func TestExpireAndTTL(t *testing.T) {
	s := NewStorage()
	s.Set("k2", "v2", 0)
	if !s.Expire("k2", 1) {
		t.Fatalf("expected expire to succeed")
	}
	if ttl := s.TTL("k2"); ttl <= 0 {
		t.Fatalf("expected ttl > 0, got %d", ttl)
	}
	// wait for expiration (poll until key is removed)
	maxWait := time.Now().Add(3 * time.Second)
	for time.Now().Before(maxWait) {
		if _, ok := s.Get("k2"); !ok {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if _, ok := s.Get("k2"); ok {
		t.Fatalf("expected k2 expired")
	}
	// 等待后台删除完成（lazy delete 是异步的）
	waitDeadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(waitDeadline) {
		if s.TTL("k2") == -2 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if ttl := s.TTL("k2"); ttl != -2 {
		t.Fatalf("expected TTL -2 for missing key, got %d", ttl)
	}
}

func TestCountAndJanitor(t *testing.T) {
	s := NewStorage()
	s.Set("a", "1", 0)
	s.Set("b", "2", 1)
	if n := s.Count(); n != 2 {
		t.Fatalf("expected count 2, got %d", n)
	}
	// start janitor and wait for b to expire
	s.StartJanitor(200 * time.Millisecond)
	time.Sleep(1500 * time.Millisecond)
	if n := s.Count(); n != 1 {
		t.Fatalf("expected count 1 after janitor, got %d", n)
	}
}
