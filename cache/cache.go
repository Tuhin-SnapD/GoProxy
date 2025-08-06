package cache

import (
	"sync"
	"time"
)

type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
	ExpiresAt  time.Time
}

type Manager struct {
	cache    sync.Map
	ttl      time.Duration
	stopChan chan struct{}
}

func New(ttl time.Duration) *Manager {
	manager := &Manager{
		ttl:      ttl,
		stopChan: make(chan struct{}),
	}
	
	// Start cleanup goroutine
	go manager.cleanup()
	
	return manager
}

func (m *Manager) Set(key string, response *Response) {
	response.ExpiresAt = time.Now().Add(m.ttl)
	m.cache.Store(key, response)
}

func (m *Manager) Get(key string) *Response {
	if value, ok := m.cache.Load(key); ok {
		response := value.(*Response)
		
		// Check if expired
		if time.Now().After(response.ExpiresAt) {
			m.cache.Delete(key)
			return nil
		}
		
		return response
	}
	
	return nil
}

func (m *Manager) Delete(key string) {
	m.cache.Delete(key)
}

func (m *Manager) Clear() {
	m.cache = sync.Map{}
}

func (m *Manager) Size() int {
	count := 0
	m.cache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func (m *Manager) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.removeExpired()
		case <-m.stopChan:
			return
		}
	}
}

func (m *Manager) removeExpired() {
	now := time.Now()
	
	m.cache.Range(func(key, value interface{}) bool {
		response := value.(*Response)
		if now.After(response.ExpiresAt) {
			m.cache.Delete(key)
		}
		return true
	})
}

func (m *Manager) Close() {
	close(m.stopChan)
} 