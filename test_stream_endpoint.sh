#!/bin/bash

# Test the /api/stream endpoint according to requirements

echo "🧪 Testing /api/stream endpoint behavior..."

# First test: unauthenticated request should fail
echo "Test 1: Unauthenticated request"
response=$(curl -s -w "%{http_code}" "http://localhost:8090/api/stream" 2>/dev/null)
http_code="${response: -3}"
body="${response%???}"

if [ "$http_code" = "401" ]; then
    echo "✅ Unauthenticated request properly rejected (401)"
else
    echo "❌ Unauthenticated request should return 401, got $http_code"
fi

# Second test: authenticated request with invalid token should fail
echo "Test 2: Invalid Bearer token"
response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer invalid_token" "http://localhost:8090/api/stream" 2>/dev/null)
http_code="${response: -3}"
body="${response%???}"

if [ "$http_code" = "401" ]; then
    echo "✅ Invalid token properly rejected (401)"
else
    echo "❌ Invalid token should return 401, got $http_code"
fi

# Check if response contains expected authentication error keywords
if echo "$body" | grep -q "authorization\|authentication\|unauthorized\|token"; then
    echo "✅ Response contains appropriate authentication error message"
else
    echo "❌ Response should contain authentication error keywords"
    echo "Response body: $body"
fi

echo ""
echo "📊 API Endpoint Tests Complete"
echo "Note: Full functionality test requires valid Soundcloud OAuth setup and user authentication"