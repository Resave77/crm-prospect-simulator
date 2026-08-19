-- Additive usage monitoring foundation for Google Maps and OpenAI.
-- Source only: do not execute automatically.
CREATE TABLE "provider_usage_events" (
    "id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    "provider" VARCHAR(32) NOT NULL,
    "feature" VARCHAR(96) NOT NULL,
    "operation" VARCHAR(96) NOT NULL,
    "api_or_model" VARCHAR(160) NOT NULL,
    "sku_category" VARCHAR(96),
    "field_mask" TEXT,
    "input_tokens" INTEGER,
    "output_tokens" INTEGER,
    "total_tokens" INTEGER,
    "request_count" INTEGER NOT NULL DEFAULT 1,
    "estimated_cost_usd" DECIMAL(12,8),
    "http_status" INTEGER,
    "success" BOOLEAN NOT NULL DEFAULT false,
    "error_code" VARCHAR(96),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "provider_usage_events_pkey" PRIMARY KEY ("id")
);

CREATE INDEX "provider_usage_events_user_id_created_at_idx" ON "provider_usage_events"("user_id", "created_at");
CREATE INDEX "provider_usage_events_provider_created_at_idx" ON "provider_usage_events"("provider", "created_at");
CREATE INDEX "provider_usage_events_feature_created_at_idx" ON "provider_usage_events"("feature", "created_at");
ALTER TABLE "provider_usage_events" ADD CONSTRAINT "provider_usage_events_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
