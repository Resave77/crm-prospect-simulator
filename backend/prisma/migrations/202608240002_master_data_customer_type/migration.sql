-- Master Data: Customer Type (B2B / B2C) sits above Segment & Category.
-- Hierarchy: Customer Type -> Segment -> Category.

CREATE TABLE "customer_types" (
  "id" UUID NOT NULL,
  "name" VARCHAR(100) NOT NULL,
  "description" TEXT NOT NULL DEFAULT '',
  "status" VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
  "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" TIMESTAMPTZ(6),
  CONSTRAINT "customer_types_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "customer_types_name_key" ON "customer_types"("name") WHERE "deleted_at" IS NULL;

ALTER TABLE "categories" ADD COLUMN "customer_type_id" UUID;
ALTER TABLE "categories" ADD CONSTRAINT "categories_customer_type_id_fkey"
  FOREIGN KEY ("customer_type_id") REFERENCES "customer_types"("id") ON DELETE SET NULL ON UPDATE CASCADE;
CREATE INDEX "categories_customer_type_id_idx" ON "categories"("customer_type_id");

INSERT INTO "customer_types" ("id", "name", "description", "status") VALUES
  (gen_random_uuid(), 'B2B', 'Business customer yang membeli produk untuk dijual kembali, digunakan dalam operasional, atau kebutuhan bisnis', 'ACTIVE'),
  (gen_random_uuid(), 'B2C', 'Individual customer yang membeli produk untuk kebutuhan konsumsi pribadi/rumah tangga', 'ACTIVE');

-- All live prospect categories are business prospects: assign them to B2B.
UPDATE "categories" SET "customer_type_id" = (SELECT "id" FROM "customer_types" WHERE "name" = 'B2B')
WHERE "deleted_at" IS NULL AND "customer_type_id" IS NULL;

-- B2C consumer segments & categories (no Place API keywords: not business places).
INSERT INTO "segments" ("id", "name", "description", "status")
SELECT gen_random_uuid(), v.segment_name, v.segment_desc, 'ACTIVE'
FROM (VALUES
  ('Direct Consumer', 'Konsumen akhir pembeli langsung'),
  ('Household', 'Konsumen rumah tangga'),
  ('Online Consumer', 'Konsumen kanal online')
) AS v(segment_name, segment_desc)
WHERE EXISTS (SELECT 1 FROM "segments" s WHERE s."name" = v.segment_name AND s."deleted_at" IS NULL) = FALSE;

INSERT INTO "categories" ("id", "segment_id", "name", "description", "place_api", "customer_type_id", "status")
SELECT gen_random_uuid(), s."id", c."cat_name", c."cat_desc", '', ct."id", 'ACTIVE'
FROM (VALUES
  ('Direct Consumer', 'Individual Consumer', 'Konsumen perorangan'),
  ('Household', 'Household Consumer', 'Konsumen rumah tangga'),
  ('Online Consumer', 'Online Consumer', 'Konsumen yang membeli melalui kanal online')
) AS c(segment_name, cat_name, cat_desc)
JOIN "segments" s ON s."name" = c.segment_name AND s."deleted_at" IS NULL
CROSS JOIN (SELECT "id" FROM "customer_types" WHERE "name" = 'B2C') ct
WHERE EXISTS (
  SELECT 1 FROM "categories" ic
  WHERE ic."segment_id" = s."id" AND ic."name" = c."cat_name" AND ic."deleted_at" IS NULL
) = FALSE;
