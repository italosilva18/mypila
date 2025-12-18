#!/bin/bash

# M2M Financeiro - Deployment Script
# Usage: ./scripts/deploy.sh [staging|production]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check arguments
if [ "$#" -ne 1 ]; then
    log_error "Usage: $0 [staging|production]"
    exit 1
fi

ENVIRONMENT=$1

if [ "$ENVIRONMENT" != "staging" ] && [ "$ENVIRONMENT" != "production" ]; then
    log_error "Environment must be 'staging' or 'production'"
    exit 1
fi

# Confirmation for production
if [ "$ENVIRONMENT" == "production" ]; then
    log_warning "You are about to deploy to PRODUCTION!"
    read -p "Are you sure? (yes/no): " -r
    echo
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        log_info "Deployment cancelled"
        exit 0
    fi
fi

log_info "Starting deployment to $ENVIRONMENT..."

# Load environment variables
if [ -f ".env.$ENVIRONMENT" ]; then
    log_info "Loading environment variables from .env.$ENVIRONMENT"
    export $(cat .env.$ENVIRONMENT | grep -v '^#' | xargs)
else
    log_error "Environment file .env.$ENVIRONMENT not found"
    exit 1
fi

# Get version
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
log_info "Deploying version: $VERSION"

# Build Docker images
log_info "Building Docker images..."
docker-compose -f docker-compose.yml build --no-cache

# Tag images
log_info "Tagging images..."
docker tag m2m-backend:latest ${REGISTRY}/${BACKEND_IMAGE_NAME}:${VERSION}
docker tag m2m-backend:latest ${REGISTRY}/${BACKEND_IMAGE_NAME}:latest
docker tag m2m-frontend:latest ${REGISTRY}/${FRONTEND_IMAGE_NAME}:${VERSION}
docker tag m2m-frontend:latest ${REGISTRY}/${FRONTEND_IMAGE_NAME}:latest

# Push to registry
log_info "Pushing images to registry..."
docker push ${REGISTRY}/${BACKEND_IMAGE_NAME}:${VERSION}
docker push ${REGISTRY}/${BACKEND_IMAGE_NAME}:latest
docker push ${REGISTRY}/${FRONTEND_IMAGE_NAME}:${VERSION}
docker push ${REGISTRY}/${FRONTEND_IMAGE_NAME}:latest

log_success "Images pushed successfully!"

# Deploy to server
if [ "$ENVIRONMENT" == "staging" ]; then
    DEPLOY_HOST=$STAGING_HOST
    DEPLOY_USER=$STAGING_USER
else
    DEPLOY_HOST=$PRODUCTION_HOST
    DEPLOY_USER=$PRODUCTION_USER
fi

log_info "Deploying to $DEPLOY_HOST..."

# SSH and deploy
ssh ${DEPLOY_USER}@${DEPLOY_HOST} << EOF
    set -e
    cd /app

    # Backup current version
    echo "Creating backup..."
    docker-compose -f docker-compose.prod.yml ps -q | xargs docker inspect --format='{{.Config.Image}}' > backup-${VERSION}.txt

    # Pull latest images
    echo "Pulling latest images..."
    export VERSION=${VERSION}
    docker-compose -f docker-compose.prod.yml pull

    # Run database migrations if needed
    # docker-compose -f docker-compose.prod.yml run --rm backend /app/migrate

    # Rolling update
    echo "Performing rolling update..."
    docker-compose -f docker-compose.prod.yml up -d --no-deps --force-recreate backend
    sleep 10
    docker-compose -f docker-compose.prod.yml up -d --no-deps --force-recreate frontend

    # Health check
    echo "Running health checks..."
    sleep 15
    curl -f http://localhost:8080/health || (echo "Backend health check failed" && exit 1)
    curl -f http://localhost:3333 || (echo "Frontend health check failed" && exit 1)

    # Cleanup old images
    echo "Cleaning up old images..."
    docker image prune -f

    echo "Deployment completed successfully!"
EOF

if [ $? -eq 0 ]; then
    log_success "Deployment to $ENVIRONMENT completed successfully!"

    # Create Git tag if production
    if [ "$ENVIRONMENT" == "production" ]; then
        log_info "Creating Git tag..."
        git tag -a "v${VERSION}" -m "Production release ${VERSION}"
        git push origin "v${VERSION}"
        log_success "Git tag created and pushed"
    fi
else
    log_error "Deployment failed!"

    # Rollback
    log_warning "Initiating rollback..."
    ssh ${DEPLOY_USER}@${DEPLOY_HOST} << EOF
        set -e
        cd /app

        # Restore from backup
        if [ -f backup-${VERSION}.txt ]; then
            echo "Rolling back to previous version..."
            # Implement rollback logic here
            docker-compose -f docker-compose.prod.yml down
            docker-compose -f docker-compose.prod.yml up -d
        fi
EOF

    exit 1
fi

log_success "All done!"
