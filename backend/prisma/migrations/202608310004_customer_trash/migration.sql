ALTER TABLE "customer_sites" ADD COLUMN "deleted_at" TIMESTAMPTZ(6);
CREATE INDEX "customer_sites_deleted_at_idx" ON "customer_sites"("deleted_at");
