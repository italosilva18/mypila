# Quick Reference Guide

## Essential Commands

### Development
```bash
make run              # Start all services
make docker-logs      # View logs
make check-health     # Check service health
make docker-down      # Stop all services
```

### Testing
```bash
make test             # Run all tests
make lint             # Run linters
make ci               # Run full CI locally
```

### Building
```bash
make build            # Build all
make docker-build     # Build Docker images
```

### Database
```bash
make backup           # Backup database
make restore          # Restore database
```

## Service URLs

| Service | Development | Production |
|---------|------------|------------|
| Frontend | http://localhost:3333 | https://yourdomain.com |
| Backend API | http://localhost:8081/api | https://yourdomain.com/api |
| MongoDB | localhost:27018 | Internal only |
| Health Check | http://localhost:8081/health | https://yourdomain.com/health |

## GitHub Actions

### Trigger CI
```bash
git add .
git commit -m "feature: new feature"
git push origin develop
```

### Deploy to Staging
```bash
git checkout main
git merge develop
git push origin main
```

### Deploy to Production
```bash
git tag -a v1.0.0 -m "Release 1.0.0"
git push origin v1.0.0
```

## Docker Commands

### Start Services
```bash
docker-compose up -d
docker-compose ps
```

### View Logs
```bash
docker-compose logs -f
docker-compose logs -f backend
docker-compose logs -f frontend
```

### Restart Services
```bash
docker-compose restart
docker-compose restart backend
```

### Stop Services
```bash
docker-compose down
docker-compose down -v  # Remove volumes too
```

## Health Checks

### Backend
```bash
curl http://localhost:8081/health
```

### Frontend
```bash
curl http://localhost:3333
```

### MongoDB
```bash
docker exec m2m-mongodb mongosh --eval "db.runCommand({ping:1})"
```

## Troubleshooting

### Services won't start
```bash
make docker-logs
make docker-down
make docker-up
```

### Port conflicts
```bash
# Change ports in .env or docker-compose.yml
netstat -tulpn | grep LISTEN
```

### Out of memory
```bash
docker stats
docker system prune -f
```

### Database issues
```bash
docker-compose restart mongodb
make backup
make restore
```

## File Locations

| Type | Location |
|------|----------|
| CI/CD Workflows | `.github/workflows/` |
| Docker Configs | `docker-compose*.yml` |
| NGINX Config | `nginx/` |
| Scripts | `scripts/` |
| Documentation | `DEVOPS*.md` |
| Build Tool | `Makefile` |

## Environment Variables

### Development (.env)
```env
MONGO_ROOT_USER=admin
MONGO_ROOT_PASSWORD=admin123
VITE_API_URL=http://localhost:8081/api
```

### Production (.env.production)
```env
MONGO_ROOT_USER=admin
MONGO_ROOT_PASSWORD=your_secure_password
JWT_SECRET=your_jwt_secret_min_32_chars
VITE_API_URL=https://yourdomain.com/api
```

## GitHub Secrets Required

- `STAGING_HOST`
- `STAGING_USER`
- `STAGING_SSH_KEY`
- `PRODUCTION_HOST`
- `PRODUCTION_USER`
- `PRODUCTION_SSH_KEY`
- `VITE_API_URL_PROD`
- `JWT_SECRET`
- `SLACK_WEBHOOK` (optional)

## Common Tasks

### Install Dependencies
```bash
make install
```

### Run Tests
```bash
make test-backend
make test-frontend
```

### Format Code
```bash
make format
```

### Clean Build Artifacts
```bash
make clean
```

### Check Version
```bash
make version
```

### View Help
```bash
make help
```

## Emergency Commands

### Rollback Deployment
```bash
VERSION=previous docker-compose -f docker-compose.prod.yml up -d
```

### Force Rebuild
```bash
docker-compose build --no-cache
docker-compose up -d --force-recreate
```

### Reset Everything
```bash
make clean-all
docker system prune -a
make install
make run
```

## Documentation Links

- [Complete DevOps Guide](DEVOPS.md)
- [Setup Instructions](DEVOPS_SETUP.md)
- [CI/CD Summary](CICD_SUMMARY.md)
- [GitHub Workflows](.github/workflows/README.md)
- [Main README](README.md)

## Support

1. Check logs: `make docker-logs`
2. Check health: `make check-health`
3. Review documentation
4. Check GitHub Actions logs

---

**Tip:** Bookmark this page for quick access!
