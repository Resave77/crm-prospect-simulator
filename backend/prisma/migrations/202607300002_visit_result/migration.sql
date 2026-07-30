-- AlterTable: Add visit_result and visit_outcome to prospect_visits
ALTER TABLE "prospect_visits" ADD COLUMN "visit_result" TEXT NOT NULL DEFAULT '';
ALTER TABLE "prospect_visits" ADD COLUMN "visit_outcome" TEXT NOT NULL DEFAULT '';
