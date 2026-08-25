-- Master Data: rebase the hierarchy on two segments, B2B and B2C, and remove
-- Customer Type entirely. Final approved structure:
--   B2B (8): Resto & Cafe, QSR / Fast Food, Bakery & Dessert,
--            Hotel & Accommodation, Catering & Event, Industry / Manufacturer,
--            Potential Institutional, Distributor / Agent
--   B2C (4): Modern Trade, Convenience Store, General Trade,
--            Toko Bahan Kue / Baking Supply

-- 1. Ensure the two canonical segments exist.
INSERT INTO "segments" ("id", "name", "description", "status")
SELECT gen_random_uuid(), v.segment_name, v.segment_desc, 'ACTIVE'
FROM (VALUES
  ('B2B', 'Business to Business'),
  ('B2C', 'Business to Consumer')
) AS v(segment_name, segment_desc)
WHERE EXISTS (
  SELECT 1 FROM "segments" s WHERE s."name" = v.segment_name AND s."deleted_at" IS NULL
) = FALSE;

-- 2. Align category labels with the approved taxonomy. Rows, IDs and Place API
--    keywords are preserved.
UPDATE "categories" SET "name" = 'Resto & Cafe' WHERE "name" = 'Resto & Café';
UPDATE "categories" SET "name" = 'Hotel & Accommodation' WHERE "name" = 'Hotels & Accommodation';
UPDATE "categories" SET "name" = 'Potential Institutional' WHERE "name" = 'Institutional';

-- 3. Re-parent the business categories under B2B.
UPDATE "categories" c
SET "segment_id" = (SELECT s."id" FROM "segments" s WHERE s."name" = 'B2B' AND s."deleted_at" IS NULL),
    "updated_at" = now()
WHERE c."deleted_at" IS NULL
  AND c."name" IN (
    'Resto & Cafe', 'QSR / Fast Food', 'Bakery & Dessert', 'Hotel & Accommodation',
    'Catering & Event', 'Industry / Manufacturer', 'Distributor / Agent'
  );

-- 4. Add the missing B2B category if it does not exist yet.
INSERT INTO "categories" ("id", "segment_id", "name", "description", "place_api", "status")
SELECT gen_random_uuid(), s."id", 'Potential Institutional',
  'Sekolah, universitas, rumah sakit, dan kantor pemerintah',
  'school, university, college, hospital, general_hospital, medical_center, government_office, corporate_office, business_center, library, museum',
  'ACTIVE'
FROM "segments" s
WHERE s."name" = 'B2B' AND s."deleted_at" IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM "categories" ic
    WHERE ic."segment_id" = s."id" AND ic."name" = 'Potential Institutional' AND ic."deleted_at" IS NULL
  );

-- 5. Re-parent the consumer-facing categories under B2C.
UPDATE "categories" c
SET "segment_id" = (SELECT s."id" FROM "segments" s WHERE s."name" = 'B2C' AND s."deleted_at" IS NULL),
    "updated_at" = now()
WHERE c."deleted_at" IS NULL
  AND c."name" IN ('Modern Trade', 'Convenience Store', 'General Trade', 'Toko Bahan Kue / Baking Supply');

-- 6. Remove the Customer Type relation from categories.
ALTER TABLE "categories" DROP CONSTRAINT IF EXISTS "categories_customer_type_id_fkey";
DROP INDEX IF EXISTS "categories_customer_type_id_idx";
ALTER TABLE "categories" DROP COLUMN IF EXISTS "customer_type_id";

-- 7. Drop the customer_types table.
DROP TABLE IF EXISTS "customer_types";

-- 8. Drop every category outside the approved structure (legacy demo rows and
--    the unused consumer placeholders).
DELETE FROM "categories"
WHERE "segment_id" NOT IN (
  SELECT s."id" FROM "segments" s WHERE s."name" IN ('B2B', 'B2C')
) OR "name" NOT IN (
  'Resto & Cafe', 'QSR / Fast Food', 'Bakery & Dessert', 'Hotel & Accommodation',
  'Catering & Event', 'Industry / Manufacturer', 'Potential Institutional', 'Distributor / Agent',
  'Modern Trade', 'Convenience Store', 'General Trade', 'Toko Bahan Kue / Baking Supply'
);

-- 9. Keep only B2B and B2C as segments.
DELETE FROM "segments" WHERE "name" NOT IN ('B2B', 'B2C');
