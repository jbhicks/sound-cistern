---
name: ralph-commands
description: User control commands for Ralph Wiggum autonomous loops
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: ralph-users
  stack: opencode-ralph-commands
---

# Ralph User Control Commands

This skill provides user intervention commands for managing Ralph Wiggum autonomous loops during execution.

## Available Commands

### `/ralph-status`
Display current Ralph loop progress, task status, and budget information.

**Usage:**
```
/ralph-status
```

**Output:**
```
🚀 Ralph Loop Status
==================
Session: session_001
State: running
Current Task: oauth_implementation (iteration 2)

📊 Task Progress:
• oauth_implementation: in_progress (1/3 success criteria)
• stream_display: pending (dependency: oauth_implementation)
• track_metadata: pending (dependency: stream_display)

💰 Budget Status:
Total: $15.00 allocated, $0.00 spent
OAuth Implementation: $4.50 allocated, $0.00 spent

⏱️  Runtime: 00:15:32
```

### `/ralph-pause`
Temporarily halt the autonomous loop for user intervention.

**Usage:**
```
/ralph-pause
```

**Effect:**
- Stops current sub-agent execution
- Prevents new sub-agents from spawning
- Preserves all current state
- Allows user to examine and modify code

### `/ralph-resume`
Continue the autonomous loop from where it was paused.

**Usage:**
```
/ralph-resume
```

**Effect:**
- Resumes autonomous loop execution
- Continues from paused state
- Maintains all context and progress

### `/ralph-hint "specific guidance"`
Add targeted feedback for the next iteration of a specific task.

**Usage:**
```
/ralph-hint "Focus on implementing PKCE code challenge generation using crypto/sha256"
```

**Effect:**
- Adds specific guidance to iteration feedback
- Provided to next sub-agent execution
- Escalates hint level for that task

### `/ralph-cost-limit [amount]`
Update the total cost limit for the Ralph loop.

**Usage:**
```
/ralph-cost-limit 20.00
```

**Effect:**
- Updates total budget allocation
- Validates against current spend
- Prevents overspending

### `/ralph-task-status [task_name]`
Show detailed status for a specific task.

**Usage:**
```
/ralph-task-status oauth_implementation
```

**Output:**
```
📋 Task: oauth_implementation
Status: in_progress
Iterations: 2/8 allocated
Last Verification: Failed (25% score)

Failed Checks:
- oauth_url_missing
- pkce_missing

Next Actions:
- Add Soundcloud OAuth URL generation
- Implement PKCE code challenge

User Hints:
- "Use crypto/rand for secure random generation"
```

### `/ralph-complete [task_name]`
Manually mark a task as complete (override verification).

**Usage:**
```
/ralph-complete oauth_implementation
```

**Effect:**
- Bypasses verification for that task
- Marks task as completed
- Allows dependent tasks to proceed
- Records manual completion in audit log

### `/ralph-reset [task_name]`
Reset a task back to pending state.

**Usage:**
```
/ralph-reset oauth_implementation
```

**Effect:**
- Clears all iteration history
- Resets progress to 0
- Useful for starting over on problematic tasks

### `/ralph-rollback [checkpoint]`
Roll back to a specific git checkpoint.

**Usage:**
```
/ralph-rollback
# or
/ralph-rollback abc1234
```

**Effect:**
- Restores git state to checkpoint
- Recovers from broken implementations
- Preserves Ralph progress tracking

## Integration with Ralph Orchestrator

These commands are designed to be used with the Ralph Orchestrator agent. When a user types a command, the orchestrator:

1. **Parses the command** and extracts parameters
2. **Updates shared state** using the context manager
3. **Modifies loop behavior** based on the command
4. **Provides feedback** to the user about the action taken

## Examples

### Example 1: Checking Status During Development
```
User: /ralph-status

Ralph: 🚀 Ralph Loop Status
==================
Session: session_001  
State: running
Current Task: oauth_implementation (iteration 3)

📊 Task Progress:
• oauth_implementation: in_progress (2/3 success criteria)
• stream_display: pending  
• track_metadata: pending

💰 Budget Status:
Total: $15.00 allocated, $2.25 spent
```

### Example 2: Providing Targeted Hint
```
User: /ralph-hint "Make sure to use the existing PocketBase migration patterns in pb_migrations/"

Ralph: 💡 Hint added for oauth_implementation
📝 Will include this guidance in next sub-agent iteration
🔄 Hint level escalated to 1
```

### Example 3: Emergency Pause
```
User: /ralph-pause

Ralph: ⏸️ Ralph loop paused
🛑 Current sub-agent will complete current operation
📋 No new sub-agents will be spawned
💬 Use /ralph-resume to continue
```

## Safety Features

### Confirmation for Destructive Actions
Commands like `/ralph-complete` and `/ralph-rollback` require confirmation:
```
Ralph: ⚠️ This will mark oauth_implementation as complete without verification.
Are you sure? [y/N]
```

### Audit Logging
All user interventions are logged:
```json
{
  "timestamp": "2025-01-25T16:30:00Z",
  "command": "ralph-hint",
  "parameters": "\"Focus on PKCE implementation\"",
  "user": "session_001"
}
```

### Budget Protection
Cost limit changes are validated against current spend:
```
Ralph: ❌ Cannot reduce cost limit below current spend ($2.25)
Minimum allowed: $2.25
```

---

These commands give users fine-grained control over the autonomous development process while maintaining the benefits of the Ralph Wiggum methodology.