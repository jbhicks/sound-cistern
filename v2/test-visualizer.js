#!/usr/bin/env node
/**
 * Test script to verify Butterchurn visualizer is rendering correctly
 * 
 * Usage: node test-visualizer.js
 * 
 * This script:
 * 1. Opens the Sound Cistern app
 * 2. Plays a track
 * 3. Opens the fullscreen visualizer
 * 4. Checks that pixels are actually changing (not black)
 * 5. Takes screenshots at intervals to verify animation
 */

const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

const BASE_URL = process.env.TEST_URL || 'http://localhost:8090';
const SCREENSHOT_DIR = path.join(__dirname, 'test-screenshots');

// Ensure screenshot directory exists
if (!fs.existsSync(SCREENSHOT_DIR)) {
  fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
}

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function runTest() {
  console.log('🎵 Starting Butterchurn Visualizer Test\n');
  
  const browser = await chromium.launch({ 
    headless: false,
    args: ['--disable-web-security'] // Allow CORS for testing
  });
  
  const context = await browser.newContext({
    viewport: { width: 1280, height: 900 }
  });
  
  const page = await context.newPage();
  
  try {
    // 1. Navigate to app
    console.log('1. Navigating to app...');
    await page.goto(`${BASE_URL}/stream`, { waitUntil: 'networkidle' });
    await sleep(2000);
    
    // 2. Click play on first track
    console.log('2. Starting playback...');
    const playButton = await page.locator('button').filter({ hasText: 'Play' }).first();
    if (await playButton.isVisible().catch(() => false)) {
      await playButton.click();
    } else {
      // Try clicking the track card directly
      const firstTrack = await page.locator('text=Midnight Drive').first();
      await firstTrack.click();
    }
    await sleep(2000);
    
    // 3. Open fullscreen visualizer by clicking artwork
    console.log('3. Opening fullscreen visualizer...');
    const artwork = await page.locator('[title="Open visualizer"]').first();
    await artwork.click();
    await sleep(3000); // Wait for animation and visualizer to initialize
    
    // 4. Test: Check if canvas is rendering (not black)
    console.log('4. Testing visualizer rendering...');
    
    let passedTests = 0;
    let failedTests = 0;
    
    // Test 1: Check canvas exists and has WebGL context
    const canvasInfo = await page.evaluate(() => {
      const canvas = document.querySelector('.fixed.top-14 canvas');
      if (!canvas) return { error: 'Canvas not found' };
      
      const gl = canvas.getContext('webgl2') || canvas.getContext('webgl');
      if (!gl) return { error: 'No WebGL context' };
      
      return {
        width: canvas.width,
        height: canvas.height,
        contextLost: gl.isContextLost(),
        framebufferBound: !!gl.getParameter(gl.FRAMEBUFFER_BINDING)
      };
    });
    
    if (canvasInfo.error) {
      console.log(`   ❌ Test 1 FAILED: ${canvasInfo.error}`);
      failedTests++;
    } else {
      console.log(`   ✓ Test 1 PASSED: Canvas exists (${canvasInfo.width}x${canvasInfo.height})`);
      console.log(`      - Context lost: ${canvasInfo.contextLost}`);
      console.log(`      - Framebuffer bound: ${canvasInfo.framebufferBound}`);
      passedTests++;
    }
    
    // Test 2: Check if pixels are changing over time (animation test)
    console.log('\n5. Testing animation (checking if pixels change)...');
    
    const samples = [];
    for (let i = 0; i < 5; i++) {
      await sleep(500); // Wait 500ms between samples
      
      const pixelData = await page.evaluate(() => {
        const canvas = document.querySelector('.fixed.top-14 canvas');
        if (!canvas) return null;
        
        const gl = canvas.getContext('webgl2') || canvas.getContext('webgl');
        if (!gl) return null;
        
        // Read center pixel
        const pixels = new Uint8Array(4);
        gl.readPixels(
          Math.floor(canvas.width / 2), 
          Math.floor(canvas.height / 2), 
          1, 1, 
          gl.RGBA, 
          gl.UNSIGNED_BYTE, 
          pixels
        );
        
        return {
          r: pixels[0],
          g: pixels[1],
          b: pixels[2],
          a: pixels[3],
          sum: pixels[0] + pixels[1] + pixels[2] + pixels[3]
        };
      });
      
      if (pixelData) {
        samples.push(pixelData);
        console.log(`   Sample ${i + 1}: RGBA(${pixelData.r}, ${pixelData.g}, ${pixelData.b}, ${pixelData.a}) - Sum: ${pixelData.sum}`);
      }
      
      // Take screenshot
      await page.screenshot({ 
        path: path.join(SCREENSHOT_DIR, `viz-frame-${i + 1}.png`),
        fullPage: false 
      });
    }
    
    // Check if any samples have non-zero pixels
    const hasVisiblePixels = samples.some(s => s.sum > 0);
    const pixelsChanging = samples.some((s, i) => {
      if (i === 0) return false;
      return s.r !== samples[i-1].r || 
             s.g !== samples[i-1].g || 
             s.b !== samples[i-1].b;
    });
    
    if (hasVisiblePixels) {
      console.log('   ✓ Test 2 PASSED: Canvas has visible pixels (not all black)');
      passedTests++;
    } else {
      console.log('   ❌ Test 2 FAILED: Canvas is completely black (all pixels are 0)');
      failedTests++;
    }
    
    if (pixelsChanging) {
      console.log('   ✓ Test 3 PASSED: Pixels are changing (animation is working)');
      passedTests++;
    } else {
      console.log('   ❌ Test 3 FAILED: Pixels are not changing (animation frozen)');
      failedTests++;
    }
    
    // Test 4: Check console for render loop messages
    console.log('\n6. Checking console logs...');
    const logs = await page.evaluate(() => {
      // This won't work for console logs, but we can check window variables
      return {
        vizStats: window._vizStats,
        butterchurnAudio: !!window._butterchurnAudio
      };
    });
    
    if (logs.vizStats) {
      console.log(`   ✓ Test 4 PASSED: Visualizer stats available`);
      console.log(`      - FPS: ${logs.vizStats.fps}`);
      console.log(`      - Frame time: ${logs.vizStats.frameMs}ms`);
      console.log(`      - Resolution: ${logs.vizStats.resW}x${logs.vizStats.resH}`);
      passedTests++;
    } else {
      console.log('   ⚠ Test 4 WARNING: No viz stats available yet');
    }
    
    // Test 5: Let it run for a bit and check again
    console.log('\n7. Running for 5 seconds to check for freezing...');
    await sleep(5000);
    
    const finalPixelData = await page.evaluate(() => {
      const canvas = document.querySelector('.fixed.top-14 canvas');
      if (!canvas) return null;
      
      const gl = canvas.getContext('webgl2') || canvas.getContext('webgl');
      if (!gl) return null;
      
      const pixels = new Uint8Array(4);
      gl.readPixels(
        Math.floor(canvas.width / 2), 
        Math.floor(canvas.height / 2), 
        1, 1, 
        gl.RGBA, 
        gl.UNSIGNED_BYTE, 
        pixels
      );
      
      return {
        r: pixels[0],
        g: pixels[1],
        b: pixels[2],
        a: pixels[3],
        sum: pixels[0] + pixels[1] + pixels[2] + pixels[3]
      };
    });
    
    if (finalPixelData && finalPixelData.sum > 0) {
      console.log(`   ✓ Test 5 PASSED: Still rendering after 5 seconds`);
      console.log(`      Final pixel: RGBA(${finalPixelData.r}, ${finalPixelData.g}, ${finalPixelData.b}, ${finalPixelData.a})`);
      passedTests++;
    } else {
      console.log('   ❌ Test 5 FAILED: Visualizer stopped/froze after 5 seconds');
      failedTests++;
    }
    
    // Final screenshot
    await page.screenshot({ 
      path: path.join(SCREENSHOT_DIR, 'viz-final.png'),
      fullPage: false 
    });
    
    // Summary
    console.log('\n' + '='.repeat(50));
    console.log('TEST SUMMARY');
    console.log('='.repeat(50));
    console.log(`Passed: ${passedTests}`);
    console.log(`Failed: ${failedTests}`);
    console.log(`Total: ${passedTests + failedTests}`);
    
    if (failedTests === 0) {
      console.log('\n✅ ALL TESTS PASSED - Visualizer is working correctly!');
    } else {
      console.log('\n❌ SOME TESTS FAILED - Visualizer has issues');
      console.log(`   Screenshots saved to: ${SCREENSHOT_DIR}`);
    }
    
    return failedTests === 0;
    
  } catch (error) {
    console.error('\n❌ Test failed with error:', error.message);
    await page.screenshot({ 
      path: path.join(SCREENSHOT_DIR, 'error.png'),
      fullPage: true 
    });
    return false;
  } finally {
    await browser.close();
  }
}

// Run the test
runTest()
  .then(success => {
    process.exit(success ? 0 : 1);
  })
  .catch(error => {
    console.error('Fatal error:', error);
    process.exit(1);
  });
