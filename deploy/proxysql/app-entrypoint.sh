#!/bin/sh
# ============================================================
# app-entrypoint.sh
# Auto-discover the local worker's ProxySQL instance and export
# MYSQL_HOST / MYSQL_PORT for the app process.
#
# How it works:
#   The app container's default gateway IP is the Docker bridge
#   gateway on the host (usually 172.x.0.1).  ProxySQL with
#   host networking binds to 0.0.0.0:6033, so the gateway IP
#   can reach it directly.
#
# If the gateway IP approach fails, fall back to DNS resolution
# of the hostname.
#
# Usage in Dockerfile:
#   COPY app-entrypoint.sh /entrypoint.sh
#   ENTRYPOINT ["/entrypoint.sh"]
#   CMD ["/your-app-binary"]
# ============================================================

set -e

# ---- Strategy 1: default gateway (docker bridge) -----------
GATEWAY=$(ip route | grep '^default' | awk '{print $3}' | head -1)
if [ -n "$GATEWAY" ]; then
    # Test if ProxySQL is reachable at the gateway IP
    if nc -z -w1 "$GATEWAY" 6033 2>/dev/null; then
        export MYSQL_HOST="$GATEWAY"
        export MYSQL_PORT="6033"
        echo "[entrypoint] ProxySQL discovered via gateway: $GATEWAY:6033"
        exec "$@"
    fi
fi

# ---- Strategy 2: try the DNS name "proxysql" --------------
if getent hosts proxysql >/dev/null 2>&1; then
    if nc -z -w1 proxysql 6033 2>/dev/null; then
        export MYSQL_HOST="proxysql"
        export MYSQL_PORT="6033"
        echo "[entrypoint] ProxySQL discovered via DNS: proxysql:6033"
        exec "$@"
    fi
fi

# ---- Strategy 3: use hostname resolution ------------------
HOST_IP=$(hostname -i 2>/dev/null | awk '{print $1}')
if [ -n "$HOST_IP" ]; then
    # The container's IP is on the overlay network.  The host's
    # IP is usually the gateway's next hop.  Try a few common
    # patterns.
    for TRIAL in "$HOST_IP" "${HOST_IP%.*}.1" "127.0.0.1"; do
        if nc -z -w1 "$TRIAL" 6033 2>/dev/null; then
            export MYSQL_HOST="$TRIAL"
            export MYSQL_PORT="6033"
            echo "[entrypoint] ProxySQL discovered at: $TRIAL:6033"
            exec "$@"
        fi
    done
fi

# ---- Fallback: use env var or default ---------------------
echo "[entrypoint] WARNING: Could not auto-discover ProxySQL."
echo "[entrypoint] Using MYSQL_HOST=${MYSQL_HOST:-10.10.10.150} MYSQL_PORT=${MYSQL_PORT:-6033}"
exec "$@"
