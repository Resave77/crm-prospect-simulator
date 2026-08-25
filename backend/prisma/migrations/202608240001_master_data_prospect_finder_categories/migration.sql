-- Master Data: align Segment & Category seed with the original Prospect Finder
-- category list. Replaces the previous demo rows so every category that used to
-- exist in Prospect Finder is available in Master Data, including the Google
-- Places keywords that used to power its search.

DELETE FROM "categories"
WHERE "name" IN (
  'Restaurant', 'Cafe', 'Bakery', 'Catering', 'Food Court',
  'Hotel', 'Resort', 'Villa', 'Guest House',
  'Supermarket', 'Minimarket', 'Retail Store',
  'Warung', 'Small Grocer',
  'Hypermarket', 'Convenience Store',
  'Cloud Kitchen', 'Central Kitchen',
  'Office', 'Institutional'
);

INSERT INTO "segments" ("id", "name", "description", "status")
SELECT gen_random_uuid(), 'Distributor & Industri', 'Distributor, agen, dan manufaktur industri', 'ACTIVE'
WHERE EXISTS (SELECT 1 FROM "segments" WHERE "name" = 'Distributor & Industri' AND "deleted_at" IS NULL) = FALSE;

INSERT INTO "categories" ("id", "segment_id", "name", "description", "place_api", "status")
SELECT gen_random_uuid(), s."id", c."name", c."description", c."keywords", 'ACTIVE'
FROM (VALUES
  ('Food & Beverage', 'Resto & Café', 'Restoran, kafe, dan coffee shop',
    'restaurant, cafe, coffee_shop, coffee_roastery, cafeteria, bistro, diner, family_restaurant, fine_dining_restaurant, buffet_restaurant, breakfast_restaurant, brunch_restaurant, food_court, gastropub, bar_and_grill, barbecue_restaurant, seafood_restaurant, steak_house, indonesian_restaurant, asian_restaurant, asian_fusion_restaurant, chinese_restaurant, japanese_restaurant, korean_restaurant, thai_restaurant, vietnamese_restaurant, malaysian_restaurant, western_restaurant, italian_restaurant, french_restaurant, mediterranean_restaurant, middle_eastern_restaurant, mexican_restaurant, indian_restaurant, vegetarian_restaurant, vegan_restaurant, hot_pot_restaurant, sushi_restaurant, ramen_restaurant, noodle_shop, bar, pub, cocktail_bar, wine_bar'),
  ('Food & Beverage', 'QSR / Fast Food', 'Restoran cepat saji dan takeaway',
    'fast_food_restaurant, meal_takeaway, meal_delivery, hamburger_restaurant, chicken_restaurant, chicken_wings_restaurant, pizza_restaurant, pizza_delivery, sandwich_shop, hot_dog_restaurant, hot_dog_stand, kebab_shop, shawarma_restaurant, taco_restaurant, burrito_restaurant, falafel_restaurant, snack_bar, salad_shop, dumpling_restaurant'),
  ('Food & Beverage', 'Bakery & Dessert', 'Toko roti, kue, dan dessert',
    'bakery, cake_shop, pastry_shop, confectionery, dessert_shop, dessert_restaurant, donut_shop, bagel_shop, candy_store, chocolate_shop, chocolate_factory, ice_cream_shop, tea_house, juice_shop'),
  ('Food & Beverage', 'Toko Bahan Kue / Baking Supply', 'Toko bahan kue dan kebutuhan bakeri',
    'food_store, grocery_store, general_store, wholesaler, supplier, warehouse_store, market, cake_shop, bakery'),
  ('Hospitality', 'Hotels & Accommodation', 'Hotel, resor, dan penginapan',
    'hotel, resort_hotel, lodging, extended_stay_hotel, motel, inn, guest_house, hostel, bed_and_breakfast, farmstay, private_guest_room, cottage'),
  ('Food Service', 'Catering & Event', 'Katering dan venue acara',
    'catering_service, banquet_hall, event_venue, wedding_venue, convention_center, community_center'),
  ('Modern Trade', 'Modern Trade', 'Supermarket, hipermarket, dan mal',
    'supermarket, hypermarket, discount_supermarket, department_store, warehouse_store, shopping_mall, grocery_store, food_store'),
  ('Retail', 'Convenience Store', 'Minimarket dan toko serba ada',
    'convenience_store, grocery_store, food_store, general_store, market'),
  ('General Trade', 'General Trade', 'Warung dan toko tradisional',
    'grocery_store, general_store, food_store, market, farmers_market, asian_grocery_store, butcher_shop, health_food_store, store'),
  ('Key Account', 'Institutional', 'Sekolah, universitas, rumah sakit, dan kantor pemerintah',
    'school, university, hospital, general_hospital, medical_center, government_office, corporate_office, business_center'),
  ('Distributor & Industri', 'Distributor / Agent', 'Distributor, agen, dan grosir',
    'wholesaler, supplier, warehouse_store, corporate_office, business_center, manufacturer, food_store'),
  ('Distributor & Industri', 'Industry / Manufacturer', 'Pabrik dan produsen industri',
    'manufacturer, supplier, corporate_office, farm, ranch, chocolate_factory, brewery, winery')
) AS c(segment_name, name, description, keywords)
JOIN "segments" s ON s."name" = c.segment_name AND s."deleted_at" IS NULL
WHERE NOT EXISTS (
  SELECT 1 FROM "categories" ic
  WHERE ic."segment_id" = s."id" AND ic."name" = c."name" AND ic."deleted_at" IS NULL
);
