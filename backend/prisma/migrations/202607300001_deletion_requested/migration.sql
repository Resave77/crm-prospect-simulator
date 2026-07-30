-- AlterTable: Add deletion_requested column to prospects
ALTER TABLE "prospects" ADD COLUMN "deletion_requested" BOOLEAN NOT NULL DEFAULT false;
