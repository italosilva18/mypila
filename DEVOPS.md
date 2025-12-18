# DevOps Documentation

## Table of Contents
- [CI/CD Pipeline](#cicd-pipeline)
- [Docker Configuration](#docker-configuration)
- [Deployment](#deployment)
- [Monitoring](#monitoring)
- [Backup & Recovery](#backup--recovery)
- [Troubleshooting](#troubleshooting)

## CI/CD Pipeline

### GitHub Actions Workflows

#### CI Pipeline (`.github/workflows/ci.yml`)
Runs on every push and pull request to `main` and `develop` branches.

**Jobs:**
- **backend-test**: Go tests, linting, static analysis, and build
- **frontend-test**: TypeScript check, npm tests, and build
- **security-scan**: Trivy vulnerability scanning
- **docker-build**: Docker image build test
- **lint-dockerfile**: Dockerfile linting with hadolint

**Triggers:**
```yaml
on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
```

#### Deploy Pipeline (`.github/workflows/deploy.yml`)
Handles building and deploying to staging and production.

**Jobs:**
- **build-and-push**: Build multi-arch Docker images and push to registry
- **deploy-staging**: Deploy to staging environment
- **deploy-production**: Deploy to production (requires manual approval)

**Triggers:**
```yaml
on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]
  workflow_dispatch:
```

### Required Secrets

Configure these secrets in GitHub repository settings:

```bash
# Container Registry
GITHUB_TOKEN  # Automatically provided

# Deployment
STAGING_HOST
STAGING_USER
STAGING_SSH_KEY
PRODUCTION_HOST
PRODUCTION_USER
PRODUCTION_SSH_KEY

# Application
VITE_API_URL_PROD
JWT_SECRET

# Notifications (optional)
SLACK_WEBHOOK
```

## Docker Configuration

### Development (`docker-compose.yml`)

**Services:**
- MongoDB (port 27018)
- Backend (port 8081)
- Frontend (port 3333)

**Features:**
- Health checks for all services
- Resource limits
- Automatic restart policies
- Log rotation
- Dependency management

**Usage:**
```bash
# Start all services
make run
# or
docker-compose up -d

# View logs
make docker-logs

# Stop services
make docker-down
```

### Production (`docker-compose.prod.yml`)

**Additional features:**
- NGINX reverse proxy with SSL support
- Service replicas for high availability
- Rolling updates with rollback capability
- Enhanced security and resource limits
- Production-grade logging

**Usage:**
```bash
# Deploy to production
./scripts/deploy.sh production

# Health check
./scripts/health-check.sh production
```

### Docker Images

#### Backend Image
- **Base**: golang:1.24-alpine (build), alpine:latest (runtime)
- **Multi-stage**: Yes
- **Size**: ~20MB
- **Platforms**: linux/amd64, linux/arm64

#### Frontend Image
- **Base**: node:20-alpine (build), nginx:alpine (runtime)
- **Multi-stage**: Yes
- **Size**: ~25MB
- **Platforms**: linux/amd64, linux/arm64

### Resource Limits

| Service  | CPU Limit | Memory Limit | CPU Reserve | Memory Reserve |
|----------|-----------|--------------|-------------|----------------|
| MongoDB  | 1.0       | 1G           | 0.5         | 512M           |
| Backend  | 0.5       | 512M         | 0.25        | 256M           |
| Frontend | 0.25      | 256M         | 0.1         | 128M           |

## Deployment

### Makefile Commands

```bash
# Development
make install          # Install dependencies
make build           # Build all components
make test            # Run all tests
make lint            # Run linters
make dev-backend     # Run backend in dev mode
make dev-frontend    # Run frontend in dev mode

# Docker
make docker-build    # Build Docker images
make docker-up       # Start all services
make docker-down     # Stop all services
make docker-logs     # View logs

# Operations
make check-health    # Check service health
make backup          # Backup database
make restore         # Restore from backup
make clean           # Clean build artifacts
make ci              # Run CI pipeline locally
```

### Manual Deployment

#### Staging
```bash
# 1. Build and tag images
docker-compose build
docker tag m2m-backend:latest registry.example.com/m2m-backend:staging
docker tag m2m-frontend:latest registry.example.com/m2m-frontend:staging

# 2. Push to registry
docker push registry.example.com/m2m-backend:staging
docker push registry.example.com/m2m-frontend:staging

# 3. Deploy to staging server
./scripts/deploy.sh staging
```

#### Production
```bash
# 1. Create release tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# 2. GitHub Actions will automatically:
#    - Build multi-arch images
#    - Run security scans
#    - Push to registry
#    - Deploy to staging

# 3. Approve production deployment in GitHub
#    Or manually deploy:
./scripts/deploy.sh production
```

### Rollback

```bash
# Rollback to previous version
docker-compose -f docker-compose.prod.yml down
VERSION=previous_version docker-compose -f docker-compose.prod.yml up -d

# Or restore from backup
./scripts/restore-backup.sh production
```

## Monitoring

### Health Checks

```bash
# All services
./scripts/health-check.sh local

# Individual services
curl http://localhost:8081/health  # Backend
curl http://localhost:3333         # Frontend
docker exec m2m-mongodb mongosh --eval "db.runCommand({ping:1})"  # MongoDB
```

### Logs

```bash
# View all logs
make docker-logs

# View specific service logs
make docker-logs-backend
make docker-logs-frontend
make docker-logs-mongodb

# Follow logs
docker-compose logs -f --tail=100
```

### Metrics

Docker Compose automatically tracks:
- CPU usage
- Memory usage
- Network I/O
- Disk I/O

View metrics:
```bash
docker stats
```

### Alerts (To Configure)

Recommended monitoring tools:
- **Prometheus + Grafana**: Metrics and dashboards
- **Sentry**: Error tracking
- **DataDog**: Full-stack monitoring
- **New Relic**: APM

## Backup & Recovery

### Automated Backups

```bash
# Manual backup
make backup
# or
./scripts/backup-db.sh local

# Scheduled backups (crontab)
0 2 * * * /path/to/scripts/backup-db.sh production
```

### Restore

```bash
# Restore from latest backup
make restore

# Restore from specific backup
docker cp ./backups/backup-20240101-120000 m2m-mongodb:/tmp/restore
docker exec m2m-mongodb mongorestore /tmp/restore
```

### Backup Strategy

- **Frequency**: Daily at 2 AM
- **Retention**: Last 10 backups
- **Location**: Local + S3 (recommended)
- **Testing**: Monthly restore tests

## Troubleshooting

### Common Issues

#### Services won't start
```bash
# Check logs
make docker-logs

# Check service status
docker-compose ps

# Restart services
make docker-restart
```

#### Database connection errors
```bash
# Check MongoDB health
docker exec m2m-mongodb mongosh --eval "db.runCommand({ping:1})"

# Check environment variables
docker-compose config

# Restart MongoDB
docker-compose restart mongodb
```

#### Out of memory
```bash
# Check memory usage
docker stats

# Increase limits in docker-compose.yml
# Or restart Docker daemon
```

#### Port conflicts
```bash
# Check ports in use
netstat -tulpn | grep LISTEN

# Change ports in docker-compose.yml or .env
```

### Performance Optimization

1. **Enable BuildKit**:
   ```bash
   export DOCKER_BUILDKIT=1
   ```

2. **Use layer caching**:
   - Workflows already use GitHub Actions cache
   - For local builds: `docker-compose build --cache`

3. **Prune unused resources**:
   ```bash
   docker system prune -f
   docker volume prune -f
   ```

### Security Best Practices

1. **Never commit secrets**:
   - Use `.env` files (gitignored)
   - Use GitHub Secrets for CI/CD

2. **Regular updates**:
   ```bash
   # Update base images
   docker-compose pull
   docker-compose build --no-cache
   ```

3. **Scan for vulnerabilities**:
   ```bash
   # Manual scan
   docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
     aquasec/trivy image m2m-backend:latest
   ```

4. **Use non-root users** (already implemented in Dockerfiles)

5. **Enable firewall rules**:
   - Only expose necessary ports
   - Use internal networks for service communication

## Infrastructure as Code

### Future Improvements

Consider implementing:

1. **Kubernetes/Helm**:
   - For orchestration at scale
   - Auto-scaling
   - Self-healing

2. **Terraform**:
   - Infrastructure provisioning
   - Multi-cloud support

3. **Ansible**:
   - Configuration management
   - Automated deployment

4. **Service Mesh**:
   - Istio or Linkerd
   - Advanced traffic management
   - Enhanced observability

## Support

For issues or questions:
- Check logs: `make docker-logs`
- Run health check: `./scripts/health-check.sh`
- Review this documentation
- Contact DevOps team
