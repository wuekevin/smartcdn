-- +goose Up

-- 凭据中心：统一管理 DNS/部署/通知/ACME CA 凭据
CREATE TABLE ssl_credentials (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name             VARCHAR(255)    NOT NULL,
    provider_type    VARCHAR(64)     NOT NULL,
    category         VARCHAR(32)     NOT NULL,
    credentials_enc  TEXT            NOT NULL,
    created_at       DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at       DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at       DATETIME(3)     NULL,
    KEY idx_ssl_cred_type (provider_type),
    KEY idx_ssl_cred_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- SSL 部署任务：记录证书被推送到哪些目标
CREATE TABLE ssl_deploy_jobs (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    cert_id          BIGINT UNSIGNED NOT NULL,
    credential_id    BIGINT UNSIGNED NULL,
    name             VARCHAR(255)    NOT NULL,
    deploy_type      VARCHAR(64)     NOT NULL,
    config_json      TEXT            NOT NULL,
    status           VARCHAR(32)     NOT NULL DEFAULT 'pending',
    last_deployed_at DATETIME(3)     NULL,
    last_error       TEXT            NULL,
    created_at       DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at       DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    KEY idx_deploy_cert (cert_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS ssl_deploy_jobs;
DROP TABLE IF EXISTS ssl_credentials;
