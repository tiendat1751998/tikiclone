#!/usr/bin/env bash
# ============================================================
# deploy-proxysql.sh
# Deploy ProxySQL on Swarm workers and clean up Keepalived
# on the database VMs.
#
# Usage:
#   ./deploy-proxysql.sh             # full deploy
#   ./deploy-proxysql.sh --skip-db   # skip keepalived cleanup
#   ./deploy-proxysql.sh --cleanup   # only cleanup keepalived
#
# Prerequisites:
#   - Docker Swarm with 1 manager + 3 workers
#   - SSH access to DB VMs (10.10.10.200-202) with sudo
#   - This script runs from the Swarm manager node
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SSH_USER="datdt"
SSH_PASS="123123"
DB_HOSTS="10.10.10.200 10.10.10.201 10.10.10.202"
WORKER_IPS="10.10.10.150 10.10.10.151 10.10.10.152"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*"; }

# ============================================================
# PHASE 1: CLEAN UP KEEPALIVED ON DB VMS
# ============================================================
cleanup_keepalived() {
    info "=== Phase 1: Removing Keepalived from DB VMs ==="

    for HOST in $DB_HOSTS; do
        info "Processing $HOST ..."

        sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_USER@$HOST" \
            "echo '$SSH_PASS' | sudo -S bash -c '
                echo \"--- $HOST: Stopping Keepalived ---\"

                # 1. Remove the VIP from the interface gracefully
                VIP=\"10.10.10.100/24\"
                if ip addr show ens33 | grep -q \"\$VIP\"; then
                    echo \"Removing VIP \$VIP from ens33 ...\"
                    ip addr del \$VIP dev ens33 2>/dev/null || true
                else
                    echo \"VIP not present on this node\"
                fi

                # 2. Stop and disable keepalived
                systemctl stop keepalived 2>/dev/null || echo \"keepalived already stopped\"
                systemctl disable keepalived 2>/dev/null || true

                # 3. Remove the package
                apt-get purge -y keepalived 2>/dev/null || echo \"keepalived not installed\"

                # 4. Remove config
                rm -f /etc/keepalived/keepalived.conf

                # 5. Allow MySQL/MongoDB direct access (remove old ufw rules if any)
                ufw delete allow 3306/tcp 2>/dev/null || true
                ufw delete allow 27017/tcp 2>/dev/null || true

                echo \"$HOST: Keepalived removed successfully\"
            '" 2>&1 | sed "s/^/    [$HOST] /"

        info "$HOST done"
    done

    info "Keepalived cleanup complete on all DB VMs"
}

# ============================================================
# PHASE 2: DEPLOY PROXYSQL STACK
# ============================================================
deploy_proxysql() {
    info "=== Phase 2: Deploying ProxySQL stack ==="

    # Create the directory if it doesn't exist
    cd "$SCRIPT_DIR"

    # Deploy the stack
    docker stack deploy -c proxysql-stack.yml proxysql 2>&1

    info "Waiting for ProxySQL containers to start ..."
    sleep 10

    # Wait for all ProxySQL containers to be running
    EXPECTED=3
    for i in $(seq 1 30); do
        RUNNING=$(docker service ls --filter name=proxysql_proxysql --format "{{.Replicas}}" 2>/dev/null | tr -d '\r')
        READY=$(echo "$RUNNING" | awk -F'/' '{print $1}' 2>/dev/null || echo "0")
        TOTAL=$(echo "$RUNNING" | awk -F'/' '{print $2}' 2>/dev/null || echo "0")

        if [ "$READY" = "$EXPECTED" ] 2>/dev/null; then
            info "All $EXPECTED ProxySQL instances are running"
            break
        fi
        if [ "$i" -eq 30 ]; then
            warn "Timed out waiting for ProxySQL. Current state: $RUNNING"
            warn "Continuing anyway — you may need to run init.sql manually."
        fi
        sleep 5
    done
}

# ============================================================
# PHASE 3: APPLY INIT SQL TO THE FIRST PROXYSQL INSTANCE
# ============================================================
apply_init_sql() {
    info "=== Phase 3: Applying ProxySQL initialization ==="

    # Check if mysql client is available on manager
    if ! command -v mysql &>/dev/null; then
        info "Installing mysql client on manager ..."
        apt-get update -qq && apt-get install -y mysql-client-8.0 -qq
    fi

    # Use the first worker as the admin entry point
    LEADER="10.10.10.150"
    info "Applying init.sql to ProxySQL on $LEADER ..."

    if mysql -h "$LEADER" -P6032 -u admin -padmin \
        < "$SCRIPT_DIR/proxysql-init.sql" 2>&1; then
        info "Init SQL applied successfully to $LEADER"
        info "Cluster sync will propagate to other workers within ~2 seconds"
    else
        error "Failed to apply init.sql. Check that ProxySQL is running on $LEADER:6032"
        error "You can retry manually:"
        error "  mysql -h $LEADER -P6032 -u admin -padmin < $SCRIPT_DIR/proxysql-init.sql"
        exit 1
    fi
}

# ============================================================
# PHASE 4: VERIFY DEPLOYMENT
# ============================================================
verify() {
    info "=== Phase 4: Verification ==="

    echo ""
    info "--- ProxySQL Services ---"
    docker service ls --filter name=proxysql --format "table {{.Name}}\t{{.Replicas}}\t{{.Mode}}"

    echo ""
    info "--- ProxySQL Backend Status ---"
    for IP in $WORKER_IPS; do
        echo "--- $IP ---"
        mysql -h "$IP" -P6032 -u admin -padmin -e "
            SELECT hostgroup_id, hostname, port, status, weight
            FROM monitor.mysql_server_connect_log
            ORDER BY hostgroup_id, hostname
            LIMIT 10;
        " 2>/dev/null || echo "    Cannot connect to $IP:6032"
    done

    echo ""
    info "--- Testing MySQL through ProxySQL ---"
    mysql -h "$LEADER" -P6033 -u tiki -ptiki_dev \
        -e "SELECT @@hostname AS server, 'ProxySQL OK' AS test;" 2>&1

    echo ""
    info "--- UFW Rules on DB VMs ---"
    for HOST in $DB_HOSTS; do
        echo "--- $HOST ---"
        sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_USER@$HOST" \
            "echo '$SSH_PASS' | sudo -S ufw status 2>/dev/null | head -5" 2>&1 | sed "s/^/    /"
    done

    echo ""
    info "=== Deployment Complete ==="
    info "App services should now use:  MYSQL_HOST=<NODE_IP>  MYSQL_PORT=6033"
    info "See app-entrypoint.sh for auto-discovery of the local ProxySQL."
    echo ""
}

# ============================================================
# MAIN
# ============================================================
main() {
    if [ "${1:-}" = "--cleanup" ]; then
        cleanup_keepalived
        exit 0
    fi

    if [ "${1:-}" != "--skip-db" ]; then
        cleanup_keepalived
    else
        warn "Skipping Keepalived cleanup (--skip-db)"
    fi

    deploy_proxysql
    apply_init_sql
    verify
}

main "$@"
