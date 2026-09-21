ALTER TABLE "prospects" ADD COLUMN IF NOT EXISTS "deleted_at" TIMESTAMPTZ(6);

CREATE INDEX IF NOT EXISTS "prospects_deleted_at_idx" ON "prospects"("deleted_at");
