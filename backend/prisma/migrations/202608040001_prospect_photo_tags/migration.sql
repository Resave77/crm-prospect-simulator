-- CreateTable
CREATE TABLE "prospect_photo_tags" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "prospect_id" UUID NOT NULL,
    "photo_name" TEXT NOT NULL,
    "category" TEXT NOT NULL,
    "updated_by" UUID,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT now(),

    CONSTRAINT "prospect_photo_tags_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "idx_prospect_photo_tags_unique" ON "prospect_photo_tags"("prospect_id", "photo_name");

-- CreateIndex
CREATE INDEX "idx_prospect_photo_tags_prospect_id" ON "prospect_photo_tags"("prospect_id");

-- AddForeignKey
ALTER TABLE "prospect_photo_tags" ADD CONSTRAINT "prospect_photo_tags_prospect_id_fkey" FOREIGN KEY ("prospect_id") REFERENCES "prospects"("id") ON DELETE CASCADE ON UPDATE NO ACTION;
