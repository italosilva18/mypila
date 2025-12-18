#!/bin/bash

# M2M Financeiro - Health Check Script
# Usage: ./scripts/health-check.sh [environment]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ENVIRONMENT=${1:-local}

case $ENVIRONMENT in
    local)
        BACKEND_URL="http://localhost:8081"
        FRONTEND_URL="http://localhost:3333"
        MONGO_HOST="localhost"
        MONGO_PORT="27018"
        ;;
    staging)
        BACKEND_URL="https://staging-api.example.com"
        FRONTEND_URL="https://staging.example.com"
        ;;
    production)
        BACKEND_URL="https://api.example.com"
        FRONTEND_URL="https://example.com"
        ;;
    *)
        echo -e "${RED}Unknown environment: $ENVIRONMENT${NC}"
        exit 1
        ;;
esac

echo -e "${YELLOW}Running health checks for $ENVIRONMENT environment...${NC}\n"

# Backend health check
echo -n "Backend health: "
if curl -sf "${BACKEND_URL}/health" > /dev/null; then
    echo -e "${GREEN}✓ Healthy${NC}"
    BACKEND_STATUS=0
else
    echo -e "${RED}✗ Unhealthy${NC}"
    BACKEND_STATUS=1
fi

# Frontend health check
echo -n "Frontend health: "
if curl -sf "${FRONTEND_URL}" > /dev/null; then
    echo -e "${GREEN}✓ Healthy${NC}"
    FRONTEND_STATUS=0
else
    echo -e "${RED}✗ Unhealthy${NC}"
    FRONTEND_STATUS=1
fi

# MongoDB health check (local only)
if [ "$ENVIRONMENT" == "local" ]; then
    echo -n "MongoDB health: "
    if docker exec m2m-mongodb mongosh --quiet --eval "db.runCommand({ping:1}).ok" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Healthy${NC}"
        MONGO_STATUS=0
    else
        echo -e "${RED}✗ Unhealthy${NC}"
        MONGO_STATUS=1
    fi
fi

# Docker containers check (local only)
if [ "$ENVIRONMENT" == "local" ]; then
    echo -e "\nDocker containers:"
    docker-compose ps
fi

# Overall status
echo -e "\n${YELLOW}Overall Status:${NC}"
TOTAL_STATUS=$((BACKEND_STATUS + FRONTEND_STATUS + ${MONGO_STATUS:-0}))

if [ $TOTAL_STATUS -eq 0 ]; then
    echo -e "${GREEN}All services are healthy!${NC}"
    exit 0
else
    echo -e "${RED}Some services are unhealthy!${NC}"
    exit 1
fi
