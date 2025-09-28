#!/bin/bash
# Add Go binaries to PATH if not already present
if [[ ":$PATH:" != *":/root/go/bin:"* ]]; then
    echo 'export PATH=$PATH:/root/go/bin' >> ~/.bashrc
    echo 'export PATH=$PATH:/root/go/bin' >> ~/.zshrc
    echo "✅ Added Go bin directory to PATH"
    echo "🔄 Please restart your shell or run: source ~/.bashrc"
fi
