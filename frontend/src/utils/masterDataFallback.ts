import type { MasterDataCategory, MasterDataSegment } from '../api/masterData'

const groups: Record<string, string> = {
  'Resto & Café': 'restaurant,cafe,coffee_shop,coffee_stand,coffee_roastery,cafeteria,bistro,diner,family_restaurant,fine_dining_restaurant,buffet_restaurant,breakfast_restaurant,brunch_restaurant,food_court,gastropub,bar,bar_and_grill,barbecue_restaurant,seafood_restaurant,steak_house,indonesian_restaurant,asian_restaurant,asian_fusion_restaurant,chinese_restaurant,japanese_restaurant,korean_restaurant,thai_restaurant,vietnamese_restaurant,malaysian_restaurant,western_restaurant,italian_restaurant,french_restaurant,mediterranean_restaurant,middle_eastern_restaurant,mexican_restaurant,indian_restaurant,vegetarian_restaurant,vegan_restaurant,hot_pot_restaurant,sushi_restaurant,ramen_restaurant,noodle_shop',
  'QSR / Fast Food': 'fast_food_restaurant,meal_takeaway,meal_delivery,hamburger_restaurant,chicken_restaurant,chicken_wings_restaurant,pizza_restaurant,pizza_delivery,sandwich_shop,hot_dog_restaurant,hot_dog_stand,kebab_shop,shawarma_restaurant,taco_restaurant,burrito_restaurant,falafel_restaurant,snack_bar,salad_shop,noodle_shop,dumpling_restaurant',
  'Bakery & Dessert': 'bakery,cake_shop,pastry_shop,confectionery,dessert_shop,dessert_restaurant,donut_shop,bagel_shop,candy_store,chocolate_shop,chocolate_factory,ice_cream_shop,tea_house,juice_shop',
  'Hotel & Accommodation': 'hotel,resort_hotel,lodging,extended_stay_hotel,motel,inn,guest_house,hostel,bed_and_breakfast,farmstay,private_guest_room,cottage',
  'Catering & Event': 'catering_service,banquet_hall,event_venue,wedding_venue,convention_center,community_center',
  'Modern Trade': 'supermarket,hypermarket,discount_supermarket,department_store,warehouse_store,shopping_mall,grocery_store,food_store',
  'Convenience Store': 'convenience_store,grocery_store,food_store,general_store,market',
  'General Trade': 'grocery_store,general_store,food_store,market,farmers_market,asian_grocery_store,butcher_shop,health_food_store,store',
  'Distributor / Agent': 'wholesaler,supplier,warehouse_store,corporate_office,business_center,manufacturer,food_store',
  'Industry / Manufacturer': 'manufacturer,supplier,corporate_office,farm,ranch,chocolate_factory,brewery,winery',
  'Toko Bahan Kue / Baking Supply': 'food_store,grocery_store,general_store,wholesaler,supplier,warehouse_store,market,cake_shop,bakery',
  'Potential Institutional': 'school,university,hospital,general_hospital,medical_center,government_office,corporate_office,business_center'
}

export const fallbackSegments: MasterDataSegment[] = ['B2B', 'B2C'].map((name, i) => ({ id: `fallback-segment-${i}`, name, description: '', status: 'ACTIVE', categoryCount: 0, createdAt: '', updatedAt: '' }))
const b2cGroups = new Set(['Modern Trade', 'Convenience Store', 'General Trade', 'Bakery & Dessert'])
export const fallbackCategories: MasterDataCategory[] = Object.entries(groups).map(([name, placeApi], i) => ({ id: `fallback-category-${i}`, segmentId: b2cGroups.has(name) ? 'fallback-segment-1' : 'fallback-segment-0', segmentName: b2cGroups.has(name) ? 'B2C' : 'B2B', name, description: 'Default CRM category', placeApi, status: 'ACTIVE', createdAt: '', updatedAt: '' }))
