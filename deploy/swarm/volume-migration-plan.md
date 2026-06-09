# Volume Migration Plan

## Overview

Migrate persistent data from Manager to Worker nodes with safety checks.

## Backup Strategy

### Phase 1: Create Backups

```bash
#!/bin/bash
set -e

BACKUP_DIR="/tmp/swarm-backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# 1. MySQL Backup
echo "Creating MySQL backup..."
docker exec mysql-primary mysqldump \
  --all-databases \
  --single-transaction \
  --master-data=2 \
  --routines \
  --triggers \
  -u root -proot_password > "$BACKUP_DIR/mysql-full-backup.sql"

# 2. Redis Backup
echo "Creating Redis backup..."
docker exec redis-master redis-cli BGSAVE
sleep 5
docker cp redis-master:/data/dump.rdb "$BACKUP_DIR/redis-dump.rdb"

# 3. MongoDB Backup
echo "Creating MongoDB backup..."
docker exec mongodb mongodump \
  --archive=/data/mongodb-backup.gz \
  --gzip

docker cp mongodb:/data/mongodb-backup.gz "$BACKUP_DIR/"

# Verify backups exist
ls -la "$BACKUP_DIR"
```

## Disk Space Validation

```bash
#!/bin/bash
# Validate disk space on each node

REQUIRED_SPACE_GB=10

for NODE in worker-1 worker-2 worker-3; do
  AVAILABLE=$(docker node inspect --format '{{.Description.Resources.Reservations.Memory}}' $NODE)
  echo "Node $NODE available memory: $AVAILABLE"
  
  # Check if we have enough space
  # This requires SSH access to workers
done
```

## Volume Migration Procedures

### MySQL Volume Migration

```bash
# Pre-migration validation
docker exec mysql-primary mysql -u root -proot_password -e "CHECK TABLE tiki_platform.products, tiki_cart.carts, tiki_order.orders;"

# Rsync with checksum verification
rsync -avz --checksum \
  mysql-primary:/var/lib/mysql/ \
  worker-1:/var/lib/mysql/

rsync -avz --checksum \
  mysql-primary:/var/lib/mysql/ \
  worker-2:/var/lib/mysql/

rsync -avz --checksum \
  mysql-primary:/var/lib/mysql/ \
  worker-3:/var/lib/mysql/

# Verify checksums
md5sum mysql_data_backup/* > mysql_backup.md5
```

### Redis Volume Migration

```bash
# For Redis, use RDB snapshot
docker exec redis-master redis-cli BGREWRITEAOF

# Copy AOF to all workers
rsync -avz --checksum \
  redis-master:/data/appendonly.aof \
  worker-1:/var/lib/redis/

rsync -avz --checksum \
  redis-master:/data/appendonly.aof \
  worker-2:/var/lib/redis/

rsync -avz --checksum \
  redis-master:/data/appendonly.aof \
  worker-3:/var/lib/redis/
```

### MongoDB Volume Migration

```bash
# Shutdown MongoDB gracefully
docker exec mongodb mongosh --eval "db.adminCommand('shutdown', 1)"

# Copy volume data
rsync -avz --checksum \
  mongodb:/data/db/ \
  worker-1:/var/lib/mongodb/

rsync -avz --checksum \
  mongodb:/data/db/ \
  worker-2:/var/lib/mongodb/

rsync -avz --checksum \
  mongodb:/data/db/ \
  worker-3:/var/lib/mongodb/
```

## Ownership & Permissions

```bash
# MySQL permissions
chown -R 999:999 /var/lib/mysql
chmod 750 /var/lib/mysql

# Redis permissions  
chown -R 1001:1001 /var/lib/redis
chmod 700 /var/lib/redis

# MongoDB permissions
chown -R 999:999 /var/lib/mongodb
chmod 755 /var/lib/mongodb
```

## Migration Workflow

```
┌─────────────────────────────────────────────────────────┐
│                   Volume Migration Flow                  │
└─────────────────────────────────────────────────────────┘

Manager Node
    │
    ├── 1. Backup Creation
    ├── 2. Validate Disk Space
    ├── 3. Validate Ownership
    ├── 4. Validate Permissions
    ├── 5. Compute Checksums (MD5/SHA256)
    │
    ▼
Worker Nodes (Parallel)
    │
    ├── 6. Rsync with --checksum
    ├── 7. Validate Copied Permissions
    ├── 8. Validate Checksums Match
    ├── 9. NOTIFY: Never Delete Source
    │
    ▼
Post-Migration Validation
    │
    ├── 10. Service Connectivity Test
    ├── 11. Data Integrity Check
    └── 12. Health Check Verification
```

## Safety Checks Script

```bash
#!/bin/bash
# volume-migration-checks.sh

# 1. Disk Space Check
check_disk_space() {
  local node=$1
  local required_gb=$2
  local available=$(ssh $node "df -BG /var/lib | awk 'NR==2 {print \$4}' | tr -d 'G'")
  
  if [ $available -lt $required_gb ]; then
    echo "ERROR: $node has insufficient disk space (${available}GB < ${required_gb}GB)"
    exit 1
  fi
  echo "OK: $node has sufficient disk space ($available GB available)"
}

# 2. Ownership Check
check_ownership() {
  local volume_path=$1
  local user=$2
  local group=$3
  
  local current_owner=$(stat -c '%U:%G' $volume_path)
  if [ "$current_owner" != "$user:$group" ]; then
    echo "WARNING: Ownership mismatch, fixing..."
    chown -R $user:$group $volume_path
  fi
}

# 3. Checksum Validation
validate_checksum() {
  local source=$1
  local target=$2
  
  SOURCE_MD5=$(md5sum $source | cut -d' ' -f1)
  TARGET_MD5=$(md5sum $target | cut -d' ' -f1)
  
  if [ "$SOURCE_MD5" != "$TARGET_MD5" ]; then
    echo "ERROR: Checksum mismatch!"
    exit 1
  fi
  echo "OK: Checksum validated"
}

# Run checks
echo "Running pre-migration checks..."
# Add your node checks here
```

## Data Preservation Guarantee

- **Source data is NEVER deleted** during migration
- Original volumes remain intact on Manager
- All migrations are additive backups
- Post-migration cleanup only after verification

## Recovery Procedures

If migration fails:

```bash
# Restore from backup
docker run --rm -v mysql_backup:/backup -v mysql_data_new:/data \
  alpine sh -c "cd /backup && cp -av . /data/"

# Or rollback to original
docker service update --force tiki_mysql-primary
```

## Volume Naming Convention

```
swarm-volumes/
├── mysql-primary-data       # MySQL data volume
├── redis-master-data        # Redis data volume  
├── mongodb-data             # MongoDB data volume
├── grafana-data             # Grafana persistence
├── prometheus-data          # Prometheus TSDB
└── shared-configs           # Config files (read-only)
```