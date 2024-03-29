package cache_store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type CacheStore struct {
	data        sync.Map
	maxCapacity int64
	cacheMemory int64
	duration    uint32
}

func NewCacheStore(maxCapacity int64, duration uint32) *CacheStore {
	maxCapacity = maxCapacity * 1024 * 1024
	return &CacheStore{maxCapacity: maxCapacity, duration: duration}
}

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

func (m *CacheStore) Set(key string, value interface{}) {
	key = m.getCacheKey(key)
	useMemory := m.cacheMemory + int64(len(fmt.Sprintf("%v", key))) + int64(len(fmt.Sprintf("%v", value)))
	if useMemory > m.maxCapacity {
		return
	}

	duration := time.Second * time.Duration(m.duration)
	expiration := time.Now().Add(duration).UnixNano()
	item := CacheItem{Value: value, Expiration: expiration}
	m.data.Store(key, item)
	m.cacheMemory = useMemory
}

func (m *CacheStore) Get(key string) (interface{}, bool) {
	key = m.getCacheKey(key)
	item, ok := m.data.Load(key)
	if !ok {
		return nil, false
	}

	cacheItem := item.(CacheItem)
	if time.Now().UnixNano() > cacheItem.Expiration {
		m.data.Delete(key)
		return nil, false
	}
	return cacheItem.Value, true
}

func (m *CacheStore) Clear() {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			m.clearExpiration()
		}
	}
}

func (m *CacheStore) clearExpiration() {
	var totalSize int64 = 0
	m.data.Range(func(key, value interface{}) bool {
		item := value.(CacheItem)
		if time.Now().UnixNano() > item.Expiration {
			m.data.Delete(key)
		} else {
			totalSize += int64(len(fmt.Sprintf("%v", key))) + int64(len(fmt.Sprintf("%v", item.Value)))
		}
		return true
	})
	m.cacheMemory = totalSize
}

func (m *CacheStore) getCacheKey(key string) string {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	hashBytes := hasher.Sum(nil)
	shortHash := hashBytes[:8]
	shortString := hex.EncodeToString(shortHash)

	return shortString
}
