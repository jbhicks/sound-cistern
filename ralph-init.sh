#!/bin/bash
# Ralph System Initialization Script

# Creates all necessary directories and files for Ralph Wiggum autonomous loops
# Ensures proper permissions and dependencies are available

set -e  # Exit on any error

echo "🚀 Initializing Ralph Wiggum System..."

# Create required directory structure
create_directories() {
    echo "📁 Creating Ralph directory structure..."
    
    directories=(
        ".ralph/logs"
        ".ralph/results" 
        ".ralph/verification_results"
        ".ralph/state"
        ".ralph/context"
        ".ralph/config"
        ".ralph/tasks"
        ".ralph/verify"
        ".opencode/agents"
        ".opencode/skills"
        "skill/ralph-autonomous-loops"
        "skill/ralph-commands"
    )
    
    for dir in "${directories[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            echo "  ✅ Created: $dir"
        else
            echo "  ℹ️  Exists: $dir"
        fi
    done
}

# Set proper permissions
set_permissions() {
    echo "🔐 Setting permissions..."
    
    # Make all scripts executable
    find .ralph -name "*.sh" -type f -exec chmod +x {} \;
    find skill -name "*.sh" -type f -exec chmod +x {} \;
    
    # Ensure data directories are writable
    chmod 755 .ralph/logs .ralph/results .ralph/verification_results .ralph/state
    
    echo "  ✅ Permissions set"
}

# Check for required dependencies
check_dependencies() {
    echo "🔍 Checking dependencies..."
    
    local missing_deps=()
    local optional_deps=()
    
    # Required dependencies
    local required=("jq" "curl" "git")
    for dep in "${required[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        else
            echo "  ✅ Found: $dep"
        fi
    done
    
    # Optional dependencies (with fallbacks)
    local optional=("bc" "yq")
    for dep in "${optional[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            optional_deps+=("$dep")
        else
            echo "  ✅ Found: $dep"
        fi
    done
    
    # Report missing dependencies
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo ""
        echo "❌ Missing required dependencies:"
        printf "  • %s\n" "${missing_deps[@]}"
        echo ""
        echo "Please install missing dependencies:"
        echo "  Ubuntu/Debian: sudo apt-get install ${missing_deps[*]}"
        echo "  macOS: brew install ${missing_deps[*]}"
        echo ""
        exit 1
    fi
    
    if [ ${#optional_deps[@]} -gt 0 ]; then
        echo ""
        echo "⚠️  Missing optional dependencies (recommended for full functionality):"
        printf "  • %s\n" "${optional_deps[@]}"
        echo ""
        echo "Optional dependencies enhance Ralph functionality:"
        echo "  bc: Floating-point budget calculations"
        echo "  yq: Advanced YAML editing"
        echo ""
    fi
}

# Initialize shared state if it doesn't exist
initialize_shared_state() {
    echo "🗂️  Initializing shared state..."
    
    if [ ! -f ".ralph/context/shared_state.yaml" ]; then
        echo "  📝 Creating default shared state..."
        cat << 'EOF' > .ralph/context/shared_state.yaml
session_id: "session_$(date +%s)"
started_at: "$(date -Iseconds)"
orchestrator_agent: "ralph-orchestrator"

loop_state:
  status: "initialized"
  current_task: "oauth_implementation"
  current_iteration: 0
  total_iterations: 0

budget:
  total_allocated: 15.00
  total_spent: 0.00
  per_task:
    oauth_implementation:
      allocated: 4.50
      spent: 0.00
    stream_display:
      allocated: 3.00
      spent: 0.00
    track_metadata:
      allocated: 2.50
      spent: 0.00

tasks:
  oauth_implementation:
    status: "pending"
    iterations: 0
    last_verification: null
    feedback_history: []
  stream_display:
    status: "pending"
    dependencies: ["oauth_implementation"]
    iterations: 0
  track_metadata:
    status: "pending"
    dependencies: ["stream_display"]
    iterations: 0

sub_agents: {}
verification_results: {}
user_interventions:
  hints: []
  pauses: []
  manual_completions: []
git_checkpoints: []
EOF
        echo "  ✅ Shared state created"
    else
        echo "  ℹ️  Shared state already exists"
    fi
}

# Initialize context files
initialize_context_files() {
    echo "📝 Initializing context files..."
    
    # Iteration feedback template
    if [ ! -f ".ralph/context/iteration_feedback.yaml" ]; then
        cat << 'EOF' > .ralph/context/iteration_feedback.yaml
current_task: "oauth_implementation"
current_iteration: 0
last_verification: null

feedback_history: []

next_iteration:
  focus_areas: []
  avoid_areas: []
  specific_hints: []
  escalation_level: 0

recommended_patterns:
  go_patterns: []
  security_patterns: []

reference_files: []
previous_results: []
EOF
        echo "  ✅ Iteration feedback template created"
    fi
    
    # Task changes log
    touch .ralph/logs/task_changes.log
    touch .ralph/logs/session.log
    
    echo "  ✅ Context files initialized"
}

# Verify Ralph configuration
verify_configuration() {
    echo "🔍 Verifying Ralph configuration..."
    
    local config_files=(
        ".ralph/tasks/structured_tasks.yaml"
        ".ralph/verify/enhanced_verify.sh"
        ".ralph/verify/verification_helpers.sh"
        ".ralph/context/context_manager.sh"
        ".opencode/agents/ralph-orchestrator.md"
        "skill/ralph-autonomous-loops/SKILL.md"
        "skill/ralph-commands/SKILL.md"
    )
    
    local missing_files=()
    
    for file in "${config_files[@]}"; do
        if [ -f "$file" ]; then
            echo "  ✅ Found: $file"
        else
            missing_files+=("$file")
        fi
    done
    
    if [ ${#missing_files[@]} -gt 0 ]; then
        echo ""
        echo "❌ Missing configuration files:"
        printf "  • %s\n" "${missing_files[@]}"
        echo ""
        echo "Please ensure all Ralph components are properly installed."
        exit 1
    fi
}

# Test basic functionality
test_system() {
    echo "🧪 Testing basic system functionality..."
    
    # Test context manager
    if .ralph/context/context_manager.sh current-task >/dev/null 2>&1; then
        echo "  ✅ Context manager working"
    else
        echo "  ❌ Context manager failed"
        return 1
    fi
    
    # Test verification helpers
    if .ralph/verify/verification_helpers.sh >/dev/null 2>&1; then
        echo "  ✅ Verification helpers working"
    else
        echo "  ❌ Verification helpers failed"
        return 1
    fi
    
    # Test JSON parsing
    if echo '{"test": "value"}' | jq . >/dev/null 2>&1; then
        echo "  ✅ JSON parsing working"
    else
        echo "  ❌ JSON parsing failed"
        return 1
    fi
    
    echo "  ✅ Basic system tests passed"
}

# Show usage instructions
show_usage() {
    echo ""
    echo "🎉 Ralph Wiggum System Initialized Successfully!"
    echo "=============================================="
    echo ""
    echo "📋 Quick Start:"
    echo "  1. Start OpenCode with Ralph Orchestrator:"
    echo "     opencode"
    echo ""
    echo "  2. Load Ralph Orchestrator agent (Tab key to switch agents)"
    echo "  3. Start autonomous development:"
    echo "     'Start Ralph autonomous development for Soundcloud MVP'"
    echo ""
    echo "🎛️  User Commands (during execution):"
    echo "  /ralph-status     - Show current progress"
    echo "  /ralph-pause      - Pause autonomous loop"
    echo "  /ralph-resume     - Resume paused loop"
    echo "  /ralph-hint       - Add specific guidance"
    echo ""
    echo "📁 Ralph Directory Structure:"
    echo "  .ralph/config/     - Configuration files"
    echo "  .ralph/context/    - Shared state and context"
    echo "  .ralph/verify/     - Verification scripts"
    echo "  .ralph/results/    - Test results"
    echo "  .ralph/logs/       - Activity logs"
    echo ""
    echo "📚 Documentation:"
    echo "  skill/ralph-autonomous-loops/SKILL.md - Core patterns"
    echo "  skill/ralph-commands/SKILL.md       - User commands"
    echo ""
    echo "🚀 Your Ralph system is ready for autonomous development!"
}

# Main initialization flow
main() {
    echo "Ralph Wiggum Autonomous Loop System"
    echo "===================================="
    echo ""
    
    create_directories
    set_permissions
    check_dependencies
    initialize_shared_state
    initialize_context_files
    verify_configuration
    test_system
    show_usage
}

# Run initialization
main "$@"