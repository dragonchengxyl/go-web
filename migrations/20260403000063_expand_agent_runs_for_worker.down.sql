DROP INDEX IF EXISTS idx_agent_runs_status_next_retry;

ALTER TABLE agent_runs
    DROP COLUMN IF EXISTS last_error_at,
    DROP COLUMN IF EXISTS next_retry_at,
    DROP COLUMN IF EXISTS max_attempts,
    DROP COLUMN IF EXISTS attempt_count;
