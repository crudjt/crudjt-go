package crudjt

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type LRUCache struct {
	cache  *lru.Cache[string, map[string]interface{}]
	wFunc  func(string)
	mutex  sync.Mutex
}

const CACHE_CAPACITY = 40_000

func NewLRUCache(wFunc func(string)) *LRUCache {
	cache, _ := lru.New[string, map[string]interface{}](CACHE_CAPACITY)
	return &LRUCache{
		cache: cache,
		wFunc: wFunc,
	}
}

func (l *LRUCache) Get(value string) map[string]interface{} {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	cachedValue, ok := l.cache.Get(value)
	if !ok {
		return nil
	}

	l.cache.Add(value, cachedValue)
	output := make(map[string]interface{})

  metadataOut := make(map[string]interface{})

	if metadata, exists := cachedValue["metadata"].(map[string]interface{}); exists {
		if ttl, exists := metadata["ttl"].(time.Time); exists {
			remainingTime := int(time.Until(ttl).Seconds())
			if remainingTime <= 0 {
				l.cache.Remove(value)
				return nil
			}
			metadataOut["ttl"] = remainingTime
		}

		if silence_read, exists := metadata["silence_read"].(int); exists {
			silence_read -= 1
			metadataOut["silence_read"] = silence_read
      metadata["silence_read"] = silence_read

			if silence_read <= 0 {
				l.cache.Remove(value)
			}
			l.wFunc(value)
		}
	}

	if len(metadataOut) > 0 {
		output["metadata"] = metadataOut
	}

	output["data"] = cachedValue["data"]

	return output
}

func (l *LRUCache) Insert(key string, value map[string]interface{}, ttl int, silence_read int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	hash := map[string]interface{}{ "data": value }

	if ttl > 0 {
		if _, exists := hash["metadata"]; !exists {
			hash["metadata"] = make(map[string]interface{})
		}
		hash["metadata"].(map[string]interface{})["ttl"] = time.Now().Add(time.Duration(ttl) * time.Second)
	}

	if silence_read > 0 {
		if _, exists := hash["metadata"]; !exists {
			hash["metadata"] = make(map[string]interface{})
		}
		hash["metadata"].(map[string]interface{})["silence_read"] = silence_read
	}

	l.cache.Add(key, hash)
}

func (l *LRUCache) ForceInsert(value string, hash map[string]interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.cache.Add(value, hash)
}

func (l *LRUCache) Delete(value string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.cache.Remove(value)
}
