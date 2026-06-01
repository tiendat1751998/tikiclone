#!/bin/bash
# Run this script ONCE with sudo to apply host-level kernel optimizations
# for 10K+ TPS load testing on Tikiclone
#
# Usage: sudo bash /home/datdt/tikiclone/scripts/apply-host-tuning.sh

set -e

echo "=== Applying kernel network optimizations ==="

# TCP connection queues
sysctl -w net.core.somaxconn=65536
sysctl -w net.ipv4.tcp_max_syn_backlog=65536
sysctl -w net.core.netdev_max_backlog=65536

# Port reuse and fast recycling
sysctl -w net.ipv4.tcp_tw_reuse=1
sysctl -w net.ipv4.tcp_fin_timeout=15

# Ephemeral port range
sysctl -w net.ipv4.ip_local_port_range="1024 65535"

# Socket buffers
sysctl -w net.core.rmem_max=16777216
sysctl -w net.core.wmem_max=16777216
sysctl -w net.ipv4.tcp_rmem="4096 87380 16777216"
sysctl -w net.ipv4.tcp_wmem="4096 87380 16777216"

# TCP memory
sysctl -w net.ipv4.tcp_mem="786432 1048576 1572864"

# Connection tracking
sysctl -w net.nf_conntrack_max=1048576

# Disable idle slow start
sysctl -w net.ipv4.tcp_slow_start_after_idle=0

# TCP keepalive
sysctl -w net.ipv4.tcp_keepalive_time=30
sysctl -w net.ipv4.tcp_keepalive_intvl=5
sysctl -w net.ipv4.tcp_keepalive_probes=3

# File descriptors
echo 2097152 > /proc/sys/fs/file-max

echo "=== Copying persistent sysctl config ==="
cp /home/datdt/tikiclone/configs/99-tikiclone-performance.conf /etc/sysctl.d/99-tikiclone-performance.conf
sysctl --system > /dev/null 2>&1

echo "=== Installing Docker daemon config for default container limits ==="
cp /home/datdt/tikiclone/configs/docker-daemon.json /etc/docker/daemon.json
systemctl reload docker 2>&1 || echo "NOTE: Could not reload Docker. Restart Docker manually."

echo ""
echo "=== Verification ==="
echo "somaxconn: $(cat /proc/sys/net/core/somaxconn)"
echo "tcp_max_syn_backlog: $(cat /proc/sys/net/ipv4/tcp_max_syn_backlog)"
echo "tcp_fin_timeout: $(cat /proc/sys/net/ipv4/tcp_fin_timeout)"
echo "ip_local_port_range: $(cat /proc/sys/net/ipv4/ip_local_port_range)"
echo "file-max: $(cat /proc/sys/fs/file-max)"
echo ""
echo "=== Done. All host optimizations applied. ==="
