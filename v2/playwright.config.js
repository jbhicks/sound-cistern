import { defineConfig, devices } from '@playwright/test'

/**
 * Sound Cistern Playwright E2E Configuration
 *
 * Tests run against the live dev server at localhost:8090.
 * The server must be running with TEST_MODE=true for auth to be bypassed.
 * Start it with: make dev-test (or make dev if already in test mode)
 *
 * Run tests: npx playwright test
 * Run with UI: npx playwright test --ui
 * Run headed: npx playwright test --headed
 */
export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: false, // run sequentially to avoid auth state conflicts
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: [['list'], ['html', { open: 'never', outputFolder: 'playwright-report' }]],

  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:8090',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'off',
    // Use a persistent context so the auth cookie survives across tests
    storageState: undefined,
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  // No webServer block — user manages the server manually
})
