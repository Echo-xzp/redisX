package command

import (
	"redisx/internal/storage"
	"testing"
)

func TestIncr(t *testing.T) {
	s := storage.NewStorage()
	s.Set("k", "1", 0)
	resp, _ := Incr(s, []string{"k"})
	if string(resp) != ":2\r\n" {
		t.Fatalf("expected :2, got %q", string(resp))
	}
}

func TestMGet(t *testing.T) {
	s := storage.NewStorage()
	s.Set("a", "1", 0)
	s.Set("b", "2", 0)
	resp, _ := MGet(s, []string{"a", "x", "b"})
	if len(resp) == 0 {
		t.Fatalf("expected response, got empty")
	}
}

func TestPersist(t *testing.T) {
	s := storage.NewStorage()
	s.Set("p", "v", 1)
	resp, _ := Persist(s, []string{"p"})
	if string(resp) != ":1\r\n" {
		t.Fatalf("expected :1, got %q", string(resp))
	}
}
