# GoProxy - High-Performance Reverse Proxy Server

A high-performance reverse proxy server written in Go with built-in caching, rate limiting, and metrics collection.

## Features

- **Reverse Proxy**: Forwards HTTP requests to configurable backend servers
- **In-Memory Caching**: Caches GET responses with configurable TTL
- **Rate Limiting**: Per-IP rate limiting with sliding window algorithm
- **Metrics**: Prometheus-style metrics endpoint with detailed statistics
- **Modular Design**: Organized into separate packages for maintainability
- **High Performance**: Uses Go's concurrency features (goroutines, channels, sync.Map)
- **Docker Support**: Complete containerization with docker-compose
- **Testing Tools**: Built-in test server and automated test scripts
- **Cross-Platform**: Windows and Unix build support

## Project Structure

```
goproxy/
├── main.go              # Main entry point and server setup
├── go.mod               # Go module file
├── README.md            # This file
├── proxy/
│   └── proxy.go         # Reverse proxy implementation
├── cache/
│   └── cache.go         # In-memory caching with TTL
├── ratelimit/
│   └── ratelimit.go     # Per-IP rate limiting
├── metrics/
│   └── metrics.go       # Metrics collection and endpoints
├── test_server.go       # Test backend server for development
├── Dockerfile           # Main application containerization
├── Dockerfile.backend   # Test backend containerization
├── docker-compose.yml   # Multi-service orchestration
├── prometheus.yml       # Prometheus metrics configuration
├── Makefile             # Unix build automation
├── build.bat            # Windows build script
├── test.sh              # Unix test script
├── test.bat             # Windows test script
└── example_config.json  # Example configuration file
```

## Quick Start

### Option 1: Docker (Recommended)

```bash
# Start all services (proxy, backend, and Prometheus)
docker-compose up

# Or run in background
docker-compose up -d
```

### Option 2: Local Build

#### Windows
```bash
# Build the project
build.bat

# Run the test backend server
test_server.exe

# In another terminal, run the proxy
goproxy.exe
```

#### Unix/Linux/macOS
```bash
# Build the project
make build

# Run the test backend server
make run-backend

# In another terminal, run the proxy
make run-proxy
```

## Installation

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (for containerized deployment)
- curl (for testing)

### Manual Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd goproxy
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the project:
```bash
# Unix/Linux/macOS
make build

# Windows
build.bat
```

## Usage

### Basic Usage

```bash
./goproxy
```

This starts the proxy server with default settings:
- Port: 8080
- Backend: http://localhost:8081
- Rate limit: 100 requests/minute per IP
- Cache TTL: 5 minutes

### Command Line Options

```bash
./goproxy -port 8080 -backend http://localhost:8081 -rate-limit 100 -cache-ttl 5m
```

Available flags:
- `-port`: Port to listen on (default: 8080)
- `-backend`: Backend URL to proxy to (default: http://localhost:8081)
- `-rate-limit`: Rate limit per IP per minute (default: 100)
- `-cache-ttl`: Cache TTL for GET responses (default: 5m)

### Example with Custom Settings

```bash
./goproxy -port 9000 -backend http://api.example.com -rate-limit 200 -cache-ttl 10m
```

## API Endpoints

### Proxy Endpoint
- **Path**: `/` (all paths except /metrics, /health)
- **Method**: All HTTP methods
- **Description**: Forwards requests to the backend server

### Metrics Endpoint
- **Path**: `/metrics`
- **Method**: GET
- **Description**: Returns Prometheus-style metrics
- **Content-Type**: text/plain

### Health Check Endpoint
- **Path**: `/health`
- **Method**: GET
- **Description**: Simple health check
- **Response**: "OK"

## Metrics

The `/metrics` endpoint provides the following metrics:

- `goproxy_total_requests`: Total number of requests processed
- `goproxy_cache_hits`: Total number of cache hits
- `goproxy_cache_misses`: Total number of cache misses
- `goproxy_blocked_requests`: Total number of blocked requests due to rate limiting
- `goproxy_cache_hit_rate`: Cache hit rate percentage
- `goproxy_average_response_time`: Average response time in milliseconds
- `goproxy_response_time_samples`: Number of response time samples
- `goproxy_uptime_seconds`: Server uptime in seconds

## Testing

### Automated Testing

#### Windows
```bash
# Run the test script
test.bat
```

#### Unix/Linux/macOS
```bash
# Make the script executable
chmod +x test.sh

# Run the test script
./test.sh
```

### Manual Testing

1. Start the test backend server:
```bash
# Unix/Linux/macOS
make run-backend

# Windows
test_server.exe
```

2. Start the proxy server:
```bash
# Unix/Linux/macOS
make run-proxy

# Windows
goproxy.exe
```

3. Test the endpoints:
```bash
# Health check
curl http://localhost:8080/health

# Metrics
curl http://localhost:8080/metrics

# Proxy to backend
curl http://localhost:8080/

# Test caching (second request should be faster)
curl http://localhost:8080/
```

## Docker Deployment

### Single Container
```bash
# Build and run the proxy
docker build -t goproxy .
docker run -p 8080:8080 goproxy
```

### Multi-Service with Docker Compose
```bash
# Start all services
docker-compose up

# Access services:
# - Proxy: http://localhost:8080
# - Backend: http://localhost:8081
# - Prometheus: http://localhost:9090
```

### Docker Compose Services

- **proxy**: Main reverse proxy server
- **backend**: Test backend server for development
- **prometheus**: Metrics collection and visualization

## Build Tools

### Makefile (Unix/Linux/macOS)
```bash
make build          # Build the project
make clean          # Clean build artifacts
make run-proxy      # Run the proxy server
make run-backend    # Run the test backend
make test           # Run tests
make help           # Show available commands
```

### Windows Batch Files
```bash
build.bat           # Build the project
test.bat            # Run tests
```

## Configuration

### Environment Variables
- `PORT`: Server port (default: 8080)
- `BACKEND_URL`: Backend URL (default: http://localhost:8081)

### Configuration File
See `example_config.json` for a comprehensive configuration example.

## Architecture

### Proxy Package
- Handles HTTP request routing
- Integrates with cache, rate limiting, and metrics
- Uses Go's `httputil.ReverseProxy` for efficient proxying
- Captures responses for caching

### Cache Package
- In-memory caching using `sync.Map`
- TTL-based expiration
- Automatic cleanup of expired entries
- Thread-safe operations

### Rate Limiting Package
- Per-IP rate limiting with sliding window algorithm
- Automatic cleanup of inactive IP limiters
- Configurable request limits and time windows

### Metrics Package
- Atomic counters for thread-safe metrics collection
- Prometheus-style metrics format
- Response time tracking
- JSON metrics endpoint for easy integration

## Performance Features

- **Concurrent Request Handling**: Uses Go's goroutines for concurrent request processing
- **Lock-Free Operations**: Uses atomic operations and sync.Map for high-performance metrics and caching
- **Memory Efficient**: Automatic cleanup of expired cache entries and inactive rate limiters
- **Connection Pooling**: Leverages Go's HTTP client connection pooling

## Monitoring and Observability

### Prometheus Integration
The project includes Prometheus configuration for metrics collection:
- Automatic metrics scraping
- Pre-configured dashboards
- Historical data retention

### Health Checks
- Built-in health check endpoint
- Docker health checks
- Graceful shutdown handling

## Configuration Examples

### Development Setup
```bash
./goproxy -port 8080 -backend http://localhost:3000 -rate-limit 1000 -cache-ttl 1m
```

### Production Setup
```bash
./goproxy -port 80 -backend https://api.production.com -rate-limit 50 -cache-ttl 10m
```

### High-Traffic Setup
```bash
./goproxy -port 443 -backend https://api.high-traffic.com -rate-limit 500 -cache-ttl 30m
```

### Docker Production
```bash
# Build production image
docker build -t goproxy:latest .

# Run with custom configuration
docker run -p 80:8080 \
  -e BACKEND_URL=https://api.production.com \
  goproxy:latest \
  ./goproxy -port 8080 -rate-limit 100 -cache-ttl 10m
```

## Troubleshooting

### Common Issues

1. **Port already in use**:
   ```bash
   # Check what's using the port
   netstat -an | grep 8080
   
   # Use a different port
   ./goproxy -port 8081
   ```

2. **Backend not reachable**:
   ```bash
   # Test backend connectivity
   curl http://localhost:8081/health
   
   # Check backend logs
   docker-compose logs backend
   ```

3. **High memory usage**:
   - Reduce cache TTL: `-cache-ttl 1m`
   - Lower rate limits: `-rate-limit 50`
   - Monitor with metrics endpoint

### Logs and Debugging

```bash
# View proxy logs
docker-compose logs proxy

# View backend logs
docker-compose logs backend

# Check metrics
curl http://localhost:8080/metrics
```

## Dependencies

- Go 1.21 or later
- Standard library packages:
  - `net/http`
  - `net/http/httputil`
  - `sync`
  - `sync/atomic`
  - `time`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is open source and available under the MIT License. 