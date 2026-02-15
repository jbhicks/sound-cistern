#!/bin/bash
# Ralph Context Manager

# Load shared state
load_shared_state() {
    if [ -f ".ralph/context/shared_state.yaml" ]; then
        export RALPH_SHARED_STATE=$(cat .ralph/context/shared_state.yaml)
    else
        echo "⚠️  Shared state not found"
        return 1
    fi
}

# Update shared state
update_shared_state() {
    local field="$1"
    local value="$2"
    
    # Handle nested fields with dot notation (e.g., "tasks.oauth.status")
    if command -v yq &> /dev/null; then
        yq eval ".$field = \"$value\"" -i .ralph/context/shared_state.yaml
    else
        echo "⚠️  yq not installed, using basic sed replacement"
        # Simple replacement for top-level fields only
        local field_name=$(echo "$field" | cut -d. -f1)
        sed -i "s/^  $field_name:.*/  $field_name: \"$value\"/" .ralph/context/shared_state.yaml
    fi
}

# Get current task
get_current_task() {
    if [ -f ".ralph/context/shared_state.yaml" ]; then
        grep "current_task:" .ralph/context/shared_state.yaml | cut -d: -f2 | tr -d ' '
    else
        echo "oauth_implementation"  # default
    fi
}

# Update task status
update_task_status() {
    local task="$1"
    local status="$2"
    echo "📝 Updating task $task status to $status"
    
    # Update shared state
    update_shared_state "tasks.$task.status" "$status"
    
    # Log the change
    echo "$(date): Task $task status updated to $status" >> .ralph/logs/task_changes.log
}

# Record verification result
record_verification_result() {
    local task="$1"
    local result_file="$2"
    
    if [ -f "$result_file" ]; then
        local status=$(jq -r '.status' "$result_file")
        local score=$(jq -r '.score' "$result_file")
        
        # Copy result to verification results
        cp "$result_file" ".ralph/verification_results/${task}_$(date +%s).json"
        
        # Update shared state
        update_shared_state "verification_results.$task" "$result_file"
        
        echo "🔍 Recorded verification for $task: $status ($score%)"
    else
        echo "❌ Verification result file not found: $result_file"
    fi
}

# Generate next iteration feedback
generate_feedback() {
    local task="$1"
    local verification_file="$2"
    
    if [ ! -f "$verification_file" ]; then
        echo "❌ Verification file not found: $verification_file"
        return 1
    fi
    
    local status=$(jq -r '.status' "$verification_file")
    local next_actions=$(jq -r '.next_actions[]' "$verification_file")
    
    # Update iteration feedback
    cat << EOF > .ralph/context/iteration_feedback.yaml
current_task: "$task"
current_iteration: $(($(grep current_iteration .ralph/context/shared_state.yaml | cut -d: -f2 | tr -d ' ') + 1))
last_verification: "$verification_file"

feedback_history:
$(cat .ralph/context/iteration_feedback.yaml | grep -A 20 "feedback_history:" | tail -n +2)

next_iteration:
  focus_areas:
$(echo "$next_actions" | sed 's/^/    - "/' | sed 's/$/"/')
  
EOF
    
    echo "📝 Generated feedback for next iteration"
}

# Check if loop should continue
should_continue_loop() {
    local completed_tasks=$(grep -c "status:.*completed" .ralph/context/shared_state.yaml || echo "0")
    local total_tasks=$(grep -c "status:" .ralph/context/shared_state.yaml | grep -v "status:.*loop_state" | grep -v "budget:" | grep -v "verification_results:" | grep -v "sub_agents:" | grep -v "user_interventions:" | grep -v "git_checkpoints:" || echo "3")
    
    if [ "$completed_tasks" -eq "$total_tasks" ]; then
        echo "🎉 All tasks completed!"
        return 1  # Don't continue
    fi
    
    # Check budget with fallback if bc not available
    local spent=$(grep "total_spent:" .ralph/context/shared_state.yaml | cut -d: -f2 | tr -d ' ')
    local allocated=$(grep "total_allocated:" .ralph/context/shared_state.yaml | cut -d: -f2 | tr -d ' ')
    
    if command -v bc &> /dev/null; then
        if (( $(echo "$spent > $allocated" | bc -l) )); then
            echo "💰 Budget exceeded!"
            return 1  # Don't continue
        fi
    else
        # Simple string comparison for integer amounts
        if [ "$spent" = "$allocated" ] || [ "${spent%.*}" -ge "${allocated%.*}" ]; then
            echo "💰 Budget exceeded!"
            return 1  # Don't continue
        fi
    fi
    
    return 0  # Continue loop
}

# Create git checkpoint
create_checkpoint() {
    local message="$1"
    
    # Check if there are changes to commit
    if ! git diff --quiet || ! git diff --cached --quiet; then
        git add .
        git commit -m "Ralph checkpoint: $message"
        local commit_hash=$(git rev-parse HEAD)
        
        echo "📸 Created git checkpoint: $commit_hash"
        update_shared_state "git_checkpoints[-1].commit_hash" "$commit_hash"
        update_shared_state "git_checkpoints[-1].timestamp" "$(date -Iseconds)"
        update_shared_state "git_checkpoints[-1].description" "$message"
    else
        echo "ℹ️  No changes to checkpoint"
    fi
}

# Main context management commands
case "${1:-help}" in
    "load")
        load_shared_state
        ;;
    "current-task")
        get_current_task
        ;;
    "update-status")
        update_task_status "$2" "$3"
        ;;
    "record-verification")
        record_verification_result "$2" "$3"
        ;;
    "generate-feedback")
        generate_feedback "$2" "$3"
        ;;
    "should-continue")
        should_continue_loop
        ;;
    "checkpoint")
        create_checkpoint "$2"
        ;;
    "help"|*)
        echo "Usage: $0 {load|current-task|update-status|record-verification|generate-feedback|should-continue|checkpoint} [args...]"
        echo ""
        echo "Commands:"
        echo "  load                    Load shared state"
        echo "  current-task            Get current task name"
        echo "  update-status TASK STATUS  Update task status"
        echo "  record-verification TASK FILE  Record verification result"
        echo "  generate-feedback TASK FILE  Generate feedback for next iteration"
        echo "  should-continue         Check if loop should continue"
        echo "  checkpoint MESSAGE      Create git checkpoint"
        ;;
esac