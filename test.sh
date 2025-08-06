#!/bin/bash

# Test script for GoProxy

echo "=== GoProxy Test Script ==="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
    fi
}

# Test health endpoint
echo "Testing health endpoint..."
curl -s -f http://localhost:8080/health > /dev/null
print_status $? "Health endpoint"

# Test metrics endpoint
echo "Testing metrics endpoint..."
curl -s -f http://localhost:8080/metrics > /dev/null
print_status $? "Metrics endpoint"

# Test proxy to backend
echo "Testing proxy to backend..."
curl -s -f http://localhost:8080/ > /dev/null
print_status $? "Proxy to backend"

# Test cache functionality
echo "Testing cache functionality..."
echo "First request (should miss cache):"
curl -s http://localhost:8080/ | head -1
echo
echo "Second request (should hit cache):"
curl -s http://localhost:8080/ | head -1
echo

# Test rate limiting
echo "Testing rate limiting..."
echo "Making 10 requests to test rate limiting..."
for i in {1..10}; do
    response=$(curl -s -w "%{http_code}" http://localhost:8080/ -o /dev/null)
    if [ "$response" = "429" ]; then
        echo -e "${YELLOW}Rate limit hit after $i requests${NC}"
        break
    fi
done

# Show metrics
echo
echo "=== Current Metrics ==="
curl -s http://localhost:8080/metrics | grep -E "(total_requests|cache_hits|cache_misses|blocked_requests)"

echo
echo "=== Test Complete ===" 