#!/bin/bash
# MySQL Backup Script
# Runs mysqldump on masterdb (10.10.10.200) via SSH
# Retention: keep last 96 backups (24h at 15-min intervals)

set -euo pipefail

BACKUP_DIR="/home/datdt/backups/mysql"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="mysql_backup_${TIMESTAMP}.sql.gz"
RETENTION_COUNT=96

echo "[$(date)] Starting MySQL backup..."

# SSH to masterdb and run mysqldump, compress and stream back
sshpass -p '123123' ssh -o ConnectTimeout=30 -o StrictHostKeyChecking=no datdt@10.10.10.200 \
  "mysqldump -u tiki -p'tiki_dev' --all-databases --single-transaction --routines --triggers --events 2>/dev/null | gzip" \
  > "${BACKUP_DIR}/${BACKUP_FILE}"

if [ $? -eq 0 ] && [ -s "${BACKUP_DIR}/${BACKUP_FILE}" ]; then
  SIZE=$(du -h "${BACKUP_DIR}/${BACKUP_FILE}" | cut -f1)
  echo "[$(date)] MySQL backup completed: ${BACKUP_FILE} (${SIZE})"
else
  echo "[$(date)] MySQL backup FAILED"
  rm -f "${BACKUP_DIR}/${BACKUP_FILE}"
  exit 1
fi

# Retention: keep only last N backups
cd "${BACKUP_DIR}"
ls -t mysql_backup_*.sql.gz 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f

REMAINING=$(ls mysql_backup_*.sql.gz 2>/dev/null | wc -l)
echo "[$(date)] Retention cleanup done. ${REMAINING} backups remaining."
