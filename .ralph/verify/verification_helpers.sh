#!/bin/bash
# Verification Helper Functions for Ralph

# Output verification result in structured JSON format
output_verification_result() {
    local test_name="$1"
    local status="$2"
    local score="$3"
    local failed_checks=("${@:4:10}")  # Up to 10 failed checks
    local feedback="${11}"
    local next_actions=("${@:12}")  # Remaining args are next actions
    
    # Convert arrays to JSON arrays
    local failed_checks_json=$(printf '"%s",' "${failed_checks[@]}" | sed 's/,$//')
    local next_actions_json=$(printf '"%s",' "${next_actions[@]}" | sed 's/,$//')
    
    # Create JSON output
    cat << EOF
{
  "test_name": "$test_name",
  "status": "$status",
  "score": $score,
  "timestamp": "$(date -Iseconds)",
  "failed_checks": [${failed_checks_json}],
  "feedback": "$feedback",
  "next_actions": [${next_actions_json}],
  "details": {
    "server": "http://localhost:8090",
    "database": "pb_data/data.db"
  }
}
EOF
}

# Generate overall summary from all test results
generate_overall_summary() {
    echo ""
    echo "📊 RALPH VERIFICATION SUMMARY"
    echo "================================"
    
    # Find all JSON result files and analyze
    local json_files=($(find .ralph/results -name "*.json" 2>/dev/null))
    
    if [ ${#json_files[@]} -eq 0 ]; then
        echo "⚠️  No test results found"
        return 1
    fi
    
    local total_tests=${#json_files[@]}
    local passed_tests=0
    local failed_tests=0
    local total_score=0
    
    for file in "${json_files[@]}"; do
        local status=$(jq -r '.status' "$file" 2>/dev/null)
        local score=$(jq -r '.score' "$file" 2>/dev/null)
        
        if [ "$status" = "passed" ]; then
            ((passed_tests++))
        else
            ((failed_tests++))
        fi
        
        total_score=$((total_score + score))
    done
    
    local average_score=$((total_score / total_tests))
    local completion_percentage=$((passed_tests * 100 / total_tests))
    
    echo "📈 Overall Progress: $completion_percentage% ($passed_tests/$total_tests tests passed)"
    echo "🎯 Average Score: $average_score%"
    echo "⏱️  Timestamp: $(date -Iseconds)"
    
    # Generate overall verdict
    if [ $completion_percentage -eq 100 ]; then
        echo ""
        echo "🎉 ALL CRITERIA MET - RALPH LOOP COMPLETE!"
        echo "🏆 Ready for user review and deployment"
        
        # Save completion state
        cat << EOF > .ralph/state/completion.json
{
  "status": "completed",
  "completion_percentage": $completion_percentage,
  "average_score": $average_score,
  "timestamp": "$(date -Iseconds)",
  "total_tests": $total_tests,
  "passed_tests": $passed_tests
}
EOF
        
    else
        echo ""
        echo "🔄 CONTINUE RALPH LOOP"
        echo "📝 Focus areas:"
        
        # Show next priority actions from failed tests
        for file in "${json_files[@]}"; do
            local status=$(jq -r '.status' "$file" 2>/dev/null)
            if [ "$status" = "failed" ]; then
                local test_name=$(jq -r '.test_name' "$file" 2>/dev/null)
                local next_actions=$(jq -r '.next_actions[]' "$file" 2>/dev/null | tr '\n' ',' | sed 's/,$//')
                echo "   • $test_name: $next_actions"
            fi
        done
    fi
    
    echo "================================"
}

# Ensure results directory exists
mkdir -p .ralph/results

# Clean up old results (keep last 5 runs)
cleanup_old_results() {
    find .ralph/results -name "*.json" -type f | sort -r | tail -n +6 | xargs rm -f
}

# Run cleanup at the beginning
cleanup_old_results