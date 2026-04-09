ALTER TABLE agent_runs
    ADD COLUMN IF NOT EXISTS lease_owner VARCHAR(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS lease_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS heartbeat_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_agent_runs_lease_expires
    ON agent_runs(status, lease_expires_at);
