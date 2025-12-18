#!/bin/bash

echo "=========================================="
echo "Security Guardian - Security Testing Suite"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo "Target: $BASE_URL"
echo ""

# Test 1: Security Headers
echo "=========================================="
echo "Test 1: Security Headers Verification"
echo "=========================================="
echo ""

HEADERS=$(curl -s -I $BASE_URL/health)

check_header() {
    local header_name=$1
    local expected=$2
    
    if echo "$HEADERS" | grep -q "$header_name: $expected"; then
        echo -e "${GREEN}✓${NC} $header_name: PRESENT"
    else
        echo -e "${RED}✗${NC} $header_name: MISSING"
    fi
}

check_header "X-Frame-Options" "DENY"
check_header "X-Content-Type-Options" "nosniff"
check_header "X-Xss-Protection" "1; mode=block"
check_header "Referrer-Policy" "strict-origin-when-cross-origin"
check_header "Permissions-Policy" "geolocation=()"

echo ""

# Test 2: Rate Limiting
echo "=========================================="
echo "Test 2: Rate Limiting (Auth Endpoint)"
echo "=========================================="
echo ""

echo "Sending 25 requests to /api/auth/login..."
SUCCESS_COUNT=0
RATE_LIMITED=0

for i in {1..25}; do
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE_URL/api/auth/login)
    if [ "$RESPONSE" = "429" ]; then
        RATE_LIMITED=$((RATE_LIMITED + 1))
    else
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    fi
done

echo "Successful requests: $SUCCESS_COUNT"
echo "Rate limited (429): $RATE_LIMITED"

if [ $RATE_LIMITED -gt 0 ]; then
    echo -e "${GREEN}✓${NC} Rate limiting is ACTIVE"
else
    echo -e "${RED}✗${NC} Rate limiting NOT working"
fi

echo ""
echo "=========================================="
echo "Test 3: Protected Routes (JWT Required)"
echo "=========================================="
echo ""

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/api/transactions)
if [ "$RESPONSE" = "401" ]; then
    echo -e "${GREEN}✓${NC} Protected routes require authentication (401)"
else
    echo -e "${RED}✗${NC} Protected routes NOT properly secured (got $RESPONSE)"
fi

echo ""
echo "=========================================="
echo "Security Testing Complete"
echo "=========================================="
