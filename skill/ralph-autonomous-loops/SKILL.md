---
name: ralph-autonomous-loops
description: Complete guide for implementing Ralph Wiggum autonomous loops with OpenCode AI agents
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: ai-developers
  stack: opencode-ralph-wiggum
---

# Ralph Wiggum Autonomous Loops for OpenCode

## Overview
This skill provides comprehensive patterns for implementing Ralph Wiggum-style autonomous development loops using OpenCode AI agents. Instead of one-shot prompts, create persistent, self-correcting loops that work until completion criteria are met.

## Core Philosophy

### The Ralph Wiggum Method
```bash
# The classic Bash loop pattern
while true; do cat PROMPT.md | opencode; done
```

### Why It Works for OpenCode
- **Persistent Context**: Each iteration sees previous work in files and git history
- **Self-Correction**: Failed attempts inform next iterations  
- **Iterative Progress**: Complex tasks broken into incremental improvements
- **Deterministic Success**: Clear completion criteria prevent infinite loops
- **Cost Control**: Built-in iteration limits and cost monitoring

## Ralph Loop Architecture

### Outer Loop Structure
```
┌──────────────────────────────────────────────┐
│                   Ralph Loop (OpenCode)              │
│  ┌────────────────────────────────────────┐  │
│  │  OpenCode Agent Loop (inner)              │  │
│  │  Agent ↔ Tools ↔ Results ... until done   │  │
│  └────────────────────────────────────────┘  │
│                         ↓                            │
│  verifyCompletion: "Is task actually complete?"  │
│                         ↓                            │
│       No? → Inject feedback → Run another  │
│       Yes? → Return final result                     │
└──────────────────────────────────────────────┘
```

### Key Components

1. **Orchestration Layer**: Main agent coordinates sub-agent execution
2. **Task Management**: Clear success criteria and progress tracking
3. **Sub-Agent System**: Focused agents with fresh contexts for specific tasks
4. **Context Management**: Shared state between orchestrator and sub-agents  
5. **Verification System**: Testable completion conditions with JSON output
6. **Feedback Injection**: Structured guidance for failed attempts
7. **Loop Control**: Iteration limits, cost controls, safety stops

## Implementation Patterns

### 1. Task Definition Structure

#### Task Definition Files
```yaml
# .ralph/tasks/current.yaml
task:
  name: "soundcloud-oauth-implementation"
  description: "Complete OAuth 2.1 flow with PKCE"
  priority: "high"
  
success_criteria:
  - name: "OAuth URL Generation"
    description: "Can generate Soundcloud authorization URLs"
    verification: "oauth_url_test"
    test_command: "./.ralph/verify/test_oauth.sh"
  
  - name: "Token Exchange"
    description: "Successfully exchange authorization codes for access tokens"
    verification: "token_exchange_test"
    test_command: "./.ralph/verify/test_token_exchange.sh"
    
blocking_issues:
  - name: "Missing OAuth endpoints"
    status: "blocked"
    
resources_needed:
  - "OAuth client credentials"
  - "PKCE implementation"
  - "Callback handling"
  - "Token refresh mechanism"
```

#### Success Criteria Design
```go
// Example verification function
func verifySoundcloudOAuth(result string, iteration int) (bool, string) {
    checks := []struct {
        name string
        test func() bool
    }{
        {"OAuth URL", func() bool {
            return strings.Contains(result, "soundcloud.com/oauth/authorize")
        }},
        {"Token Exchange", func() bool {
            return strings.Contains(result, "access_token") && 
                   strings.Contains(result, "expires_in")
        }},
        {"PKCE Support", func() bool {
            return strings.Contains(result, "code_challenge") && 
                   strings.Contains(result, "code_verifier")
        }},
    }
    
    passed := 0
    for _, check := range checks {
        if check.test() {
            passed++
        }
    }
    
    complete := passed == len(checks)
    
    if complete {
        return true, "All OAuth criteria verified in iteration " + strconv.Itoa(iteration)
    } else {
        // Generate specific feedback
        var failedChecks []string
        for _, check := range checks {
            if !check.test() {
                failedChecks = append(failedChecks, check.name)
            }
        }
        
        feedback := fmt.Sprintf("Iteration %d failed: %v", iteration, failedChecks)
        return false, feedback
    }
}
```

### 2. Loop Control Implementation

#### Iteration Management
```bash
# .ralph/control/loop_control.sh
#!/bin/bash

# Ralph Loop Control for OpenCode
# Usage: ./loop_control.sh [start|stop|status|configure]

MAX_ITERATIONS=${MAX_ITERATIONS:-50}
COST_LIMIT=${COST_LIMIT:-15.00}

case "$1" in
    start)
        echo "🚀 Starting Ralph loop..."
        echo "Max iterations: $MAX_ITERATIONS"
        echo "Cost limit: $COST_LIMIT"
        
        # Create lock file
        echo "$(date):STARTED" > .ralph/state/loop.lock
        
        # Start the loop
        opencode ralph-autonomous-loops "Implement soundcloud OAuth flow" \
            --max-iterations "$MAX_ITERATIONS" \
            --cost-limit "$COST_LIMIT" \
            --verbose
        ;;
        
    stop)
        echo "🛑 Stopping Ralph loop..."
        # Remove lock file
        rm -f .ralph/state/loop.lock
        ;;
        
    status)
        if [ -f ".ralph/state/loop.lock" ]; then
            echo "🔄 Ralph loop is RUNNING"
            echo "Lock file: $(cat .ralph/state/loop.lock)"
        else
            echo "⏸️ Ralph loop is IDLE"
        ;;
        
    configure)
        echo "⚙️ Ralph loop configuration:"
        echo "Current max iterations: ${MAX_ITERATIONS:-50}"
        echo "Current cost limit: ${COST_LIMIT:-15.00}"
        ;;
        
    *)
        echo "Usage: $0 {start|stop|status|configure}"
        echo "Commands:"
        echo "  start    - Start autonomous loop"
        echo "  stop     - Stop running loop"
        echo "  status   - Check loop status"
        echo "  configure - Configure loop parameters"
        exit 1
        ;;
esac
```

#### State Management
```yaml
# .ralph/state/current_session.yaml
session:
  id: "session_001"
  task: "soundcloud-oauth-implementation"
  started_at: "2025-01-25T15:30:00Z"
  current_iteration: 1
  max_iterations: 50
  cost_limit: 15.00
  total_cost: 0.00
  
progress:
  iterations_completed: 1
  last_completion_check: false
  context_updates: 0
  
status:
  loop_state: "FAILED"
  last_error: "Missing OAuth endpoints in main.go"
  
next_actions:
  - "Add /auth/soundcloud route to main.go"
  - "Implement OAuth callback handler"
  - "Create soundcloud_users migration"
  - "Test OAuth flow with verification script"
```

### 3. Context Management System

#### File-Based Context
```markdown
<!-- .ralph/context/iteration_summary.md -->
## Iteration 1 Summary (2025-01-25 15:45)

### Attempted Changes
- Add OAuth URL generation endpoint
- Implement token exchange logic
- Create basic callback handler

### Results
❌ **OAuth URL Generation**: FAILED
  - Endpoint `/auth/soundcloud` not found in main.go
  - Error: Route not implemented

❌ **Token Exchange**: NOT TESTED
  - Could not test without OAuth URLs

❌ **Callback Handling**: NOT TESTED
  - No implementation found

### Issues Identified
1. Missing route: `/auth/soundcloud` endpoint not implemented
2. Incomplete OAuth flow: No PKCE support
3. No token storage: No database schema for tokens

### Files Modified
- `main.go`: Attempted to add OAuth routes (failed)
- `views/oauth.templ`: Created OAuth template (not used)

### Next Actions
1. Implement missing `/auth/soundcloud` route
2. Add PKCE (Proof Key for Code Exchange) support
3. Complete OAuth callback handler
4. Add proper error handling and state management

### Context for Next Iteration
The next iteration should focus on implementing the missing OAuth endpoints with proper PKCE support. The main.go file needs these routes:

```go
// Add to main.go
app.Router.GET("/auth/soundcloud", func(c echo.Context) error {
    // OAuth URL generation
})

app.Router.GET("/auth/soundcloud/callback", func(c echo.Context) error {
    // OAuth callback handling
})
```

---

<!-- .ralph/context/pending_hints.md -->
## Hints for Next Iteration

### Critical Focus Areas
1. **PKCE Implementation**: Use SHA256 for code challenges, not plain text
2. **State Management**: Store OAuth state securely with expiration
3. **Error Handling**: Comprehensive error responses for OAuth failures
4. **Token Storage**: Encrypt tokens in database, not plain text
5. **Testing**: Each component should be individually testable

### Recommended Implementation Order
1. Implement OAuth URL generation with PKCE verifier
2. Add callback handler with state validation
3. Create token exchange with code verifier
4. Add proper error handling and redirects
5. Test each component independently before integration

### Code Patterns to Use
```go
// PKCE implementation
func generatePKCE() (string, string, string) {
    verifier := generateRandomString(128)
    challenge := sha256.Sum256([]byte(verifier))
    return base64.RawURLEncoding.EncodeToString(challenge), verifier
}

// Secure state management
func createOAuthState() (string, string) error {
    state := generateRandomString(32)
    expires := time.Now().Add(10 * time.Minute)
    
    // Store state securely
    return state, nil
}

// Token storage
func storeTokenEncrypted(userID, accessToken string) error {
    encrypted := encryptToken(accessToken)
    return db.SaveUserToken(userID, encrypted)
}
```
```

### 4. Verification Systems

#### Automated Testing Scripts
```bash
# .ralph/verify/test_oauth.sh
#!/bin/bash

echo "🔍 Testing OAuth URL generation..."

# Test OAuth endpoint
response=$(curl -s "http://localhost:8090/auth/soundcloud")

# Check for PKCE parameters
if echo "$response" | grep -q "code_challenge"; then
    echo "✅ PKCE support detected"
    pkce_support=0
else
    echo "❌ Missing PKCE support"
    pkce_support=1
fi

# Check for state parameter
if echo "$response" | grep -q "state="; then
    echo "✅ State parameter present"
    state_param=0
else
    echo "❌ Missing state parameter"
    state_param=1
fi

# Check for proper redirect URI
if echo "$response" | grep -q "redirect_uri=http://localhost:8090"; then
    echo "✅ Correct redirect URI"
    redirect_uri=0
else
    echo "❌ Incorrect redirect URI"
    redirect_uri=1
fi

# Overall result
if [ $pkce_support -eq 0 ] && [ $state_param -eq 0 ] && [ $redirect_uri -eq 0 ]; then
    echo "✅ OAuth URL generation working correctly"
    exit 0
else
    echo "❌ OAuth URL generation has issues"
    exit 1
fi
```

### 5. Progress Tracking

#### Session Management
```yaml
# .ralph/progress/session_history.yaml
sessions:
  - id: "session_001"
    task: "soundcloud-oauth-implementation"
    start_time: "2025-01-25T15:30:00Z"
    end_time: "2025-01-25T16:00:00Z"
    iterations: 1
    result: "failed"
    final_cost: 0.50
    
  - id: "session_002" 
    task: "soundcloud-oauth-implementation"
    start_time: "2025-01-26T09:00:00Z"
    end_time: "2025-01-26T14:30:00Z"
    iterations: 3
    result: "completed"
    final_cost: 2.25
```

## OpenCode Integration Patterns

### Orchestration Architecture

The Ralph system uses a **two-tier agent architecture**:

1. **Ralph Orchestrator (Main Agent)**
   - Runs in user's TUI session
   - Loads Ralph configuration and manages state
   - Spawns focused sub-agents for specific tasks
   - Runs verification and decides loop continuation

2. **Sub-Agent Workers**
   - Fresh context windows for focused tasks
   - Execute specific PRD items
   - Report results back to orchestrator

### Main Agent Instructions

```markdown
## Ralph Orchestrator Instructions

You are the Ralph Orchestrator agent. Your role is to coordinate autonomous development loops using sub-agents.

### Your Workflow:
1. Load Ralph configuration from `.ralph/` directory
2. Identify next actionable task (considering dependencies)
3. Spawn sub-agent using the `task` tool with focused prompt
4. Run verification script after sub-agent completion
5. Update progress and decide whether to continue or complete

### Sub-Agent Spawning Pattern:
```javascript
task({
  subagent_type: "build",
  prompt: "Implement [specific task description]. Focus only on this component.",
  session_id: "task-name-iteration-X",
  command: "ralph-execute-task oauth_implementation"
})
```

### User Commands:
- `/ralph-status` - Show current progress and task status
- `/ralph-pause` - Pause autonomous loop
- `/ralph-resume` - Resume paused loop
- `/ralph-hint "guidance"` - Add specific feedback for next iteration

### Cost Control:
- Track total cost across all sub-agent sessions
- Respect per-task and overall budget limits
- Abort loops approaching cost limits

### Verification Integration:
- Run verification scripts after each sub-agent completion
- Parse JSON output for structured feedback
- Inject specific feedback into next iteration if failed
```

### Sub-Agent Instructions

```markdown
## Ralph Sub-Agent Instructions

You are a Ralph sub-agent focused on a specific task implementation.

### Your Constraints:
- Fresh context window - no previous iteration knowledge
- Complete only the assigned task component
- Respect iteration limits and cost controls
- Report progress clearly back to orchestrator

### Your Workflow:
1. Read task description and any provided feedback
2. Implement the specific component
3. Test your implementation thoroughly
4. Report completion status and any issues

### Success Reporting:
End with clear status summary:
```
✅ Task Component: OAuth URL Generation
📝 Changes Made: Added /auth/soundcloud route with PKCE
🔍 Testing: Manual testing shows proper URL generation
⚠️ Issues: Token exchange not yet implemented
```
```

### Usage Examples

#### Start Ralph Loop (Main Agent)
```bash
opencode

# In TUI:
"Start Ralph autonomous development for Soundcloud MVP"
```

#### Manual Sub-Agent Execution (for testing)
```bash
opencode ralph-autonomous-loops "Implement OAuth endpoints only" \
    --max-iterations 5 \
    --cost-limit 2.00
```

#### Status Monitoring in TUI
```bash
ralph-status
ralph-pause
ralph-resume
```

## Best Practices

### Success Criteria Design
1. **Specific & Testable**: Each criterion should be binary and verifiable
2. **Incremental**: Build on previous iterations
3. **Comprehensive**: Cover all aspects of the task
4. **Atomic**: Each iteration should make measurable progress

### Context Management
1. **File-based State**: Use `.ralph/context/` for persistent data
2. **Git Integration**: Commit after successful iterations
3. **Rollback Support**: Save checkpoints for recovery
4. **Documentation**: Maintain iteration logs for analysis

### Safety Controls
1. **Iteration Limits**: Always set `--max-iterations`
2. **Cost Monitoring**: Use `--cost-limit` for budget control
3. **Manual Override**: Human intervention commands (`--stop`, `--pause`)
4. **Verification Gates**: Multiple validation checks per iteration

## Troubleshooting

### Common Issues
1. **Infinite Loops**: Missing or incorrect completion criteria
2. **Context Loss**: Not maintaining state between iterations
3. **Cost Overruns**: No iteration limits set
4. **Stuck Agents**: Poor success criteria or feedback loops

### Recovery Strategies
1. **Git Checkpoints**: Rollback to last working state
2. **Context Injection**: Provide specific guidance for stuck agents
3. **Manual Intervention**: Always include override capabilities
4. **Incremental Backoff**: Reduce complexity when stuck

## Example: Complete Soundcloud Integration

### Task Definition
```yaml
task:
  name: "complete-soundcloud-integration"
  description: "Full Soundcloud OAuth integration with feed aggregation"
  priority: "critical"
  
success_criteria:
  - name: "OAuth Flow"
    description: "Complete OAuth 2.1 implementation with PKCE"
    verification: "oauth_integration_test"
    
  - name: "Track Management"
    description: "Fetch and store Soundcloud tracks"
    verification: "track_storage_test"
    
  - name: "Feed Generation"
    description: "Generate RSS feeds from aggregated tracks"
    verification: "feed_generation_test"
    
completion_promise: "All Soundcloud features implemented and tested with comprehensive error handling"
```

### Implementation Plan
```yaml
iterations:
  1: "OAuth URL generation with PKCE"
  2: "OAuth callback implementation"
  3: "Token exchange and storage"
  4: "Track fetching API client"
  5: "Stream display template"
  6: "Basic filtering functionality"
  7: "RSS feed generation"
  8: "Error handling and testing"
```

---

This skill provides the foundation for implementing Ralph Wiggum autonomous loops with OpenCode, enabling AI agents to work continuously on complex tasks with self-correction, persistent context, and systematic progress tracking.