# E2E Testing Strategy for OAuth Apps

## Recommended Approach: Hybrid Testing with Mocks

For your Sound Cistern app, I recommend a **hybrid approach** combining:

1. **API-level unit tests** with mocked Soundcloud responses
2. **Browser-based e2e tests** using Playwright with test user bypass
3. **Manual verification scripts** (what you already have)

## Implementation Strategy

### 1. Test User Bypass (Simplest)

Add a test mode that bypasses OAuth entirely:

```go
// In your main.go, add test mode detection
isTestMode := os.Getenv("TEST_MODE") == "true"

// Modify your auth middleware
func testAuthMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            if !isTestMode {
                return next(c)
            }
            
            // Create/find test user automatically
            testUser := createOrGetTestUser(app)
            c.Set(apis.ContextAuthRecordKey, testUser)
            return next(c)
        }
    }
}
```

### 2. Mock Soundcloud API

Create an HTTP test server that mocks Soundcloud responses:

```go
func mockSoundcloudAPI() *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/me/activities":
            json.NewEncoder(w).Encode(mockActivitiesResponse())
        case "/me":
            json.NewEncoder(w).Encode(mockUserResponse())
        case "/oauth2/token":
            json.NewEncoder(w).Encode(mockTokenResponse())
        }
    }))
}
```

### 3. Playwright E2E Tests

```javascript
// e2e/stream.spec.js
import { test, expect } from '@playwright/test';

test.describe('Stream Page', () => {
  test.beforeEach(async ({ page }) => {
    // Set test mode and navigate
    await page.goto('http://localhost:8090?test_mode=true');
  });

  test('displays tracks from Soundcloud', async ({ page }) => {
    await expect(page.locator('.track-item')).toHaveCount(5);
    await expect(page.locator('.track-title').first()).toContainText('Test Track');
  });

  test('favorites toggle works', async ({ page }) => {
    const favoriteButton = page.locator('.favorite-btn').first();
    await favoriteButton.click();
    await expect(favoriteButton).toHaveClass('favorited');
  });
});
```

## Running Tests

```bash
# Unit tests (API level)
go test -v ./...

# E2E tests (browser level)  
TEST_MODE=true npx playwright test

# Manual verification
./test_stream_endpoint.sh
```

## Benefits of This Approach

✅ **Fast**: No real OAuth flows or API calls  
✅ **Reliable**: No external service dependencies  
✅ **Comprehensive**: Tests both API and UI layers  
✅ **Maintainable**: Clear separation of concerns  

## What You Already Have

Your Ralph verification scripts are perfect for **manual testing** and **CI checks**. The hybrid approach above would complement them with automated e2e coverage.

Would you like me to help you implement the Playwright setup or the test user bypass mechanism?