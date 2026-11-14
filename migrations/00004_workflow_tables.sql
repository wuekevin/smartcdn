-- +goose Up

-- DAG workflow templates
CREATE TABLE workflow_templates (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512) NOT NULL DEFAULT '',
    graph_json  LONGTEXT     NOT NULL,
    is_active   TINYINT(1)   NOT NULL DEFAULT 1,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  DATETIME     NULL,
    UNIQUE KEY ux_wf_template_name (name),
    KEY idx_wf_template_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Workflow run instances
CREATE TABLE workflow_runs (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    template_id   BIGINT UNSIGNED NOT NULL,
    template_name VARCHAR(128)    NOT NULL,
    trigger_type  VARCHAR(64)     NOT NULL,
    trigger_ref   VARCHAR(255)    NOT NULL DEFAULT '',
    status        VARCHAR(32)     NOT NULL DEFAULT 'pending',
    context_json  TEXT            NULL,
    started_at    DATETIME        NULL,
    finished_at   DATETIME        NULL,
    last_error    TEXT            NULL,
    retry_count   INT             NOT NULL DEFAULT 0,
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_wf_run_template (template_id),
    KEY idx_wf_run_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Individual node executions within a run
CREATE TABLE workflow_node_runs (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    run_id      BIGINT UNSIGNED NOT NULL,
    node_id     VARCHAR(128)    NOT NULL,
    node_type   VARCHAR(64)     NOT NULL,
    status      VARCHAR(32)     NOT NULL DEFAULT 'pending',
    input_json  TEXT            NULL,
    output_json TEXT            NULL,
    started_at  DATETIME        NULL,
    finished_at DATETIME        NULL,
    error       TEXT            NULL,
    retry_count INT             NOT NULL DEFAULT 0,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_wf_node_run (run_id),
    UNIQUE KEY ux_wf_node_run_node (run_id, node_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- +goose Down
DROP TABLE IF EXISTS workflow_node_runs;
DROP TABLE IF EXISTS workflow_runs;
DROP TABLE IF EXISTS workflow_templates;
