#!/bin/bash
# Remove Cloudflare tunnel service and configuration

set -e

echo "🛑 Stopping Cloudflare tunnel..."

# Try to stop via systemctl first (may require sudo)
if systemctl is-active --quiet cloudflared 2>/dev/null || systemctl is-active --quiet cloudflared.service 2>/dev/null; then
    echo "Found systemd service, stopping..."
    sudo systemctl stop cloudflared || sudo systemctl stop cloudflared.service || true
    sudo systemctl disable cloudflared 2>/dev/null || sudo systemctl disable cloudflared.service 2>/dev/null || true
    echo "✅ Service stopped and disabled"
else
    echo "No systemd service found, killing process..."
    # Kill any running cloudflared processes
    sudo pkill -f "cloudflared.*tunnel" 2>/dev/null || true
    sudo pkill -9 -f cloudflared 2>/dev/null || true
    echo "✅ Processes killed"
fi

# Remove systemd service file if it exists
if [ -f /etc/systemd/system/cloudflared.service ]; then
    echo "Removing systemd service file..."
    sudo rm -f /etc/systemd/system/cloudflared.service
    sudo systemctl daemon-reload 2>/dev/null || true
    echo "✅ Service file removed"
fi

# Remove cloudflared configuration
echo "🗑️  Removing Cloudflare configuration..."
if [ -d /etc/cloudflared ]; then
    sudo rm -rf /etc/cloudflared
    echo "✅ Removed /etc/cloudflared"
fi

if [ -d ~/.cloudflared ]; then
    rm -rf ~/.cloudflared
    echo "✅ Removed ~/.cloudflared"
fi

# Optionally uninstall cloudflared binary
echo ""
echo "📦 Cloudflare tunnel removed!"
echo ""
echo "To completely uninstall cloudflared binary:"
echo "  sudo apt remove cloudflared    # Debian/Ubuntu"
echo "  sudo yum remove cloudflared    # RHEL/CentOS"
echo "  brew uninstall cloudflared     # macOS"
echo ""
echo "Current status:"
ps aux | grep -E "cloudflared|tunnel" | grep -v grep || echo "✅ No cloudflared processes running"
