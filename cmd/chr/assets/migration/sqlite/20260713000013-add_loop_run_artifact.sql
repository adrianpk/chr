-- +migrate Up
ALTER TABLE loop_runs
ADD COLUMN artifact_id text NOT NULL DEFAULT '';

ALTER TABLE loop_runs
ADD COLUMN artifact_dir text NOT NULL DEFAULT '';

CREATE INDEX loop_runs_artifact_id_idx
ON loop_runs (artifact_id);

-- +migrate Down
DROP INDEX IF EXISTS loop_runs_artifact_id_idx;
