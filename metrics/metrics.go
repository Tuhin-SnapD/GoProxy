package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Collector struct {
	totalRequests    int64
	cacheHits        int64
	cacheMisses      int64
	blockedRequests  int64
	responseTimes    []time.Duration
	responseTimeMutex sync.RWMutex
}

func New() *Collector {
	return &Collector{
		responseTimes: make([]time.Duration, 0, 1000),
	}
}

func (c *Collector) IncrementTotalRequests() {
	atomic.AddInt64(&c.totalRequests, 1)
}

func (c *Collector) IncrementCacheHits() {
	atomic.AddInt64(&c.cacheHits, 1)
}

func (c *Collector) IncrementCacheMisses() {
	atomic.AddInt64(&c.cacheMisses, 1)
}

func (c *Collector) IncrementBlockedRequests() {
	atomic.AddInt64(&c.blockedRequests, 1)
}

func (c *Collector) RecordResponseTime(duration time.Duration) {
	c.responseTimeMutex.Lock()
	defer c.responseTimeMutex.Unlock()
	
	// Keep only the last 1000 response times
	if len(c.responseTimes) >= 1000 {
		c.responseTimes = c.responseTimes[1:]
	}
	c.responseTimes = append(c.responseTimes, duration)
}

func (c *Collector) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	
	totalRequests := atomic.LoadInt64(&c.totalRequests)
	cacheHits := atomic.LoadInt64(&c.cacheHits)
	cacheMisses := atomic.LoadInt64(&c.cacheMisses)
	blockedRequests := atomic.LoadInt64(&c.blockedRequests)
	
	// Calculate cache hit rate
	var cacheHitRate float64
	if totalRequests > 0 {
		cacheHitRate = float64(cacheHits) / float64(cacheHits+cacheMisses) * 100
	}
	
	// Calculate average response time
	c.responseTimeMutex.RLock()
	var avgResponseTime time.Duration
	if len(c.responseTimes) > 0 {
		total := time.Duration(0)
		for _, rt := range c.responseTimes {
			total += rt
		}
		avgResponseTime = total / time.Duration(len(c.responseTimes))
	}
	responseTimeCount := len(c.responseTimes)
	c.responseTimeMutex.RUnlock()
	
	// Generate metrics in Prometheus format
	metrics := fmt.Sprintf(`# HELP goproxy_total_requests Total number of requests processed
# TYPE goproxy_total_requests counter
goproxy_total_requests %d

# HELP goproxy_cache_hits Total number of cache hits
# TYPE goproxy_cache_hits counter
goproxy_cache_hits %d

# HELP goproxy_cache_misses Total number of cache misses
# TYPE goproxy_cache_misses counter
goproxy_cache_misses %d

# HELP goproxy_blocked_requests Total number of blocked requests due to rate limiting
# TYPE goproxy_blocked_requests counter
goproxy_blocked_requests %d

# HELP goproxy_cache_hit_rate Cache hit rate percentage
# TYPE goproxy_cache_hit_rate gauge
goproxy_cache_hit_rate %.2f

# HELP goproxy_average_response_time Average response time in milliseconds
# TYPE goproxy_average_response_time gauge
goproxy_average_response_time %.2f

# HELP goproxy_response_time_samples Number of response time samples
# TYPE goproxy_response_time_samples gauge
goproxy_response_time_samples %d

# HELP goproxy_uptime_seconds Server uptime in seconds
# TYPE goproxy_uptime_seconds counter
goproxy_uptime_seconds %.0f
`,
		totalRequests,
		cacheHits,
		cacheMisses,
		blockedRequests,
		cacheHitRate,
		float64(avgResponseTime.Microseconds())/1000.0, // Convert to milliseconds
		responseTimeCount,
		float64(time.Since(startTime).Seconds()),
	)
	
	w.Write([]byte(metrics))
}

// Simple JSON metrics endpoint
func (c *Collector) HandleJSONMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	totalRequests := atomic.LoadInt64(&c.totalRequests)
	cacheHits := atomic.LoadInt64(&c.cacheHits)
	cacheMisses := atomic.LoadInt64(&c.cacheMisses)
	blockedRequests := atomic.LoadInt64(&c.blockedRequests)
	
	var cacheHitRate float64
	if totalRequests > 0 {
		cacheHitRate = float64(cacheHits) / float64(cacheHits+cacheMisses) * 100
	}
	
	c.responseTimeMutex.RLock()
	var avgResponseTime time.Duration
	if len(c.responseTimes) > 0 {
		total := time.Duration(0)
		for _, rt := range c.responseTimes {
			total += rt
		}
		avgResponseTime = total / time.Duration(len(c.responseTimes))
	}
	c.responseTimeMutex.RUnlock()
	
	json := fmt.Sprintf(`{
  "total_requests": %d,
  "cache_hits": %d,
  "cache_misses": %d,
  "blocked_requests": %d,
  "cache_hit_rate": %.2f,
  "average_response_time_ms": %.2f,
  "uptime_seconds": %.0f
}`,
		totalRequests,
		cacheHits,
		cacheMisses,
		blockedRequests,
		cacheHitRate,
		float64(avgResponseTime.Microseconds())/1000.0,
		float64(time.Since(startTime).Seconds()),
	)
	
	w.Write([]byte(json))
}

var startTime = time.Now() 