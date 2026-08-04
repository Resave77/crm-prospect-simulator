ALTER TABLE "prospect_comments"
ADD COLUMN "attachments" JSONB NOT NULL DEFAULT '[]'::jsonb;
