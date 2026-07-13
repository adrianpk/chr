-- +migrate Up
ALTER TABLE execution_records
ADD COLUMN artifact_id text NOT NULL DEFAULT '';

CREATE INDEX execution_records_artifact_id_idx
ON execution_records (artifact_id);

-- +migrate Down
DROP INDEX IF EXISTS execution_records_artifact_id_idx;
