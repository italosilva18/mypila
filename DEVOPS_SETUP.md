# DevOps Setup - Configuration Summary

## Created Files

### GitHub Actions Workflows
```
.github/workflows/
├── ci.yml          # Continuous Integration pipeline
└── deploy.yml      # Deployment pipeline
```

### Docker Configuration
```
├── docker-compose.yml           # Development environment (updated)
├── docker-compose.prod.yml      # Production environment (new)
├── backend/.dockerignore        # Backend build exclusions (updated)
├── frontend/.dockerignore       # Frontend build exclusions (updated)
└── .env.production.example      # Production environment template (new)
```

### NGINX Configuration
```
nginx/
├── nginx.conf           # Main NGINX configuration
└── conf.d/
    └── default.conf     # Site configuration with reverse proxy
```

### Automation Scripts
```
scripts/
├── deploy.sh           # Deployment automation script
├── health-check.sh     # Service health monitoring
└── backup-db.sh        # Database backup script
```

### Build Automation
```
Makefile                # Build and operations automation
```

### Documentation
```
├── DEVOPS.md           # Complete DevOps documentation
└── DEVOPS_SETUP.md     # This file
```

## Quick Start

### 1. Development
```bash
# Install dependencies
make install

# Start services
make run

# Check health
make check-health

# View logs
make docker-logs
```

### 2. Testing
```bash
# Run all tests
make test

# Run linters
make lint

# Run CI locally
make ci
```

### 3. Building
```bash
# Build all components
make build

# Build Docker images
make docker-build
```

## GitHub Actions Setup

### Required Repository Secrets

Go to: `Settings > Secrets and variables > Actions > New repository secret`

Add the following secrets:

#### Deployment Secrets
- `STAGING_HOST`: Staging server hostname
- `STAGING_USER`: SSH user for staging
- `STAGING_SSH_KEY`: Private SSH key for staging
- `PRODUCTION_HOST`: Production server hostname
- `PRODUCTION_USER`: SSH user for production
- `PRODUCTION_SSH_KEY`: Private SSH key for production

#### Application Secrets
- `VITE_API_URL_PROD`: Production API URL (e.g., https://api.example.com)
- `JWT_SECRET`: JWT secret key (min 32 chars)

#### Notifications (Optional)
- `SLACK_WEBHOOK`: Slack webhook URL for notifications

### GitHub Container Registry

The workflows use GitHub Container Registry (ghcr.io) by default.

**Enable in repository:**
1. Go to `Settings > Actions > General`
2. Scroll to "Workflow permissions"
3. Select "Read and write permissions"
4. Check "Allow GitHub Actions to create and approve pull requests"
5. Save

### Enable GitHub Packages

1. Go to your repository
2. Click on "Packages" in the right sidebar
3. Images will appear after first successful workflow run

## CI/CD Pipeline Overview

### CI Pipeline (Automatic on Push/PR)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Steps:**
1. Backend testing (go test, vet, staticcheck)
2. Frontend testing (TypeScript check, build)
3. Security scanning (Trivy)
4. Docker image building
5. Dockerfile linting (hadolint)

**Duration:** ~5-10 minutes

### Deploy Pipeline (Automatic/Manual)

**Triggers:**
- Push to `main` branch → Deploy to staging
- Git tag `v*` → Deploy to production
- Manual trigger → Choose staging or production

**Steps:**
1. Build multi-architecture images (amd64/arm64)
2. Push to GitHub Container Registry
3. Scan images for vulnerabilities
4. Deploy to staging (automatic)
5. Deploy to production (manual approval)
6. Run smoke tests
7. Send notifications

**Duration:** ~15-20 minutes

## Docker Compose Configurations

### Development (docker-compose.yml)

**Features:**
- Local development optimized
- Hot reload support
- Debug ports exposed
- Minimal resource limits

**Services:**
- MongoDB: port 27018
- Backend: port 8081
- Frontend: port 3333

**Start:**
```bash
make run
# or
docker-compose up -d
```

### Production (docker-compose.prod.yml)

**Features:**
- Production optimized
- Service replicas (2x backend, 2x frontend)
- NGINX reverse proxy
- SSL/TLS support ready
- Enhanced monitoring
- Rolling updates
- Automatic restarts

**Services:**
- MongoDB (internal only)
- Backend (internal only)
- Frontend (internal only)
- NGINX: ports 80, 443

**Start:**
```bash
VERSION=1.0.0 docker-compose -f docker-compose.prod.yml up -d
```

## Makefile Commands Reference

### Development
```bash
make help              # Show all commands
make install           # Install dependencies
make dev-backend       # Run backend in dev mode
make dev-frontend      # Run frontend in dev mode
```

### Building
```bash
make build             # Build all components
make docker-build      # Build Docker images
```

### Testing
```bash
make test              # Run all tests
make test-backend      # Run backend tests only
make test-frontend     # Run frontend tests only
make lint              # Run all linters
make format            # Format code
make ci                # Run full CI pipeline locally
```

### Docker Operations
```bash
make docker-up         # Start services
make docker-down       # Stop services
make docker-restart    # Restart services
make docker-logs       # View all logs
make docker-ps         # Show running containers
```

### Database Operations
```bash
make backup            # Backup MongoDB
make restore           # Restore from backup
make migrate           # Run migrations
```

### Maintenance
```bash
make clean             # Clean build artifacts
make clean-all         # Deep clean (includes node_modules)
make check-health      # Check service health
make stats             # Show project statistics
make version           # Show version info
```

## Deployment Workflow

### Staging Deployment

1. **Automatic:** Push to `main` branch
   ```bash
   git checkout main
   git merge develop
   git push origin main
   ```

2. **Manual:** Use GitHub Actions UI
   - Go to Actions tab
   - Select "Deploy Pipeline"
   - Click "Run workflow"
   - Select "staging"

### Production Deployment

1. **Create Release Tag:**
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

2. **Approve Deployment:**
   - GitHub Actions will build images
   - Go to Actions tab
   - Find the running workflow
   - Approve the production deployment

3. **Alternative - Manual Script:**
   ```bash
   ./scripts/deploy.sh production
   ```

### Rollback

If deployment fails:

1. **Automatic:** Pipeline will rollback automatically
2. **Manual:**
   ```bash
   # SSH to server
   ssh user@production-host

   # Rollback to previous version
   cd /app
   VERSION=previous docker-compose -f docker-compose.prod.yml up -d
   ```

## Health Monitoring

### Automated Checks

All services have health checks configured:

- **MongoDB:** ping command every 10s
- **Backend:** HTTP GET /health every 15s
- **Frontend:** HTTP GET / every 30s

### Manual Checks

```bash
# Local environment
make check-health

# Staging/Production
./scripts/health-check.sh staging
./scripts/health-check.sh production
```

### Health Endpoints

- Backend: `GET /health` → `{"status": "ok"}`
- Frontend: `GET /` → 200 OK
- MongoDB: `db.runCommand({ping:1})`

## Backup Strategy

### Automated Backups

Configure cron job on production server:

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /app/scripts/backup-db.sh production
```

### Manual Backup

```bash
# Local
make backup

# Production
./scripts/backup-db.sh production
```

### Restore Backup

```bash
# From latest backup
make restore

# From specific backup
BACKUP=backup-20240101-120000 make restore
```

## Security Checklist

- [ ] Configure all GitHub Secrets
- [ ] Use strong passwords in .env files
- [ ] Enable 2FA on GitHub
- [ ] Restrict SSH access to deployment servers
- [ ] Configure firewall rules
- [ ] Set up SSL/TLS certificates (Let's Encrypt recommended)
- [ ] Enable GitHub security alerts
- [ ] Configure Dependabot
- [ ] Set up audit logging
- [ ] Regular security scans with Trivy

## Monitoring Setup (Recommended)

### Metrics Collection
- **Prometheus:** Container metrics
- **Grafana:** Dashboards and visualization

### Error Tracking
- **Sentry:** Application error tracking
- **New Relic:** APM and performance monitoring

### Logging
- **ELK Stack:** Centralized logging
- **Loki:** Log aggregation

### Uptime Monitoring
- **UptimeRobot:** Service availability
- **Pingdom:** Performance monitoring

## Resource Requirements

### Development
- **CPU:** 2 cores minimum
- **RAM:** 4GB minimum
- **Disk:** 10GB available

### Production (Recommended)
- **CPU:** 4 cores minimum
- **RAM:** 8GB minimum
- **Disk:** 50GB available (including backups)
- **Network:** 100 Mbps minimum

## Troubleshooting

### Common Issues

#### 1. CI Pipeline Fails

```bash
# Check logs in GitHub Actions
# Run CI locally to debug
make ci
```

#### 2. Docker Build Fails

```bash
# Clear cache and rebuild
docker system prune -a
make docker-build
```

#### 3. Services Won't Start

```bash
# Check logs
make docker-logs

# Restart services
make docker-restart
```

#### 4. Health Checks Failing

```bash
# Check service status
make docker-ps

# Check individual service logs
make docker-logs-backend
make docker-logs-frontend
```

## Next Steps

1. **Configure GitHub Secrets** (see above)
2. **Test CI Pipeline:** Create a test commit
3. **Review NGINX config:** Adjust for your domain
4. **Set up SSL certificates:** Use Let's Encrypt
5. **Configure monitoring:** Set up metrics collection
6. **Schedule backups:** Configure cron jobs
7. **Test deployment:** Deploy to staging
8. **Document runbooks:** Create incident response procedures

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [NGINX Documentation](https://nginx.org/en/docs/)
- [Makefile Tutorial](https://makefiletutorial.com/)

## Support

For questions or issues:
1. Check [DEVOPS.md](DEVOPS.md) for detailed documentation
2. Review logs: `make docker-logs`
3. Run health check: `make check-health`
4. Check GitHub Actions logs

---

**Last Updated:** 2025-12-16
**Version:** 1.0.0
