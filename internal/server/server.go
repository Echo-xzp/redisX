package server

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"redisx/internal/protocol"
	"redisx/internal/storage"
)

// Server 集成了存储与协议处理

type Server struct {
	addr  string
	ln    net.Listener
	store *storage.Storage
	// connection/resource limits
	MaxConns       int
	connLimiter    chan struct{}
	ConnTimeout    time.Duration
	MaxMemoryBytes int64

	connCount uint64
	startTime time.Time
}

func NewServer(addr string) *Server {
	s := &Server{addr: addr, store: storage.NewStorage(), startTime: time.Now()}
	// 启动后台清理过期键
	s.store.StartJanitor(time.Second * 1)
	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln
	// 初始化资源限制
	if s.MaxConns > 0 {
		s.connLimiter = make(chan struct{}, s.MaxConns)
	}
	if s.MaxMemoryBytes > 0 {
		s.store.SetMaxMemory(s.MaxMemoryBytes)
	}
	log.Printf("redisx server listening on %s", s.addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			// 当 listener 被关闭时，优雅退出
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Println("accept error:", err)
			continue
		}
		// 连接数限制：尝试非阻塞获取信号量
		if s.MaxConns > 0 {
			select {
			case s.connLimiter <- struct{}{}:
				// acquired
			default:
				// 拒绝新的连接
				conn.Write([]byte("-ERR max connections\r\n"))
				conn.Close()
				continue
			}
		}
		atomic.AddUint64(&s.connCount, 1)
		go func(c net.Conn) {
			defer func() {
				atomic.AddUint64(&s.connCount, ^uint64(0))
				if s.MaxConns > 0 {
					<-s.connLimiter
				}
			}()
			s.handleConn(c)
		}(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		// 设置读写超时（如果配置了）
		if s.ConnTimeout > 0 {
			_ = conn.SetDeadline(time.Now().Add(s.ConnTimeout))
		}
		cmd, args, err := protocol.ParseRequest(reader)
		if err != nil {
			if err == protocol.ErrClosed {
				return
			}
			// 如果是超时错误，给出明确消息
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				conn.Write([]byte("-ERR connection timeout\r\n"))
				return
			}
			// 协议错误，返回错误并关闭连接
			conn.Write([]byte(fmt.Sprintf("-ERR %v\r\n", err)))
			return
		}
		switch strings.ToUpper(cmd) {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "QUIT":
			conn.Write([]byte("+OK\r\n"))
			return
		case "SET":
			if len(args) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'SET' command\r\n"))
				continue
			}
			key := args[0]
			value := args[1]
			// 支持 EX seconds 和 PX milliseconds
			ttl := int64(0)
			i := 2
			bad := false
			for i < len(args) {
				op := strings.ToUpper(args[i])
				switch op {
				case "EX":
					if i+1 >= len(args) {
						conn.Write([]byte("-ERR syntax error\r\n"))
						bad = true
						break
					}
					sec, err := strconv.ParseInt(args[i+1], 10, 64)
					if err != nil {
						conn.Write([]byte("-ERR invalid expire time\r\n"))
						bad = true
						break
					}
					ttl = sec
					i += 2
				case "PX":
					if i+1 >= len(args) {
						conn.Write([]byte("-ERR syntax error\r\n"))
						bad = true
						break
					}
					ms, err := strconv.ParseInt(args[i+1], 10, 64)
					if err != nil {
						conn.Write([]byte("-ERR invalid expire time\r\n"))
						bad = true
						break
					}
					// 转换为秒（向上取整）
					ttl = (ms + 999) / 1000
					i += 2
				default:
					i++
				}
				if bad {
					break
				}
			}
			if bad {
				continue
			}
			// 如果存在 PX 参数则按毫秒设置；否则按秒设置
			pxMillis := int64(0)
			for j := 2; j < len(args); j++ {
				if strings.ToUpper(args[j]) == "PX" && j+1 < len(args) {
					if ms, err := strconv.ParseInt(args[j+1], 10, 64); err == nil {
						pxMillis = ms
					}
				}
			}
			// 如果 storage 配置了 max memory 则使用 TrySet 系列以进行内存检查
			if s.store.GetMaxMemory() > 0 {
				if pxMillis > 0 {
					if !s.store.TrySetWithMs(key, value, pxMillis) {
						conn.Write([]byte("-ERR max memory reached\r\n"))
						continue
					}
				} else {
					if !s.store.TrySet(key, value, ttl) {
						conn.Write([]byte("-ERR max memory reached\r\n"))
						continue
					}
				}
			} else {
				if pxMillis > 0 {
					s.store.SetWithMs(key, value, pxMillis)
				} else {
					s.store.Set(key, value, ttl)
				}
			}
			conn.Write([]byte("+OK\r\n"))
		case "GET":
			if len(args) < 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'GET' command\r\n"))
				continue
			}
			key := args[0]
			if v, ok := s.store.Get(key); ok {
				// Bulk string: $<len>\r\n<bytes>\r\n
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)))
			} else {
				conn.Write([]byte("$-1\r\n"))
			}
		case "DEL":
			count := 0
			for _, k := range args {
				if s.store.Delete(k) {
					count++
				}
			}
			conn.Write([]byte(fmt.Sprintf(":%d\r\n", count)))
		case "EXISTS":
			count := 0
			for _, k := range args {
				if s.store.Exists(k) {
					count++
				}
			}
			conn.Write([]byte(fmt.Sprintf(":%d\r\n", count)))
		case "EXPIRE":
			if len(args) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'EXPIRE' command\r\n"))
				continue
			}
			key := args[0]
			ttl, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				conn.Write([]byte("-ERR invalid expire time\r\n"))
				continue
			}
			if s.store.Expire(key, ttl) {
				conn.Write([]byte(":1\r\n"))
			} else {
				conn.Write([]byte(":0\r\n"))
			}
		case "TTL":
			if len(args) < 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'TTL' command\r\n"))
				continue
			}
			key := args[0]
			ttl := s.store.TTL(key)
			conn.Write([]byte(fmt.Sprintf(":%d\r\n", ttl)))
		case "INFO":
			info := fmt.Sprintf("# Server\r\nredis_version:redisX-0.1.0\r\nconnected_clients:%d\r\nkeys:%d\r\nuptime_in_seconds:%d\r\n", atomic.LoadUint64(&s.connCount), s.store.Count(), int(time.Since(s.startTime).Seconds()))
			conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(info), info)))
		default:
			conn.Write([]byte("-ERR unknown command\r\n"))
		}
	}
}
