package protocol

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

var ErrClosed = io.EOF

// ParseRequest 支持两种简单情况：
// 1) 行协议（inline commands），如: "PING\r\n" -> ["PING"]
// 2) RESP 数组（支持批量字符串）：*N\r\n$len\r\n...\r\n
func ParseRequest(r *bufio.Reader) (string, []string, error) {
	// Peek first byte
	b, err := r.Peek(1)
	if err != nil {
		return "", nil, err
	}
	if b[0] == '*' {
		// RESP Array
		line, err := r.ReadString('\n')
		if err != nil {
			return "", nil, err
		}
		line = strings.TrimSpace(line)
		n, err := strconv.Atoi(strings.TrimPrefix(line, "*"))
		if err != nil {
			return "", nil, errors.New("invalid array header")
		}
		args := make([]string, 0, n)
		for i := 0; i < n; i++ {
			// Read $len header
			h, err := r.ReadString('\n')
			if err != nil {
				return "", nil, err
			}
			h = strings.TrimSpace(h)
			if !strings.HasPrefix(h, "$") {
				return "", nil, errors.New("invalid bulk string header")
			}
			llen, err := strconv.Atoi(strings.TrimPrefix(h, "$"))
			if err != nil {
				return "", nil, errors.New("invalid bulk len")
			}
			buf := make([]byte, llen+2)
			if _, err := io.ReadFull(r, buf); err != nil {
				return "", nil, err
			}
			arg := string(buf[:llen])
			args = append(args, arg)
		}
		if len(args) == 0 {
			return "", nil, errors.New("empty command")
		}
		cmd := args[0]
		return cmd, args[1:], nil
	}
	// Inline: read a line
	line, err := r.ReadString('\n')
	if err != nil {
		return "", nil, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return "", nil, errors.New("empty line")
	}
	parts := strings.Fields(line)
	cmd := parts[0]
	return cmd, parts[1:], nil
}
