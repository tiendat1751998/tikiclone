#!/bin/bash
# MongoDB Backup Script
# Runs mongodump on replica set via SSH to any member
# Retention: keep last 96 backups (24h at 15-min intervals)

set -euo pipefail

BACKUP_DIR="/home/datdt/backups/mongodb"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="mongodb_backup_${TIMESTAMP}"
TEMP_DIR="/tmp/mongo_backup_${TIMESTAMP}"
RETENTION_COUNT=96

echo "[$(date)] Starting MongoDB backup..."

# Create temp dir on remote, run mongodump, compress and stream back
sshpass -p '123123' ssh -o ConnectTimeout=30 -o StrictHostKeyChecking=no datdt@10.10.10.200 \
  "mkdir -p ${TEMP_DIR} && /usr/local/bin/mongodump --host rs_tiki/10.10.10.200:27017,10.10.10.201:27017,10.10.10.202:27017 --db tiki_catalog --out ${TEMP_DIR} 2>/dev/null && tar czf - -C ${TEMP_DIR} . && rm -rf ${TEMP_DIR}" \
  > "${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"

if [ $? -eq 0 ] && [ -s "${BACKUP_DIR}/${BACKUP_NAME}.tar.gz" ]; then
  SIZE=$(du -h "${BACKUP_DIR}/${BACKUP_NAME}.tar.gz" | cut -f1)
  echo "[$(date)] MongoDB backup completed: ${BACKUP_NAME}.tar.gz (${SIZE})"
else
  echo "[$(date)] MongoDB backup FAILED"
  rm -f "${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"
  exit 1
fi

# Retention: keep only last N backups
cd "${BACKUP_DIR}"
ls -t mongodb_backup_*.tar.gz 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f

REMAINING=$(ls mongodb_backup_*.tar.gz 2>/dev/null | wc -l)
echo "[$(date)] Retention cleanup done. ${REMAINING} backups remaining."
