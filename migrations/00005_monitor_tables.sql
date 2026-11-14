-- +goose Up

-- Time-series metric data points
CREATE TABLE metrics (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    source_type  VARCHAR(32)     NOT NULL,
    source_id    BIGINT UNSIGNED NOT NULL,
    type         VARCHAR(64)     NOT NULL,
    value        DOUBLE          NOT NULL,
    tags         VARCHAR(512)    NOT NULL DEFAULT '{}',
    collected_at DATETIME(3)     NOT NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_metric_source (source_type, source_id, type, collected_at),
    KEY idx_metric_window (type, collected_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  PARTITION BY RANGE (TO_DAYS(collected_at)) (
    PARTITION p_default VALUES LESS THAN MAXVALUE
  );

-- Alert rule definitions
CREATE TABLE alert_rules (
    id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name           VARCHAR(128)    NOT NULL,
    description    VARCHAR(512)    NOT NULL DEFAULT '',
    source_type    VARCHAR(32)     NOT NULL,
    source_id      BIGINT UNSIGNED NULL,
    metric_type    VARCHAR(64)     NOT NULL,
    `condition`    VARCHAR(16)     NOT NULL,
    threshold      DOUBLE          NOT NULL,
    window_minutes INT             NOT NULL DEFAULT 5,
    for_minutes    INT             NOT NULL DEFAULT 1,
    severity       VARCHAR(32)     NOT NULL DEFAULT 'warning',
    enabled        TINYINT(1)      NOT NULL DEFAULT 1,
    channel_ids    TEXT            NOT NULL DEFAULT '[]',
    created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at     DATETIME        NULL,
    KEY idx_alert_rule_enabled (enabled, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Notification channels (webhook / email / dingtalk / slack)
CREATE TABLE alert_channels (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(128)    NOT NULL,
    type        VARCHAR(32)     NOT NULL,
    config_json TEXT            NOT NULL DEFAULT '{}',
    enabled     TINYINT(1)      NOT NULL DEFAULT 1,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  DATETIME        NULL,
    KEY idx_alert_channel_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Fired alert instances
CREATE TABLE alerts (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_id       BIGINT UNSIGNED NOT NULL,
    rule_name     VARCHAR(128)    NOT NULL,
    source_type   VARCHAR(32)     NOT NULL,
    source_id     BIGINT UNSIGNED NOT NULL,
    metric_type   VARCHAR(64)     NOT NULL,
    trigger_value DOUBLE          NOT NULL,
    threshold     DOUBLE          NOT NULL,
    severity      VARCHAR(32)     NOT NULL,
    status        VARCHAR(32)     NOT NULL DEFAULT 'firing',
    fired_at      DATETIME        NOT NULL,
    resolved_at   DATETIME        NULL,
    message       TEXT            NOT NULL DEFAULT '',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_alert_rule_source (rule_id, source_id, status),
    KEY idx_alert_status (status, fired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- +goose Down
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS alert_channels;
DROP TABLE IF EXISTS alert_rules;
DROP TABLE IF EXISTS metrics;
