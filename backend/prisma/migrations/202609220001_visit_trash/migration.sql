ALTER TABLE prospect_visits ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS prospect_visits_deleted_at_idx ON prospect_visits (deleted_at);
