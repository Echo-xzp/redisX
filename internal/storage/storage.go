package storage

import (
	"sync"
	"time"
)

// Entry 表示存储的值及过期时间（秒级）
type Entry struct {
	Value    string
	ExpireAt int64 // Unix 时间戳，0 表示永不过期
}

type Storage struct {
	mu   sync.RWMutex
	data map[string]*Entry
}

func NewStorage() *Storage {
	return &Storage{data: make(map[string]*Entry)}
}

func (s *Storage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	if !ok {
		return "", false
	}
	// 过期检查
	if v.ExpireAt != 0 && time.Now().Unix() >= v.ExpireAt {
		// 懒删除
		go s.Delete(key)
		return "", false
	}
	return v.Value, true
}

func (s *Storage) Set(key, value string, ttlSeconds int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp := int64(0)
	if ttlSeconds > 0 {
		exp = time.Now().Unix() + ttlSeconds
	}
	s.data[key] = &Entry{Value: value, ExpireAt: exp}
}

func (s *Storage) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; ok {
		delete(s.data, key)
		return true
	}
	return false
}

func (s *Storage) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}

// Expire 为现有键设置过期时间，返回是否设置成功（键存在）
func (s *Storage) Expire(key string, ttlSeconds int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.data[key]; ok {
		if ttlSeconds > 0 {
			e.ExpireAt = time.Now().Unix() + ttlSeconds
		} else {
			e.ExpireAt = 0
		}
		return true
	}
	return false
}

// TTL 返回剩余秒数：-2 表示键不存在，-1 表示键存在但无过期
func (s *Storage) TTL(key string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.data[key]; ok {
		if e.ExpireAt == 0 {
			return -1
		}
		rem := e.ExpireAt - time.Now().Unix()
		if rem < 0 {
			return -2
		}
		return rem
	}
	return -2
}

// Count 返回当前键数量
func (s *Storage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// StartJanitor 启动后台清理过期键的 goroutine
func (s *Storage) StartJanitor(interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			now := time.Now().Unix()
			// 扫描并删除过期键
			s.mu.Lock()
			for k, v := range s.data {
				if v.ExpireAt != 0 && now >= v.ExpireAt {
					delete(s.data, k)
				}
			}
			s.mu.Unlock()
		}
	}()
}
