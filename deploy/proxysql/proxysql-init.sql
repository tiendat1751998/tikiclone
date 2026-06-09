-- ============================================================
-- proxysql-init.sql
-- ProxySQL initialization — run once via the admin interface
-- (port 6032) on ONE node.  Cluster sync propagates to all.
--
-- Usage:
--   mysql -h <NODE_IP> -P6032 -u admin -padmin < proxysql-init.sql
--
-- This script is idempotent: safe to run multiple times or
-- on every node if cluster sync is not yet configured.
-- ============================================================

-- ============================================================
-- 1. MYSQL SERVERS (BACKEND DATABASE TIER)
-- ----------------------------------------------------------
-- Hostgroup 10 = Writer  (MySQL Master, read_only=0)
-- Hostgroup 20 = Reader  (MySQL Slaves, read_only=1)
--
-- ProxySQL monitors the read_only variable and auto-moves
-- servers between hostgroups if a slave is promoted to master
-- (read_only→0) or a master demoted (read_only→1).
-- This enables automatic failover WITHOUT any external script.
-- ============================================================

-- Clear existing servers (idempotent)
DELETE FROM mysql_servers;

-- Writer hostgroup: MySQL Master
INSERT INTO mysql_servers (hostgroup_id, hostname, port, weight, max_connections, max_replication_lag, comment)
VALUES (10, '10.10.10.200', 3306, 100, 500, 0, 'MySQL Master - Writer');

-- Reader hostgroup: MySQL Slaves
INSERT INTO mysql_servers (hostgroup_id, hostname, port, weight, max_connections, max_replication_lag, comment)
VALUES (20, '10.10.10.201', 3306, 100, 500, 10, 'MySQL Slave 1 - Reader');
INSERT INTO mysql_servers (hostgroup_id, hostname, port, weight, max_connections, max_replication_lag, comment)
VALUES (20, '10.10.10.202', 3306, 100, 500, 10, 'MySQL Slave 2 - Reader');


-- ============================================================
-- 2. REPLICATION HOSTGROUPS
-- ----------------------------------------------------------
-- Tells ProxySQL which hostgroups represent writer/reader pairs.
-- ProxySQL periodically checks @@read_only on each server.
--   read_only=0 → assigned to writer_hostgroup (10)
--   read_only=1 → assigned to reader_hostgroup (20)
--
-- This means:
--   - If masterdb (10.200) goes down and workerdb1 gets
--     promoted (read_only→0), ProxySQL auto-moves it to HG 10.
--   - When masterdb comes back, it gets moved back.
-- ============================================================
DELETE FROM mysql_replication_hostgroups;
INSERT INTO mysql_replication_hostgroups (writer_hostgroup, reader_hostgroup, comment)
VALUES (10, 20, 'MySQL GTID Replication: HG10=writer, HG20=reader');


-- ============================================================
-- 3. MYSQL USERS
-- ----------------------------------------------------------
-- Application database credentials.  Each user maps to:
--   - default_hostgroup: where unmatched queries go (writer)
--   - transaction_persistent: keeps transactions on the same
--     hostgroup (prevents breaking BEGIN…COMMIT)
--
-- ProxySQL will also auto-create a 'monitor' user if
-- mysql-monitor_username is set (handled in bootstrap.cnf).
-- For safety, we create it here explicitly.
-- ============================================================
DELETE FROM mysql_users;

-- Application user: tiki / tiki_dev
-- Default hostgroup = writer (10).  SELECT rules below
-- will redirect reads to HG 20.
INSERT INTO mysql_users (username, password, default_hostgroup, transaction_persistent, active, max_connections)
VALUES ('tiki', 'tiki_dev', 10, 1, 1, 500);

-- Root user (for admin/debug purposes, not used by apps)
INSERT INTO mysql_users (username, password, default_hostgroup, transaction_persistent, active)
VALUES ('root', '123123', 10, 1, 1);


-- ============================================================
-- 4. MYSQL QUERY RULES (READ/WRITE SPLITTING)
-- ----------------------------------------------------------
-- Rules are evaluated top-down by rule_id.
-- The first match wins (apply=1 stops further evaluation).
--
-- Rule priority order:
--   1. SELECT … FOR UPDATE → writer  (must lock rows on master)
--   2. SELECT …             → reader (offload reads to slaves)
--   3. Everything else       → writer (default_hostgroup from user)
--
-- Regex notes:
--   ^SELECT\s     — case-sensitive start-of-statement
--   .*FOR UPDATE$ — must end with FOR UPDATE
--   (?i)          — inline case-insensitive flag for safety
-- ============================================================
DELETE FROM mysql_query_rules;

-- Rule 1: SELECT … FOR UPDATE → writer (HG 10)
-- Must be BEFORE the generic SELECT rule.
INSERT INTO mysql_query_rules
    (rule_id, active, match_pattern, destination_hostgroup, apply, comment)
VALUES
    (1, 1, '^SELECT.*FOR UPDATE', 10, 1, 'SELECT FOR UPDATE → writer');

-- Rule 2: SELECT … → reader (HG 20)
-- Matches any SELECT statement not caught by rule 1.
INSERT INTO mysql_query_rules
    (rule_id, active, match_pattern, destination_hostgroup, apply, comment)
VALUES
    (2, 1, '^SELECT ', 20, 1, 'SELECT → reader');

-- Rule 3: SHOW statements → reader (optional, reduces master load)
INSERT INTO mysql_query_rules
    (rule_id, active, match_pattern, destination_hostgroup, apply, comment)
VALUES
    (3, 1, '^SHOW ', 20, 1, 'SHOW → reader');


-- ============================================================
-- 5. PROXYSQL CLUSTER
-- ----------------------------------------------------------
-- Each ProxySQL instance registers its peers so config changes
-- on one node are replicated to all others.
--
-- NOTE: proxysql_servers is already set in bootstrap.cnf.
-- We include it here as a fallback in case the bootstrap file
-- is not mounted, or to update the list at runtime.
-- ============================================================
DELETE FROM proxysql_servers;
INSERT INTO proxysql_servers (hostname, port, weight, comment)
VALUES ('10.10.10.150', 6032, 0, 'worker1-proxysql');
INSERT INTO proxysql_servers (hostname, port, weight, comment)
VALUES ('10.10.10.151', 6032, 0, 'worker2-proxysql');
INSERT INTO proxysql_servers (hostname, port, weight, comment)
VALUES ('10.10.10.152', 6032, 0, 'worker3-proxysql');


-- ============================================================
-- 6. APPLY CONFIGURATION
-- ----------------------------------------------------------
-- LOAD … TO RUNTIME  — activates config without restart
-- SAVE … TO DISK     — persists config to SQLite DB (survives restart)
--
-- Order matters:
--   1. Servers    (backends must exist before users can connect)
--   2. Users      (auth must work before queries flow)
--   3. Rules      (routing logic for queries)
--   4. Cluster    (peers must be known for sync)
-- ============================================================
LOAD MYSQL SERVERS TO RUNTIME;
SAVE MYSQL SERVERS TO DISK;

LOAD MYSQL USERS TO RUNTIME;
SAVE MYSQL USERS TO DISK;

LOAD MYSQL QUERY RULES TO RUNTIME;
SAVE MYSQL QUERY RULES TO DISK;

LOAD PROXYSQL SERVERS TO RUNTIME;
SAVE PROXYSQL SERVERS TO DISK;

-- Verify
SELECT 'ProxySQL configured successfully' AS status;
SELECT hostgroup_id, hostname, port, status FROM mysql_servers;
SELECT username, default_hostgroup, active FROM mysql_users;
SELECT rule_id, match_pattern, destination_hostgroup FROM mysql_query_rules ORDER BY rule_id;
