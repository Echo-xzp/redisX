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
	mu         sync.RWMutex
	data       map[string]*Entry
	maxMemory  int64 // bytes, 0 means no limit
	totalBytes int64 // current total bytes used by values
}

func NewStorage() *Storage {
	return &Storage{data: make(map[string]*Entry)}
}

// SetMaxMemory sets a soft memory limit in bytes (0 disables limit).
func (s *Storage) SetMaxMemory(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxMemory = bytes
}

// MemoryUsage returns the current approximate memory usage in bytes.
func (s *Storage) MemoryUsage() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalBytes
}

// GetMaxMemory returns the configured max memory (0 means disabled).
func (s *Storage) GetMaxMemory() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.maxMemory
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
	oldLen := 0
	if e, ok := s.data[key]; ok {
		// 如果旧值未过期，计算旧长度
		if e.ExpireAt == 0 || time.Now().UnixMilli() < e.ExpireAt {
			oldLen = len(e.Value)
		}
	}
	delta := int64(len(value) - oldLen)
	// 更新总字节数
	s.totalBytes += delta
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
	oldLen := 0
	if e, ok := s.data[key]; ok {
		if e.ExpireAt == 0 || time.Now().UnixMilli() < e.ExpireAt {
			oldLen = len(e.Value)
		}
	}
	delta := int64(len(value) - oldLen)
	s.totalBytes += delta
	s.data[key] = &Entry{Value: value, ExpireAt: exp}
}

func (s *Storage) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.data[key]; ok {
		// 调整内存计数
		s.totalBytes -= int64(len(e.Value))
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
	oldLen := 0
	if e, ok := s.data[key]; ok {
		if e.ExpireAt != 0 && time.Now().UnixMilli() >= e.ExpireAt {
			// expired
			// 调整内存计数
			s.totalBytes -= int64(len(e.Value))
			delete(s.data, key)
			cur = 0
		} else {
			val := e.Value
			oldLen = len(val)
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
	newVal := strconv.FormatInt(cur, 10)
	// 更新总字节数
	s.totalBytes += int64(len(newVal) - oldLen)
	if e, ok := s.data[key]; ok {
		e.Value = newVal
	} else {
		s.data[key] = &Entry{Value: newVal, ExpireAt: 0}
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

// TrySet tries to set a key given seconds TTL, honoring maxMemory if set.
// Returns true if set succeeded, false if rejected due to memory limit.
func (s *Storage) TrySet(key, value string, ttlSeconds int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp := int64(0)
	if ttlSeconds > 0 {
		exp = time.Now().UnixMilli() + ttlSeconds*1000
	}
	oldLen := 0
	if e, ok := s.data[key]; ok {
		if e.ExpireAt == 0 || time.Now().UnixMilli() < e.ExpireAt {
			oldLen = len(e.Value)
		}
	}
	delta := int64(len(value) - oldLen)
	if s.maxMemory > 0 && s.totalBytes+delta > s.maxMemory {
		return false
	}
	s.totalBytes += delta
	s.data[key] = &Entry{Value: value, ExpireAt: exp}
	return true
}

// TrySetWithMs tries to set a key given ms TTL, honoring maxMemory if set.
func (s *Storage) TrySetWithMs(key, value string, ttlMillis int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp := int64(0)
	if ttlMillis > 0 {
		exp = time.Now().UnixMilli() + ttlMillis
	}
	oldLen := 0
	if e, ok := s.data[key]; ok {
		if e.ExpireAt == 0 || time.Now().UnixMilli() < e.ExpireAt {
			oldLen = len(e.Value)
		}
	}
	delta := int64(len(value) - oldLen)
	if s.maxMemory > 0 && s.totalBytes+delta > s.maxMemory {
		return false
	}
	s.totalBytes += delta
	s.data[key] = &Entry{Value: value, ExpireAt: exp}
	return true
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
					// 调整内存计数
					s.totalBytes -= int64(len(v.Value))
					delete(s.data, k)
				}
			}
			s.mu.Unlock()
		}
	}()
}

<<<<<<< HEAD
=======
// SetMaxMemory sets a soft memory limit in bytes (0 disables limit).
func (s *Storage) SetMaxMemory(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxMemory = bytes
}

// MemoryUsage returns the current approximate memory usage in bytes.
func (s *Storage) MemoryUsage() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalBytes
}

// GetMaxMemory returns the configured max memory (0 means disabled).
func (s *Storage) GetMaxMemory() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.maxMemory
}

// Get retrieves the value for the given key. If the key is expired it will be
// removed and the function returns ("", false).
>>>>>>> f42e614 (feat(limits): add max-conns, connection timeout, max-memory enforcement)
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
<<<<<<< HEAD
		exp = time.Now().Unix() + ttlSeconds
=======
		exp = time.Now().UnixMilli() + ttlSeconds*1000
	}
	oldLen := 0
	if e, ok := s.data[key]; ok {
		// 如果旧值未过期，计算旧长度
		if e.ExpireAt == 0 || time.Now().UnixMilli() < e.ExpireAt {
			oldLen = len(e.Value)
		}
	}
	delta := int64(len(value) - oldLen)
	// 更新总字节数
	s.totalBytes += delta
	s.data[key] = &Entry{Value: value, ExpireAt: exp}
}

// SetWithMs sets the value and ttl in milliseconds (0 means no expiration).
func (s *Storage) SetWithMs(key, value string, ttlMillis int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp := int64(0)
	if ttlMillis > 0 {
		exp = time.Now().UnixMilli() + ttlMillis
>>>>>>> f42e614 (feat(limits): add max-conns, connection timeout, max-memory enforcement)
	}
	oldLen := 0
	if e, ok := s.data[key]; ok {
		if e.ExpireAt == 0 || time.Now().UnixMilli() < e.ExpireAt {
			oldLen = len(e.Value)
		}
	}
	delta := int64(len(value) - oldLen)
	s.totalBytes += delta
	s.data[key] = &Entry{Value: value, ExpireAt: exp}
}

func (s *Storage) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.data[key]; ok {
		// 调整内存计数
		s.totalBytes -= int64(len(e.Value))
		delete(s.data, key)
		return true
	}
	return false
}

<<<<<<< HEAD
=======
// IncrBy atomically increments the integer value of a key by delta. If the key
// does not exist it is set to delta. Returns the new value or an error if the
// current value is not an integer.
func (s *Storage) IncrBy(key string, delta int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var cur int64
	oldLen := 0
	if e, ok := s.data[key]; ok {
		if e.ExpireAt != 0 && time.Now().UnixMilli() >= e.ExpireAt {
			// expired
			// 调整内存计数
			s.totalBytes -= int64(len(e.Value))
			delete(s.data, key)
			cur = 0
		} else {
			val := e.Value
			oldLen = len(val)
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
	newVal := strconv.FormatInt(cur, 10)
	// 更新总字节数
	s.totalBytes += int64(len(newVal) - oldLen)
	if e, ok := s.data[key]; ok {
		e.Value = newVal
	} else {
		s.data[key] = &Entry{Value: newVal, ExpireAt: 0}
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

// TrySet tries to set a key given seconds TTL, honoring maxMemory if set.
// Returns true if set succeeded, false if rejected due to memory limit.
func (s *Storage) TrySet(key, value string, ttlSeconds int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp := int64(0)
	if ttlSeconds > 0 {
		exp = time.Now().UnixMilli() + ttlSeconds*1000
	}
	oldLen := 0
	if e, ok := s.data[key]; ok {
		if e.ExpireAt == 0 || time.Now().UnixMilli() < e.ExpireAt {
			oldLen = len(e.Value)
		}
	}
	delta := int64(len(value) - oldLen)
	if s.maxMemory > 0 && s.totalBytes+delta > s.maxMemory {
		return false
	}
	s.totalBytes += delta
	s.data[key] = &Entry{Value: value, ExpireAt: exp}
	return true
}

// TrySetWithMs tries to set a key given ms TTL, honoring maxMemory if set.
func (s *Storage) TrySetWithMs(key, value string, ttlMillis int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp := int64(0)
	if ttlMillis > 0 {
		exp = time.Now().UnixMilli() + ttlMillis
	}
	oldLen := 0
	if e, ok := s.data[key]; ok {
		if e.ExpireAt == 0 || time.Now().UnixMilli() < e.ExpireAt {
			oldLen = len(e.Value)
		}
	}
	delta := int64(len(value) - oldLen)
	if s.maxMemory > 0 && s.totalBytes+delta > s.maxMemory {
		return false
	}
	s.totalBytes += delta
	s.data[key] = &Entry{Value: value, ExpireAt: exp}
	return true
}

>>>>>>> f42e614 (feat(limits): add max-conns, connection timeout, max-memory enforcement)
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
					// 调整内存计数
					s.totalBytes -= int64(len(v.Value))
					delete(s.data, k)
				}
			}
			s.mu.Unlock()
		}
	}()
}
