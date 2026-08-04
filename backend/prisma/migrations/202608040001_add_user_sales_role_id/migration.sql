ALTER TABLE "users" ADD COLUMN "sales_role_id" UUID;

CREATE INDEX "users_sales_role_id_idx" ON "users"("sales_role_id");

ALTER TABLE "users" ADD CONSTRAINT "users_sales_role_id_fkey" FOREIGN KEY ("sales_role_id") REFERENCES "sales_roles"("id") ON DELETE SET NULL ON UPDATE CASCADE;
