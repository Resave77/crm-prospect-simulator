-- Additive reconciliation metadata only. Source file; do not execute automatically.
ALTER TABLE "provider_usage_events"
  ADD COLUMN "credential_alias" VARCHAR(128),
  ADD COLUMN "environment" VARCHAR(32),
  ADD COLUMN "cached_tokens" INTEGER;
