#!/bin/bash
set -euo pipefail

OUT_DIR="${1:-/tmp/home-assistant-diagnostics-$(date +%Y%m%d-%H%M%S)}"
mkdir -p "$OUT_DIR"

exec > >(tee "$OUT_DIR/report.txt") 2>&1

section() {
  printf '\n=== %s ===\n' "$1"
}

echo "Home Assistant diagnostics report"
echo "Output: $OUT_DIR"

section "local backend"
if curl -fsSI http://127.0.0.1:8123/ >/tmp/ha_headers.txt 2>/dev/null; then
  cat /tmp/ha_headers.txt
  rm -f /tmp/ha_headers.txt
else
  echo "No response from http://127.0.0.1:8123/"
fi

section "processes"
ps -eo pid,ppid,%cpu,%mem,stat,etime,cmd --sort=-%cpu | grep -Ei 'homeassistant|home assistant|hass|supervisor|8123|chrome|chromium' | head -n 80 || true

section "docker"
if command -v docker >/dev/null 2>&1; then
  docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' | grep -Ei 'home|hass|assistant|8123' || echo "No matching docker containers"
else
  echo "docker not installed"
fi

section "systemd"
systemctl --user --no-pager --all list-units | grep -Ei 'home|hass|assistant' || true
systemctl --no-pager --all list-units | grep -Ei 'home|hass|assistant' || true

section "chrome session urls for HA"
SESSION_DIR="$HOME/.config/google-chrome/Default/Sessions"
if [[ -d "$SESSION_DIR" ]]; then
  strings "$SESSION_DIR"/Session_* "$SESSION_DIR"/Tabs_* 2>/dev/null \
    | grep -Eo 'https?://[^ ]+' \
    | grep -Ei '8123|home assistant|hass|lovelace|hacs|media-browser|onboarding' \
    | sort -u \
    | tee "$OUT_DIR/ha-urls.txt"
else
  echo "No Chrome session directory found"
fi

section "chrome window title"
if command -v hyprctl >/dev/null 2>&1; then
  hyprctl clients 2>/dev/null | grep -A12 -B2 -i 'google chrome\|chromium' || true
else
  echo "hyprctl not available"
fi

section "summary"
if curl -fsSI http://127.0.0.1:8123/ >/dev/null 2>&1; then
  echo "Home Assistant is reachable locally."
else
  echo "Home Assistant is not reachable on 127.0.0.1:8123."
fi
