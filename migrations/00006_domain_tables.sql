-- +goose Up

-- Fix dns_accounts: rename columns to match entity struct (MySQL 8.0+)
ALTER TABLE dns_accounts RENAME COLUMN provider_code TO provider_type;
ALTER TABLE dns_accounts MODIFY COLUMN provider_type VARCHAR(64) NOT NULL;
ALTER TABLE dns_accounts RENAME COLUMN is_enabled TO is_active;

CREATE TABLE IF NOT EXISTS domains (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name            VARCHAR(253)    NOT NULL,
    account_id      BIGINT UNSIGNED NOT NULL,
    icp_filed       TINYINT(1)      NOT NULL DEFAULT 0,
    max_publish_ips INT             NOT NULL DEFAULT 1,
    is_active       TINYINT(1)      NOT NULL DEFAULT 1,
    created_at      DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at      DATETIME(3),
    PRIMARY KEY (id),
    UNIQUE KEY uq_domains_name (name),
    KEY idx_domains_account_id (account_id),
    KEY idx_domains_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS dns_records (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    domain_id   BIGINT UNSIGNED NOT NULL,
    provider_id VARCHAR(255),
    host        VARCHAR(253)    NOT NULL,
    type        VARCHAR(16)     NOT NULL,
    value       VARCHAR(4096)   NOT NULL,
    ttl         INT             NOT NULL DEFAULT 600,
    line        VARCHAR(64),
    priority    INT             NOT NULL DEFAULT 0,
    is_external TINYINT(1)      NOT NULL DEFAULT 0,
    created_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at  DATETIME(3),
    PRIMARY KEY (id),
    KEY idx_dns_records_domain_id (domain_id),
    KEY idx_dns_records_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS dns_snapshots (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    domain_id      BIGINT UNSIGNED NOT NULL,
    snapshot_json  JSON            NOT NULL,
    trigger_event  VARCHAR(128),
    created_at     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_dns_snapshots_domain_id (domain_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cdn_accounts (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name            VARCHAR(128)    NOT NULL,
    provider_type   VARCHAR(64)     NOT NULL,
    credentials_enc TEXT            NOT NULL,
    is_active       TINYINT(1)      NOT NULL DEFAULT 1,
    created_at      DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at      DATETIME(3),
    PRIMARY KEY (id),
    KEY idx_cdn_accounts_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cdn_domains (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    account_id       BIGINT UNSIGNED NOT NULL,
    domain_name      VARCHAR(253)    NOT NULL,
    provider_zone_id VARCHAR(255),
    cname            VARCHAR(512),
    status           VARCHAR(32)     NOT NULL DEFAULT 'pending',
    origin_addr      VARCHAR(512),
    origin_port      INT             NOT NULL DEFAULT 80,
    protocol         VARCHAR(16)     NOT NULL DEFAULT 'follow',
    is_active        TINYINT(1)      NOT NULL DEFAULT 1,
    created_at       DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at       DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at       DATETIME(3),
    PRIMARY KEY (id),
    UNIQUE KEY uq_cdn_domains_domain_name (domain_name),
    KEY idx_cdn_domains_account_id (account_id),
    KEY idx_cdn_domains_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cdn_custom_hostnames (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    cdn_domain_id BIGINT UNSIGNED NOT NULL,
    hostname      VARCHAR(253)    NOT NULL,
    provider_id   VARCHAR(255),
    status        VARCHAR(64),
    dcv_txt_name  VARCHAR(512),
    dcv_txt_value VARCHAR(512),
    dcv_http_path VARCHAR(512),
    dcv_http_value VARCHAR(512),
    ssl_status    VARCHAR(64),
    created_at    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_cdn_custom_hostnames_cdn_domain_id (cdn_domain_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cdn_nodes (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name             VARCHAR(128)    NOT NULL,
    ip               VARCHAR(45)     NOT NULL,
    isp              VARCHAR(64)     NOT NULL,
    region           VARCHAR(64)     NOT NULL,
    origin_addr      VARCHAR(512),
    origin_port      INT             NOT NULL DEFAULT 80,
    is_active        TINYINT(1)      NOT NULL DEFAULT 1,
    auth_token       VARCHAR(128)    NOT NULL,
    auth_confirmed   TINYINT(1)      NOT NULL DEFAULT 0,
    https_status     VARCHAR(32),
    https_code       INT,
    tls_domain       VARCHAR(253),
    https_latency_ms BIGINT,
    https_checked_at DATETIME(3),
    created_at       DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at       DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at       DATETIME(3),
    PRIMARY KEY (id),
    UNIQUE KEY uq_cdn_nodes_ip (ip),
    UNIQUE KEY uq_cdn_nodes_auth_token (auth_token),
    KEY idx_cdn_nodes_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS node_agents (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    node_id        BIGINT UNSIGNED NOT NULL,
    client_cert_sn VARCHAR(128),
    last_seen_at   DATETIME(3),
    cpu_percent    DOUBLE          NOT NULL DEFAULT 0,
    mem_percent    DOUBLE          NOT NULL DEFAULT 0,
    https_status   VARCHAR(32),
    created_at     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uq_node_agents_node_id (node_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS node_tasks (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    node_id       BIGINT UNSIGNED NOT NULL,
    type          VARCHAR(64)     NOT NULL,
    payload       TEXT,
    status        VARCHAR(32)     NOT NULL DEFAULT 'pending',
    result_msg    TEXT,
    dispatched_at DATETIME(3),
    finished_at   DATETIME(3),
    created_at    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_node_tasks_node_id (node_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS node_pools (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name       VARCHAR(128)    NOT NULL,
    is_active  TINYINT(1)      NOT NULL DEFAULT 1,
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    PRIMARY KEY (id),
    UNIQUE KEY uq_node_pools_name (name),
    KEY idx_node_pools_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS node_pool_members (
    id       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    pool_id  BIGINT UNSIGNED NOT NULL,
    node_id  BIGINT UNSIGNED NOT NULL,
    priority INT             NOT NULL DEFAULT 10,
    PRIMARY KEY (id),
    UNIQUE KEY ux_pool_node (pool_id, node_id),
    KEY idx_node_pool_members_node_id (node_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS probe_agents (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name           VARCHAR(128)    NOT NULL,
    isp            VARCHAR(64)     NOT NULL,
    region         VARCHAR(64)     NOT NULL,
    client_cert_sn VARCHAR(128),
    is_active      TINYINT(1)      NOT NULL DEFAULT 1,
    last_seen_at   DATETIME(3),
    created_at     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uq_probe_agents_cert_sn (client_cert_sn)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS probe_results (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    probe_agent_id BIGINT UNSIGNED NOT NULL,
    target_ip      VARCHAR(45)     NOT NULL,
    isp            VARCHAR(64)     NOT NULL,
    latency_ms     BIGINT          NOT NULL DEFAULT 0,
    ttfb_ms        BIGINT          NOT NULL DEFAULT 0,
    https_status   INT             NOT NULL DEFAULT 0,
    success        TINYINT(1)      NOT NULL DEFAULT 0,
    probe_at       DATETIME(3)     NOT NULL,
    created_at     DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_probe_results_agent (probe_agent_id),
    KEY idx_probe_results_target_ip (target_ip),
    KEY idx_probe_results_probe_at (probe_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ip_candidates (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    domain_id     BIGINT UNSIGNED NOT NULL,
    node_id       BIGINT UNSIGNED NOT NULL,
    ip            VARCHAR(45)     NOT NULL,
    isp           VARCHAR(64)     NOT NULL,
    status        VARCHAR(32)     NOT NULL DEFAULT 'candidate',
    fail_streak   INT             NOT NULL DEFAULT 0,
    score         DOUBLE          NOT NULL DEFAULT 0,
    last_score_at DATETIME(3),
    created_at    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_ip_candidates_domain_id (domain_id),
    KEY idx_ip_candidates_node_id (node_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS scheduler_circuit_breakers (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    domain_id    BIGINT UNSIGNED NOT NULL,
    isp          VARCHAR(64)     NOT NULL,
    tripped      TINYINT(1)      NOT NULL DEFAULT 0,
    switch_count INT             NOT NULL DEFAULT 0,
    window_start DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    tripped_at   DATETIME(3),
    created_at   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY ux_cb_domain_isp (domain_id, isp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS publish_jobs (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    domain_id    BIGINT UNSIGNED NOT NULL,
    status       VARCHAR(32)     NOT NULL DEFAULT 'pending',
    target_host  VARCHAR(253)    NOT NULL,
    target_type  VARCHAR(16)     NOT NULL,
    target_line  VARCHAR(64),
    payload_json JSON            NOT NULL,
    snapshot_id  BIGINT UNSIGNED,
    dry_run      TINYINT(1)      NOT NULL DEFAULT 0,
    step_logs    JSON,
    error_msg    TEXT,
    created_at   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    finished_at  DATETIME(3),
    PRIMARY KEY (id),
    KEY idx_publish_jobs_domain_id (domain_id),
    KEY idx_publish_jobs_snapshot_id (snapshot_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS migration_jobs (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    domain_id         BIGINT UNSIGNED NOT NULL,
    source_account_id BIGINT UNSIGNED NOT NULL,
    target_account_id BIGINT UNSIGNED NOT NULL,
    status            VARCHAR(32)     NOT NULL DEFAULT 'pending',
    snapshot_id       BIGINT UNSIGNED,
    step_logs         JSON,
    error_msg         TEXT,
    created_at        DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at        DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    finished_at       DATETIME(3),
    PRIMARY KEY (id),
    KEY idx_migration_jobs_domain_id (domain_id),
    KEY idx_migration_jobs_snapshot_id (snapshot_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS multi_dns_configs (
    id                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    domain_id            BIGINT UNSIGNED NOT NULL,
    primary_account_id   BIGINT UNSIGNED NOT NULL,
    secondary_account_id BIGINT UNSIGNED NOT NULL,
    sync_enabled         TINYINT(1)      NOT NULL DEFAULT 1,
    created_at           DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at           DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uq_multi_dns_configs_domain (domain_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS multi_dns_configs;
DROP TABLE IF EXISTS migration_jobs;
DROP TABLE IF EXISTS publish_jobs;
DROP TABLE IF EXISTS scheduler_circuit_breakers;
DROP TABLE IF EXISTS ip_candidates;
DROP TABLE IF EXISTS probe_results;
DROP TABLE IF EXISTS probe_agents;
DROP TABLE IF EXISTS node_pool_members;
DROP TABLE IF EXISTS node_pools;
DROP TABLE IF EXISTS node_tasks;
DROP TABLE IF EXISTS node_agents;
DROP TABLE IF EXISTS cdn_nodes;
DROP TABLE IF EXISTS cdn_custom_hostnames;
DROP TABLE IF EXISTS cdn_domains;
DROP TABLE IF EXISTS cdn_accounts;
DROP TABLE IF EXISTS dns_snapshots;
DROP TABLE IF EXISTS dns_records;
DROP TABLE IF EXISTS domains;
