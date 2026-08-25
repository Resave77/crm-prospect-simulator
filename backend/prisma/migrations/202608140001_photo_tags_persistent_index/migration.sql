-- Additive identity support for Google photo collections that can be refreshed.
ALTER TABLE "prospect_photo_tags"
    ADD COLUMN "photo_index" INTEGER;

CREATE INDEX "idx_prospect_photo_tags_prospect_photo_index"
    ON "prospect_photo_tags"("prospect_id", "photo_index");
