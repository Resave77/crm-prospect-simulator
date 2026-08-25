-- Master Data: Place API mapping for categories.
-- Optional Google Places API keywords used for prospect search.

ALTER TABLE "categories" ADD COLUMN "place_api" TEXT NOT NULL DEFAULT '';

UPDATE "categories" SET "place_api" = v.keywords
FROM (VALUES
  ('Restaurant', 'restaurant'),
  ('Cafe', 'cafe, coffee_shop'),
  ('Bakery', 'bakery, cake_shop'),
  ('Catering', 'catering_service'),
  ('Food Court', 'food_court'),
  ('Hotel', 'hotel, lodging, motel, inn'),
  ('Resort', 'resort, lodging'),
  ('Villa', 'villa, lodging'),
  ('Guest House', 'guest_house, lodging'),
  ('Supermarket', 'supermarket, grocery_store'),
  ('Minimarket', 'convenience_store, grocery_store'),
  ('Retail Store', 'store, department_store'),
  ('Warung', 'restaurant, grocery_store, market'),
  ('Small Grocer', 'grocery_store, convenience_store'),
  ('Hypermarket', 'hypermarket, supermarket, department_store'),
  ('Convenience Store', 'convenience_store, grocery_store'),
  ('Cloud Kitchen', 'meal_delivery, meal_takeaway'),
  ('Central Kitchen', 'meal_takeaway, food_processor'),
  ('Office', 'corporate_office, business_center'),
  ('Institutional', 'school, university, hospital, government_office')
) AS v(name, keywords)
WHERE "categories"."name" = v.name;
