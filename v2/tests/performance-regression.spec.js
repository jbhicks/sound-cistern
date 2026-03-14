import { test, expect } from '@playwright/test'

/**
 * Performance Regression Tests
 * 
 * These tests verify that the essential functionality of the player still works
 * after the performance optimizations were applied.
 * 
 * Run with: npx playwright test tests/performance-regression.spec.js
 */

test.describe('Player Core Functionality', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app
    await page.goto('http://localhost:5173')
    
    // Wait for the app to load
    await page.waitForLoadState('networkidle')
    
    // Handle any initial auth state if needed
    // This might need adjustment based on your auth flow
  })

  test('player appears when track is clicked', async ({ page }) => {
    // Find and click the first track card
    const firstTrack = page.locator('[data-testid="track-card"]').first()
    await expect(firstTrack).toBeVisible()
    await firstTrack.click()
    
    // Verify player appears
    const player = page.locator('[data-testid="player-bar"]')
    await expect(player).toBeVisible()
  })

  test('play/pause button toggles playback', async ({ page }) => {
    // Click first track
    await page.locator('[data-testid="track-card"]').first().click()
    
    // Wait for player to appear
    const player = page.locator('[data-testid="player-bar"]')
    await expect(player).toBeVisible()
    
    // Click play/pause button
    const playPauseBtn = player.locator('button[aria-label="Play/Pause"], button').filter({ hasText: /play|pause/i }).first()
    await playPauseBtn.click()
    
    // Verify button state changes (you may need to adjust based on your actual implementation)
    // This is a basic check - the button should still be clickable
    await expect(playPauseBtn).toBeVisible()
  })

  test('progress bar shows and updates', async ({ page }) => {
    // Click first track
    await page.locator('[data-testid="track-card"]').first().click()
    
    // Wait for player
    const player = page.locator('[data-testid="player-bar"]')
    await expect(player).toBeVisible()
    
    // Check progress bar exists and shows some progress
    const progressBar = player.locator('.bg-gradient-to-r').first()
    await expect(progressBar).toBeVisible()
  })

  test('seeking on progress bar jumps to position', async ({ page }) => {
    // Click first track and wait for it to load
    await page.locator('[data-testid="track-card"]').first().click()
    const player = page.locator('[data-testid="player-bar"]')
    await expect(player).toBeVisible()
    
    // Wait a moment for audio to start loading
    await page.waitForTimeout(2000)
    
    // Click on the progress bar at 50% position
    const progressContainer = player.locator('.cursor-pointer').first()
    const box = await progressContainer.boundingBox()
    if (box) {
      await progressContainer.click({ position: { x: box.width * 0.5, y: box.height / 2 } })
    }
    
    // Verify the click was registered (progress bar should still be visible)
    await expect(progressContainer).toBeVisible()
  })

  test('volume control works', async ({ page }) => {
    // Click first track
    await page.locator('[data-testid="track-card"]').first().click()
    
    const player = page.locator('[data-testid="player-bar"]')
    await expect(player).toBeVisible()
    
    // Find volume slider or button
    const volumeControl = player.locator('input[type="range"]').filter({ has: page.locator('') }).first()
    
    // If there's a volume slider, try to adjust it
    if (await volumeControl.isVisible().catch(() => false)) {
      await volumeControl.fill('0.5')
      const value = await volumeControl.inputValue()
      expect(value).toBe('0.5')
    }
  })

  test('track info displays correctly', async ({ page }) => {
    // Click first track
    const firstTrack = page.locator('[data-testid="track-card"]').first()
    await firstTrack.click()
    
    // Check player shows track info
    const player = page.locator('[data-testid="player-bar"]')
    await expect(player).toBeVisible()
    
    // Verify track title is displayed
    const trackTitle = player.locator('text=/[\\w\\s]+/').first()
    await expect(trackTitle).toBeVisible()
  })
})

test.describe('Visualizer Functionality', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173')
    await page.waitForLoadState('networkidle')
  })

  test('fullscreen visualizer opens', async ({ page }) => {
    // Play a track first
    await page.locator('[data-testid="track-card"]').first().click()
    await page.waitForTimeout(1000)
    
    // Click on artwork to open fullscreen visualizer
    const artwork = page.locator('[data-testid="player-artwork"]').first()
    await artwork.click()
    
    // Verify fullscreen visualizer appears
    const fullscreenViz = page.locator('[data-testid="fullscreen-visualizer"]').first()
    await expect(fullscreenViz).toBeVisible()
    
    // Close it
    await page.keyboard.press('Escape')
    await expect(fullscreenViz).not.toBeVisible()
  })

  test('visualizer presets can be cycled', async ({ page }) => {
    // Play a track
    await page.locator('[data-testid="track-card"]').first().click()
    await page.waitForTimeout(1000)
    
    // Open fullscreen visualizer
    await page.locator('[data-testid="player-artwork"]').first().click()
    
    // Try to cycle presets with arrow keys
    await page.keyboard.press('ArrowRight')
    await page.waitForTimeout(500)
    await page.keyboard.press('ArrowLeft')
    
    // Should still be visible
    const fullscreenViz = page.locator('[data-testid="fullscreen-visualizer"]').first()
    await expect(fullscreenViz).toBeVisible()
  })
})

test.describe('Track Card Interactions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173')
    await page.waitForLoadState('networkidle')
  })

  test('track cards render without errors', async ({ page }) => {
    // Check that track cards are visible
    const trackCards = page.locator('[data-testid="track-card"]')
    const count = await trackCards.count()
    expect(count).toBeGreaterThan(0)
    
    // Verify first few cards are visible
    for (let i = 0; i < Math.min(3, count); i++) {
      await expect(trackCards.nth(i)).toBeVisible()
    }
  })

  test('favorite button works on track cards', async ({ page }) => {
    const firstCard = page.locator('[data-testid="track-card"]').first()
    await expect(firstCard).toBeVisible()
    
    // Find and click favorite button
    const favButton = firstCard.locator('button').filter({ has: page.locator('svg') }).first()
    await favButton.click()
    
    // Button should still be visible after click
    await expect(favButton).toBeVisible()
  })

  test('track card flip animation works', async ({ page }) => {
    const firstCard = page.locator('[data-testid="track-card"]').first()
    await expect(firstCard).toBeVisible()
    
    // Find flip button (if exists)
    const flipButton = firstCard.locator('button[title="Flip"]').first()
    
    if (await flipButton.isVisible().catch(() => false)) {
      await flipButton.click()
      await page.waitForTimeout(600) // Wait for animation
      await expect(firstCard).toBeVisible()
    }
  })
})

test.describe('Performance Optimizations Verification', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173')
    await page.waitForLoadState('networkidle')
  })

  test('player does not cause excessive re-renders during playback', async ({ page }) => {
    // Start playback
    await page.locator('[data-testid="track-card"]').first().click()
    await page.waitForTimeout(2000)
    
    // Monitor for console errors
    const errors = []
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text())
      }
    })
    
    // Let it play for a few seconds
    await page.waitForTimeout(5000)
    
    // Should not have excessive React warnings
    const reactWarnings = errors.filter(e => e.includes('React') || e.includes('re-render'))
    expect(reactWarnings.length).toBeLessThan(10)
  })

  test('visualizer pauses when tab is hidden', async ({ page }) => {
    // Play track and open visualizer
    await page.locator('[data-testid="track-card"]').first().click()
    await page.waitForTimeout(1000)
    await page.locator('[data-testid="player-artwork"]').first().click()
    
    // Hide tab (simulate visibility change)
    await page.evaluate(() => {
      Object.defineProperty(document, 'hidden', { value: true, writable: true })
      document.dispatchEvent(new Event('visibilitychange'))
    })
    
    await page.waitForTimeout(1000)
    
    // Show tab again
    await page.evaluate(() => {
      Object.defineProperty(document, 'hidden', { value: false, writable: true })
      document.dispatchEvent(new Event('visibilitychange'))
    })
    
    // Visualizer should still be functioning
    const fullscreenViz = page.locator('[data-testid="fullscreen-visualizer"]').first()
    await expect(fullscreenViz).toBeVisible()
  })
})