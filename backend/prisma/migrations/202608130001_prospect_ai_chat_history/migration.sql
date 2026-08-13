CREATE TABLE "prospect_ai_chats" (
    "id" UUID NOT NULL,
    "prospect_id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    "message" TEXT NOT NULL,
    "answer" TEXT NOT NULL,
    "skill" VARCHAR(64) NOT NULL,
    "insight" TEXT,
    "why" TEXT,
    "recommended_action" TEXT,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "prospect_ai_chats_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "prospect_ai_chats_prospect_id_fkey" FOREIGN KEY ("prospect_id") REFERENCES "prospects"("id") ON DELETE CASCADE ON UPDATE NO ACTION,
    CONSTRAINT "prospect_ai_chats_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE RESTRICT ON UPDATE NO ACTION
);

CREATE INDEX "prospect_ai_chats_prospect_id_created_at_idx" ON "prospect_ai_chats"("prospect_id", "created_at");
CREATE INDEX "prospect_ai_chats_user_id_created_at_idx" ON "prospect_ai_chats"("user_id", "created_at");
