-- Master Data: Segment & Category
-- Segments are top-level business classifications; categories belong to a segment.

CREATE TABLE "segments" (
  "id" UUID NOT NULL,
  "name" VARCHAR(100) NOT NULL,
  "description" TEXT NOT NULL DEFAULT '',
  "status" VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
  "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT "segments_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "segments_name_key" ON "segments"("name");

CREATE TABLE "categories" (
  "id" UUID NOT NULL,
  "segment_id" UUID NOT NULL,
  "name" VARCHAR(100) NOT NULL,
  "description" TEXT NOT NULL DEFAULT '',
  "status" VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
  "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT "categories_pkey" PRIMARY KEY ("id"),
  CONSTRAINT "categories_segment_id_fkey" FOREIGN KEY ("segment_id") REFERENCES "segments"("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

CREATE INDEX "categories_segment_id_idx" ON "categories"("segment_id");
CREATE UNIQUE INDEX "categories_segment_id_name_key" ON "categories"("segment_id", "name");

-- Seed default master data
INSERT INTO "segments" ("id", "name", "description", "status") VALUES
  (gen_random_uuid(), 'Food & Beverage', 'Bisnis makanan dan minuman', 'ACTIVE'),
  (gen_random_uuid(), 'Hospitality', 'Hotel, resort, dan akomodasi', 'ACTIVE'),
  (gen_random_uuid(), 'Retail', 'Bisnis ritel dan toko', 'ACTIVE'),
  (gen_random_uuid(), 'General Trade', 'Warung dan toko tradisional', 'ACTIVE'),
  (gen_random_uuid(), 'Modern Trade', 'Ritel modern dan franchise', 'ACTIVE'),
  (gen_random_uuid(), 'Food Service', 'Jasa boga dan dapur skala besar', 'ACTIVE'),
  (gen_random_uuid(), 'Key Account', 'Akun strategis nasional', 'ACTIVE');

INSERT INTO "categories" ("id", "segment_id", "name", "description", "status")
SELECT gen_random_uuid(), s."id", c."name", c."description", 'ACTIVE'
FROM (VALUES
  ('Food & Beverage', 'Restaurant', 'Restaurant / tempat makan'),
  ('Food & Beverage', 'Cafe', 'Kafe dan coffee shop'),
  ('Food & Beverage', 'Bakery', 'Toko roti dan kue'),
  ('Food & Beverage', 'Catering', 'Jasa katering'),
  ('Food & Beverage', 'Food Court', 'Pusat jajanan food court'),
  ('Hospitality', 'Hotel', 'Hotel dan penginapan'),
  ('Hospitality', 'Resort', 'Resort dan villa liburan'),
  ('Hospitality', 'Villa', 'Villa dan homestay'),
  ('Hospitality', 'Guest House', 'Guest house dan losmen'),
  ('Retail', 'Supermarket', 'Supermarket dan grocery'),
  ('Retail', 'Minimarket', 'Minimarket'),
  ('Retail', 'Retail Store', 'Toko ritel lainnya'),
  ('General Trade', 'Warung', 'Warung makan dan sembako'),
  ('General Trade', 'Small Grocer', 'Toko kelontong kecil'),
  ('Modern Trade', 'Hypermarket', 'Hypermarket'),
  ('Modern Trade', 'Convenience Store', 'Convenience store / franchise'),
  ('Food Service', 'Cloud Kitchen', 'Dapur produksi tanpa ruang makan'),
  ('Food Service', 'Central Kitchen', 'Dapur pusat'),
  ('Key Account', 'Office', 'Kantor dan institusi bisnis'),
  ('Key Account', 'Institutional', 'Institusi pemerintah dan pendidikan')
) AS c(segment_name, name, description)
JOIN "segments" s ON s."name" = c.segment_name;
