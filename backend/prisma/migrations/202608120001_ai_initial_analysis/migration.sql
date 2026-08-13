CREATE TABLE "prospect_ai_analyses" (
    "prospect_id" UUID NOT NULL,
    "summary_json" JSONB,
    "menu_json" JSONB,
    "status" VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    "error_code" VARCHAR(64) NOT NULL DEFAULT '',
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "prospect_ai_analyses_pkey" PRIMARY KEY ("prospect_id"),
    CONSTRAINT "prospect_ai_analyses_prospect_id_fkey" FOREIGN KEY ("prospect_id") REFERENCES "prospects"("id") ON DELETE CASCADE ON UPDATE NO ACTION
);
