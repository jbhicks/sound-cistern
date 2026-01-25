# Coolify Deployment Guide for Sound Cistern

## Deployment Readiness ✅

The application has been reviewed and is ready for Coolify deployment with the following features:

### ✅ **Completed Preparations**
- **Build Verification**: Application compiles successfully
- **Dependencies**: All Go modules properly configured (PocketBase v0.22.0, Templ v0.3.960)
- **Database**: Embedded SQLite database (no external database required)
- **Security**: No hardcoded secrets, proper OAuth configuration
- **Docker**: Multi-stage Dockerfile optimized for production
- **Health Check**: `/api/health` endpoint available via PocketBase
- **Environment**: Production-ready configuration templates

### 🔧 **Configuration Requirements**

#### 1. **Environment Variables** (Set in Coolify)
```bash
# Soundcloud OAuth
SOUNDCLOUD_CLIENT_ID=your_client_id
SOUNDCLOUD_CLIENT_SECRET=your_client_secret
SOUNDCLOUD_REDIRECT_URI=https://your-coolify-domain.com/auth/callback

# Application
GO_ENV=production
LOG_LEVEL=info
```

#### 2. **Soundcloud App Configuration**
- Update redirect URI in Soundcloud Developer Console to: `https://your-coolify-domain.com/auth/callback`
- Ensure your Coolify domain matches exactly

#### 3. **Data Persistence**
- Mount `/pb_data` directory as persistent volume in Coolify
- This stores SQLite database, logs, and uploaded files
- **Important**: Ensure backups of this directory

### 🚀 **Coolify Deployment Steps**

#### 1. **Create New Project**
- In Coolify dashboard, create a new project
- Select "Git Repository" as source
- Connect to your Git repository

#### 2. **Configure Build Settings**
- **Build Command**: `make build`
- **Start Command**: `./sound-cistern serve --http=0.0.0.0:8090`
- **Port**: `8090`
- **Environment Variables**: Add all required variables above
- **Persistent Volume**: Mount `/pb_data` for data persistence

#### 3. **Deploy**
- Deploy the application
- Monitor build logs for any issues
- Test the `/api/health` endpoint once deployed

### 🧪 **Post-Deployment Testing**

#### 1. **Basic Functionality**
```bash
# Health check
curl https://your-coolify-domain.com/api/health

# Home page
curl https://your-coolify-domain.com/

# Admin dashboard
open https://your-coolify-domain.com/_/
```

#### 2. **Database**
- SQLite database automatically created in `/pb_data/data.db`
- Migrations run automatically on startup
- Access admin UI at `/_/` to verify database structure

#### 3. **OAuth Flow**
- Test complete Soundcloud authentication flow
- Verify callback handling
- Check session management

### 🔍 **Monitoring & Troubleshooting**

#### Health Check Endpoint
- **URL**: `https://your-domain.com/api/health`
- **Response**: PocketBase health status

#### Common Issues
1. **Data Loss**: Ensure `/pb_data` is mounted as persistent volume
2. **OAuth Redirect**: Ensure Soundcloud redirect URI matches deployed domain
3. **Build Failures**: Check Go 1.23 compatibility and dependencies
4. **Port Binding**: Verify port 8090 is accessible
5. **File Permissions**: Ensure write permissions on `/pb_data`

#### Logs
- Application logs available in Coolify dashboard
- PocketBase logs in `/pb_data/logs.db`
- Monitor for any error patterns

### 🔒 **Security Considerations**

- **HTTPS**: Coolify provides SSL certificates automatically
- **Secrets**: All sensitive data via environment variables
- **Admin Access**: Secure admin panel at `/_/` with strong password
- **Data Files**: Ensure `/pb_data` directory is not publicly accessible
- **CORS**: OAuth flow properly configured for domain

### 📈 **Scaling & Performance**

- **Current Setup**: Single instance with embedded SQLite
- **Database**: SQLite suitable for small to medium traffic (thousands of concurrent users)
- **For Higher Scale**: Consider migrating to PostgreSQL (PocketBase supports both)
- **CDN**: Static assets in `/public` directory
- **Backups**: Regular backups of `/pb_data` directory

### 🔄 **Updates & Maintenance**

#### Deploying Updates
1. Push changes to Git repository
2. Coolify will auto-deploy if configured
3. Monitor deployment logs
4. Test critical functionality post-deployment

#### Database Backups
- Backup `/pb_data/data.db` regularly
- Use SQLite backup tools or simple file copy when app is stopped
- Consider automated backup scripts

#### Migrations
- PocketBase migrations in `/pb_migrations` run automatically on startup
- No manual migration commands needed

### 📞 **Support & Resources**

- **Coolify Documentation**: https://coolify.io/docs/
- **PocketBase Documentation**: https://pocketbase.io/docs/
- **Templ Documentation**: https://templ.guide/
- **Soundcloud API**: https://developers.soundcloud.com/
- **Project Repository**: Your Git repository

---

**Status**: ✅ Ready for deployment
**Last Updated**: 2025-10-21
**Version**: 1.0.0