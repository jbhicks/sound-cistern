---
name: go-saas-template
description: Complete guide for building SaaS applications with Go, PocketBase, and modern web stack
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: go-developers
  stack: go-saas-pocketbase
---

# Go SaaS Application Development Skill

## Overview
This skill provides comprehensive patterns and best practices for building Software-as-a-Service (SaaS) applications using Go, PocketBase, and modern web technologies. Covers authentication, multi-tenancy, subscription management, billing, monitoring, and deployment strategies.

## SaaS Architecture Patterns

### Multi-Tenant Design
```go
package saas

import (
    "context"
    "database/sql"
    "github.com/pocketbase/pocketbase/core"
)

type Tenant struct {
    ID          string    `db:"id" json:"id"`
    Name        string    `db:"name" json:"name"`
    Subdomain    string    `db:"subdomain" json:"subdomain"`
    Domain      string    `db:"domain" json:"domain"`
    Plan        string    `db:"plan" json:"plan"`
    Status      string    `db:"status" json:"status"`
    CreatedAt   time.Time `db:"created_at" json:"created_at"`
    ExpiresAt   *time.Time `db:"expires_at" json:"expires_at,omitempty"`
}

type TenantContext struct {
    context.Context
    Tenant *Tenant
    User   *models.Record
}

// Tenant isolation middleware
func TenantMiddleware(app *pocketbase.PocketBase) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Extract tenant from subdomain or custom domain
            host := c.Request().Host
            tenantID := extractTenantID(host)
            
            if tenantID == "" {
                return next(c) // Main site
            }
            
            // Load tenant from database
            tenant, err := getTenantByID(app, tenantID)
            if err != nil {
                return c.JSON(http.StatusNotFound, map[string]string{
                    "error": "Tenant not found",
                })
            }
            
            // Check tenant status
            if tenant.Status != "active" {
                return c.JSON(http.StatusForbidden, map[string]string{
                    "error": "Tenant is not active",
                })
            }
            
            // Check subscription expiration
            if tenant.ExpiresAt != nil && tenant.ExpiresAt.Before(time.Now()) {
                return c.JSON(http.StatusForbidden, map[string]string{
                    "error": "Subscription expired",
                })
            }
            
            // Add tenant to context
            tenantContext := &TenantContext{
                Context: c.Request().Context(),
                Tenant:  tenant,
            }
            
            c.SetRequest(context.WithValue(c.Request().Context(), "tenant", tenant))
            return next(c)
        }
    }
}

func extractTenantID(host string) string {
    // Extract subdomain for multi-tenant routing
    parts := strings.Split(host, ".")
    if len(parts) >= 2 && parts[0] != "www" {
        return parts[0]
    }
    
    // For custom domains, look up by domain
    return getTenantByDomain(host)
}
```

### Tenant Database Collections
```go
// pb_migrations/1696000005_create_tenants.go
package migrations

import (
    "github.com/pocketbase/pocketbase/core"
)

func init() {
    core.OnMigrate().Register(func(db core.DB) error {
        collection := &core.Collection{
            Name: "tenants",
            Type: core.CollectionTypeBase,
        }

        collection.Fields.Add(
            &core.TextField{
                Name:     "name",
                Required: true,
                Min:      2,
                Max:      100,
            },
            &core.TextField{
                Name:     "subdomain",
                Required: true,
                Unique:   true,
                Pattern:  "^[a-z0-9-]+$",
                Min:      3,
                Max:      50,
            },
            &core.TextField{
                Name:     "domain",
                Unique:   true,
            },
            &core.TextField{
                Name:     "plan",
                Required: true,
                Default:   "free",
            },
            &core.SelectField{
                Name:      "status",
                Required:  true,
                MaxSelect: 1,
                Values:    []string{"active", "suspended", "expired"},
                Default:   "active",
            },
            &core.TextField{
                Name:     "settings",
                Type:      core.FieldTypeJson,
            },
            &core.DateTimeField{
                Name:     "expires_at",
            },
            &core.TextField{
                Name:     "stripe_customer_id",
            },
            &core.TextField{
                Name:     "owner_id",
                Required: true,
            },
        )

        // Access rules - owner can manage, system can read
        createRule := "@request.auth.id != '' && @request.auth.verified = true"
        collection.CreateRule = &createRule
        
        updateRule := "owner_id = @request.auth.id"
        collection.UpdateRule = &updateRule
        
        deleteRule := "owner_id = @request.auth.id"
        collection.DeleteRule = &deleteRule

        return app.Dao().SaveCollection(collection)
    }, nil)
}
```

## Subscription Management

### Subscription Plans
```go
type SubscriptionPlan struct {
    ID               string    `json:"id"`
    Name             string    `json:"name"`
    Price            float64   `json:"price"`
    Interval         string    `json:"interval"` // monthly, yearly
    Features         []Feature `json:"features"`
    MaxUsers         int       `json:"max_users"`
    MaxStorage       int64     `json:"max_storage"` // in bytes
    MaxAPIRequests   int       `json:"max_api_requests"`
}

type Feature struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Enabled     bool   `json:"enabled"`
}

var Plans = map[string]SubscriptionPlan{
    "free": {
        ID:             "free",
        Name:           "Free",
        Price:          0,
        Interval:        "monthly",
        MaxUsers:       3,
        MaxStorage:      1024 * 1024 * 1024, // 1GB
        MaxAPIRequests: 1000,
        Features: []Feature{
            {Name: "Basic Features", Description: "Core functionality", Enabled: true},
            {Name: "Support", Description: "Community support", Enabled: true},
            {Name: "API Access", Description: "Limited API access", Enabled: true},
        },
    },
    "pro": {
        ID:             "pro",
        Name:           "Professional",
        Price:          29.99,
        Interval:        "monthly",
        MaxUsers:       50,
        MaxStorage:      100 * 1024 * 1024 * 1024, // 100GB
        MaxAPIRequests: 100000,
        Features: []Feature{
            {Name: "Basic Features", Description: "Core functionality", Enabled: true},
            {Name: "Priority Support", Description: "Email support within 24h", Enabled: true},
            {Name: "Advanced Analytics", Description: "Detailed analytics and reports", Enabled: true},
            {Name: "Unlimited API Access", Description: "No rate limiting", Enabled: true},
            {Name: "Custom Integrations", Description: "Webhooks and integrations", Enabled: true},
        },
    },
    "enterprise": {
        ID:             "enterprise",
        Name:           "Enterprise",
        Price:          199.99,
        Interval:        "monthly",
        MaxUsers:       -1, // unlimited
        MaxStorage:      -1, // unlimited
        MaxAPIRequests:  -1, // unlimited
        Features: []Feature{
            {Name: "Basic Features", Description: "Core functionality", Enabled: true},
            {Name: "Dedicated Support", Description: "24/7 phone and email support", Enabled: true},
            {Name: "Advanced Analytics", Description: "Enterprise-grade analytics", Enabled: true},
            {Name: "Unlimited API Access", Description: "No rate limiting", Enabled: true},
            {Name: "Custom Integrations", Description: "Custom development support", Enabled: true},
            {Name: "SLA Guarantee", Description: "99.9% uptime guarantee", Enabled: true},
            {Name: "Custom Domain", Description: "White-label options", Enabled: true},
        },
    },
}
```

### Stripe Integration
```go
package billing

import (
    "github.com/stripe/stripe-go/v78"
    "github.com/stripe/stripe-go/v78/price"
    "github.com/stripe/stripe-go/v78/checkout/session"
    "github.com/stripe/stripe-go/v78/customer"
    "github.com/stripe/stripe-go/v78/subscription"
)

type BillingService struct {
    app         *pocketbase.PocketBase
    stripeClient *stripe.Client
}

func NewBillingService(app *pocketbase.PocketBase, stripeKey string) *BillingService {
    stripeClient := stripe.New(stripeKey, nil)
    return &BillingService{
        app:         app,
        stripeClient: stripeClient,
    }
}

func (bs *BillingService) CreateCheckoutSession(tenantID, planID, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
    plan, exists := Plans[planID]
    if !exists {
        return nil, fmt.Errorf("invalid plan: %s", planID)
    }

    // Get or create Stripe customer
    customerID, err := bs.getOrCreateStripeCustomer(tenantID)
    if err != nil {
        return nil, err
    }

    // Create Stripe price if not exists
    stripePrice, err := bs.getOrCreatePrice(planID, plan)
    if err != nil {
        return nil, err
    }

    params := &stripe.CheckoutSessionParams{
        Customer:     stripe.String(customerID),
        PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
        LineItems: []*stripe.CheckoutSessionLineItemParams{
            {
                Price:    stripe.String(stripePrice.ID),
                Quantity: stripe.Int64(1),
            },
        },
        Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
        SuccessURL: stripe.String(successURL),
        CancelURL:  stripe.String(cancelURL),
        Metadata: map[string]string{
            "tenant_id": tenantID,
            "plan_id":   planID,
        },
    }

    return session.New(params)
}

func (bs *BillingService) getOrCreateStripeCustomer(tenantID string) (string, error) {
    // Load tenant from database
    tenant, err := getTenantByID(bs.app, tenantID)
    if err != nil {
        return "", err
    }

    // Return existing customer ID
    if tenant.StripeCustomerID != "" {
        return tenant.StripeCustomerID, nil
    }

    // Create new customer
    params := &stripe.CustomerParams{
        Email: stripe.String(tenant.OwnerEmail),
        Metadata: map[string]string{
            "tenant_id": tenantID,
        },
    }

    customer, err := customer.New(params)
    if err != nil {
        return "", err
    }

    // Update tenant with customer ID
    tenant.StripeCustomerID = customer.ID
    updateTenant(bs.app, tenant)

    return customer.ID, nil
}
```

### Webhook Handling
```go
func (bs *BillingService) HandleStripeWebhook(payload []byte, signature string) error {
    event, err := webhook.ConstructEvent(payload, signature, os.Getenv("STRIPE_WEBHOOK_SECRET"))
    if err != nil {
        return fmt.Errorf("failed to construct webhook event: %w", err)
    }

    switch event.Type {
    case "checkout.session.completed":
        return bs.handleCheckoutCompleted(event.Data.Object.(*stripe.CheckoutSession))
    case "invoice.payment_succeeded":
        return bs.handlePaymentSucceeded(event.Data.Object.(*stripe.Invoice))
    case "invoice.payment_failed":
        return bs.handlePaymentFailed(event.Data.Object.(*stripe.Invoice))
    case "customer.subscription.deleted":
        return bs.handleSubscriptionDeleted(event.Data.Object.(*stripe.Subscription))
    default:
        return nil // Ignore other events
    }
}

func (bs *BillingService) handleCheckoutCompleted(session *stripe.CheckoutSession) error {
    tenantID := session.Metadata["tenant_id"]
    planID := session.Metadata["plan_id"]
    
    // Load tenant
    tenant, err := getTenantByID(bs.app, tenantID)
    if err != nil {
        return err
    }
    
    // Update tenant subscription
    tenant.Plan = planID
    tenant.Status = "active"
    
    if session.Subscription != nil {
        tenant.ExpiresAt = time.Unix(session.Subscription.CurrentPeriodEnd, 0)
    }
    
    return updateTenant(bs.app, tenant)
}
```

## User Management

### Multi-Tenant User System
```go
type TenantUser struct {
    ID       string    `db:"id" json:"id"`
    TenantID string    `db:"tenant_id" json:"tenant_id"`
    UserID   string    `db:"user_id" json:"user_id"`
    Role     string    `db:"role" json:"role"`
    Status   string    `db:"status" json:"status"`
    JoinedAt time.Time `db:"joined_at" json:"joined_at"`
}

// User invitation system
func (bs *BillingService) InviteUser(tenantID, email, role string) error {
    // Check tenant permissions
    tenant, err := getTenantByID(bs.app, tenantID)
    if err != nil {
        return err
    }
    
    if !bs.canInviteUsers(tenant.Plan) {
        return fmt.Errorf("plan does not allow user invitations")
    }
    
    // Check user limit
    currentUsers, err := bs.getTenantUserCount(tenantID)
    if err != nil {
        return err
    }
    
    plan, exists := Plans[tenant.Plan]
    if !exists || (plan.MaxUsers > 0 && currentUsers >= plan.MaxUsers) {
        return fmt.Errorf("user limit exceeded for plan %s", tenant.Plan)
    }
    
    // Create invitation
    invitation := &Invitation{
        ID:        generateID(),
        TenantID:   tenantID,
        Email:      email,
        Role:       role,
        Status:     "pending",
        ExpiresAt:  time.Now().Add(7 * 24 * time.Hour), // 7 days
        CreatedAt:  time.Now(),
    }
    
    return saveInvitation(bs.app, invitation)
}

func (bs *BillingService) AcceptInvitation(invitationID, userID string) error {
    invitation, err := getInvitationByID(bs.app, invitationID)
    if err != nil {
        return err
    }
    
    if invitation.Status != "pending" {
        return fmt.Errorf("invitation is no longer valid")
    }
    
    if invitation.ExpiresAt.Before(time.Now()) {
        return fmt.Errorf("invitation has expired")
    }
    
    // Create tenant user relationship
    tenantUser := &TenantUser{
        ID:       generateID(),
        TenantID: invitation.TenantID,
        UserID:   userID,
        Role:     invitation.Role,
        Status:   "active",
        JoinedAt:  time.Now(),
    }
    
    // Update invitation status
    invitation.Status = "accepted"
    saveInvitation(bs.app, invitation)
    
    return saveTenantUser(bs.app, tenantUser)
}
```

## Analytics and Monitoring

### Usage Tracking
```go
type UsageTracker struct {
    app *pocketbase.PocketBase
}

type UsageMetrics struct {
    TenantID       string    `json:"tenant_id"`
    MetricType     string    `json:"metric_type"` // api_requests, storage, users
    Value          int64     `json:"value"`
    Period         string    `json:"period"`     // daily, monthly
    RecordedAt     time.Time `json:"recorded_at"`
}

func (ut *UsageTracker) TrackAPIRequest(tenantID string) error {
    return ut.recordUsage(tenantID, "api_requests", 1, "daily")
}

func (ut *UsageTracker) TrackStorageUsage(tenantID string, bytes int64) error {
    return ut.recordUsage(tenantID, "storage", bytes, "daily")
}

func (ut *UsageTracker) recordUsage(tenantID, metricType string, value int64, period string) error {
    usage := &UsageMetrics{
        TenantID:   tenantID,
        MetricType: metricType,
        Value:      value,
        Period:     period,
        RecordedAt: time.Now(),
    }
    
    return saveUsageMetrics(ut.app, usage)
}

func (ut *UsageTracker) GetUsageReport(tenantID, period string) (*UsageReport, error) {
    tenant, err := getTenantByID(ut.app, tenantID)
    if err != nil {
        return nil, err
    }
    
    plan, exists := Plans[tenant.Plan]
    if !exists {
        return nil, fmt.Errorf("unknown plan: %s", tenant.Plan)
    }
    
    // Get usage metrics
    metrics, err := getUsageMetrics(ut.app, tenantID, period)
    if err != nil {
        return nil, err
    }
    
    report := &UsageReport{
        Plan:     plan,
        Period:   period,
        Metrics:  metrics,
        Exceeded: make(map[string]bool),
    }
    
    // Check limits
    for _, metric := range metrics {
        switch metric.MetricType {
        case "api_requests":
            if plan.MaxAPIRequests > 0 && metric.Value > int64(plan.MaxAPIRequests) {
                report.Exceeded["api_requests"] = true
            }
        case "storage":
            if plan.MaxStorage > 0 && metric.Value > plan.MaxStorage {
                report.Exceeded["storage"] = true
            }
        case "users":
            if plan.MaxUsers > 0 && metric.Value > int64(plan.MaxUsers) {
                report.Exceeded["users"] = true
            }
        }
    }
    
    return report, nil
}

type UsageReport struct {
    Plan     SubscriptionPlan `json:"plan"`
    Period   string           `json:"period"`
    Metrics  []UsageMetrics   `json:"metrics"`
    Exceeded map[string]bool  `json:"exceeded"`
}
```

### Performance Monitoring
```go
type HealthChecker struct {
    app *pocketbase.PocketBase
}

type HealthStatus struct {
    Status   string                 `json:"status"`
    Checks  map[string]CheckResult   `json:"checks"`
    Uptime  time.Duration           `json:"uptime"`
    Version string                 `json:"version"`
}

type CheckResult struct {
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
    Latency int64  `json:"latency_ms,omitempty"`
}

func (hc *HealthChecker) CheckHealth() *HealthStatus {
    status := &HealthStatus{
        Status:  "healthy",
        Checks:  make(map[string]CheckResult),
        Version: os.Getenv("APP_VERSION"),
    }
    
    // Database connectivity
    start := time.Now()
    err := hc.app.Dao().DB().Ping()
    latency := time.Since(start).Milliseconds()
    
    if err != nil {
        status.Checks["database"] = CheckResult{
            Status:  "unhealthy",
            Message: "Database connection failed",
        }
        status.Status = "unhealthy"
    } else {
        status.Checks["database"] = CheckResult{
            Status:  "healthy",
            Latency: latency,
        }
    }
    
    // Memory usage
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    memoryMB := float64(memStats.Alloc) / 1024 / 1024
    
    if memoryMB > 1000 { // 1GB threshold
        status.Checks["memory"] = CheckResult{
            Status:  "warning",
            Message: fmt.Sprintf("High memory usage: %.2f MB", memoryMB),
        }
        if status.Status == "healthy" {
            status.Status = "warning"
        }
    } else {
        status.Checks["memory"] = CheckResult{
            Status:  "healthy",
            Message: fmt.Sprintf("Memory usage: %.2f MB", memoryMB),
        }
    }
    
    // Active tenants
    activeTenants, err := getActiveTenantCount(hc.app)
    if err != nil {
        status.Checks["tenants"] = CheckResult{
            Status:  "unhealthy",
            Message: "Failed to check tenant count",
        }
        status.Status = "unhealthy"
    } else {
        status.Checks["tenants"] = CheckResult{
            Status:  "healthy",
            Message: fmt.Sprintf("Active tenants: %d", activeTenants),
        }
    }
    
    return status
}
```

## Security Implementation

### Multi-Tenant Security
```go
type SecurityManager struct {
    app *pocketbase.PocketBase
}

func (sm *SecurityManager) TenantIsolationMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            tenant := getTenantFromContext(c)
            if tenant == nil {
                return next(c) // Skip for main site
            }
            
            // Ensure queries are tenant-scoped
            c.Set("tenant_filter", fmt.Sprintf("tenant_id = '%s'", tenant.ID))
            
            // Rate limiting per tenant
            if err := sm.checkRateLimit(c, tenant.ID); err != nil {
                return c.JSON(http.StatusTooManyRequests, map[string]string{
                    "error": "Rate limit exceeded",
                })
            }
            
            return next(c)
        }
    }
}

func (sm *SecurityManager) checkRateLimit(c echo.Context, tenantID string) error {
    // Get tenant's plan limits
    tenant, err := getTenantByID(sm.app, tenantID)
    if err != nil {
        return err
    }
    
    plan, exists := Plans[tenant.Plan]
    if !exists || plan.MaxAPIRequests <= 0 {
        return nil // No limit
    }
    
    // Check current usage
    currentUsage, err := getCurrentAPIUsage(sm.app, tenantID, time.Hour)
    if err != nil {
        return err
    }
    
    if currentUsage >= int64(plan.MaxAPIRequests) {
        return fmt.Errorf("rate limit exceeded")
    }
    
    return nil
}

func (sm *SecurityManager) EncryptSensitiveData(data string) (string, error) {
    key := []byte(os.Getenv("MASTER_ENCRYPTION_KEY"))
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
    
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}
```

## Deployment and Scaling

### Docker Configuration
```dockerfile
# Dockerfile for multi-tenant SaaS
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/pb_migrations ./pb_migrations/
COPY --from=builder /app/public ./public/

EXPOSE 8090
CMD ["./main", "serve", "--http=0.0.0.0:8090"]
```

### Docker Compose for Production
```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8090:8090"
    environment:
      - GO_ENV=production
      - DATABASE_URL=sqlite:///data/pb_data.db
      - SOUNDCLOUD_CLIENT_ID=${SOUNDCLOUD_CLIENT_ID}
      - SOUNDCLOUD_CLIENT_SECRET=${SOUNDCLOUD_CLIENT_SECRET}
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY}
      - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET}
      - REDIS_URL=${REDIS_URL}
    volumes:
      - ./pb_data:/data
      - ./logs:/logs
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8090/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl
    depends_on:
      - app
    restart: unless-stopped

volumes:
  redis_data:
```

### Kubernetes Deployment
```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: saas-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: saas-app
  template:
    metadata:
      labels:
        app: saas-app
    spec:
      containers:
      - name: app
        image: your-registry/saas-app:latest
        ports:
        - containerPort: 8090
        env:
        - name: GO_ENV
          value: "production"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: database-url
        - name: SOUNDCLOUD_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: soundcloud-client-id
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8090
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Testing Strategies

### Integration Testing
```go
func TestTenantIsolation(t *testing.T) {
    app := setupTestApp(t)
    
    // Create two tenants
    tenant1 := createTestTenant(t, app, "tenant1")
    tenant2 := createTestTenant(t, app, "tenant2")
    
    // Create users for each tenant
    user1 := createTestUser(t, app, "user1@tenant1.com")
    user2 := createTestUser(t, app, "user2@tenant2.com")
    
    // Test data isolation
    req1 := httptest.NewRequest("GET", "/api/posts", nil)
    req1.Header.Set("Host", "tenant1.yourapp.com")
    req1.Header.Set("Authorization", "Bearer "+user1.Token)
    
    req2 := httptest.NewRequest("GET", "/api/posts", nil)
    req2.Header.Set("Host", "tenant2.yourapp.com")
    req2.Header.Set("Authorization", "Bearer "+user2.Token)
    
    // Should only return posts from respective tenants
    resp1 := testEndpoint(t, app, req1)
    resp2 := testEndpoint(t, app, req2)
    
    posts1 := parsePosts(resp1.Body)
    posts2 := parsePosts(resp2.Body)
    
    // Verify isolation
    for _, post := range posts1 {
        assert.Equal(t, tenant1.ID, post.TenantID)
    }
    
    for _, post := range posts2 {
        assert.Equal(t, tenant2.ID, post.TenantID)
    }
}
```

### Load Testing
```go
func TestConcurrentRequests(t *testing.T) {
    app := setupTestApp(t)
    tenant := createTestTenant(t, app, "load-test")
    
    var wg sync.WaitGroup
    var errors []error
    var mu sync.Mutex
    
    // Simulate 100 concurrent requests
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            req := httptest.NewRequest("GET", "/api/posts", nil)
            req.Header.Set("Host", "load-test.yourapp.com")
            
            resp, err := testRequest(app, req)
            if err != nil {
                mu.Lock()
                errors = append(errors, err)
                mu.Unlock()
            }
            
            if resp.StatusCode != http.StatusOK {
                mu.Lock()
                errors = append(errors, fmt.Errorf("status: %d", resp.StatusCode))
                mu.Unlock()
            }
        }()
    }
    
    wg.Wait()
    
    // Check that most requests succeeded
    assert.LessOrEqual(t, len(errors), 5) // Allow <5% failure rate
}
```

## Monitoring and Observability

### Structured Logging
```go
type Logger struct {
    app *pocketbase.PocketBase
}

type LogEntry struct {
    Level     string                 `json:"level"`
    Message  string                 `json:"message"`
    TenantID string                 `json:"tenant_id,omitempty"`
    UserID   string                 `json:"user_id,omitempty"`
    RequestID string                 `json:"request_id"`
    Context  map[string]interface{} `json:"context,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
}

func (l *Logger) LogRequest(level, message string, c echo.Context, context map[string]interface{}) {
    entry := LogEntry{
        Level:     level,
        Message:  message,
        TenantID: getTenantIDFromContext(c),
        UserID:   getUserIDFromContext(c),
        RequestID: c.Response().Header().Get("X-Request-ID"),
        Context:  context,
        Timestamp: time.Now(),
    }
    
    jsonEntry, _ := json.Marshal(entry)
    fmt.Printf("%s\n", jsonEntry)
}
```

### Metrics Collection
```go
type MetricsCollector struct {
    prometheusRegistry *prometheus.Registry
}

func NewMetricsCollector() *MetricsCollector {
    registry := prometheus.NewRegistry()
    
    // Register metrics
    registry.MustRegister(prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "tenant", "status"},
    ))
    
    registry.MustRegister(prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint", "tenant"},
    ))
    
    return &MetricsCollector{
        prometheusRegistry: registry,
    }
}

func (mc *MetricsCollector) RecordRequest(method, endpoint, tenant string, status int, duration time.Duration) {
    // Increment request counter
    prometheus.Counter.WithLabelValues(method, endpoint, tenant, strconv.Itoa(status)).Inc()
    
    // Record duration histogram
    prometheus.Histogram.WithLabelValues(method, endpoint, tenant).Observe(duration.Seconds())
}
```

## Best Practices

### Development Guidelines
1. **Tenant Isolation**: Always scope database queries to tenant
2. **Plan Enforcement**: Implement feature flags based on subscription plans
3. **Security First**: Encrypt sensitive data, implement proper RBAC
4. **Monitoring**: Comprehensive logging, metrics, and health checks
5. **Scalability**: Design for horizontal scaling from day one

### Production Considerations
1. **Database Optimization**: Use connection pooling, query optimization
2. **Caching Strategy**: Implement Redis for session and API caching
3. **Load Balancing**: Use nginx or cloud load balancers
4. **Backup Strategy**: Regular database backups and disaster recovery
5. **Security Hardening**: Regular security audits, dependency updates

### Performance Optimization
1. **Lazy Loading**: Load tenant data on-demand
2. **Connection Pooling**: Reuse database connections efficiently
3. **Async Processing**: Background jobs for heavy operations
4. **CDN Integration**: Static assets delivered via CDN
5. **Database Indexing**: Proper indexes for multi-tenant queries

---

This skill provides comprehensive patterns for building scalable, secure SaaS applications with Go and PocketBase, covering the complete development lifecycle from architecture to deployment.