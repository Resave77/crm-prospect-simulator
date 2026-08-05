-- Google Places photo references rotate on every Place Details request, so tags
-- keyed by photo_name can never match photos from a fresh fetch. The photo order
-- within a place is stable, so key tags by photo_index (position) instead.
ALTER TABLE "prospect_photo_tags" ADD COLUMN "photo_index" INTEGER;

-- Existing rows reference rotated tokens and can no longer be matched to a photo.
DELETE FROM "prospect_photo_tags";

ALTER TABLE "prospect_photo_tags" ALTER COLUMN "photo_index" SET NOT NULL;

DROP INDEX "idx_prospect_photo_tags_unique";
DROP INDEX "idx_prospect_photo_tags_prospect_id";

ALTER TABLE "prospect_photo_tags" DROP COLUMN "photo_name";

CREATE UNIQUE INDEX "idx_prospect_photo_tags_unique" ON "prospect_photo_tags"("prospect_id", "photo_index");
CREATE INDEX "idx_prospect_photo_tags_prospect_id" ON "prospect_photo_tags"("prospect_id");
