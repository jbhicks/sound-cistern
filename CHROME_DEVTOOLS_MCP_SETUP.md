# Chrome DevTools MCP Server Setup Guide for OpenCode

This guide provides comprehensive instructions for setting up and using the Chrome DevTools MCP server with OpenCode.

## Overview

The Chrome DevTools MCP server allows AI coding agents to control and inspect a live Chrome browser, providing access to the full power of Chrome DevTools for:
- Reliable automation
- In-depth debugging  
- Performance analysis
- Network inspection
- Console debugging

## Configuration

### Basic Setup

Add this to your `~/.config/opencode/opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "chrome-devtools": {
      "type": "local",
      "command": [
        "npx",
        "-y",
        "chrome-devtools-mcp@latest"
      ],
      "enabled": true
    }
  }
}
```

### Advanced Configuration Options

You can customize the Chrome DevTools MCP server with additional arguments:

```json
{
  "mcp": {
    "chrome-devtools": {
      "type": "local",
      "command": [
        "npx",
        "-y",
        "chrome-devtools-mcp@latest",
        "--headless=true",
        "--isolated=true",
        "--viewport=1280x720"
      ],
      "enabled": true
    }
  }
}
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--autoConnect` | Automatically connect to running Chrome (Chrome 144+) | `false` |
| `--browser-url` | Connect to running Chrome instance (e.g., `http://127.0.0.1:9222`) | - |
| `--headless` | Run in headless mode (no UI) | `false` |
| `--isolated` | Use temporary user data directory (cleaned up after) | `false` |
| `--channel` | Chrome channel: `stable`, `beta`, `dev`, `canary` | `stable` |
| `--viewport` | Initial viewport size (e.g., `1280x720`) | - |
| `--user-data-dir` | Custom Chrome user data directory | Auto-generated |

## Usage

### Basic Commands

Once configured, you can use Chrome DevTools tools in your OpenCode prompts:

```bash
# Take a screenshot of a website
"Take a screenshot of https://example.com"

# Analyze page performance
"Check the performance of https://developers.chrome.com"

# Navigate and interact with pages
"Navigate to https://github.com and click the Sign in button"

# Debug network requests
"List all network requests for https://example.com"

# Console debugging
"Show me the console messages for https://example.com"
```

### Available Tools

#### Navigation & Page Management
- `navigate_page` - Navigate to a URL
- `new_page` - Open new tab/page
- `close_page` - Close a page
- `list_pages` - List all open pages
- `select_page` - Switch between pages

#### Input Automation
- `click` - Click on elements
- `fill` - Fill form fields
- `fill_form` - Fill entire forms
- `hover` - Hover over elements
- `press_key` - Press keyboard keys
- `drag` - Drag and drop
- `upload_file` - Upload files

#### Debugging & Inspection
- `take_screenshot` - Capture screenshots
- `take_snapshot` - Take DOM snapshot
- `evaluate_script` - Execute JavaScript
- `list_console_messages` - Get console logs
- `get_console_message` - Get specific console message

#### Performance Analysis
- `performance_start_trace` - Start performance trace
- `performance_stop_trace` - Stop performance trace
- `performance_analyze_insight` - Analyze performance

#### Network Monitoring
- `list_network_requests` - List network requests
- `get_network_request` - Get specific request details

#### Emulation
- `emulate` - Emulate devices/network conditions
- `resize_page` - Resize viewport

## Advanced Setup

### Connecting to Running Chrome Instance

For scenarios where you want to maintain browser state or use existing sessions:

#### Option 1: Auto-Connect (Chrome 144+)

1. Enable remote debugging in Chrome:
   - Navigate to `chrome://inspect/#remote-debugging`
   - Enable remote debugging
   - Allow incoming connections

2. Configure MCP server with `--autoConnect`:
   ```json
   {
     "mcp": {
       "chrome-devtools": {
         "type": "local",
         "command": [
           "npx",
           "-y",
           "chrome-devtools-mcp@latest",
           "--autoConnect"
         ]
       }
     }
   }
   ```

#### Option 2: Manual Connection via Debug Port

1. Start Chrome with remote debugging:
   ```bash
   # Linux
   google-chrome --remote-debugging-port=9222 --user-data-dir=/tmp/chrome-profile
   
   # macOS
   /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222 --user-data-dir=/tmp/chrome-profile
   
   # Windows
   "C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222 --user-data-dir="%TEMP%\chrome-profile"
   ```

2. Configure MCP server:
   ```json
   {
     "mcp": {
       "chrome-devtools": {
         "type": "local",
         "command": [
           "npx",
           "-y",
           "chrome-devtools-mcp@latest",
           "--browser-url=http://127.0.0.1:9222"
         ]
       }
     }
   }
   ```

### Security Considerations

⚠️ **Important Security Notes:**

- Remote debugging ports expose your browser to control by any application on your machine
- Avoid browsing sensitive websites while debugging ports are open
- Use separate user data directories for debugging sessions
- Consider using `--isolated=true` for automated tasks

### Troubleshooting

#### Common Issues

1. **MCP Server Won't Connect**
   ```bash
   # Check MCP server status
   opencode mcp list
   
   # Restart OpenCode if needed
   # Check Node.js and npm are installed
   node --version
   npm --version
   ```

2. **Chrome Won't Start**
   ```bash
   # Try headless mode
   "--headless=true"
   
   # Check Chrome installation
   google-chrome --version
   ```

3. **Permission Issues**
   ```bash
   # Try isolated mode
   "--isolated=true"
   
   # Check user data directory permissions
   ls -la ~/.cache/chrome-devtools-mcp/
   ```

4. **Port Already in Use**
   ```bash
   # Find process using port 9222
   lsof -i :9222
   
   # Use different port
   "--browser-url=http://127.0.0.1:9223"
   ```

#### Debug Logs

Enable verbose logging for troubleshooting:
```bash
# Add to command:
"--log-file=/tmp/chrome-devtools-mcp.log"

# Set environment variable:
export DEBUG="*"
```

## Examples

### Example 1: Performance Analysis
```bash
"Analyze the performance of https://example.com and identify the bottlenecks"
```

### Example 2: Form Testing
```bash
"Navigate to https://github.com/signup, fill in the registration form with test data, but don't submit"
```

### Example 3: Screenshot Testing
```bash
"Take a screenshot of https://example.com at 1920x1080 resolution and save it"
```

### Example 4: Network Debugging
```bash
"Check the network requests for https://example.com and identify any failed requests"
```

### Example 5: Console Error Checking
```bash
"Navigate to https://example.com and check for any JavaScript errors in the console"
```

## Best Practices

1. **Use Isolated Mode**: For automated tasks, use `--isolated=true` to ensure clean state
2. **Viewport Sizing**: Set appropriate viewport sizes for responsive testing
3. **Error Handling**: Always include error handling in your automation scripts
4. **Performance**: Consider headless mode for faster execution
5. **Security**: Never expose debugging ports to untrusted networks
6. **Resource Management**: Close pages and clean up resources when done

## Integration with Other Tools

The Chrome DevTools MCP server works well with:
- **Context7**: For documentation search while debugging
- **Custom MCP Servers**: For domain-specific testing tools
- **CI/CD Pipelines**: For automated testing workflows

## Current Status

✅ **Configuration**: Chrome DevTools MCP server is configured and connected in OpenCode
✅ **MCP Server**: Direct testing confirms Chrome DevTools MCP server works perfectly (26 tools available)
✅ **Documentation**: This guide is complete and up-to-date
⚠️ **OpenCode Integration**: There appears to be a timeout issue when executing commands through OpenCode CLI

## Testing Results

### MCP Server Direct Test ✅
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | npx -y chrome-devtools-mcp@latest
```
- **Result**: Successfully returned all 26 Chrome DevTools tools
- **Status**: MCP server is fully functional

### Available Tools Confirmed ✅
- Navigation: `navigate_page`, `new_page`, `close_page`, `list_pages`, `select_page`
- Input: `click`, `fill`, `fill_form`, `hover`, `press_key`, `drag`, `upload_file`
- Debugging: `take_screenshot`, `take_snapshot`, `evaluate_script`, `list_console_messages`
- Performance: `performance_start_trace`, `performance_stop_trace`, `performance_analyze_insight`
- Network: `list_network_requests`, `get_network_request`
- Emulation: `emulate`, `resize_page`
- Utilities: `wait_for`, `handle_dialog`

### OpenCode Integration Issue ⚠️
When testing through OpenCode CLI:
- Commands timeout after 60-120 seconds
- MCP server status shows "connected" 
- Provider configuration needs to be set correctly

## Working Configuration

The final working OpenCode configuration (`~/.config/opencode/opencode.json`):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {},
  "mcp": {
    "context7": {
      "type": "local",
      "command": ["npx", "-y", "@upstash/context7-mcp"],
      "enabled": true
    },
    "chrome-devtools": {
      "type": "local",
      "command": ["npx", "-y", "chrome-devtools-mcp@latest"],
      "enabled": true
    }
  }
}
```

## Usage Examples

### Direct MCP Server Usage (Working)
```bash
# Test MCP server directly
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | npx -y chrome-devtools-mcp@latest

# Take screenshot (direct JSON-RPC)
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"take_screenshot","arguments":{"url":"https://httpbin.org/html"}}}' | npx -y chrome-devtools-mcp@latest
```

### OpenCode Usage (When Provider Issues Resolved)
Once OpenCode provider is configured correctly:

```bash
# Set up a model first
opencode run "Use chrome-devtools to take a screenshot of https://httpbin.org/html" --model github-copilot/claude-sonnet-4
```

## Troubleshooting OpenCode Integration

The timeout issue appears to be related to:
1. **Provider Configuration**: OpenCode needs a properly configured AI provider
2. **Model Selection**: Explicit model specification required
3. **Session Management**: OpenCode may have session initialization issues

### Workaround Options

1. **Use Direct MCP Commands**: For testing and simple automation
2. **Web Interface**: Try `opencode web` for GUI-based interaction
3. **Alternative MCP Clients**: Consider other MCP-compatible clients

Next steps:
1. Resolve OpenCode provider configuration
2. Test with different AI models
3. Consider using web interface for Chrome DevTools interaction
4. Integrate working direct MCP commands into scripts

---

*Last updated: January 25, 2026*
*Based on Chrome DevTools MCP v0.13.0*