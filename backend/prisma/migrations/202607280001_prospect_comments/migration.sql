-- CreateTable
CREATE TABLE "prospect_comments" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "prospect_id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    "content" TEXT NOT NULL,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT now(),

    CONSTRAINT "prospect_comments_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "idx_prospect_comments_prospect_id" ON "prospect_comments"("prospect_id", "created_at");

-- CreateIndex
CREATE INDEX "idx_prospect_comments_user_id" ON "prospect_comments"("user_id");

-- AddForeignKey
ALTER TABLE "prospect_comments" ADD CONSTRAINT "prospect_comments_prospect_id_fkey" FOREIGN KEY ("prospect_id") REFERENCES "prospects"("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "prospect_comments" ADD CONSTRAINT "prospect_comments_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE RESTRICT ON UPDATE NO ACTION;
