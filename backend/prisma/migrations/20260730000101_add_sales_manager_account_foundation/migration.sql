-- AlterEnum
ALTER TYPE "UserRole" ADD VALUE 'SALES_MANAGER';

-- DropForeignKey
ALTER TABLE "prospect_comments" DROP CONSTRAINT "prospect_comments_prospect_id_fkey";

-- DropForeignKey
ALTER TABLE "prospect_comments" DROP CONSTRAINT "prospect_comments_user_id_fkey";

-- DropIndex
DROP INDEX "prospects_industry_group_status_idx";

-- AlterTable
ALTER TABLE "customer_sites" ALTER COLUMN "source_prospect_id" SET NOT NULL,
ALTER COLUMN "source_google_place_id" SET NOT NULL,
ALTER COLUMN "updated_at" DROP DEFAULT;

-- AlterTable
ALTER TABLE "parent_companies" ALTER COLUMN "updated_at" DROP DEFAULT;

-- AlterTable
ALTER TABLE "prospect_comments" ALTER COLUMN "id" DROP DEFAULT,
ALTER COLUMN "updated_at" DROP DEFAULT;

-- AlterTable
ALTER TABLE "prospect_visits" ALTER COLUMN "updated_at" DROP DEFAULT;

-- AlterTable
ALTER TABLE "prospects" ALTER COLUMN "updated_at" DROP DEFAULT;

-- AlterTable
ALTER TABLE "refresh_sessions" ALTER COLUMN "updated_at" DROP DEFAULT;

-- AlterTable
ALTER TABLE "users" ADD COLUMN     "created_by" UUID,
ADD COLUMN     "employee_id" TEXT,
ADD COLUMN     "manager_id" UUID,
ADD COLUMN     "must_change_password" BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN     "phone" TEXT DEFAULT '',
ADD COLUMN     "updated_by" UUID,
ALTER COLUMN "updated_at" DROP DEFAULT;

-- CreateIndex
CREATE UNIQUE INDEX "customer_sites_source_prospect_id_key" ON "customer_sites"("source_prospect_id");

-- CreateIndex
CREATE UNIQUE INDEX "customer_sites_source_google_place_id_key" ON "customer_sites"("source_google_place_id");

-- CreateIndex
CREATE UNIQUE INDEX "users_employee_id_key" ON "users"("employee_id");

-- AddForeignKey
ALTER TABLE "users" ADD CONSTRAINT "users_manager_id_fkey" FOREIGN KEY ("manager_id") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "users" ADD CONSTRAINT "users_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "users" ADD CONSTRAINT "users_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "customer_sites" ADD CONSTRAINT "customer_sites_source_prospect_id_fkey" FOREIGN KEY ("source_prospect_id") REFERENCES "prospects"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "prospect_comments" ADD CONSTRAINT "prospect_comments_prospect_id_fkey" FOREIGN KEY ("prospect_id") REFERENCES "prospects"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "prospect_comments" ADD CONSTRAINT "prospect_comments_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- RenameIndex
ALTER INDEX "idx_prospect_comments_prospect_id" RENAME TO "prospect_comments_prospect_id_created_at_idx";

-- RenameIndex
ALTER INDEX "idx_prospect_comments_user_id" RENAME TO "prospect_comments_user_id_idx";
