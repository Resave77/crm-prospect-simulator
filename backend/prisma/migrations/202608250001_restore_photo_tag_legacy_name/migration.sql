-- Preserve the legacy photo_name contract while photo_index is the stable key.
ALTER TABLE "prospect_photo_tags"
  ADD COLUMN IF NOT EXISTS "photo_name" TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS "idx_prospect_photo_tags_legacy_name"
  ON "prospect_photo_tags"("prospect_id", "photo_name");
