ALTER TABLE "activity_logs"
  ADD COLUMN "hidden_at" TIMESTAMPTZ(6),
  ADD COLUMN "hidden_by" UUID,
  ADD COLUMN "hide_reason" TEXT;

ALTER TABLE "provider_usage_events"
  ADD COLUMN "hidden_at" TIMESTAMPTZ(6),
  ADD COLUMN "hidden_by" UUID,
  ADD COLUMN "hide_reason" TEXT;

CREATE INDEX "activity_logs_hidden_at_created_at_idx"
  ON "activity_logs"("hidden_at", "created_at");

CREATE INDEX "provider_usage_events_hidden_at_created_at_idx"
  ON "provider_usage_events"("hidden_at", "created_at");
