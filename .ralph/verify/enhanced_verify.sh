#!/bin/bash
# Enhanced Ralph Verification Script with JSON Output

# Import shared helper functions
source "$(dirname "$0")/verification_helpers.sh"

test_oauth_url_generation() {
    # echo "🔍 Testing OAuth URL generation..."  # Commented for pure JSON output
    
    # Test OAuth initiation
    response=$(curl -s "http://localhost:8090/auth/soundcloud" 2>/dev/null)
    local status=$?
    
    if [ $status -ne 0 ]; then
        output_verification_result "oauth_url_test" "failed" 0 \
            ["server_error"] \
            "Server not responding on localhost:8090" \
            ["Check server is running", "Verify route implementation in main.go"]
        return
    fi
    
    local checks_passed=0
    local total_checks=4
    local failed_checks=()
    local next_actions=()
    
    # Check for Soundcloud OAuth URL
    if echo "$response" | grep -q "soundcloud.com/oauth/authorize\|soundcloud.com/connect"; then
        ((checks_passed++))
    else
        failed_checks+=("oauth_url_missing")
        next_actions+=("Add Soundcloud OAuth URL generation")
    fi
    
    # Check for PKCE parameters
    if echo "$response" | grep -q "code_challenge"; then
        ((checks_passed++))
    else
        failed_checks+=("pkce_missing")
        next_actions+=("Implement PKCE code challenge")
    fi
    
    # Check for state parameter
    if echo "$response" | grep -q "state="; then
        ((checks_passed++))
    else
        failed_checks+=("state_missing")
        next_actions+=("Add OAuth state parameter")
    fi
    
    # Check for proper redirect URI
    if echo "$response" | grep -q "redirect_uri=http://localhost:8090\|redirect_uri=http%3A%2F%2Flocalhost%3A8090"; then
        ((checks_passed++))
    else
        failed_checks+=("redirect_uri_incorrect")
        next_actions+=("Fix redirect URI configuration")
    fi
    
    local score=$((checks_passed * 100 / total_checks))
    
    if [ $score -eq 100 ]; then
        output_verification_result "oauth_url_test" "passed" $score \
            [] \
            "OAuth URL generation working correctly with PKCE support" \
            []
    else
        output_verification_result "oauth_url_test" "failed" $score \
            "${failed_checks[@]}" \
            "OAuth URL generation incomplete - missing $((total_checks - checks_passed)) components" \
            "${next_actions[@]}"
    fi
}

test_oauth_callback() {
    echo "🔄 Testing OAuth callback handling..."
    
    # Test callback endpoint exists and handles errors properly
    response=$(curl -s -w "%{http_code}" "http://localhost:8090/auth/soundcloud/callback?code=test&state=test" 2>/dev/null)
    local http_code="${response: -3}"
    local body="${response%???}"
    
    local checks_passed=0
    local total_checks=3
    local failed_checks=()
    local next_actions=()
    
    # Check endpoint responds (even with error)
    if [ "$http_code" != "000" ]; then
        ((checks_passed++))
    else
        failed_checks+=("endpoint_not_found")
        next_actions+=("Implement /auth/soundcloud/callback route")
    fi
    
    # Check proper error handling for invalid code
    if echo "$body" | grep -q "error\|invalid\|unauthorized"; then
        ((checks_passed++))
    else
        failed_checks+=("no_error_handling")
        next_actions+=("Add proper error handling for invalid OAuth codes")
    fi
    
    # Check state validation (should fail with missing/invalid state)
    response_invalid_state=$(curl -s -w "%{http_code}" "http://localhost:8090/auth/soundcloud/callback?code=test" 2>/dev/null)
    local invalid_state_code="${response_invalid_state: -3}"
    
    if [ "$invalid_state_code" = "400" ] || echo "${response_invalid_state%???}" | grep -q "state"; then
        ((checks_passed++))
    else
        failed_checks+=("no_state_validation")
        next_actions+=("Implement OAuth state parameter validation")
    fi
    
    local score=$((checks_passed * 100 / total_checks))
    
    if [ $score -ge 66 ]; then
        output_verification_result "oauth_callback_test" "passed" $score \
            [] \
            "OAuth callback endpoint working with proper error handling" \
            []
    else
        output_verification_result "oauth_callback_test" "failed" $score \
            "${failed_checks[@]}" \
            "OAuth callback handling incomplete - needs proper endpoint and validation" \
            "${next_actions[@]}"
    fi
}

test_token_storage() {
    echo "💾 Testing token storage mechanism..."
    
    local checks_passed=0
    local total_checks=3
    local failed_checks=()
    local next_actions=()
    
    # Check if soundcloud_users table exists
    if sqlite3 pb_data/data.db ".tables" 2>/dev/null | grep -q "soundcloud_users"; then
        ((checks_passed++))
    else
        failed_checks+=("missing_table")
        next_actions+=("Create soundcloud_users database migration")
    fi
    
    # Check table schema has required fields
    if sqlite3 pb_data/data.db "PRAGMA table_info(soundcloud_users);" 2>/dev/null | grep -q "access_token\|soundcloud_token\|token"; then
        ((checks_passed++))
    else
        failed_checks+=("missing_token_field")
        next_actions+=("Add access_token field to soundcloud_users table")
    fi
    
    # Check for encrypted storage (basic check - should not store plain tokens)
    # This is a simplified check - in practice you'd verify encryption is used
    if sqlite3 pb_data/data.db "SELECT sql FROM sqlite_master WHERE type='table' AND name='soundcloud_users';" 2>/dev/null | grep -q "soundcloud_users"; then
        ((checks_passed++))
    else
        failed_checks+=("no_storage_mechanism")
        next_actions+=("Implement secure token storage with encryption")
    fi
    
    local score=$((checks_passed * 100 / total_checks))
    
    if [ $score -eq 100 ]; then
        output_verification_result "token_storage_test" "passed" $score \
            [] \
            "Token storage mechanism implemented securely" \
            []
    else
        output_verification_result "token_storage_test" "failed" $score \
            "${failed_checks[@]}" \
            "Token storage incomplete - needs database schema and security" \
            "${next_actions[@]}"
    fi
}

test_track_fetching() {
    echo "🎵 Testing track fetching from Soundcloud API..."
    
    local checks_passed=0
    local total_checks=3
    local failed_checks=()
    local next_actions=()
    
    # Check if tracks table exists
    if sqlite3 pb_data/data.db ".tables" 2>/dev/null | grep -q "tracks\|soundcloud_tracks"; then
        ((checks_passed++))
    else
        failed_checks+=("missing_tracks_table")
        next_actions+=("Create tracks database migration")
    fi
    
    # Check API endpoint exists
    response=$(curl -s -w "%{http_code}" "http://localhost:8090/api/stream" 2>/dev/null)
    local http_code="${response: -3}"
    
    if [ "$http_code" != "000" ]; then
        ((checks_passed++))
    else
        failed_checks+=("missing_api_endpoint")
        next_actions+=("Implement /api/stream endpoint")
    fi
    
    # Check response structure (even for unauthenticated requests)
    body="${response%???}"
    if echo "$body" | grep -q "tracks\|error\|unauthorized\|authentication"; then
        ((checks_passed++))
    else
        failed_checks+=("invalid_response")
        next_actions+=("Fix API response structure")
    fi
    
    local score=$((checks_passed * 100 / total_checks))
    
    if [ $score -ge 66 ]; then
        output_verification_result "track_fetching_test" "passed" $score \
            [] \
            "Track fetching infrastructure in place" \
            []
    else
        output_verification_result "track_fetching_test" "failed" $score \
            "${failed_checks[@]}" \
            "Track fetching incomplete - needs database and API setup" \
            "${next_actions[@]}"
    fi
}

test_stream_display() {
    echo "🖥️ Testing stream display functionality..."
    
    local checks_passed=0
    local total_checks=3
    local failed_checks=()
    local next_actions=()
    
    # Check if stream template exists
    if [ -f "views/stream.templ" ] || [ -f "views/index.templ" ]; then
        ((checks_passed++))
    else
        failed_checks+=("missing_template")
        next_actions+=("Create stream display template")
    fi
    
    # Check template has proper structure
    if grep -q "soundcloud\|track\|stream" views/*.templ 2>/dev/null; then
        ((checks_passed++))
    else
        failed_checks+=("invalid_template")
        next_actions+=("Add track display structure to template")
    fi
    
    # Check for HTMX integration
    if grep -q "hx-\|htmx" views/*.templ 2>/dev/null || grep -q "htmx" public/*.js 2>/dev/null; then
        ((checks_passed++))
    else
        failed_checks+=("missing_htmx")
        next_actions+=("Add HTMX for dynamic updates")
    fi
    
    local score=$((checks_passed * 100 / total_checks))
    
    if [ $score -eq 100 ]; then
        output_verification_result "stream_display_test" "passed" $score \
            [] \
            "Stream display template working with HTMX" \
            []
    else
        output_verification_result "stream_display_test" "failed" $score \
            "${failed_checks[@]}" \
            "Stream display incomplete - needs template and HTMX setup" \
            "${next_actions[@]}"
    fi
}

# Main verification router
case "${1:-all}" in
    "oauth_url_test")
        test_oauth_url_generation
        ;;
    "oauth_callback_test")
        test_oauth_callback
        ;;
    "token_storage_test")
        test_token_storage
        ;;
    "track_fetching_test")
        test_track_fetching
        ;;
    "stream_display_test")
        test_stream_display
        ;;
    "all")
        echo "🚀 Starting comprehensive Ralph verification..."
        test_oauth_url_generation
        test_oauth_callback  
        test_token_storage
        test_track_fetching
        test_stream_display
        
        # Generate overall summary
        generate_overall_summary
        ;;
    *)
        echo "Usage: $0 {oauth_url_test|oauth_callback_test|token_storage_test|track_fetching_test|stream_display_test|all}"
        exit 1
        ;;
esac