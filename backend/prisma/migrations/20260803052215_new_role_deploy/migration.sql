/*
  Warnings:

  - You are about to drop the column `visit_outcome` on the `prospect_visits` table. All the data in the column will be lost.
  - You are about to drop the column `visit_result` on the `prospect_visits` table. All the data in the column will be lost.

*/
-- DropForeignKey
ALTER TABLE "sales_roles" DROP CONSTRAINT "sales_roles_created_by_fkey";

-- DropForeignKey
ALTER TABLE "sales_roles" DROP CONSTRAINT "sales_roles_updated_by_fkey";

-- DropForeignKey
ALTER TABLE "sales_structure_assignments" DROP CONSTRAINT "sales_structure_assignments_created_by_fkey";

-- DropForeignKey
ALTER TABLE "sales_structure_assignments" DROP CONSTRAINT "sales_structure_assignments_parent_user_id_fkey";

-- DropForeignKey
ALTER TABLE "sales_structure_assignments" DROP CONSTRAINT "sales_structure_assignments_user_id_fkey";

-- AlterTable
ALTER TABLE "prospect_visits" DROP COLUMN "visit_outcome",
DROP COLUMN "visit_result";

-- AlterTable
ALTER TABLE "sales_roles" ALTER COLUMN "updated_at" DROP DEFAULT;

-- AlterTable
ALTER TABLE "sales_structure_assignments" ALTER COLUMN "updated_at" DROP DEFAULT;

-- CreateIndex
CREATE INDEX "sales_roles_level_idx" ON "sales_roles"("level");
