ALTER TABLE "users" ADD COLUMN "deleted_at" TIMESTAMPTZ(6);

CREATE INDEX "users_deleted_at_idx" ON "users"("deleted_at");
