package proxy

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"goproxy/cache"
	"goproxy/metrics"
	"goproxy/ratelimit"
)

type ReverseProxy struct {
	backendURL      string
	cacheManager    *cache.Manager
	rateLimiter     *ratelimit.Manager
	metricsCollector *metrics.Collector
	proxy           *httputil.ReverseProxy
}

func New(backendURL string, cacheManager *cache.Manager, rateLimiter *ratelimit.Manager, metricsCollector *metrics.Collector) *ReverseProxy {
	backend, err := url.Parse(backendURL)
	if err != nil {
		log.Fatalf("Invalid backend URL: %v", err)
	}

	proxy := &ReverseProxy{
		backendURL:       backendURL,
		cacheManager:     cacheManager,
		rateLimiter:      rateLimiter,
		metricsCollector: metricsCollector,
	}

	// Create reverse proxy
	proxy.proxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = backend.Scheme
			req.URL.Host = backend.Host
			req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
			req.Host = backend.Host
		},
		ModifyResponse: proxy.modifyResponse,
		ErrorHandler:   proxy.errorHandler,
	}

	return proxy
}

func (rp *ReverseProxy) HandleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Extract client IP
	clientIP := getClientIP(r)
	
	// Update metrics
	rp.metricsCollector.IncrementTotalRequests()
	
	// Check rate limit
	if !rp.rateLimiter.Allow(clientIP) {
		rp.metricsCollector.IncrementBlockedRequests()
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	
	// Handle GET requests with caching
	if r.Method == http.MethodGet {
		rp.handleGetRequest(w, r, clientIP)
		return
	}
	
	// Handle other methods directly
	rp.proxy.ServeHTTP(w, r)
	
	// Record response time
	rp.metricsCollector.RecordResponseTime(time.Since(start))
}

func (rp *ReverseProxy) handleGetRequest(w http.ResponseWriter, r *http.Request, clientIP string) {
	// Create cache key
	cacheKey := r.URL.String()
	
	// Try to get from cache
	if cachedResponse := rp.cacheManager.Get(cacheKey); cachedResponse != nil {
		rp.metricsCollector.IncrementCacheHits()
		
		// Copy cached response to client
		for key, values := range cachedResponse.Headers {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(cachedResponse.StatusCode)
		w.Write(cachedResponse.Body)
		return
	}
	
	rp.metricsCollector.IncrementCacheMisses()
	
	// Create a custom response writer to capture the response
	responseWriter := &responseCapture{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
		body:          &bytes.Buffer{},
	}
	
	// Forward request to backend
	rp.proxy.ServeHTTP(responseWriter, r)
	
	// Cache successful GET responses
	if responseWriter.statusCode == http.StatusOK {
		cachedResponse := &cache.Response{
			StatusCode: responseWriter.statusCode,
			Headers:    responseWriter.headers,
			Body:       responseWriter.body.Bytes(),
		}
		rp.cacheManager.Set(cacheKey, cachedResponse)
	}
}

func (rp *ReverseProxy) modifyResponse(resp *http.Response) error {
	// Add custom headers
	resp.Header.Set("X-Proxy-Server", "goproxy")
	resp.Header.Set("X-Proxy-Timestamp", time.Now().Format(time.RFC3339))
	return nil
}

func (rp *ReverseProxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Proxy error: %v", err)
	http.Error(w, "Backend service unavailable", http.StatusServiceUnavailable)
}

func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check for X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	
	// Use remote address
	return r.RemoteAddr
}

// responseCapture captures the response for caching
type responseCapture struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       *bytes.Buffer
}

func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
}

func (rc *responseCapture) Write(data []byte) (int, error) {
	rc.body.Write(data)
	return rc.ResponseWriter.Write(data)
}

func (rc *responseCapture) Header() http.Header {
	return rc.headers
} 