-- Master Data: final Google Place Types (place_api) per category.
-- Replaces the previous keyword lists with the approved taxonomy mapping.
-- Only updates place_api; category rows, segments and status are preserved.

UPDATE "categories" SET "place_api" = 'restaurant, cafe, coffee_shop, coffee_stand, coffee_roastery, cafeteria, bistro, diner, family_restaurant, fine_dining_restaurant, buffet_restaurant, breakfast_restaurant, brunch_restaurant, food_court, gastropub, bar_and_grill, barbecue_restaurant, seafood_restaurant, steak_house, indonesian_restaurant, asian_restaurant, asian_fusion_restaurant, chinese_restaurant, japanese_restaurant, korean_restaurant, thai_restaurant, vietnamese_restaurant, malaysian_restaurant, western_restaurant, italian_restaurant, french_restaurant, mediterranean_restaurant, middle_eastern_restaurant, mexican_restaurant, indian_restaurant, vegetarian_restaurant, vegan_restaurant, hot_pot_restaurant, sushi_restaurant, ramen_restaurant, noodle_shop' WHERE "name" = 'Resto & Cafe' AND "deleted_at" IS NULL;

UPDATE "categories" SET "place_api" = 'fast_food_restaurant, meal_takeaway, meal_delivery, hamburger_restaurant, chicken_restaurant, chicken_wings_restaurant, pizza_restaurant, pizza_delivery, sandwich_shop, hot_dog_restaurant, hot_dog_stand, kebab_shop, shawarma_restaurant, taco_restaurant, burrito_restaurant, falafel_restaurant, snack_bar, salad_shop, noodle_shop, dumpling_restaurant' WHERE "name" = 'QSR / Fast Food' AND "deleted_at" IS NULL;

UPDATE "categories" SET "place_api" = 'bakery, cake_shop, pastry_shop, confectionery, dessert_shop, dessert_restaurant, donut_shop, bagel_shop, candy_store, chocolate_shop, chocolate_factory, ice_cream_shop, tea_house, juice_shop' WHERE "name" = 'Bakery & Dessert' AND "deleted_at" IS NULL;

UPDATE "categories" SET "place_api" = 'hotel, resort_hotel, lodging, extended_stay_hotel, motel, inn, guest_house, hostel, bed_and_breakfast, farmstay, private_guest_room, cottage' WHERE "name" = 'Hotel & Accommodation' AND "deleted_at" IS NULL;

UPDATE "categories" SET "place_api" = 'catering_service, banquet_hall, event_venue, wedding_venue, convention_center, community_center' WHERE "name" = 'Catering & Event' AND "deleted_at" IS NULL;

UPDATE "categories" SET "place_api" = 'supermarket, hypermarket, discount_supermarket, department_store, warehouse_store, shopping_mall, grocery_store, food_store' WHERE "name" = 'Modern Trade' AND "deleted_at" IS NULL;

UPDATE "categories" SET "place_api" = 'convenience_store, grocery_store, food_store, general_store, market' WHERE "name" = 'Convenience Store' AND "deleted_at" IS NULL;

UPDATE "categories" SET "place_api" = 'grocery_store, general_store, food_store, market, farmers_market, asian_grocery_store, butcher_shop, health_food_store, store' WHERE "name" = 'General Trade' AND "deleted_at" IS NULL;

UPDATE "categories" SET "place_api" = 'wholesaler, supplier, warehouse_store, corporate_office, business_center, manufacturer, food_store' WHERE "name" = 'Distributor / Agent' AND "deleted_at" IS NULL;

UPDATE "categories" SET "place_api" = 'manufacturer, supplier, corporate_office, farm, ranch, chocolate_factory, brewery, winery' WHERE "name" = 'Industry / Manufacturer' AND "deleted_at" IS NULL;

UPDATE "categories" SET "place_api" = 'food_store, grocery_store, general_store, wholesaler, supplier, warehouse_store, market, cake_shop, bakery' WHERE "name" = 'Toko Bahan Kue / Baking Supply' AND "deleted_at" IS NULL;

UPDATE "categories" SET "place_api" = 'school, university, hospital, general_hospital, medical_center, government_office, corporate_office, business_center' WHERE "name" = 'Potential Institutional' AND "deleted_at" IS NULL;
