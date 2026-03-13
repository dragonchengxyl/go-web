DROP INDEX IF EXISTS idx_reports_action_taken;

ALTER TABLE reports
DROP COLUMN IF EXISTS action_taken;
