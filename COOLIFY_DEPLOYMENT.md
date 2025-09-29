# Coolify Deployment Guide for Sound Cistern

## Deployment Readiness ✅

The application has been reviewed and is ready for Coolify deployment with the following features:

### ✅ **Completed Preparations**
- **Build Verification**: Application compiles successfully
- **Dependencies**: All Go modules properly configured
- **Database**: Migrations ready, supports environment variables
- **Security**: No hardcoded secrets, proper OAuth configuration
- **Docker**: Multi-stage Dockerfile optimized for production
- **Health Check**: `/health` endpoint added for monitoring
- **Environment**: Production-ready configuration templates

### 🔧 **Configuration Requirements**

#### 1. **Environment Variables** (Set in Coolify)
```bash
# Database
DATABASE_URL=postgres://username:password@hostname:5432/sound_cistern_production?sslmode=require

# Soundcloud OAuth
SOUNDCLOUD_REDIRECT_URI=https://your-coolify-domain.com/auth/callback

# Application
GO_ENV=production
LOG_LEVEL=info
SESSION_SECRET=your-secure-random-string
```

#### 2. **Soundcloud App Configuration**
- Update redirect URI in Soundcloud Developer Console to: `https://your-coolify-domain.com/auth/callback`
- Ensure your Coolify domain matches exactly

#### 3. **Database Setup**
- Coolify will need PostgreSQL service configured
- Run migrations after deployment: `buffalo db migrate up`

### 🚀 **Coolify Deployment Steps**

#### 1. **Create New Project**
- In Coolify dashboard, create a new project
- Select "Git Repository" as source
- Connect to your Git repository

#### 2. **Configure Build Settings**
- **Build Command**: `buffalo build --static -o /bin/app`
- **Start Command**: `./bin/app`
- **Port**: `3000`
- **Environment Variables**: Add all required variables above

#### 3. **Add Database Service**
- Add PostgreSQL service to your project
- Configure `DATABASE_URL` to point to the PostgreSQL service

#### 4. **Deploy**
- Deploy the application
- Monitor build logs for any issues
- Test the `/health` endpoint once deployed

### 🧪 **Post-Deployment Testing**

#### 1. **Basic Functionality**
```bash
# Health check
curl https://your-coolify-domain.com/health

# Home page
curl https://your-coolify-domain.com/

# OAuth initiation
curl -I https://your-coolify-domain.com/auth/soundcloud
```

#### 2. **Database Connectivity**
- Verify database connection in application logs
- Check that migrations have run successfully

#### 3. **OAuth Flow**
- Test complete Soundcloud authentication flow
- Verify callback handling
- Check session management

### 🔍 **Monitoring & Troubleshooting**

#### Health Check Endpoint
- **URL**: `https://your-domain.com/health`
- **Response**: `{"status": "healthy", "service": "sound-cistern", "version": "1.0.0"}`

#### Common Issues
1. **Database Connection**: Check `DATABASE_URL` format and credentials
2. **OAuth Redirect**: Ensure Soundcloud redirect URI matches deployed domain
3. **Build Failures**: Check Go version compatibility and dependencies
4. **Port Binding**: Verify port 3000 is accessible

#### Logs
- Application logs available in Coolify dashboard
- Database logs in PostgreSQL service logs
- Monitor for any error patterns

### 🔒 **Security Considerations**

- **HTTPS**: Coolify provides SSL certificates automatically
- **Secrets**: All sensitive data via environment variables
- **Session Security**: Secure session secret configured
- **CORS**: OAuth flow properly configured for domain

### 📈 **Scaling & Performance**

- **Current Setup**: Single instance suitable for small to medium traffic
- **Database**: PostgreSQL can be scaled independently
- **Caching**: Consider Redis for session storage in high-traffic scenarios
- **CDN**: Static assets served directly by Buffalo

### 🔄 **Updates & Maintenance**

#### Deploying Updates
1. Push changes to Git repository
2. Coolify will auto-deploy if configured
3. Monitor deployment logs
4. Test critical functionality post-deployment

#### Database Migrations
- Run `buffalo db migrate up` after code deployments with schema changes
- Backup database before major migrations

### 📞 **Support & Resources**

- **Coolify Documentation**: https://coolify.io/docs/
- **Buffalo Framework**: https://gobuffalo.io/
- **Soundcloud API**: https://developers.soundcloud.com/
- **Project Repository**: Your Git repository

---

**Status**: ✅ Ready for deployment
**Last Updated**: $(date)
**Version**: 1.0.0