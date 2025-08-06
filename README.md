# GoProxy - Simple and Fast Web Proxy Server

**GoProxy** is a web server that acts as a middleman between your computer and other websites. Think of it like a smart assistant that helps make web requests faster and more secure.

## What Does GoProxy Do?

Imagine you're trying to get information from a website, but it's slow or you want to access it more efficiently. GoProxy sits between you and that website, helping by:

- **Making things faster** - It remembers (caches) information you've already requested, so you don't have to wait for it again
- **Protecting against overload** - It limits how many requests can be made at once to prevent the server from getting overwhelmed
- **Keeping track of usage** - It provides statistics about how much traffic is going through
- **Adding security** - It can help protect your backend services from direct access

## Key Features

### 🚀 **Speed Boost**
- **Smart Caching**: Remembers responses and serves them instantly on repeat requests
- **High Performance**: Built with Go for fast, efficient operation

### 🛡️ **Protection**
- **Rate Limiting**: Prevents too many requests from overwhelming your server
- **Request Filtering**: Can block or limit access based on rules you set

### 📊 **Monitoring**
- **Real-time Metrics**: See exactly how much traffic you're handling
- **Performance Stats**: Track response times, cache hits, and more
- **Health Checks**: Easy way to verify everything is working

### 🔧 **Easy to Use**
- **Simple Setup**: Just run one command to get started
- **Flexible Configuration**: Easy to customize for your needs
- **Cross-Platform**: Works on Windows, Mac, and Linux

## Quick Start Guide

### Option 1: The Easiest Way (Docker)

If you have Docker installed, this is the simplest approach:

```bash
# Start everything with one command
docker-compose up
```

This will start:
- The proxy server (accessible at http://localhost:8080)
- A test backend server (accessible at http://localhost:8081)
- A monitoring dashboard (accessible at http://localhost:9090)

### Option 2: Manual Installation

#### Step 1: Install Go
First, you need to install Go on your computer:
- **Windows**: Download from https://golang.org/dl/ and run the installer
- **Mac**: Use `brew install go` or download from the website
- **Linux**: Use your package manager (e.g., `sudo apt install golang-go`)

#### Step 2: Build the Project
```bash
# Download the project
git clone https://github.com/Tuhin-SnapD/GoProxy
cd goproxy

# Install dependencies and build
go mod tidy
go build -o goproxy.exe main.go
go build -o test_server.exe test_server.go
```

#### Step 3: Run the Servers
```bash
# Start the test backend server (in one terminal)
./test_server.exe

# Start the proxy server (in another terminal)
./goproxy.exe
```

## How to Use GoProxy

### Basic Usage
Once running, GoProxy will be available at `http://localhost:8080`. Here's what you can do:

1. **Access your backend**: Any request to `http://localhost:8080/` will be forwarded to your backend server
2. **Check health**: Visit `http://localhost:8080/health` to see if everything is working
3. **View metrics**: Visit `http://localhost:8080/metrics` to see performance statistics

### Customizing Settings
You can customize how GoProxy behaves:

```bash
# Run with custom settings
./goproxy.exe -port 9000 -backend http://my-website.com -rate-limit 200 -cache-ttl 10m
```

**What these settings mean:**
- `-port 9000`: Run on port 9000 instead of 8080
- `-backend http://my-website.com`: Forward requests to your website
- `-rate-limit 200`: Allow 200 requests per minute per user
- `-cache-ttl 10m`: Keep cached responses for 10 minutes

## Understanding the Metrics

When you visit `/metrics`, you'll see information like this:

```
goproxy_total_requests 150        # Total number of requests handled
goproxy_cache_hits 120           # Requests served from cache (fast!)
goproxy_cache_misses 30          # Requests that had to go to backend
goproxy_blocked_requests 5       # Requests blocked due to rate limiting
goproxy_cache_hit_rate 80.00     # Percentage of requests served from cache
```

**What this tells you:**
- **High cache hit rate** (like 80%+) means your proxy is working efficiently
- **Low blocked requests** means your rate limits are reasonable
- **Increasing total requests** shows your service is being used

## Common Use Cases

### 1. **Website Performance**
Use GoProxy in front of your website to:
- Speed up page loads with caching
- Reduce load on your main server
- Protect against traffic spikes

### 2. **API Protection**
Place GoProxy in front of your API to:
- Limit how many requests each user can make
- Cache frequently requested data
- Monitor API usage patterns

### 3. **Development Testing**
Use GoProxy during development to:
- Test how your app handles high traffic
- Simulate real-world conditions
- Debug performance issues

## Testing Your Setup

### Quick Test
Run the built-in test script to verify everything works:

```bash
# Windows
./test.bat

# Linux/Mac
./test.sh
```

### Manual Testing
You can also test manually:

```bash
# Test health endpoint
curl http://localhost:8080/health

# Test proxy (should return backend response)
curl http://localhost:8080/

# Test caching (second request should be faster)
curl http://localhost:8080/

# Check metrics
curl http://localhost:8080/metrics
```

## Troubleshooting

### Common Issues

**"Port already in use"**
- Another program is using the same port
- Solution: Use a different port with `-port 8081`

**"Backend not reachable"**
- Your backend server isn't running
- Solution: Start your backend server first

**"High memory usage"**
- Cache is storing too much data
- Solution: Reduce cache TTL with `-cache-ttl 1m`

### Getting Help
- Check the logs for error messages
- Verify your backend server is running
- Test with the built-in test scripts
- Check the metrics endpoint for clues

## Configuration Examples

### For Development
```bash
./goproxy.exe -port 8080 -backend http://localhost:3000 -rate-limit 1000 -cache-ttl 1m
```
- High rate limits for testing
- Short cache time for fresh data

### For Production
```bash
./goproxy.exe -port 80 -backend https://my-production-site.com -rate-limit 50 -cache-ttl 10m
```
- Lower rate limits for security
- Longer cache time for performance

### For High Traffic
```bash
./goproxy.exe -port 443 -backend https://my-busy-site.com -rate-limit 500 -cache-ttl 30m
```
- Higher rate limits for busy sites
- Long cache time for maximum performance

## What's Inside GoProxy?

GoProxy is built with several components that work together:

### 🔄 **Proxy Component**
- Handles incoming web requests
- Forwards them to your backend server
- Captures responses for caching

### 💾 **Cache Component**
- Stores responses in memory
- Automatically removes old data
- Makes repeat requests super fast

### 🚦 **Rate Limiter**
- Tracks requests per user
- Blocks users who make too many requests
- Protects your server from overload

### 📈 **Metrics Component**
- Counts requests, cache hits, etc.
- Provides performance statistics
- Helps you understand usage patterns

## Performance Tips

### For Best Performance
1. **Use appropriate cache TTL**: Longer for static content, shorter for dynamic
2. **Set reasonable rate limits**: High enough for normal use, low enough to prevent abuse
3. **Monitor metrics**: Watch cache hit rates and response times
4. **Scale horizontally**: Run multiple proxy instances for high traffic

### Memory Usage
- Cache size depends on TTL and request volume
- Monitor memory usage in production
- Adjust cache TTL if memory usage is too high

## Security Considerations

### Rate Limiting
- Prevents abuse and DoS attacks
- Configurable per your needs
- Tracks by IP address

### Headers
- Adds proxy identification headers
- Preserves original request information
- Can be customized for your needs

## Getting Started Checklist

- [ ] Install Go (if not using Docker)
- [ ] Download and build GoProxy
- [ ] Start your backend server
- [ ] Start GoProxy
- [ ] Test the health endpoint
- [ ] Test proxy functionality
- [ ] Check metrics
- [ ] Configure for your needs

## Need More Help?

If you're having trouble:
1. Check the troubleshooting section above
2. Look at the example configurations
3. Test with the built-in test scripts
4. Check the metrics for clues about what's happening

## Contributing

Want to help improve GoProxy?
1. Fork the repository
2. Make your changes
3. Test thoroughly
4. Submit a pull request

## License

This project is open source and available under the MIT License. 
