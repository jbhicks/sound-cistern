#!/bin/bash
set -euo pipefail

OUT_DIR="${1:-/tmp/browser-diagnostics-$(date +%Y%m%d-%H%M%S)}"
mkdir -p "$OUT_DIR"

exec > >(tee "$OUT_DIR/report.txt") 2>&1

section() {
  printf '\n=== %s ===\n' "$1"
}

save() {
  local name="$1"
  shift
  "$@" > "$OUT_DIR/$name" 2>&1 || true
}

echo "Browser diagnostics report"
echo "Output: $OUT_DIR"

section "system"
uptime
free -h
uname -a

section "chrome processes"
save chrome-processes.txt ps -eo pid,ppid,%cpu,%mem,stat,etime,cmd --sort=-%cpu
grep -Ei 'chrome|chromium' "$OUT_DIR/chrome-processes.txt" | head -n 80 || true

section "chrome tree"
save chrome-tree.txt ps -eo pid,ppid,%cpu,%mem,stat,etime,cmd --forest
grep -A4 -B2 -Ei 'google-chrome|chromium|chrome-devtools-mcp' "$OUT_DIR/chrome-tree.txt" | head -n 120 || true

section "chrome tab domains from session files"
SESSION_DIR="$HOME/.config/google-chrome/Default/Sessions"
if [[ -d "$SESSION_DIR" ]]; then
  strings "$SESSION_DIR"/Session_* "$SESSION_DIR"/Tabs_* 2>/dev/null \
    | grep -Eo 'https?://[^ ]+' \
    | sed 's#https\?://##' \
    | cut -d/ -f1 \
    | sort \
    | uniq -c \
    | sort -nr \
    | head -n 50 \
    | tee "$OUT_DIR/tab-domains.txt"
else
  echo "No session directory found at $SESSION_DIR"
fi

section "chrome session urls"
if [[ -d "$SESSION_DIR" ]]; then
  strings "$SESSION_DIR"/Session_* "$SESSION_DIR"/Tabs_* 2>/dev/null \
    | grep -Eo 'https?://[^ ]+' \
    | sort -u \
    | tee "$OUT_DIR/tab-urls.txt"
fi

section "chrome crash reports"
CRASH_DIR="$HOME/.config/google-chrome/Crash Reports"
if [[ -d "$CRASH_DIR" ]]; then
  find "$CRASH_DIR" -maxdepth 2 -type f | sort | tee "$OUT_DIR/crash-files.txt"
else
  echo "No crash report directory found at $CRASH_DIR"
fi

section "journal hints"
save journal-hints.txt journalctl -b --no-pager
grep -Ei 'oom|killed process|segfault|chrome|chromium|gpu|amdgpu|crashpad' "$OUT_DIR/journal-hints.txt" | tail -n 200 || true

section "remote debugging"
if curl -fsS http://127.0.0.1:9222/json/version >/dev/null 2>&1; then
  curl -fsS http://127.0.0.1:9222/json/version | tee "$OUT_DIR/remote-version.json"
  curl -fsS http://127.0.0.1:9222/json/list | tee "$OUT_DIR/remote-tabs.json"
else
  echo "No debugger on 127.0.0.1:9222"
fi

section "hyprland clients"
if command -v hyprctl >/dev/null 2>&1; then
  hyprctl clients | tee "$OUT_DIR/hypr-clients.txt"
  grep -A6 -B2 -i 'chrome\|chromium' "$OUT_DIR/hypr-clients.txt" || true
else
  echo "hyprctl not available"
fi

section "summary"
echo "Report saved to: $OUT_DIR"
echo "Top domains:"
head -n 20 "$OUT_DIR/tab-domains.txt" 2>/dev/null || true
echo "\nTop chrome processes:"
grep -Ei 'chrome|chromium' "$OUT_DIR/chrome-processes.txt" | head -n 20 || true
echo "\nCrash reports:"
cat "$OUT_DIR/crash-files.txt" 2>/dev/null || true