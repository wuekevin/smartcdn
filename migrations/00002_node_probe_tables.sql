-- +goose Up

-- CDN Nodes (self-hosted reverse-proxy nodes)
CREATE TABLE cdn_nodes (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name             VARCHAR(128) NOT NULL,
    ip               VARCHAR(45)  NOT NULL,
    isp              VARCHAR(64)  NOT NULL DEFAULT '',
    region           VARCHAR(64)  NOT NULL DEFAULT '',
    origin_addr      VARCHAR(512) NOT NULL DEFAULT '',
    origin_port      INT          NOT NULL DEFAULT 80,
    is_active        TINYINT(1)   NOT NULL DEFAULT 1,
    auth_token       VARCHAR(128) NOT NULL,
    auth_confirmed   TINYINT(1)   NOT NULL DEFAULT 0,
    https_status     VARCHAR(32)  NOT NULL DEFAULT 'pending',
    https_code       INT          NOT NULL DEFAULT 0,
    tls_domain       VARCHAR(253) NOT NULL DEFAULT '',
    https_latency_ms BIGINT       NOT NULL DEFAULT 0,
    https_checked_at DATETIME     NULL,
    created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at       DATETIME     NULL,
    UNIQUE KEY ux_node_ip (ip),
    UNIQUE KEY ux_node_token (auth_token),
    KEY idx_node_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Node Agent registrations and heartbeats
CREATE TABLE node_agents (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    node_id         BIGINT UNSIGNED NOT NULL,
    client_cert_s_n VARCHAR(128)    NOT NULL DEFAULT '',
    last_seen_at    DATETIME        NULL,
    cpu_percent     DOUBLE          NOT NULL DEFAULT 0,
    mem_percent     DOUBLE          NOT NULL DEFAULT 0,
    https_status    VARCHAR(32)     NOT NULL DEFAULT '',
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY ux_agent_node (node_id),
    KEY idx_agent_seen (last_seen_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Node management tasks
CREATE TABLE node_tasks (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    node_id       BIGINT UNSIGNED NOT NULL,
    type          VARCHAR(64)     NOT NULL,
    payload       TEXT            NULL,
    status        VARCHAR(32)     NOT NULL DEFAULT 'pending',
    result_msg    TEXT            NULL,
    dispatched_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at   DATETIME        NULL,
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_task_node_status (node_id, status),
    KEY idx_task_node_pending (node_id, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Node resource pools
CREATE TABLE node_pools (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(128) NOT NULL,
    is_active  TINYINT(1)   NOT NULL DEFAULT 1,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME     NULL,
    UNIQUE KEY ux_pool_name (name),
    KEY idx_pool_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Pool membership (M2M)
CREATE TABLE node_pool_members (
    id       BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    pool_id  BIGINT UNSIGNED NOT NULL,
    node_id  BIGINT UNSIGNED NOT NULL,
    priority INT             NOT NULL DEFAULT 10,
    UNIQUE KEY ux_pool_node (pool_id, node_id),
    KEY idx_member_node (node_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Probe Agents (distributed measurement nodes)
CREATE TABLE probe_agents (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    isp             VARCHAR(64)  NOT NULL DEFAULT '',
    region          VARCHAR(64)  NOT NULL DEFAULT '',
    client_cert_s_n VARCHAR(128) NOT NULL DEFAULT '',
    is_active       TINYINT(1)   NOT NULL DEFAULT 1,
    last_seen_at    DATETIME     NULL,
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY ux_probe_cert (client_cert_s_n),
    KEY idx_probe_seen (last_seen_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Probe measurement results
CREATE TABLE probe_results (
    id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    probe_agent_id BIGINT UNSIGNED NOT NULL,
    target_ip      VARCHAR(45)     NOT NULL,
    isp            VARCHAR(64)     NOT NULL DEFAULT '',
    latency_ms     BIGINT          NOT NULL DEFAULT 0,
    ttfb_ms        BIGINT          NOT NULL DEFAULT 0,
    https_status   INT             NOT NULL DEFAULT 0,
    success        TINYINT(1)      NOT NULL DEFAULT 0,
    probe_at       DATETIME        NOT NULL,
    created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_result_ip_time (target_ip, probe_at),
    KEY idx_result_ip_isp_time (target_ip, isp, probe_at),
    KEY idx_result_agent (probe_agent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- IP candidates for scheduler
CREATE TABLE ip_candidates (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    domain_id     BIGINT UNSIGNED NOT NULL,
    node_id       BIGINT UNSIGNED NOT NULL,
    ip            VARCHAR(45)     NOT NULL,
    isp           VARCHAR(64)     NOT NULL DEFAULT 'default',
    status        VARCHAR(32)     NOT NULL DEFAULT 'candidate',
    fail_streak   INT             NOT NULL DEFAULT 0,
    score         DOUBLE          NOT NULL DEFAULT 0,
    last_score_at DATETIME        NULL,
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_candidate_domain (domain_id),
    KEY idx_candidate_domain_isp_status (domain_id, isp, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Scheduler circuit breakers (per domain/ISP)
CREATE TABLE scheduler_circuit_breakers (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    domain_id    BIGINT UNSIGNED NOT NULL,
    isp          VARCHAR(64)     NOT NULL DEFAULT 'default',
    tripped      TINYINT(1)      NOT NULL DEFAULT 0,
    switch_count INT             NOT NULL DEFAULT 0,
    window_start DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tripped_at   DATETIME        NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY ux_cb_domain_isp (domain_id, isp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- +goose Down
DROP TABLE IF EXISTS scheduler_circuit_breakers;
DROP TABLE IF EXISTS ip_candidates;
DROP TABLE IF EXISTS probe_results;
DROP TABLE IF EXISTS probe_agents;
DROP TABLE IF EXISTS node_pool_members;
DROP TABLE IF EXISTS node_pools;
DROP TABLE IF EXISTS node_tasks;
DROP TABLE IF EXISTS node_agents;
DROP TABLE IF EXISTS cdn_nodes;
