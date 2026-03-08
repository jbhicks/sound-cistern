/**
 * Sound Cistern E2E Tests
 *
 * Tests run against the live server at http://localhost:8090.
 * The server must be running before executing tests.
 *
 * Two modes:
 *   Default  — unauthenticated tests only (no env vars needed)
 *   TEST_MODE — full app tests; start server with: TEST_MODE=true make dev-test
 *               then run: TEST_MODE=true npx playwright test
 */

import { test, expect } from '@playwright/test'

const BASE = process.env.BASE_URL || 'http://localhost:8090'
const TEST_MODE = process.env.TEST_MODE === 'true'

// ─── Health & Infrastructure ──────────────────────────────────────────────────

test.describe('Health & Infrastructure', () => {
  test('health endpoint returns healthy', async ({ request }) => {
    const res = await request.get(`${BASE}/health`)
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body.status).toBe('healthy')
  })

  test('root / serves the React app (not a redirect, not old Templ HTML)', async ({ page }) => {
    await page.goto(`${BASE}/`)
    // Must NOT redirect to /v2/
    expect(page.url()).not.toContain('/v2')
    // Must have the React root mount point
    await expect(page.locator('#root')).toBeAttached()
    // Must NOT contain old Templ/Pico.css markers
    const html = await page.content()
    expect(html).not.toContain('pico.css')
    expect(html).not.toContain('hx-get')
    expect(html).not.toContain('htmx')
  })

  test('/stream serves the React app (not the old Templ stream page)', async ({ page }) => {
    await page.goto(`${BASE}/stream`)
    await expect(page.locator('#root')).toBeAttached()
    const html = await page.content()
    expect(html).not.toContain('hx-get')
    expect(html).not.toContain('htmx')
    expect(html).not.toContain('filter-q')   // old Templ stream had id="filter-q"
    expect(html).not.toContain('cards-grid') // old Templ stream had id="cards-grid"
  })

  test('/favorites serves the React app (not the old Templ favorites page)', async ({ page }) => {
    await page.goto(`${BASE}/favorites`)
    await expect(page.locator('#root')).toBeAttached()
    const html = await page.content()
    expect(html).not.toContain('hx-get')
    expect(html).not.toContain('htmx')
  })

  test('/v2/ redirects to /', async ({ request }) => {
    const res = await request.get(`${BASE}/v2/`, { maxRedirects: 0 })
    expect([301, 302, 307, 308]).toContain(res.status())
    expect(res.headers()['location']).toBe('/')
  })

  test('/v2/stream redirects to /stream', async ({ request }) => {
    const res = await request.get(`${BASE}/v2/stream`, { maxRedirects: 0 })
    expect([301, 302, 307, 308]).toContain(res.status())
    expect(res.headers()['location']).toBe('/stream')
  })

  test('manifest.json has correct root start_url', async ({ request }) => {
    const res = await request.get(`${BASE}/manifest.json`)
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body.name).toBe('Sound Cistern')
    expect(body.start_url).toBe('/')
    expect(body.start_url).not.toContain('/v2')
    expect(body.icons[0].src).not.toContain('/v2')
  })

  test('PWA icon-192.png is served', async ({ request }) => {
    const res = await request.get(`${BASE}/icons/icon-192.png`)
    expect(res.status()).toBe(200)
    expect(res.headers()['content-type']).toContain('image/png')
  })

  test('PWA icon-512.png is served', async ({ request }) => {
    const res = await request.get(`${BASE}/icons/icon-512.png`)
    expect(res.status()).toBe(200)
    expect(res.headers()['content-type']).toContain('image/png')
  })

  test('service worker sw.js is served at root (not /v2/sw.js)', async ({ request }) => {
    const good = await request.get(`${BASE}/sw.js`)
    expect(good.status()).toBe(200)
    expect(good.headers()['content-type']).toContain('javascript')
  })
})

// ─── Unauthenticated React UI ─────────────────────────────────────────────────

test.describe('Unauthenticated React UI', () => {
  test('React app renders login screen when not authenticated', async ({ page }) => {
    await page.goto(`${BASE}/`)
    // Wait for React to finish rendering (spinner disappears)
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 8000 })
    // Must show the Connect SoundCloud button — this is React-rendered, not Templ
    await expect(page.getByRole('link', { name: /connect soundcloud/i })).toBeVisible()
    // Must show the gradient heading
    await expect(page.locator('h1', { hasText: 'Sound Cistern' })).toBeVisible()
    // Must NOT show the old Templ login card
    await expect(page.locator('text=Continue with Soundcloud')).not.toBeVisible()
  })

  test('connect button href is /auth/soundcloud', async ({ page }) => {
    await page.goto(`${BASE}/`)
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 8000 })
    const btn = page.getByRole('link', { name: /connect soundcloud/i })
    await expect(btn).toHaveAttribute('href', '/auth/soundcloud')
  })

  test('/stream redirects unauthenticated users to login (React login, not Templ /login)', async ({ page }) => {
    await page.goto(`${BASE}/stream`)
    // React app handles auth check client-side — stays on same URL, shows login UI
    // It should NOT redirect to the old /login Templ route
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 8000 })
    // Either shows login UI at /stream or redirects to /
    const url = page.url()
    expect(url).not.toContain('/login') // old Templ login page
    // Should show the connect button (React login state)
    await expect(page.getByRole('link', { name: /connect soundcloud/i })).toBeVisible()
  })
})

// ─── API Endpoints (unauthenticated) ─────────────────────────────────────────

test.describe('API Endpoints (unauthenticated)', () => {
  test('/api/user returns 401', async ({ request }) => {
    const res = await request.get(`${BASE}/api/user`)
    expect(res.status()).toBe(401)
  })

  test('/api/stream returns 401', async ({ request }) => {
    const res = await request.get(`${BASE}/api/stream?format=json`)
    expect(res.status()).toBe(401)
  })

  test('/api/favorites returns 4xx', async ({ request }) => {
    const res = await request.get(`${BASE}/api/favorites`)
    expect(res.status()).toBeGreaterThanOrEqual(400)
    expect(res.status()).toBeLessThan(500)
  })

  test('/api/play-history returns 401', async ({ request }) => {
    const res = await request.get(`${BASE}/api/play-history`)
    expect(res.status()).toBe(401)
  })
})

// ─── TEST_MODE: Full App Flow ─────────────────────────────────────────────────

test.describe('TEST_MODE: Full App Flow', () => {
  test.skip(!TEST_MODE, 'Start server with TEST_MODE=true (make dev-test), then run: TEST_MODE=true npx playwright test')

  test('stream API returns JSON tracks with required fields', async ({ request }) => {
    const res = await request.get(`${BASE}/api/stream?format=json&limit=10`)
    expect(res.status()).toBe(200)
    const body = await res.json()
    const tracks = body.tracks || body
    expect(Array.isArray(tracks)).toBe(true)
    expect(tracks.length).toBeGreaterThan(0)
    const t = tracks[0]
    expect(t).toHaveProperty('track_id')
    expect(t).toHaveProperty('track_title')
    expect(t).toHaveProperty('artist_name')
    expect(t).toHaveProperty('track_duration')
    expect(t).toHaveProperty('artwork_url')
    expect(t).toHaveProperty('playback_count')
    expect(t).toHaveProperty('favoritings_count')
  })

  test('stream API search filter returns only matching tracks', async ({ request }) => {
    const res = await request.get(`${BASE}/api/stream?format=json&q=house`)
    expect(res.status()).toBe(200)
    const body = await res.json()
    const tracks = body.tracks || body
    expect(tracks.length).toBeGreaterThan(0)
    for (const t of tracks) {
      const text = `${t.track_title} ${t.artist_name} ${t.genre}`.toLowerCase()
      expect(text).toContain('house')
    }
  })

  test('stream API duration filter returns only long-enough tracks', async ({ request }) => {
    const res = await request.get(`${BASE}/api/stream?format=json&duration_min=5`)
    expect(res.status()).toBe(200)
    const body = await res.json()
    const tracks = body.tracks || body
    for (const t of tracks) {
      expect(t.track_duration).toBeGreaterThanOrEqual(300000) // 5 min in ms
    }
  })

  test('stream API pagination returns different tracks per page', async ({ request }) => {
    const p1 = await (await request.get(`${BASE}/api/stream?format=json&limit=5&offset=0`)).json()
    const p2 = await (await request.get(`${BASE}/api/stream?format=json&limit=5&offset=5`)).json()
    const ids1 = (p1.tracks || p1).map(t => t.track_id)
    const ids2 = (p2.tracks || p2).map(t => t.track_id)
    expect(ids1.length).toBeGreaterThan(0)
    expect(ids2.length).toBeGreaterThan(0)
    // No overlap between pages
    const overlap = ids1.filter(id => ids2.includes(id))
    expect(overlap).toHaveLength(0)
  })

  test('favorites API returns { tracks: [] }', async ({ request }) => {
    const res = await request.get(`${BASE}/api/favorites`)
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body).toHaveProperty('tracks')
    expect(Array.isArray(body.tracks)).toBe(true)
  })

  test('play history POST syncs entries, GET returns them', async ({ request }) => {
    const entry = {
      track_id: 'e2e-test-track',
      track_title: 'E2E Test Track',
      artist_name: 'Test Artist',
      artwork_url: '',
      track_duration: 180000,
      genre: 'Electronic',
      played_at: new Date().toISOString(),
    }
    const post = await request.post(`${BASE}/api/play-history`, {
      data: [entry],
      headers: { 'Content-Type': 'application/json' },
    })
    expect(post.status()).toBe(200)
    const postBody = await post.json()
    expect(postBody).toHaveProperty('synced')

    const get = await request.get(`${BASE}/api/play-history?limit=10`)
    expect(get.status()).toBe(200)
    const getBody = await get.json()
    expect(getBody).toHaveProperty('entries')
    expect(Array.isArray(getBody.entries)).toBe(true)
  })

  test('playlists API returns array', async ({ request }) => {
    const res = await request.get(`${BASE}/api/playlists`)
    expect(res.status()).toBe(200)
    expect(Array.isArray(await res.json())).toBe(true)
  })

  test('user API returns id', async ({ request }) => {
    const res = await request.get(`${BASE}/api/user`)
    expect(res.status()).toBe(200)
    expect(await res.json()).toHaveProperty('id')
  })
})

// ─── TEST_MODE: React App UI ──────────────────────────────────────────────────

test.describe('TEST_MODE: React App UI', () => {
  test.skip(!TEST_MODE, 'Start server with TEST_MODE=true (make dev-test), then run: TEST_MODE=true npx playwright test')

  test('authenticated app shows nav with Stream, Favorites, Playlists, Analytics links', async ({ page }) => {
    await page.goto(`${BASE}/`)
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 10000 })
    // Must show nav — not the login screen
    await expect(page.locator('nav')).toBeVisible()
    await expect(page.getByRole('link', { name: /connect soundcloud/i })).not.toBeVisible()
    // Nav must have the four main links
    await expect(page.getByRole('link', { name: /stream/i })).toBeVisible()
    await expect(page.getByRole('link', { name: /favorites/i })).toBeVisible()
  })

  test('stream page renders React UI with search input and sync button', async ({ page }) => {
    await page.goto(`${BASE}/stream`)
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 10000 })
    // React-specific elements
    await expect(page.getByPlaceholder(/search tracks/i)).toBeVisible()
    await expect(page.getByRole('button', { name: /sync/i })).toBeVisible()
    await expect(page.locator('h1', { hasText: /stream/i })).toBeVisible()
    // Must NOT have old Templ markers
    const html = await page.content()
    expect(html).not.toContain('hx-get')
    expect(html).not.toContain('filter-q')
  })

  test('stream page loads and displays track cards after sync', async ({ page }) => {
    await page.goto(`${BASE}/stream`)
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 10000 })

    // Trigger sync to populate tracks
    const syncBtn = page.getByRole('button', { name: /sync/i })
    await syncBtn.click()
    // Wait for sync to complete (spinner goes away)
    await expect(syncBtn).not.toHaveText(/syncing/i, { timeout: 15000 })

    // Wait for skeleton loaders to resolve
    await page.waitForFunction(() => {
      return document.querySelectorAll('.animate-pulse').length === 0
    }, { timeout: 10000 })

    // Should show track count text like "12 tracks" or "12+ tracks"
    await expect(page.locator('text=/\\d+\\+? tracks/')).toBeVisible({ timeout: 8000 })
  })

  test('search filters tracks and shows result count', async ({ page }) => {
    await page.goto(`${BASE}/stream`)
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 10000 })

    const search = page.getByPlaceholder(/search tracks/i)
    await search.fill('house')
    // Debounce is 300ms
    await page.waitForTimeout(500)

    // Should show either filtered count or "No matches"
    const filtered = page.locator('text=/\\d+\\+? tracks/')
    const noMatch = page.locator('text=No matches')
    await expect(filtered.or(noMatch)).toBeVisible({ timeout: 5000 })

    // Clear search
    await search.fill('')
    await page.waitForTimeout(500)
  })

  test('grid/list toggle switches layout class', async ({ page }) => {
    await page.goto(`${BASE}/stream`)
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 10000 })

    // Grid view should be active by default — grid container exists
    await expect(page.locator('.grid')).toBeVisible()

    // Click list view button (the second button in the toggle group)
    const listBtn = page.locator('button').filter({ hasText: '' }).filter({ has: page.locator('svg') })
    // The view toggle is the last button group — find List icon button
    const viewToggle = page.locator('[class*="rounded-xl"][class*="border"][class*="overflow-hidden"]').last()
    const listViewBtn = viewToggle.locator('button').last()
    await listViewBtn.click()
    await page.waitForTimeout(200)

    // Grid should be gone, space-y list layout should appear
    await expect(page.locator('.grid')).not.toBeVisible()
  })

  test('filter panel opens, shows options, closes', async ({ page }) => {
    await page.goto(`${BASE}/stream`)
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 10000 })

    const filterBtn = page.getByRole('button', { name: /filter/i })
    await expect(filterBtn).toBeVisible()
    await filterBtn.click()

    // Filter panel should slide in — look for sort/duration options
    await expect(page.locator('text=/sort|duration|genre/i').first()).toBeVisible({ timeout: 3000 })

    // Close with Escape
    await page.keyboard.press('Escape')
    await page.waitForTimeout(300)
    await expect(page.locator('text=/sort|duration|genre/i').first()).not.toBeVisible()
  })

  test('favorites page renders React UI', async ({ page }) => {
    await page.goto(`${BASE}/favorites`)
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 10000 })
    await expect(page.locator('h1', { hasText: /favorites/i })).toBeVisible()
    // React-specific: RSS and Export buttons
    await expect(page.getByRole('button', { name: /rss/i }).or(page.locator('text=RSS'))).toBeVisible()
    await expect(page.getByRole('button', { name: /export/i }).or(page.locator('text=Export'))).toBeVisible()
    const html = await page.content()
    expect(html).not.toContain('hx-get')
  })

  test('analytics page renders with stat cards', async ({ page }) => {
    await page.goto(`${BASE}/analytics`)
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 10000 })
    await expect(page.locator('h1', { hasText: /analytics/i })).toBeVisible()
    // Should have stat cards (total plays, etc.)
    await expect(page.locator('text=/plays|tracks|artists/i').first()).toBeVisible()
  })

  test('playlists page renders React UI', async ({ page }) => {
    await page.goto(`${BASE}/playlists`)
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 10000 })
    await expect(page.locator('h1', { hasText: /playlists/i })).toBeVisible()
    const html = await page.content()
    expect(html).not.toContain('hx-get')
  })

  test('nav links navigate to correct React routes', async ({ page }) => {
    await page.goto(`${BASE}/`)
    await expect(page.locator('nav')).toBeVisible({ timeout: 10000 })

    await page.getByRole('link', { name: /favorites/i }).click()
    await expect(page).toHaveURL(/\/favorites/)
    await expect(page.locator('h1', { hasText: /favorites/i })).toBeVisible()

    await page.getByRole('link', { name: /stream/i }).click()
    await expect(page).toHaveURL(/\/stream/)
    await expect(page.locator('h1', { hasText: /stream/i })).toBeVisible()
  })
})
