package proxy

import (
    "bytes"
    "log"
    "net"
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
    backendParsed   *url.URL
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
        backendParsed:    backend,
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
            if req.TLS != nil {
                req.Header.Set("X-Forwarded-Proto", "https")
            } else {
                req.Header.Set("X-Forwarded-Proto", "http")
            }
			req.Host = backend.Host
		},
		ModifyResponse: proxy.modifyResponse,
		ErrorHandler:   proxy.errorHandler,
	}

    // Tune transport for better performance
    transport := &http.Transport{
        Proxy:                 http.ProxyFromEnvironment,
        ForceAttemptHTTP2:     true,
        MaxIdleConns:          512,
        MaxIdleConnsPerHost:   256,
        IdleConnTimeout:       120 * time.Second,
        TLSHandshakeTimeout:   10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    }
    proxy.proxy.Transport = transport
    proxy.proxy.FlushInterval = 100 * time.Millisecond

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
    rp.metricsCollector.AddRequestLog(metrics.RequestLogEntry{
            Timestamp:  time.Now(),
            Method:     r.Method,
            Path:       r.URL.String(),
            Status:     http.StatusTooManyRequests,
            ClientIP:   clientIP,
            DurationMs: float64(time.Since(start).Microseconds()) / 1000.0,
            CacheHit:   false,
            Bytes:      0,
        Host:       rp.backendParsed.Host,
        Scheme:     rp.backendParsed.Scheme,
        UserAgent:  r.UserAgent(),
        Referer:    r.Referer(),
        ContentType: "",
        CacheTTLRemainingMs: 0,
        })
		return
	}
	
	// Handle GET requests with caching
	if r.Method == http.MethodGet {
		rp.handleGetRequest(w, r, clientIP)
		return
	}
	
    // Handle other methods directly with capture for logging
    capture := &responseCapture{
        ResponseWriter: w,
        statusCode:     http.StatusOK,
        headers:        make(http.Header),
        body:           &bytes.Buffer{},
    }
    rp.proxy.ServeHTTP(capture, r)
    duration := time.Since(start)
    rp.metricsCollector.RecordResponseTime(duration)
    rp.metricsCollector.AddRequestLog(metrics.RequestLogEntry{
        Timestamp:  time.Now(),
        Method:     r.Method,
        Path:       r.URL.String(),
        Status:     capture.statusCode,
        ClientIP:   clientIP,
        DurationMs: float64(duration.Microseconds()) / 1000.0,
        CacheHit:   false,
        Bytes:      capture.body.Len(),
        Host:       rp.backendParsed.Host,
        Scheme:     rp.backendParsed.Scheme,
        UserAgent:  r.UserAgent(),
        Referer:    r.Referer(),
        ContentType: capture.headers.Get("Content-Type"),
        CacheTTLRemainingMs: 0,
    })
}

func (rp *ReverseProxy) handleGetRequest(w http.ResponseWriter, r *http.Request, clientIP string) {
    start := time.Now()
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
        _, _ = w.Write(cachedResponse.Body)
        duration := time.Since(start)
        rp.metricsCollector.RecordResponseTime(duration)
        // compute remaining TTL
        remaining := time.Until(cachedResponse.ExpiresAt)
        if remaining < 0 { remaining = 0 }
        rp.metricsCollector.AddRequestLog(metrics.RequestLogEntry{
            Timestamp:  time.Now(),
            Method:     r.Method,
            Path:       r.URL.String(),
            Status:     cachedResponse.StatusCode,
            ClientIP:   clientIP,
            DurationMs: float64(duration.Microseconds()) / 1000.0,
            CacheHit:   true,
            Bytes:      len(cachedResponse.Body),
            Host:       rp.backendParsed.Host,
            Scheme:     rp.backendParsed.Scheme,
            UserAgent:  r.UserAgent(),
            Referer:    r.Referer(),
            ContentType: http.Header(cachedResponse.Headers).Get("Content-Type"),
            CacheTTLRemainingMs: float64(remaining.Microseconds()) / 1000.0,
        })
		return
	}
	
	rp.metricsCollector.IncrementCacheMisses()
	
	// Create a custom response writer to capture the response
	responseWriter := &responseCapture{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		headers:        make(http.Header),
        body:           &bytes.Buffer{},
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

    duration := time.Since(start)
    rp.metricsCollector.RecordResponseTime(duration)
    rp.metricsCollector.AddRequestLog(metrics.RequestLogEntry{
        Timestamp:  time.Now(),
        Method:     r.Method,
        Path:       r.URL.String(),
        Status:     responseWriter.statusCode,
        ClientIP:   clientIP,
        DurationMs: float64(duration.Microseconds()) / 1000.0,
        CacheHit:   false,
        Bytes:      responseWriter.body.Len(),
        Host:       rp.backendParsed.Host,
        Scheme:     rp.backendParsed.Scheme,
        UserAgent:  r.UserAgent(),
        Referer:    r.Referer(),
        ContentType: responseWriter.headers.Get("Content-Type"),
        CacheTTLRemainingMs: 0,
    })
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
	
    // Use remote address (strip port if present)
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err == nil && host != "" {
        return host
    }
    return r.RemoteAddr
}

// responseCapture captures the response for caching
type responseCapture struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       *bytes.Buffer
    captured   bool
}

func (rc *responseCapture) WriteHeader(statusCode int) {
    rc.statusCode = statusCode
    rc.captureHeaders()
    rc.ResponseWriter.WriteHeader(statusCode)
}

func (rc *responseCapture) Write(data []byte) (int, error) {
    if !rc.captured {
        if rc.statusCode == 0 {
            rc.statusCode = http.StatusOK
        }
        rc.captureHeaders()
    }
    rc.body.Write(data)
    return rc.ResponseWriter.Write(data)
}

func (rc *responseCapture) Header() http.Header {
    // delegate to underlying so headers are actually sent to client
    return rc.ResponseWriter.Header()
}

// captureHeaders copies headers from the underlying header map once
func (rc *responseCapture) captureHeaders() {
    if rc.captured {
        return
    }
    if rc.headers != nil {
        for k, v := range rc.ResponseWriter.Header() {
            vv := make([]string, len(v))
            copy(vv, v)
            rc.headers[k] = vv
        }
    }
    rc.captured = true
}