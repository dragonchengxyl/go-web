DROP INDEX IF EXISTS idx_agent_runs_lease_expires;

ALTER TABLE agent_runs
    DROP COLUMN IF EXISTS heartbeat_at,
    DROP COLUMN IF EXISTS lease_expires_at,
    DROP COLUMN IF EXISTS lease_owner;
