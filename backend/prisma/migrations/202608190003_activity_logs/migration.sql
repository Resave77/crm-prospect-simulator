-- Source-only migration. Do not execute as part of this task.
CREATE TABLE "activity_logs" (
    "id" UUID NOT NULL,
    "user_id" UUID,
    "request_id" VARCHAR(128) NOT NULL,
    "method" VARCHAR(12) NOT NULL,
    "endpoint" TEXT NOT NULL,
    "request_body" JSONB NOT NULL,
    "query_params" JSONB NOT NULL,
    "response_body" JSONB NOT NULL,
    "response_status" INTEGER NOT NULL,
    "domain" VARCHAR(32) NOT NULL,
    "additional_trace" JSONB NOT NULL,
    "duration_ms" INTEGER,
    "ip" VARCHAR(64),
    "user_agent" TEXT,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "activity_logs_pkey" PRIMARY KEY ("id")
);
CREATE INDEX "activity_logs_user_id_created_at_idx" ON "activity_logs"("user_id", "created_at");
CREATE INDEX "activity_logs_request_id_idx" ON "activity_logs"("request_id");
CREATE INDEX "activity_logs_domain_created_at_idx" ON "activity_logs"("domain", "created_at");
ALTER TABLE "activity_logs" ADD CONSTRAINT "activity_logs_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "provider_usage_events" ADD COLUMN "request_id" VARCHAR(128);
CREATE INDEX "provider_usage_events_request_id_idx" ON "provider_usage_events"("request_id");
