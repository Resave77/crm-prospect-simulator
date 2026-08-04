-- Restore visit completion fields used by checkout, monitoring, dashboard, and reports.
-- IF NOT EXISTS keeps this migration safe for databases where the earlier drop was not applied.
ALTER TABLE "prospect_visits"
  ADD COLUMN IF NOT EXISTS "visit_result" TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS "visit_outcome" TEXT NOT NULL DEFAULT '';
