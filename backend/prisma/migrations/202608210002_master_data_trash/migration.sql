-- Master Data Trash: soft delete support for segments & categories.
-- Deleted records stay in the table with deleted_at set so they can be
-- listed in Trash and restored later.

ALTER TABLE "segments" ADD COLUMN "deleted_at" TIMESTAMPTZ(6);
ALTER TABLE "categories" ADD COLUMN "deleted_at" TIMESTAMPTZ(6);

-- Unique names only apply to live (non-deleted) records so trashed items
-- never block newly created ones.
DROP INDEX IF EXISTS "segments_name_key";
CREATE UNIQUE INDEX "segments_name_key" ON "segments"("name") WHERE "deleted_at" IS NULL;

DROP INDEX IF EXISTS "categories_segment_id_name_key";
CREATE UNIQUE INDEX "categories_segment_id_name_key" ON "categories"("segment_id", "name") WHERE "deleted_at" IS NULL;
