-- +goose Up
-- MySQL JSON columns reject empty string '' (Error 3140).
-- Go zero-value string is "", so any INSERT without an explicit capabilities_json
-- or step_logs value causes a runtime failure.  Storing these as TEXT avoids
-- MySQL's strict JSON validation while keeping the data identical on disk.

ALTER TABLE dns_accounts   MODIFY COLUMN capabilities_json TEXT;
ALTER TABLE publish_jobs   MODIFY COLUMN step_logs         TEXT;
ALTER TABLE migration_jobs MODIFY COLUMN step_logs         TEXT;

-- +goose Down
ALTER TABLE dns_accounts   MODIFY COLUMN capabilities_json JSON;
ALTER TABLE publish_jobs   MODIFY COLUMN step_logs         JSON;
ALTER TABLE migration_jobs MODIFY COLUMN step_logs         JSON;
