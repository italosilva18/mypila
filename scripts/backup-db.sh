#!/bin/bash

# M2M Financeiro - Database Backup Script
# Usage: ./scripts/backup-db.sh [environment]

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ENVIRONMENT=${1:-local}
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_NAME="backup-${ENVIRONMENT}-${TIMESTAMP}"

echo -e "${BLUE}Starting database backup...${NC}"

# Create backup directory
mkdir -p ${BACKUP_DIR}

# Backup based on environment
case $ENVIRONMENT in
    local)
        echo -e "${YELLOW}Backing up local database...${NC}"
        docker exec m2m-mongodb mongodump --out=/tmp/backup --db=m2m_financeiro
        docker cp m2m-mongodb:/tmp/backup ${BACKUP_DIR}/${BACKUP_NAME}
        docker exec m2m-mongodb rm -rf /tmp/backup
        ;;
    staging|production)
        echo -e "${YELLOW}Backing up ${ENVIRONMENT} database...${NC}"
        # SSH to server and backup
        if [ "$ENVIRONMENT" == "staging" ]; then
            HOST=$STAGING_HOST
            USER=$STAGING_USER
        else
            HOST=$PRODUCTION_HOST
            USER=$PRODUCTION_USER
        fi

        ssh ${USER}@${HOST} "docker exec m2m-mongodb-prod mongodump --out=/tmp/backup --db=m2m_financeiro"
        scp -r ${USER}@${HOST}:/tmp/backup ${BACKUP_DIR}/${BACKUP_NAME}
        ssh ${USER}@${HOST} "rm -rf /tmp/backup"
        ;;
    *)
        echo "Unknown environment: $ENVIRONMENT"
        exit 1
        ;;
esac

# Compress backup
echo -e "${YELLOW}Compressing backup...${NC}"
cd ${BACKUP_DIR}
tar -czf ${BACKUP_NAME}.tar.gz ${BACKUP_NAME}
rm -rf ${BACKUP_NAME}

# Get backup size
BACKUP_SIZE=$(du -h ${BACKUP_NAME}.tar.gz | cut -f1)

echo -e "${GREEN}Backup completed successfully!${NC}"
echo -e "Backup file: ${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"
echo -e "Backup size: ${BACKUP_SIZE}"

# Keep only last 10 backups
echo -e "${YELLOW}Cleaning old backups...${NC}"
ls -t ${BACKUP_DIR}/backup-${ENVIRONMENT}-*.tar.gz | tail -n +11 | xargs -r rm
echo -e "${GREEN}Old backups cleaned!${NC}"

# Optional: Upload to cloud storage
# echo "Uploading to S3..."
# aws s3 cp ${BACKUP_DIR}/${BACKUP_NAME}.tar.gz s3://your-bucket/backups/

exit 0
