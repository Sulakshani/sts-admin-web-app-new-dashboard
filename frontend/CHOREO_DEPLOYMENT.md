# Choreo Deployment Guide - Admin Dashboard

This guide explains how to deploy the SmartTransit Admin Dashboard to Choreo platform.

## Prerequisites

1. **Choreo Account**: Sign up at [https://console.choreo.dev/](https://console.choreo.dev/)
2. **GitHub Repository**: Your code must be in a GitHub repository (already done ✓)
3. **Backend API**: Ensure your Go backend is already deployed and accessible

## Project Structure

The following files are configured for Choreo deployment:

- `Dockerfile` - Multi-stage build for production deployment
- `nginx.conf` - NGINX configuration for serving the Angular SPA
- `.dockerignore` - Excludes unnecessary files from Docker build

## Deployment Steps

### 1. Create Component in Choreo

1. Log in to [Choreo Console](https://console.choreo.dev/)
2. Click **"Create"** → **"Component"**
3. Select **"Web Application"**
4. Choose **"Bring Your Own Container Image"** (BYOI)

### 2. Connect GitHub Repository

1. **Repository**: Select `BusLounge/sts-admin-web-app`
2. **Branch**: `main`
3. **Build Context Path**: `/` (root directory)
4. **Dockerfile Path**: `Dockerfile`

### 3. Configure Build Settings

**Build Configuration:**
- **Buildpack Type**: Docker
- **Port**: `8080`
- **Health Check Endpoint**: `/health`

**Build Command** (if using buildpack instead of Docker):
```bash
npm run build
```

**Build Output Directory** (if using buildpack):
```
dist/my-angular-app/browser
```

### 4. Environment Variables

Add the following environment variables in Choreo:

| Variable Name | Value | Description |
|--------------|-------|-------------|
| `BACKEND_API_URL` | `https://your-backend-api.choreo.dev/api/v1` | Your deployed Go backend URL |

**Note**: You'll need to update the Angular auth service to use this environment variable.

### 5. Update API URL in Angular

#### Option 1: Using Environment Files (Recommended)

1. Create `src/environments/environment.prod.ts`:
```typescript
export const environment = {
  production: true,
  apiUrl: 'https://your-backend-api.choreo.dev/api/v1'
};
```

2. Create `src/environments/environment.ts`:
```typescript
export const environment = {
  production: false,
  apiUrl: 'http://localhost:8080/api/v1'
};
```

3. Update `src/app/core/services/admin-auth.service.ts`:
```typescript
import { environment } from '../../../environments/environment';

export class AdminAuthService {
  private readonly API_URL = `${environment.apiUrl}/admin/auth`;
  // ... rest of the code
}
```

4. Update `angular.json` to use environment files:
```json
"configurations": {
  "production": {
    "fileReplacements": [
      {
        "replace": "src/environments/environment.ts",
        "with": "src/environments/environment.prod.ts"
      }
    ],
    // ... rest of config
  }
}
```

#### Option 2: Runtime Configuration (Alternative)

Use `window.location.origin` or a config file loaded at runtime.

### 6. CORS Configuration

Ensure your Go backend allows requests from your Choreo-deployed frontend:

In `DEPLOYsms-auth-backend-go/internal/config/config.go` or environment variables:

```yaml
cors:
  allowed_origins:
    - "https://your-admin-dashboard.choreo.dev"
    - "http://localhost:4200"  # for local development
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "Content-Type"
    - "Authorization"
```

### 7. Deploy

1. Click **"Deploy"** in Choreo console
2. Choreo will:
   - Clone your repository
   - Build the Docker image using the Dockerfile
   - Deploy the container
   - Expose it on port 8080

3. Wait for deployment to complete (usually 2-5 minutes)

### 8. Access Your Application

Once deployed, Choreo will provide you with a URL:
```
https://your-admin-dashboard-<unique-id>.choreo.dev
```

## Testing the Deployment

1. **Health Check**: Visit `https://your-app.choreo.dev/health`
   - Should return: `healthy`

2. **Login Page**: Visit `https://your-app.choreo.dev/login`
   - Try logging in with: `admin@smarttransit.com` / `Admin@123`

3. **API Connection**: Check browser console for any CORS or API connection errors

## Docker Build & Test Locally

Before deploying to Choreo, test the Docker build locally:

```bash
# Build the Docker image
docker build -t admin-dashboard .

# Run the container
docker run -p 8080:8080 admin-dashboard

# Access the app
open http://localhost:8080
```

## Troubleshooting

### Build Fails

**Issue**: `npm install` fails
- **Solution**: Make sure all dependencies in `package.json` are correct
- **Check**: Node version compatibility (requires Node 18+)

**Issue**: Angular build fails
- **Solution**: Run `npm run build` locally first to catch errors
- **Check**: TypeScript errors in the code

### Runtime Issues

**Issue**: Application shows blank page
- **Solution**: Check browser console for errors
- **Check**: Verify nginx.conf is correct
- **Check**: Ensure build output path matches Dockerfile COPY command

**Issue**: API calls fail with CORS error
- **Solution**: Add your Choreo URL to backend CORS allowed origins
- **Check**: Backend is deployed and accessible

**Issue**: Login doesn't work
- **Solution**: Verify API URL is correctly configured
- **Check**: Database migration was run (admin_users table exists)
- **Check**: Backend admin authentication endpoints are working

### Permission Errors

**Issue**: nginx permission denied
- **Solution**: Ensure UID 10014 is used (required by Choreo)
- **Check**: All necessary directories have correct ownership in Dockerfile

## Key Features of This Deployment

✅ **Multi-stage build** - Optimized image size (build artifacts not included)
✅ **Non-root user** - Runs as UID 10014 (Choreo requirement)
✅ **Unprivileged nginx** - Enhanced security
✅ **SPA routing** - All routes fall back to index.html
✅ **Static asset caching** - Optimized performance
✅ **Health check** - Choreo can monitor app health
✅ **Gzip compression** - Faster load times
✅ **Security headers** - XSS, clickjacking protection

## Monitoring & Logs

Access logs in Choreo Console:
1. Go to your component
2. Click **"Logs"** tab
3. View real-time nginx and application logs

## Updates & Redeployment

To deploy updates:
1. Push changes to `main` branch on GitHub
2. Choreo will auto-deploy (if auto-deploy is enabled)
3. Or manually trigger deployment from Choreo console

## Production Checklist

Before going to production:

- [ ] Change default admin password
- [ ] Configure production backend API URL
- [ ] Set up CORS properly
- [ ] Enable HTTPS (Choreo provides this automatically)
- [ ] Test all authentication flows
- [ ] Verify all dashboard features work
- [ ] Set up monitoring and alerts
- [ ] Configure backup strategy for admin user data

## Support

For issues:
- **Choreo Documentation**: https://wso2.com/choreo/docs/
- **GitHub Issues**: https://github.com/BusLounge/sts-admin-web-app/issues

## Additional Resources

- [Choreo Web App Deployment Guide](https://wso2.com/choreo/docs/deploy/deploy-a-web-application/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [NGINX Configuration Guide](https://nginx.org/en/docs/)
