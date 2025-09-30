# Sound Cistern OAuth Redirect URI Implementation Plan

## Overview
This plan outlines the steps to properly configure the Soundcloud OAuth redirect URI for deployment on Coolify. The redirect URI is essential for the OAuth 2.0 flow, ensuring secure authentication and callback handling.

## Current Status
- Soundcloud app configured with placeholder redirect URI: `http://jbhicks.dev/auth/callback`
- Application code uses environment variables for configuration
- Need to update for Coolify deployment with generated URL

## Prerequisites
- Coolify deployment URL (e.g., `https://your-app-name.coolify-instance.com`)
- Access to Soundcloud Developer Console
- Application deployed to Coolify environment

## Implementation Steps

### 1. Determine Coolify Deployment URL
**Action**: Obtain the exact URL assigned by Coolify for your application.
- This will be in the format: `https://[app-name].[coolify-instance].com`
- Ensure it uses HTTPS (required for OAuth security)

### 2. Update Soundcloud App Configuration
**Action**: Modify redirect URI in Soundcloud Developer Console
- Navigate to: https://developers.soundcloud.com/
 - Select your app (Client ID: configured in environment variables)
- Update "Redirect URI" field to:
  ```
  https://your-coolify-deployment-url.com/auth/callback
  ```
- Save changes
- **Note**: Exact match required (case-sensitive, no trailing slashes)

### 3. Update Application Environment Variables
**Action**: Modify `.env` file with new redirect URI
```bash
SOUNDCLOUD_REDIRECT_URI=https://your-coolify-deployment-url.com/auth/callback
```

### 4. Update Application Code (if needed)
**Action**: Ensure code uses environment variable correctly
- Current code in `/actions/soundcloud.go` already uses `envy.Get("SOUNDCLOUD_REDIRECT_URI", default)`
- No changes needed unless default fallback needs updating
- Verify both `SoundcloudAuth` and `SoundcloudCallback` handlers use the same URI

### 5. Security and HTTPS Configuration
**Action**: Ensure HTTPS is properly configured
- Coolify typically handles SSL certificates automatically
- Verify redirect URI uses `https://` scheme
- Update any load balancer or proxy configurations if necessary

## Testing Procedures

### 1. Local Testing (with updated .env)
```bash
# Set environment variables (configured in .env or deployment platform)
export SOUNDCLOUD_REDIRECT_URI=https://your-coolify-deployment-url.com/auth/callback

# Start application
go run ./cmd/app

# Test OAuth flow
curl -I http://localhost:3000/auth/soundcloud
# Should return 302 redirect to Soundcloud with correct parameters
```

### 2. Deployed Testing
- Deploy updated application to Coolify
- Test complete OAuth flow:
  1. Visit: `https://your-coolify-deployment-url.com/auth/soundcloud`
  2. Complete Soundcloud authentication
  3. Verify callback to `/auth/callback`
  4. Check session creation and feed access

### 3. Edge Case Testing
- Test with invalid/missing authorization codes
- Verify error handling for mismatched redirect URIs
- Test on different browsers/devices

## Deployment Considerations

### Environment-Specific Configuration
- Use different redirect URIs for development vs. production
- Consider using environment-specific `.env` files
- Document all configured URIs in project README

### Monitoring and Logging
- Enable logging for OAuth callback attempts
- Monitor for authentication failures
- Set up alerts for OAuth-related errors

### Rollback Plan
- Keep previous redirect URI configuration as backup
- Have quick revert process if issues arise
- Test rollback in staging environment first

## Future Steps (Permanent Domain)

When purchasing a permanent domain:
1. Purchase domain (e.g., `soundcistern.yourdomain.com`)
2. Update DNS to point to Coolify instance
3. Update Soundcloud redirect URI to new domain
4. Update application environment variables
5. Redeploy and test
6. Remove old redirect URIs from Soundcloud app

## Success Criteria
- OAuth authentication completes successfully
- Users can access `/feed` after authentication
- No security errors or redirect mismatches
- Application handles token refresh properly
- All tests pass in deployed environment

## Timeline
- **Immediate**: Update Soundcloud configuration once Coolify URL is known
- **Short-term**: Deploy and test with temporary URL
- **Long-term**: Migrate to permanent domain when available

## Resources
- Soundcloud OAuth 2.0 Documentation: https://developers.soundcloud.com/docs/api/reference#authentication
- Coolify Documentation: https://coolify.io/docs/
- OAuth 2.0 Security Best Practices: https://tools.ietf.org/html/rfc6819