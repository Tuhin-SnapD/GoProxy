package ratelimit

import (
	"sync"
	"time"
)

type Request struct {
	Timestamp time.Time
}

type IPLimiter struct {
	requests []Request
	limit    int
	window   time.Duration
	mutex    sync.RWMutex
}

type Manager struct {
	limiters sync.Map
	limit    int
	window   time.Duration
	stopChan chan struct{}
}

func New(requestsPerMinute int) *Manager {
	manager := &Manager{
		limit:    requestsPerMinute,
		window:   time.Minute,
		stopChan: make(chan struct{}),
	}
	
	// Start cleanup goroutine
	go manager.cleanup()
	
	return manager
}

func (m *Manager) Allow(ip string) bool {
	limiter := m.getOrCreateLimiter(ip)
	return limiter.Allow()
}

func (m *Manager) getOrCreateLimiter(ip string) *IPLimiter {
	// Try to get existing limiter
	if value, ok := m.limiters.Load(ip); ok {
		return value.(*IPLimiter)
	}
	
	// Create new limiter
	limiter := &IPLimiter{
		requests: make([]Request, 0),
		limit:    m.limit,
		window:   m.window,
	}
	
	// Store the limiter
	m.limiters.Store(ip, limiter)
	
	return limiter
}

func (m *Manager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.removeInactiveLimiters()
		case <-m.stopChan:
			return
		}
	}
}

func (m *Manager) removeInactiveLimiters() {
	cutoff := time.Now().Add(-10 * time.Minute)
	
	m.limiters.Range(func(key, value interface{}) bool {
		limiter := value.(*IPLimiter)
		limiter.mutex.RLock()
		hasRecentRequests := false
		
		for _, req := range limiter.requests {
			if req.Timestamp.After(cutoff) {
				hasRecentRequests = true
				break
			}
		}
		limiter.mutex.RUnlock()
		
		if !hasRecentRequests {
			m.limiters.Delete(key)
		}
		
		return true
	})
}

func (m *Manager) Close() {
	close(m.stopChan)
}

func (l *IPLimiter) Allow() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	now := time.Now()
	windowStart := now.Add(-l.window)
	
	// Remove old requests outside the window
	l.removeOldRequests(windowStart)
	
	// Check if we're under the limit
	if len(l.requests) < l.limit {
		l.requests = append(l.requests, Request{Timestamp: now})
		return true
	}
	
	return false
}

func (l *IPLimiter) removeOldRequests(windowStart time.Time) {
	validRequests := make([]Request, 0, len(l.requests))
	
	for _, req := range l.requests {
		if req.Timestamp.After(windowStart) {
			validRequests = append(validRequests, req)
		}
	}
	
	l.requests = validRequests
}

func (l *IPLimiter) GetCurrentCount() int {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	now := time.Now()
	windowStart := now.Add(-l.window)
	
	count := 0
	for _, req := range l.requests {
		if req.Timestamp.After(windowStart) {
			count++
		}
	}
	
	return count
} 