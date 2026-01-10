package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func startServer(t *testing.T) *Server {
	s := NewServer(":0")
	go func() {
		if err := s.Start(); err != nil {
			t.Fatalf("start server: %v", err)
		}
	}()
	// 等待 listener 就绪
	for i := 0; i < 50; i++ {
		if s.ln != nil {
			return s
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server failed to start")
	return nil
}

func writeReq(conn net.Conn, parts ...string) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "*%d\r\n", len(parts))
	for _, p := range parts {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(p), p)
	}
	_, err := conn.Write(b.Bytes())
	return err
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line, nil
}

func readBulk(r *bufio.Reader) (string, error) {
	// 读 header
	header, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	header = strings.TrimSpace(header)
	if header == "$-1" {
		return "", nil
	}
	if len(header) == 0 || header[0] != '$' {
		return "", fmt.Errorf("invalid bulk header: %q", header)
	}
	n, err := strconv.Atoi(strings.TrimPrefix(header, "$"))
	if err != nil {
		return "", fmt.Errorf("invalid bulk size: %v", err)
	}
	buf := make([]byte, n+2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func TestSetWithEX(t *testing.T) {
	s := startServer(t)
	defer s.ln.Close()

	conn, err := net.Dial("tcp", s.ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	r := bufio.NewReader(conn)

	if err := writeReq(conn, "SET", "kex", "vex", "EX", "1"); err != nil {
		t.Fatalf("write set: %v", err)
	}
	line, _ := readLine(r)
	if line != "+OK\r\n" {
		t.Fatalf("expected +OK, got %q", line)
	}

	if err := writeReq(conn, "GET", "kex"); err != nil {
		t.Fatalf("write get: %v", err)
	}
	val, _ := readBulk(r)
	if val != "vex" {
		t.Fatalf("expected vex, got %q", val)
	}

	// 等待 1.2 秒后应过期
	time.Sleep(1200 * time.Millisecond)
	if err := writeReq(conn, "GET", "kex"); err != nil {
		t.Fatalf("write get: %v", err)
	}
	line, _ = readLine(r)
	if line != "$-1\r\n" {
		t.Fatalf("expected $-1, got %q", line)
	}
}

func TestSetWithPX(t *testing.T) {
	s := startServer(t)
	defer s.ln.Close()

	conn, err := net.Dial("tcp", s.ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	r := bufio.NewReader(conn)

	if err := writeReq(conn, "SET", "kpx", "vpx", "PX", "500"); err != nil {
		t.Fatalf("write set: %v", err)
	}
	line, _ := readLine(r)
	if line != "+OK\r\n" {
		t.Fatalf("expected +OK, got %q", line)
	}

	// 检查 PTTL 接口返回毫秒级剩余时间
	if err := writeReq(conn, "PTTL", "kpx"); err != nil {
		t.Fatalf("write pttl: %v", err)
	}
	line, _ = readLine(r)
	if len(line) == 0 || line[0] != ':' {
		t.Fatalf("expected integer reply for PTTL, got %q", line)
	}
	pttlStr := strings.TrimSpace(strings.TrimPrefix(line, ":"))
	pttlVal, err := strconv.Atoi(pttlStr)
	if err != nil {
		t.Fatalf("invalid PTTL value: %v", err)
	}
	if pttlVal <= 0 || pttlVal > 500 {
		t.Fatalf("expected pttl between 0 and 500, got %d", pttlVal)
	}

	if err := writeReq(conn, "GET", "kpx"); err != nil {
		t.Fatalf("write get: %v", err)
	}
	val, _ := readBulk(r)
	if val != "vpx" {
		t.Fatalf("expected vpx, got %q", val)
	}

	// 等待 700ms 后应过期
	time.Sleep(700 * time.Millisecond)
	if err := writeReq(conn, "GET", "kpx"); err != nil {
		t.Fatalf("write get: %v", err)
	}
	line, _ = readLine(r)
	if line != "$-1\r\n" {
		t.Fatalf("expected $-1, got %q", line)
	}
}

func TestPExpireAndPTTL_Server(t *testing.T) {
	s := startServer(t)
	defer s.ln.Close()

	conn, err := net.Dial("tcp", s.ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	r := bufio.NewReader(conn)

	if err := writeReq(conn, "SET", "pp", "vp"); err != nil {
		t.Fatalf("write set: %v", err)
	}
	_, _ = readLine(r)

	if err := writeReq(conn, "PEXPIRE", "pp", "500"); err != nil {
		t.Fatalf("write pexpire: %v", err)
	}
	line, _ := readLine(r)
	if line != ":1\r\n" {
		t.Fatalf("expected :1, got %q", line)
	}

	if err := writeReq(conn, "PTTL", "pp"); err != nil {
		t.Fatalf("write pttl: %v", err)
	}
	line, _ = readLine(r)
	pttlStr := strings.TrimSpace(strings.TrimPrefix(line, ":"))
	pttlVal, err := strconv.Atoi(pttlStr)
	if err != nil {
		t.Fatalf("invalid PTTL value: %v", err)
	}
	if pttlVal <= 0 || pttlVal > 500 {
		t.Fatalf("expected pttl between 0 and 500, got %d", pttlVal)
	}

	time.Sleep(700 * time.Millisecond)
	if err := writeReq(conn, "GET", "pp"); err != nil {
		t.Fatalf("write get: %v", err)
	}
	line, _ = readLine(r)
	if line != "$-1\r\n" {
		t.Fatalf("expected $-1, got %q", line)
	}

	// PTTL/TTL 应返回 -2
	if err := writeReq(conn, "PTTL", "pp"); err != nil {
		t.Fatalf("write pttl: %v", err)
	}
	line, _ = readLine(r)
	if line != ":-2\r\n" {
		t.Fatalf("expected :-2 for missing key, got %q", line)
	}
	if err := writeReq(conn, "TTL", "pp"); err != nil {
		t.Fatalf("write ttl: %v", err)
	}
	line, _ = readLine(r)
	if line != ":-2\r\n" {
		t.Fatalf("expected :-2 for missing key TTL, got %q", line)
	}
}

func TestInfoContainsKeys(t *testing.T) {
	s := startServer(t)
	defer s.ln.Close()

	conn, err := net.Dial("tcp", s.ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	r := bufio.NewReader(conn)

	if err := writeReq(conn, "SET", "i1", "1", "EX", "1"); err != nil {
		t.Fatalf("write set: %v", err)
	}
	_, _ = readLine(r)

	if err := writeReq(conn, "INFO"); err != nil {
		t.Fatalf("write info: %v", err)
	}
	info, err := readBulk(r)
	if err != nil {
		t.Fatalf("read bulk: %v", err)
	}
	if !bytes.Contains([]byte(info), []byte("keys:")) {
		t.Fatalf("INFO output missing keys: %q", info)
	}
}
