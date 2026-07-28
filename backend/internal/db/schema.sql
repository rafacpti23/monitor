-- P-mon SQLite Schema

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- Whitelabel: each company owns its branding and its internal users.
CREATE TABLE IF NOT EXISTS companies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL DEFAULT 'P-mon',
    system_name TEXT NOT NULL DEFAULT 'P-mon',
    logo_url TEXT NOT NULL DEFAULT '',
    accent_color TEXT NOT NULL DEFAULT '#00e676',
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_id INTEGER REFERENCES companies(id) ON DELETE CASCADE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'admin' CHECK(role IN ('admin','member','viewer')),
    whatsapp_number TEXT NOT NULL DEFAULT '',
    timezone TEXT NOT NULL DEFAULT 'UTC',
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_users_company_id ON users(company_id);

CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    ip TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS server_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '#00e676'
);
CREATE INDEX IF NOT EXISTS idx_server_groups_user_id ON server_groups(user_id);

CREATE TABLE IF NOT EXISTS servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id INTEGER REFERENCES server_groups(id) ON DELETE SET NULL,
    name TEXT NOT NULL DEFAULT '',
    server_key TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL DEFAULT 'linux',
    os TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('online','offline','pending')),
    on_map INTEGER NOT NULL DEFAULT 0,
    lat REAL NOT NULL DEFAULT 0,
    lng REAL NOT NULL DEFAULT 0,
    hostname TEXT NOT NULL DEFAULT '',
    agent_version TEXT NOT NULL DEFAULT '',
    interval_seconds INTEGER NOT NULL DEFAULT 0,
    last_seen DATETIME,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_servers_user_id ON servers(user_id);
CREATE INDEX IF NOT EXISTS idx_servers_server_key ON servers(server_key);
CREATE INDEX IF NOT EXISTS idx_servers_group_id ON servers(group_id);
CREATE INDEX IF NOT EXISTS idx_servers_status ON servers(status);

CREATE TABLE IF NOT EXISTS server_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    timestamp DATETIME NOT NULL DEFAULT (datetime('now')),
    cpu_percent REAL NOT NULL DEFAULT 0,
    ram_total INTEGER NOT NULL DEFAULT 0,
    ram_used INTEGER NOT NULL DEFAULT 0,
    disk_total INTEGER NOT NULL DEFAULT 0,
    disk_used INTEGER NOT NULL DEFAULT 0,
    load1 REAL NOT NULL DEFAULT 0,
    load5 REAL NOT NULL DEFAULT 0,
    load15 REAL NOT NULL DEFAULT 0,
    net_rx INTEGER NOT NULL DEFAULT 0,
    net_tx INTEGER NOT NULL DEFAULT 0,
    uptime_seconds INTEGER NOT NULL DEFAULT 0,
    raw_data TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS idx_server_metrics_server_id ON server_metrics(server_id);
CREATE INDEX IF NOT EXISTS idx_server_metrics_timestamp ON server_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_server_metrics_server_ts ON server_metrics(server_id, timestamp);

CREATE TABLE IF NOT EXISTS server_services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'unknown',
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(server_id, name)
);
CREATE INDEX IF NOT EXISTS idx_server_services_server_id ON server_services(server_id);

CREATE TABLE IF NOT EXISTS website_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '#00e676'
);
CREATE INDEX IF NOT EXISTS idx_website_groups_user_id ON website_groups(user_id);

CREATE TABLE IF NOT EXISTS websites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id INTEGER REFERENCES website_groups(id) ON DELETE SET NULL,
    name TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    method TEXT NOT NULL DEFAULT 'GET',
    check_interval_sec INTEGER NOT NULL DEFAULT 60,
    search_string TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('up','down','pending')),
    last_checked DATETIME,
    last_response_code INTEGER NOT NULL DEFAULT 0,
    last_response_time_ms INTEGER NOT NULL DEFAULT 0,
    ssl_expires_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_websites_user_id ON websites(user_id);
CREATE INDEX IF NOT EXISTS idx_websites_group_id ON websites(group_id);
CREATE INDEX IF NOT EXISTS idx_websites_status ON websites(status);

CREATE TABLE IF NOT EXISTS website_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    website_id INTEGER NOT NULL REFERENCES websites(id) ON DELETE CASCADE,
    timestamp DATETIME NOT NULL DEFAULT (datetime('now')),
    response_code INTEGER NOT NULL DEFAULT 0,
    response_time_ms INTEGER NOT NULL DEFAULT 0,
    status_ok INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_website_history_website_id ON website_history(website_id);
CREATE INDEX IF NOT EXISTS idx_website_history_timestamp ON website_history(timestamp);
CREATE INDEX IF NOT EXISTS idx_website_history_website_ts ON website_history(website_id, timestamp);

CREATE TABLE IF NOT EXISTS checks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'ping' CHECK(type IN ('ping','tcp','http','dns','ssl_expiry')),
    target TEXT NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 0,
    expected_response TEXT NOT NULL DEFAULT '',
    interval_sec INTEGER NOT NULL DEFAULT 60,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('up','down','pending')),
    last_checked DATETIME,
    last_result_ok INTEGER NOT NULL DEFAULT 0,
    last_response_time_ms INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_checks_user_id ON checks(user_id);
CREATE INDEX IF NOT EXISTS idx_checks_type ON checks(type);
CREATE INDEX IF NOT EXISTS idx_checks_status ON checks(status);

CREATE TABLE IF NOT EXISTS check_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    check_id INTEGER NOT NULL REFERENCES checks(id) ON DELETE CASCADE,
    timestamp DATETIME NOT NULL DEFAULT (datetime('now')),
    result_ok INTEGER NOT NULL DEFAULT 0,
    response_time_ms INTEGER NOT NULL DEFAULT 0,
    details TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_check_history_check_id ON check_history(check_id);
CREATE INDEX IF NOT EXISTS idx_check_history_timestamp ON check_history(timestamp);
CREATE INDEX IF NOT EXISTS idx_check_history_check_ts ON check_history(check_id, timestamp);

-- PAPI WhatsApp panels: each panel exposes a list of instances via /api/instances.
CREATE TABLE IF NOT EXISTS papi_panels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    base_url TEXT NOT NULL DEFAULT 'https://papi.api.br',
    panel_token TEXT NOT NULL DEFAULT '',
    check_interval_sec INTEGER NOT NULL DEFAULT 60,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('ok','error','pending')),
    last_checked DATETIME,
    last_error TEXT NOT NULL DEFAULT '',
    total_instances INTEGER NOT NULL DEFAULT 0,
    connected_instances INTEGER NOT NULL DEFAULT 0,
    channels TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_papi_panels_user_id ON papi_panels(user_id);

-- PAPI instances discovered per panel (cache for individual status tracking).
CREATE TABLE IF NOT EXISTS papi_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    panel_id INTEGER NOT NULL REFERENCES papi_panels(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    instance_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    phone_number TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    last_seen DATETIME,
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(panel_id, instance_id)
);
CREATE INDEX IF NOT EXISTS idx_papi_instances_panel_id ON papi_instances(panel_id);
CREATE INDEX IF NOT EXISTS idx_papi_instances_user_id ON papi_instances(user_id);

CREATE TABLE IF NOT EXISTS alert_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    monitor_type TEXT NOT NULL DEFAULT '' CHECK(monitor_type IN ('server','website','check')),
    monitor_id INTEGER NOT NULL DEFAULT 0,
    alert_type TEXT NOT NULL DEFAULT '' CHECK(alert_type IN ('nodata','cpu','ram','disk','load','service_down','docker_down','website_down','ssl_expiry','ping_latency')),
    comparison TEXT NOT NULL DEFAULT '>=' CHECK(comparison IN ('>=','<=','>','<','==','!=')),
    threshold TEXT NOT NULL DEFAULT '0',
    occurrences INTEGER NOT NULL DEFAULT 1,
    cooldown_min INTEGER NOT NULL DEFAULT 30,
    status TEXT NOT NULL DEFAULT 'enabled' CHECK(status IN ('enabled','disabled')),
    channels TEXT NOT NULL DEFAULT '[]'
);
CREATE INDEX IF NOT EXISTS idx_alert_rules_user_id ON alert_rules(user_id);
CREATE INDEX IF NOT EXISTS idx_alert_rules_monitor ON alert_rules(monitor_type, monitor_id);

CREATE TABLE IF NOT EXISTS alert_channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL DEFAULT 'email' CHECK(type IN ('email','whatsapp','webhook')),
    name TEXT NOT NULL DEFAULT '',
    config TEXT NOT NULL DEFAULT '{}',
    enabled INTEGER NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_alert_channels_user_id ON alert_channels(user_id);

CREATE TABLE IF NOT EXISTS incidents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    monitor_type TEXT NOT NULL DEFAULT '',
    monitor_id INTEGER NOT NULL DEFAULT 0,
    alert_type TEXT NOT NULL DEFAULT '',
    severity TEXT NOT NULL DEFAULT 'warning' CHECK(severity IN ('info','warning','critical')),
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active','acknowledged','resolved')),
    start_time DATETIME NOT NULL DEFAULT (datetime('now')),
    end_time DATETIME,
    acknowledged_by INTEGER REFERENCES users(id),
    acknowledged_at DATETIME,
    resolved_at DATETIME,
    ignored INTEGER NOT NULL DEFAULT 0,
    comment TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_incidents_user_id ON incidents(user_id);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
CREATE INDEX IF NOT EXISTS idx_incidents_monitor ON incidents(monitor_type, monitor_id);
CREATE INDEX IF NOT EXISTS idx_incidents_start_time ON incidents(start_time);

CREATE TABLE IF NOT EXISTS notifications_sent (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    incident_id INTEGER NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    channel_id INTEGER NOT NULL REFERENCES alert_channels(id) ON DELETE CASCADE,
    sent_at DATETIME NOT NULL DEFAULT (datetime('now')),
    success INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_notifications_sent_incident_id ON notifications_sent(incident_id);

CREATE TABLE IF NOT EXISTS system_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL DEFAULT 0,
    action TEXT NOT NULL DEFAULT '',
    details TEXT NOT NULL DEFAULT '',
    ip TEXT NOT NULL DEFAULT '',
    timestamp DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_system_logs_user_id ON system_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_system_logs_timestamp ON system_logs(timestamp);
