-- +goose Up
-- Baseline migration: no-op, exercises the migration pipeline end-to-end.
-- Application schema tables (installations, diagnoses, anthropic_keys, etc.) are owned by later feature changes.

-- +goose Down
-- No-op down migration.
