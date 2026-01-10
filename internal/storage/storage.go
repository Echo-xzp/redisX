package storage

import (
	"strconv"
	"sync"
	"time"
)

// Entry 表示存储的值及过期时间（Unix 毫秒）
type Entry struct {
	Value    string
	ExpireAt int64 // Unix 毫秒时间戳，0 表示永不过期
}

type Storage struct {
	mu   sync.RWMutex
	data map[string]*Entry
}

func NewStorage() *Storage {
	return &Storage{data: make(map[string]*Entry)}
}

// Get retrieves the value for the given key. If the key is expired it will be
// removed and the function returns ("", false).
func (s *Storage) Get(key string) (string, bool) {
	s.mu.RLock()
	v, ok := s.data[key]
	if !ok {
		s.mu.RUnlock()
		return "", false
	}
	// 过期检查（使用毫秒精度）
	if v.ExpireAt != 0 && time.Now().UnixMilli() >= v.ExpireAt {
		// 升级为写锁以删除已过期的键
		s.mu.RUnlock()
		s.mu.Lock()
		if vv, ok2 := s.data[key]; ok2 {
			if vv.ExpireAt != 0 && time.Now().UnixMilli() >= vv.ExpireAt {
				delete(s.data, key)
				s.mu.Unlock()
				return "", false
			}
			val := vv.Value
			s.mu.Unlock()
			return val, true
		}
		s.mu.Unlock()
		return "", false
	}
	val := v.Value
	s.mu.RUnlock()
	return val, true
}

// Set sets the value and ttl in seconds (0 means no expiration).
func (s *Storage) Set(key, value string, ttlSeconds int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp := int64(0)
	if ttlSeconds > 0 {
		exp = time.Now().UnixMilli() + ttlSeconds*1000
	}
	s.data[key] = &Entry{Value: value, ExpireAt: exp}
}

// SetWithMs sets the value and ttl in milliseconds (0 means no expiration).
func (s *Storage) SetWithMs(key, value string, ttlMillis int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp := int64(0)
	if ttlMillis > 0 {
		exp = time.Now().UnixMilli() + ttlMillis
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

// IncrBy atomically increments the integer value of a key by delta. If the key
// does not exist it is set to delta. Returns the new value or an error if the
// current value is not an integer.
func (s *Storage) IncrBy(key string, delta int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var cur int64
	if e, ok := s.data[key]; ok {
		if e.ExpireAt != 0 && time.Now().UnixMilli() >= e.ExpireAt {
			// expired
			delete(s.data, key)
			cur = 0
		} else {
			val := e.Value
			if val == "" {
				cur = 0
			} else {
				v, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return 0, err
				}
				cur = v
			}
		}
	}
	cur += delta
	if e, ok := s.data[key]; ok {
		e.Value = strconv.FormatInt(cur, 10)
	} else {
		s.data[key] = &Entry{Value: strconv.FormatInt(cur, 10), ExpireAt: 0}
	}
	return cur, nil
}

// Persist removes the expiration from a key. Returns true if the timeout was removed.
func (s *Storage) Persist(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.data[key]; ok {
		if e.ExpireAt != 0 {
			e.ExpireAt = 0
			return true
		}
		return false
	}
	return false
}

func (s *Storage) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}

// Expire sets TTL in seconds for an existing key.
func (s *Storage) Expire(key string, ttlSeconds int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.data[key]; ok {
		if ttlSeconds > 0 {
			e.ExpireAt = time.Now().UnixMilli() + ttlSeconds*1000
		} else {
			e.ExpireAt = 0
		}
		return true
	}
	return false
}

// PExpire sets TTL in milliseconds for an existing key.
func (s *Storage) PExpire(key string, ttlMillis int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.data[key]; ok {
		if ttlMillis > 0 {
			e.ExpireAt = time.Now().UnixMilli() + ttlMillis
		} else {
			e.ExpireAt = 0
		}
		return true
	}
	return false
}

// TTL returns remaining seconds: -2 key not exist, -1 key exists but no expiry.
func (s *Storage) TTL(key string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.data[key]; ok {
		if e.ExpireAt == 0 {
			return -1
		}
		rem := e.ExpireAt - time.Now().UnixMilli()
		if rem < 0 {
			return -2
		}
		// 返回向上取整的秒数
		return (rem + 999) / 1000
	}
	return -2
}

// PTTL 返回剩余毫秒数：-2 表示键不存在，-1 表示键存在但无过期
func (s *Storage) PTTL(key string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.data[key]; ok {
		if e.ExpireAt == 0 {
			return -1
		}
		rem := e.ExpireAt - time.Now().UnixMilli()
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

// StartJanitor 启动后台清理过期键的 goroutine（使用毫秒比较）
func (s *Storage) StartJanitor(interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			now := time.Now().UnixMilli()
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
