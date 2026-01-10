package protocol

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseInline(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("PING\r\n"))
	cmd, args, err := ParseRequest(r)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if cmd != "PING" {
		t.Fatalf("expected PING, got %s", cmd)
	}
	if len(args) != 0 {
		t.Fatalf("expected 0 args, got %d", len(args))
	}
}

func TestParseArray(t *testing.T) {
	in := "*2\r\n$4\r\nPING\r\n$4\r\nTEST\r\n"
	r := bufio.NewReader(strings.NewReader(in))
	cmd, args, err := ParseRequest(r)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if cmd != "PING" {
		t.Fatalf("expected PING, got %s", cmd)
	}
	if len(args) != 1 || args[0] != "TEST" {
		t.Fatalf("expected args [TEST], got %v", args)
	}
}
