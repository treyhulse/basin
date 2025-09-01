# Railway Deployment Guide

This guide will help you deploy your Go backend to Railway with a PostgreSQL database.

## Prerequisites

1. A Railway account (https://railway.app)
2. Your Go backend code in a Git repository
3. Basic understanding of Railway's deployment process

## Step 1: Set Up Railway Project

1. Go to [Railway.app](https://railway.app) and create a new project
2. Choose "Deploy from GitHub repo" and select your repository
3. Railway will automatically detect it's a Go project

## Step 2: Add PostgreSQL Database

1. In your Railway project dashboard, click "New Service"
2. Select "Database" → "PostgreSQL"
3. Railway will automatically create a PostgreSQL database service
4. Note the service name (e.g., "postgresql")

## Step 3: Configure Environment Variables

Railway automatically provides these environment variables:

- `DATABASE_URL`: Full PostgreSQL connection string
- `PORT`: Port number for your application
- `RAILWAY_STATIC_URL`: Static asset URL (if needed)

### Optional Environment Variables

You can add these in Railway's Variables tab:

```
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY=24h
SERVER_MODE=production
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=secure-password
ADMIN_FIRST_NAME=Admin
ADMIN_LAST_NAME=User
```

## Step 4: Deploy Your Application

1. Railway will automatically deploy your application when you push to your main branch
2. The deployment will use the `DATABASE_URL` environment variable automatically
3. Your Go application will connect to the PostgreSQL database using the parsed connection string

## Step 5: Verify Deployment

1. Check the deployment logs in Railway dashboard
2. Test your health endpoint: `https://your-app.railway.app/health`
3. Verify database connection by checking logs for "Successfully connected to database"

## Database Migration

If you need to run database migrations:

1. Add a migration script to your Railway project
2. Or run migrations manually using Railway's CLI:

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login to Railway
railway login

# Connect to your project
railway link

# Run migrations (if you have a migration script)
railway run go run migrations/migrate.go
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check that `DATABASE_URL` is set correctly
   - Verify the PostgreSQL service is running
   - Check SSL mode (Railway requires SSL)

2. **Port Binding Issues**
   - Ensure your app listens on the `PORT` environment variable
   - Check that Railway's `PORT` variable is set

3. **SSL Connection Issues**
   - Railway requires SSL connections
   - The code automatically sets `sslmode=require` for Railway deployments

### Debugging

1. Check Railway deployment logs
2. Verify environment variables are set correctly
3. Test database connection locally with the same `DATABASE_URL`

## Environment Variable Priority

The application uses this priority for database configuration:

1. `DATABASE_URL` (Railway provides this)
2. Individual `DB_*` variables (for local development)
3. Default values

## Security Notes

1. **Never commit sensitive environment variables** to your repository
2. **Use Railway's Variables tab** for sensitive configuration
3. **Rotate JWT secrets** regularly in production
4. **Use strong passwords** for admin accounts

## Monitoring

1. Use Railway's built-in monitoring
2. Check application logs regularly
3. Monitor database performance
4. Set up alerts for critical errors

## Scaling

1. Railway automatically scales based on traffic
2. Monitor resource usage in the dashboard
3. Consider upgrading your PostgreSQL plan if needed

## Support

If you encounter issues:

1. Check Railway's documentation: https://docs.railway.app
2. Review your application logs
3. Verify environment variable configuration
4. Test database connectivity
