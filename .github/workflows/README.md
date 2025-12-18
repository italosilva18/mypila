# GitHub Actions Workflows

This directory contains automated CI/CD pipelines for the M2M Financeiro project.

## Workflows

### 1. CI Pipeline (`ci.yml`)

Continuous Integration pipeline that runs on every push and pull request.

**When it runs:**
- Push to `main` or `develop` branches
- Pull requests targeting `main` or `develop` branches

**What it does:**
- Runs Go tests with coverage
- Performs static analysis (go vet, staticcheck, gofmt)
- Runs TypeScript checks and frontend tests
- Builds Docker images
- Scans for security vulnerabilities
- Lints Dockerfiles

**Example trigger:**
```bash
git add .
git commit -m "Add new feature"
git push origin develop
```

### 2. Deploy Pipeline (`deploy.yml`)

Deployment pipeline for staging and production environments.

**When it runs:**
- Push to `main` branch (auto-deploys to staging)
- Git tags matching `v*` (e.g., v1.0.0)
- Manual trigger via GitHub UI

**What it does:**
- Builds multi-architecture Docker images
- Pushes images to GitHub Container Registry
- Deploys to staging automatically
- Deploys to production with manual approval
- Runs health checks and smoke tests
- Creates GitHub releases for tagged versions

**Example triggers:**

Automatic staging deployment:
```bash
git checkout main
git merge develop
git push origin main
```

Production release:
```bash
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

Manual deployment:
1. Go to Actions tab
2. Select "Deploy Pipeline"
3. Click "Run workflow"
4. Choose environment (staging/production)

## Setup Requirements

### 1. Enable GitHub Actions

In repository settings:
- Go to `Settings > Actions > General`
- Enable "Allow all actions and reusable workflows"
- Set workflow permissions to "Read and write permissions"
- Save changes

### 2. Configure Secrets

Go to `Settings > Secrets and variables > Actions` and add:

**Required:**
- `STAGING_HOST` - Staging server hostname
- `STAGING_USER` - SSH username for staging
- `STAGING_SSH_KEY` - Private SSH key for staging
- `PRODUCTION_HOST` - Production server hostname
- `PRODUCTION_USER` - SSH username for production
- `PRODUCTION_SSH_KEY` - Private SSH key for production
- `VITE_API_URL_PROD` - Production API URL
- `JWT_SECRET` - JWT secret key

**Optional:**
- `SLACK_WEBHOOK` - Slack notifications webhook

### 3. Enable GitHub Packages

Docker images are pushed to GitHub Container Registry.

**Make repository public or configure access:**
1. Go to repository settings
2. Enable "Packages" if not already enabled
3. After first workflow run, go to Packages tab
4. Configure package visibility and access

## Status Badges

Add these badges to your README.md:

```markdown
[![CI](https://github.com/YOUR_USERNAME/YOUR_REPO/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_USERNAME/YOUR_REPO/actions/workflows/ci.yml)
[![Deploy](https://github.com/YOUR_USERNAME/YOUR_REPO/actions/workflows/deploy.yml/badge.svg)](https://github.com/YOUR_USERNAME/YOUR_REPO/actions/workflows/deploy.yml)
```

## Monitoring Workflows

### View Workflow Runs

1. Go to repository Actions tab
2. Select workflow from left sidebar
3. Click on specific run to view details

### Debug Failed Workflows

1. Click on failed workflow run
2. Click on failed job
3. Expand failed step to view logs
4. Fix issue and push new commit

### Re-run Workflows

1. Go to failed workflow run
2. Click "Re-run all jobs" or "Re-run failed jobs"

## Workflow Outputs

### Artifacts

CI workflow uploads:
- Backend binary
- Frontend build artifacts
- Coverage reports

Access artifacts:
1. Go to workflow run
2. Scroll to "Artifacts" section
3. Download artifacts

### Docker Images

Images are tagged with:
- `latest` - Latest main branch
- `<branch>-<sha>` - Branch and commit SHA
- `v1.0.0` - Git tag version
- `1.0` - Major.minor version

Pull images:
```bash
docker pull ghcr.io/YOUR_USERNAME/YOUR_REPO/backend:latest
docker pull ghcr.io/YOUR_USERNAME/YOUR_REPO/frontend:latest
```

## Advanced Usage

### Manual Deploy to Staging

```bash
# Trigger via GitHub CLI
gh workflow run deploy.yml -f environment=staging

# Or use curl
curl -X POST \
  -H "Accept: application/vnd.github.v3+json" \
  -H "Authorization: token YOUR_TOKEN" \
  https://api.github.com/repos/YOUR_USERNAME/YOUR_REPO/actions/workflows/deploy.yml/dispatches \
  -d '{"ref":"main","inputs":{"environment":"staging"}}'
```

### Skip CI for Commits

Add `[skip ci]` to commit message:
```bash
git commit -m "Update documentation [skip ci]"
```

### Approval Required Deployments

Production deployments require manual approval:
1. Go to Actions tab
2. Click on running Deploy workflow
3. Click "Review deployments"
4. Select "production" environment
5. Click "Approve and deploy"

## Workflow Customization

### Modify Triggers

Edit workflow files to change when they run:

```yaml
on:
  push:
    branches: [ main, develop, feature/* ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
```

### Add New Jobs

Add jobs to existing workflows:

```yaml
jobs:
  my-new-job:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run custom script
        run: ./scripts/my-script.sh
```

### Environment Variables

Add environment variables:

```yaml
env:
  MY_VAR: value

jobs:
  build:
    env:
      JOB_VAR: value
```

## Best Practices

1. **Keep secrets secure** - Never commit secrets to repository
2. **Use caching** - Cache dependencies for faster builds
3. **Fail fast** - Stop pipeline on first failure
4. **Matrix builds** - Test multiple versions/platforms
5. **Artifact retention** - Set appropriate retention period
6. **Resource limits** - Be mindful of GitHub Actions minutes

## Troubleshooting

### Issue: Workflow doesn't trigger

**Solution:**
- Check branch name matches trigger configuration
- Ensure workflows are enabled in repository settings
- Verify workflow YAML syntax

### Issue: Docker build fails

**Solution:**
- Check Dockerfile syntax
- Verify build context
- Review build logs for specific errors
- Test build locally first

### Issue: Deployment fails

**Solution:**
- Verify all secrets are configured
- Check SSH connectivity to servers
- Review deployment logs
- Test deployment script locally

### Issue: Out of GitHub Actions minutes

**Solution:**
- Upgrade to paid plan
- Optimize workflows to run faster
- Use self-hosted runners
- Reduce frequency of scheduled workflows

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Workflow Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Actions Marketplace](https://github.com/marketplace?type=actions)
- [Self-hosted Runners](https://docs.github.com/en/actions/hosting-your-own-runners)

## Support

For issues with workflows:
1. Check workflow logs in Actions tab
2. Review this documentation
3. Test workflow steps locally
4. Check [DEVOPS.md](../../DEVOPS.md) for more details
