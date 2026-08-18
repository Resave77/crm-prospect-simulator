ALTER TABLE "users"
  ADD COLUMN "timezone" TEXT DEFAULT 'Asia/Jakarta',
  ADD COLUMN "city" TEXT,
  ADD COLUMN "province" TEXT,
  ADD COLUMN "district" TEXT,
  ADD COLUMN "job_title" TEXT,
  ADD COLUMN "position_grade" TEXT,
  ADD COLUMN "sub_department" TEXT,
  ADD COLUMN "join_date" DATE,
  ADD COLUMN "gender" TEXT,
  ADD COLUMN "date_of_birth" DATE,
  ADD COLUMN "avatar_path" TEXT;

CREATE TABLE "user_phone_numbers" (
  "id" UUID NOT NULL,
  "user_id" UUID NOT NULL,
  "phone_number" TEXT NOT NULL,
  "label" TEXT,
  "is_primary" BOOLEAN NOT NULL DEFAULT false,
  "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT "user_phone_numbers_pkey" PRIMARY KEY ("id"),
  CONSTRAINT "user_phone_numbers_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX "user_phone_numbers_user_id_idx" ON "user_phone_numbers"("user_id");
