#!/bin/bash
# Quick verification that the butterchurn visualizer fix is in place

echo "Checking Butterchurn Visualizer Fix..."
echo "======================================"

# Check if the fix is in the code
if grep -q "Don't check for existing WebGL context here" v2/src/components/ButterchurnVisualizer.jsx; then
    echo "✓ Fix is present in source code"
else
    echo "✗ Fix NOT found in source code"
    exit 1
fi

# Check that the problematic line was removed
if grep -q "const existingGL = canvas.getContext('webgl2')" v2/src/components/ButterchurnVisualizer.jsx; then
    echo "✗ Old problematic code still present"
    exit 1
else
    echo "✓ Problematic code removed"
fi

echo ""
echo "Fix Summary:"
echo "- Removed canvas.getContext('webgl2') check that was preventing 2D context creation"
echo "- Butterchurn can now successfully get 2D context on the visible canvas"
echo ""
echo "Next step: Rebuild the frontend to apply changes"
echo "Run: make v2-build"
