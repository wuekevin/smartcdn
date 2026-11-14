-- +goose Up

-- ACME accounts (one per email/provider)
CREATE TABLE acme_accounts (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email            VARCHAR(255) NOT NULL,
    provider         VARCHAR(64)  NOT NULL DEFAULT 'letsencrypt',
    priv_key_enc     TEXT         NOT NULL DEFAULT '',
    registration_url VARCHAR(512) NOT NULL DEFAULT '',
    created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY ux_acme_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- SSL/TLS certificates
CREATE TABLE certificates (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    domain_id       BIGINT UNSIGNED NOT NULL,
    domain_name     VARCHAR(253)    NOT NULL,
    s_a_ns          TEXT            NOT NULL DEFAULT '',
    status          VARCHAR(32)     NOT NULL DEFAULT 'pending',
    acme_provider   VARCHAR(64)     NOT NULL DEFAULT 'letsencrypt',
    challenge       VARCHAR(32)     NOT NULL DEFAULT 'dns-01',
    account_email   VARCHAR(255)    NOT NULL,
    private_key_enc TEXT            NOT NULL DEFAULT '',
    cert_chain_p_e_m LONGTEXT       NOT NULL DEFAULT '',
    not_before      DATETIME        NULL,
    not_after       DATETIME        NULL,
    auto_renew      TINYINT(1)      NOT NULL DEFAULT 1,
    last_issued_at  DATETIME        NULL,
    last_error      TEXT            NOT NULL DEFAULT '',
    d_n_s_account_i_d BIGINT UNSIGNED NULL,
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME        NULL,
    KEY idx_cert_domain (domain_id),
    KEY idx_cert_status_renew (status, auto_renew, not_after),
    KEY idx_cert_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Certificate deploy targets
CREATE TABLE cert_deploy_targets (
    id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    certificate_id BIGINT UNSIGNED NOT NULL,
    target_type    VARCHAR(64)     NOT NULL,
    target_id      VARCHAR(255)    NOT NULL,
    deployed_at    DATETIME        NULL,
    status         VARCHAR(32)     NOT NULL DEFAULT 'pending',
    last_error     TEXT            NOT NULL DEFAULT '',
    created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_deploy_cert (certificate_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- +goose Down
DROP TABLE IF EXISTS cert_deploy_targets;
DROP TABLE IF EXISTS certificates;
DROP TABLE IF EXISTS acme_accounts;
