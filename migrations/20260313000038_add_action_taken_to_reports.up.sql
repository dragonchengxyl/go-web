ALTER TABLE reports
ADD COLUMN IF NOT EXISTS action_taken VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_reports_action_taken
ON reports(action_taken);
