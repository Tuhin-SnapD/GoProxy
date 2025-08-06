package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goproxy/cache"
	"goproxy/metrics"
	"goproxy/proxy"
	"goproxy/ratelimit"
)

type Config struct {
	Port            string
	BackendURL      string
	RateLimitPerMin int
	CacheTTL        time.Duration
}

func main() {
	config := parseFlags()
	
	// Initialize components
	cacheManager := cache.New(config.CacheTTL)
	rateLimiter := ratelimit.New(config.RateLimitPerMin)
	metricsCollector := metrics.New()
	
	// Create reverse proxy
	reverseProxy := proxy.New(config.BackendURL, cacheManager, rateLimiter, metricsCollector)
	
	// Setup HTTP server
	mux := http.NewServeMux()
	
	// Proxy routes
	mux.HandleFunc("/", reverseProxy.HandleRequest)
	
	// Metrics endpoint
	mux.HandleFunc("/metrics", metricsCollector.HandleMetrics)
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start server in a goroutine
	go func() {
		log.Printf("Starting goproxy server on port %s", config.Port)
		log.Printf("Backend URL: %s", config.BackendURL)
		log.Printf("Rate limit: %d requests/min", config.RateLimitPerMin)
		log.Printf("Cache TTL: %v", config.CacheTTL)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
	
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// Cleanup
	cacheManager.Close()
	rateLimiter.Close()
	
	log.Println("Server stopped")
}

func parseFlags() *Config {
	port := flag.String("port", "8080", "Port to listen on")
	backendURL := flag.String("backend", "http://localhost:8081", "Backend URL to proxy to")
	rateLimitPerMin := flag.Int("rate-limit", 100, "Rate limit per IP per minute")
	cacheTTL := flag.Duration("cache-ttl", 5*time.Minute, "Cache TTL for GET responses")
	
	flag.Parse()
	
	return &Config{
		Port:            *port,
		BackendURL:      *backendURL,
		RateLimitPerMin: *rateLimitPerMin,
		CacheTTL:        *cacheTTL,
	}
} 