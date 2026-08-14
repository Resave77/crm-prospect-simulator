<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Slider from 'primevue/slider'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { useToast } from 'primevue/usetoast'
import * as crmApi from '../../../api/crm'
import { getPlacePhotoBlob } from '../../../api/crm'
import type { CustomerMarker, MenuImage, PlaceDetails, PlacePhoto, PlaceResult, SalesExecutiveOption } from '../../../types/crm'

const toast = useToast()

const categoryOptions = [
  { key: 'resto_cafe', label: 'Resto & Café', icon: '🍽️' },
  { key: 'qsr_fast_food', label: 'QSR / Fast Food', icon: '🍔' },
  { key: 'bakery_dessert', label: 'Bakery & Dessert', icon: '🎂' },
  { key: 'hotels_accommodation', label: 'Hotels & Accommodation', icon: '🏨' },
  { key: 'catering_event', label: 'Catering & Event', icon: '🎪' },
  { key: 'modern_trade', label: 'Modern Trade', icon: '🛒' },
  { key: 'convenience_store', label: 'Convenience Store', icon: '🏪' },
  { key: 'general_trade', label: 'General Trade', icon: '🏬' },
  { key: 'distributor_agent', label: 'Distributor / Agent', icon: '📦' },
  { key: 'industry_manufacturer', label: 'Industry / Manufacturer', icon: '🏭' },
  { key: 'baking_supply', label: 'Toko Bahan Kue / Baking Supply', icon: '🥣' },
  { key: 'institutional', label: 'Institutional', icon: '🏫' },
] as const
const ratingOptions = [
  { value: 0, label: 'All' }, { value: 3, label: '3★' }, { value: 4, label: '4★' }, { value: 4.5, label: '4.5★' }, { value: 4.8, label: '4.8★' },
] as const
const savedFilterOptions = [
  { value: 'all', label: 'All' }, { value: 'saved', label: 'Saved' }, { value: 'unsaved', label: 'Unsaved' },
] as const
const menuFilterOptions = [
  { value: 'all', label: 'All' }, { value: 'likely', label: 'Likely Has Menu' }, { value: 'ready', label: 'Menu Ready' }, { value: 'not_ready', label: 'Menu Not Ready' },
] as const
const keyword = ref('')
const categories = ref<string[]>(categoryOptions.map(o => o.key))
const radius = ref(5000)
const minRating = ref(0)
const savedFilter = ref<'all' | 'saved' | 'unsaved'>('all')
const menuFilter = ref<'all' | 'likely' | 'ready' | 'not_ready'>('all')
const savedPlaceIds = ref<Set<string>>(new Set())
const queried = ref(false)
const latitude = ref(0)
const longitude = ref(0)
const geoResolved = ref(false)
const results = ref<PlaceResult[]>([])
const resultSearch = ref('')
const selected = ref<PlaceResult | null>(null)
const customerMarkers = ref<CustomerMarker[]>([])
const customersLoading = ref(false)
const placeDetails = ref<PlaceDetails | null>(null)
const placeDetailsLoading = ref(false)
const placeDetailsError = ref('')
const menuImages = ref<MenuImage[]>([])
const menuImagesLoading = ref(false)
const menuImagesError = ref('')
const visiblePhotoCount = ref(1)
const sales = ref<SalesExecutiveOption[]>([])
const salesExecutiveId = ref('')
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const success = ref('')
const detailOpen = ref(false)
const previewOpen = ref(false)
const previewPhoto = ref<{ url: string; alt: string; attribution?: string } | null>(null)
const photoObjectUrls = ref<Record<string, string>>({})
const failedPhotoNames = ref<Set<string>>(new Set())
const pinOpen = ref(false)
const pinName = ref('')
const pinLat = ref(0)
const pinLng = ref(0)
const pinSaving = ref(false)
const pinError = ref('')
const pinDragging = ref(false)
const mapElement = ref<HTMLElement | null>(null)
const resultsScroll = ref<HTMLElement | null>(null)
let map: L.Map | null = null
let searchCircle: L.Circle | null = null
let centerMarker: L.Marker | null = null
let customPinMarker: L.Marker | null = null
let draggingCenter = false
let zooming = false
let customerLayer: L.LayerGroup | null = null
const markers = new Map<string, L.Marker>()
const customerMarkerMap = new Map<string, L.Marker>()

const selectedSalesCount = computed(() => {
  const exec = sales.value.find(s => s.id === salesExecutiveId.value)
  return exec?.activeProspectCount ?? 0
})
const selectedPlaceCategory = computed(() =>
  placeDetails.value?.placeCategory?.trim() || selected.value?.category?.trim() || '',
)
const pinCategory = 'Custom Pin'

// Places API does not expose the Google Maps UI's Menu/Photo grouping.
// Ignore legacy `isMenu` values that were inferred from landscape dimensions;
// those values incorrectly classified storefront and street photos as menus.
const menuPhotos = computed<PlacePhoto[]>(() => [])
const regularPhotos = computed(() => placeDetails.value?.photos ?? [])
const visiblePhotos = computed(() => regularPhotos.value.slice(0, visiblePhotoCount.value))
const hasMorePhotos = computed(() => visiblePhotoCount.value < regularPhotos.value.length)

let photoLoadToken = 0

function revokePlacePhotoObjectUrls() {
  photoLoadToken += 1
  Object.values(photoObjectUrls.value).forEach((url) => URL.revokeObjectURL(url))
  photoObjectUrls.value = {}
  failedPhotoNames.value = new Set()
}

function resolvedPlacePhotoUrl(photo: PlacePhoto) {
  return photoObjectUrls.value[photo.name] ?? ''
}

function canShowPlacePhoto(photo: PlacePhoto) {
  return !!resolvedPlacePhotoUrl(photo) && !failedPhotoNames.value.has(photo.name)
}

async function loadPlacePhotoObjectUrls(photos: PlacePhoto[], reset = false) {
  if (reset) revokePlacePhotoObjectUrls()
  const token = photoLoadToken
  const unloaded = photos.filter(photo => !photoObjectUrls.value[photo.name] && !failedPhotoNames.value.has(photo.name))
  if (!unloaded.length) return

  await Promise.all(unloaded.map(async (photo) => {
    try {
      const blob = await getPlacePhotoBlob(photo.photoUrl)
      if (token !== photoLoadToken) {
        return
      }
      const objectUrl = URL.createObjectURL(blob)
      photoObjectUrls.value = { ...photoObjectUrls.value, [photo.name]: objectUrl }
    } catch {
      if (token !== photoLoadToken) return
      const next = new Set(failedPhotoNames.value)
      next.add(photo.name)
      failedPhotoNames.value = next
    }
  }))
}

async function loadNextPlacePhoto() {
  if (!hasMorePhotos.value) return
  visiblePhotoCount.value += 1
  const photo = regularPhotos.value[visiblePhotoCount.value - 1]
  if (photo) await loadPlacePhotoObjectUrls([photo])
}

async function loadMenuImages() {
  if (!placeDetails.value || menuImagesLoading.value) return
  const query = menuImageQuery()
  if (!query) return
  menuImagesLoading.value = true
  menuImagesError.value = ''
  try {
    menuImages.value = (await crmApi.getMenuImages(query, 3)).slice(0, 3)
    if (selected.value) selected.value.hasMenuPhotos = menuImages.value.length > 0
  } catch (caught) {
    menuImagesError.value = crmError(caught)
    menuImages.value = []
  } finally {
    menuImagesLoading.value = false
  }
}

function menuImageQuery() {
  const d = placeDetails.value
  if (!d) return ''
  const types = d.placeTypes ?? []
  const foodKeywords: Array<[string, string]> = [
    ['fast_food_restaurant', 'menu fast food'],
    ['hamburger_restaurant', 'menu hamburger'],
    ['pizza_restaurant', 'menu pizza'],
    ['pizza_delivery', 'menu pizza delivery'],
    ['sandwich_shop', 'menu sandwich'],
    ['kebab_shop', 'menu kebab'],
    ['shawarma_restaurant', 'menu shawarma'],
    ['taco_restaurant', 'menu taco'],
    ['burrito_restaurant', 'menu burrito'],
    ['falafel_restaurant', 'menu falafel'],
    ['hot_dog_restaurant', 'menu hot dog'],
    ['hot_dog_stand', 'menu hot dog'],
    ['snack_bar', 'snack menu'],
    ['meal_takeaway', 'takeaway food menu'],
    ['meal_delivery', 'delivery food menu'],
    ['food_court', 'menu food court'],
    ['dessert_restaurant', 'dessert menu'],
    ['ice_cream_shop', 'ice cream menu'],
    ['bakery', 'bakery menu'],
    ['pastry_shop', 'pastry menu'],
    ['cake_shop', 'cake menu'],
    ['confectionery', 'dessert menu'],
    ['dessert_shop', 'dessert menu'],
    ['donut_shop', 'donut menu'],
    ['bagel_shop', 'bagel menu'],
    ['candy_store', 'candy menu'],
    ['coffee_shop', 'coffee menu'],
    ['coffee_roastery', 'coffee menu'],
    ['cafe', 'cafe menu'],
    ['cafeteria', 'cafeteria menu'],
    ['restaurant', 'restaurant menu'],
    ['bistro', 'bistro menu'],
    ['diner', 'diner menu'],
    ['family_restaurant', 'family restaurant menu'],
    ['fine_dining_restaurant', 'fine dining menu'],
    ['buffet_restaurant', 'buffet menu'],
    ['breakfast_restaurant', 'breakfast menu'],
    ['brunch_restaurant', 'brunch menu'],
    ['gastropub', 'gastropub menu'],
    ['bar_and_grill', 'bar grill menu'],
    ['barbecue_restaurant', 'barbecue menu'],
    ['seafood_restaurant', 'seafood menu'],
    ['steak_house', 'steak house menu'],
    ['indonesian_restaurant', 'indonesian food menu'],
    ['asian_restaurant', 'asian food menu'],
    ['asian_fusion_restaurant', 'asian fusion menu'],
    ['chinese_restaurant', 'chinese food menu'],
    ['japanese_restaurant', 'japanese food menu'],
    ['korean_restaurant', 'korean food menu'],
    ['thai_restaurant', 'thai food menu'],
    ['vietnamese_restaurant', 'vietnamese food menu'],
    ['malaysian_restaurant', 'malaysian food menu'],
    ['western_restaurant', 'western food menu'],
    ['italian_restaurant', 'italian food menu'],
    ['french_restaurant', 'french food menu'],
    ['mediterranean_restaurant', 'mediterranean food menu'],
    ['middle_eastern_restaurant', 'middle eastern food menu'],
    ['mexican_restaurant', 'mexican food menu'],
    ['indian_restaurant', 'indian food menu'],
    ['vegetarian_restaurant', 'vegetarian menu'],
    ['vegan_restaurant', 'vegan menu'],
    ['hot_pot_restaurant', 'hot pot menu'],
    ['sushi_restaurant', 'sushi menu'],
    ['ramen_restaurant', 'ramen menu'],
    ['noodle_shop', 'noodle menu'],
    ['catering_service', 'catering menu'],
    ['banquet_hall', 'event catering menu'],
    ['event_venue', 'event menu'],
    ['wedding_venue', 'wedding menu'],
    ['convention_center', 'meeting menu'],
    ['community_center', 'community menu'],
    ['supermarket', 'grocery menu'],
    ['hypermarket', 'food menu'],
    ['discount_supermarket', 'food menu'],
    ['department_store', 'food menu'],
    ['warehouse_store', 'food menu'],
    ['shopping_mall', 'food menu'],
    ['grocery_store', 'grocery menu'],
    ['food_store', 'food store menu'],
    ['convenience_store', 'store menu'],
    ['general_store', 'store menu'],
    ['market', 'market menu'],
    ['farmers_market', 'market menu'],
    ['asian_grocery_store', 'asian grocery menu'],
    ['butcher_shop', 'butcher menu'],
    ['health_food_store', 'health food menu'],
    ['wholesaler', 'wholesale food menu'],
    ['supplier', 'supplier menu'],
    ['manufacturer', 'manufacturer menu'],
    ['business_center', 'business catering'],
    ['corporate_office', 'office catering'],
    ['hospital', 'hospital menu'],
    ['general_hospital', 'hospital menu'],
    ['medical_center', 'medical center menu'],
    ['government_office', 'office catering'],
    ['school', 'canteen menu'],
    ['university', 'campus menu'],
  ]
  const keyword = foodKeywords.find(([type]) => types.includes(type))?.[1] ?? 'restaurant menu'
  return `"${d.placeName}" ${d.formattedAddress} ${keyword} menu harga daftar harga`.trim()
}

watch(
  () => (placeDetails.value?.photos ?? []).map((photo) => `${photo.name}:${photo.photoUrl}`).join('|'),
  () => {
    visiblePhotoCount.value = 1
    loadPlacePhotoObjectUrls((placeDetails.value?.photos ?? []).slice(0, 1), true)
  },
  { immediate: true },
)

const filteredResults = ref<PlaceResult[]>([])

const menuBearingPrimaryTypes = new Set([
  'restaurant', 'cafe', 'coffee_shop', 'coffee_roastery', 'cafeteria', 'bistro', 'diner',
  'family_restaurant', 'fine_dining_restaurant', 'buffet_restaurant', 'breakfast_restaurant',
  'brunch_restaurant', 'food_court', 'gastropub', 'bar_and_grill', 'barbecue_restaurant',
  'seafood_restaurant', 'steak_house', 'indonesian_restaurant', 'asian_restaurant',
  'asian_fusion_restaurant', 'chinese_restaurant', 'japanese_restaurant', 'korean_restaurant',
  'thai_restaurant', 'vietnamese_restaurant', 'malaysian_restaurant', 'western_restaurant',
  'italian_restaurant', 'french_restaurant', 'mediterranean_restaurant', 'middle_eastern_restaurant',
  'mexican_restaurant', 'indian_restaurant', 'vegetarian_restaurant', 'vegan_restaurant',
  'hot_pot_restaurant', 'sushi_restaurant', 'ramen_restaurant', 'noodle_shop',
  'fast_food_restaurant', 'meal_takeaway', 'meal_delivery', 'hamburger_restaurant',
  'chicken_restaurant', 'chicken_wings_restaurant', 'pizza_restaurant', 'pizza_delivery',
  'sandwich_shop', 'hot_dog_restaurant', 'hot_dog_stand', 'kebab_shop', 'shawarma_restaurant',
  'taco_restaurant', 'burrito_restaurant', 'falafel_restaurant', 'snack_bar', 'salad_shop',
  'dumpling_restaurant', 'bakery', 'cake_shop', 'pastry_shop', 'confectionery', 'dessert_shop',
  'dessert_restaurant', 'donut_shop', 'bagel_shop', 'candy_store', 'chocolate_shop',
  'ice_cream_shop', 'tea_house', 'juice_shop', 'catering_service',
])

function likelyHasMenu(item: PlaceResult) {
  return menuBearingPrimaryTypes.has(item.placeTypes?.[0] ?? '')
}

watch([results, resultSearch, minRating, savedFilter, menuFilter, savedPlaceIds], () => {
  const q = resultSearch.value.toLowerCase().trim()
  const min = minRating.value
  const mode = savedFilter.value
  const menuMode = menuFilter.value
  filteredResults.value = results.value.filter(r => {
    if (q && !(r.name.toLowerCase().includes(q) || r.category.toLowerCase().includes(q) || r.address.toLowerCase().includes(q))) return false
    if (min && r.rating < min) return false
    if (mode === 'saved' && !(r.isCustomer || savedPlaceIds.value.has(r.googlePlaceId))) return false
    if (mode === 'unsaved' && (r.isCustomer || savedPlaceIds.value.has(r.googlePlaceId))) return false
    if (menuMode === 'likely' && !likelyHasMenu(r)) return false
    if (menuMode === 'ready' && !r.hasMenuPhotos) return false
    if (menuMode === 'not_ready' && r.hasMenuPhotos) return false
    return true
  })
}, { immediate: true })

watch(filteredResults, () => {
  if (map && results.value.length) renderMarkers()
})

function selectAllCategories() {
  categories.value = categoryOptions.map(o => o.key)
}

function openPhotoPreview(url: string, alt: string, attribution?: string) {
  previewPhoto.value = { url, alt, attribution }
  previewOpen.value = true
}

function markerIcon(item: PlaceResult, active = false) {
  if (item.isCustomer) {
    return customerMarkerIcon(active)
  }
  const safeIcon = /^pi pi-[a-z-]+$/.test(item.markerIcon) ? item.markerIcon : 'pi pi-map-marker'
  const safeColor = /^#[0-9a-f]{6}$/i.test(item.markerColor) ? item.markerColor : '#e63946'
  return L.divIcon({
    className: 'finder-leaflet-icon-host',
    html: `<span class="finder-leaflet-marker${active ? ' is-selected' : ''}" style="--marker-color:${safeColor}"><i class="${safeIcon}"></i></span>`,
    iconSize: active ? [44, 50] : [36, 42],
    iconAnchor: active ? [22, 48] : [18, 40],
    popupAnchor: [0, -42],
  })
}

function customerMarkerIcon(active = false) {
  return L.divIcon({
    className: 'finder-leaflet-icon-host',
    html: `<span class="finder-leaflet-marker is-customer${active ? ' is-selected' : ''}" title="Existing customer"><b>Y</b></span>`,
    iconSize: active ? [46, 52] : [38, 44],
    iconAnchor: active ? [23, 50] : [19, 42],
    popupAnchor: [0, -44],
  })
}

function centerMarkerIcon() {
  return L.divIcon({
    className: 'finder-center-icon-host',
    html: '<span class="finder-center-marker"><i class="pi pi-crosshairs"></i></span><span class="finder-center-ring"></span>',
    iconSize: [34, 34],
    iconAnchor: [17, 17],
  })
}

function customPinIcon(active = false) {
  return L.divIcon({
    className: 'finder-leaflet-icon-host',
    html: `<span class="finder-leaflet-marker is-custom-pin${active ? ' is-selected' : ''}" title="Custom pin"><i class="pi pi-map-marker"></i></span>`,
    iconSize: active ? [44, 50] : [36, 42],
    iconAnchor: active ? [22, 48] : [18, 40],
    popupAnchor: [0, -42],
  })
}

function renderPinMarker() {
  if (!map) return
  customPinMarker?.remove()
  customPinMarker = L.marker([pinLat.value, pinLng.value], { icon: customPinIcon(), draggable: true, keyboard: true, zIndexOffset: 800, title: 'Pin location' })
    .bindTooltip('Drag to position', { direction: 'top', offset: [0, -28] })
    .on('dragstart', () => { pinDragging.value = true })
    .on('drag', () => {
      const pos = customPinMarker!.getLatLng()
      pinLat.value = pos.lat
      pinLng.value = pos.lng
    })
    .on('dragend', () => { pinDragging.value = false })
    .addTo(map)
}

watch([pinLat, pinLng], () => {
  if (map && !pinDragging.value) customPinMarker?.setLatLng([pinLat.value, pinLng.value])
})

function openPinForm() {
  if (!map) return
  detailOpen.value = false
  pinName.value = ''
  pinError.value = ''
  pinLat.value = latitude.value
  pinLng.value = longitude.value
  pinOpen.value = true
  nextTick(() => {
    renderPinMarker()
    map?.flyTo([pinLat.value, pinLng.value], Math.max(map.getZoom(), 15), { duration: 0.5 })
  })
}

function closePinForm() {
  pinOpen.value = false
  customPinMarker?.remove()
  customPinMarker = null
}

async function savePin() {
  if (!pinName.value.trim()) {
    toast.add({ severity: 'warn', summary: 'Missing name', detail: 'Enter a name for the pin.', life: 4000 })
    return
  }
  if (!salesExecutiveId.value) {
    toast.add({ severity: 'warn', summary: 'Missing information', detail: 'Select a Sales Executive before saving.', life: 4000 })
    return
  }
  pinError.value = ''
  pinSaving.value = true
  try {
    const lat = pinLat.value
    const lng = pinLng.value
    const result: PlaceResult = {
      googlePlaceId: `CUSTOM_PIN_${lat.toFixed(6)},${lng.toFixed(6)}`,
      name: pinName.value.trim(),
      category: 'Custom Pin',
      address: `${lat.toFixed(6)}, ${lng.toFixed(6)}`,
      distance: 0,
      rating: 0,
      userRatingCount: 0,
      businessStatus: 'OPERATIONAL',
      latitude: lat,
      longitude: lng,
      phone: '',
      website: '',
      googleMapsUrl: '',
      markerCategory: 'custom',
      markerColor: '#7c3aed',
      markerIcon: 'pi pi-map-marker',
      placeTypes: [],
    }
    const item = await crmApi.saveProspect(result, pinCategory, salesExecutiveId.value)
    toast.add({ severity: 'success', summary: 'Pin saved', detail: `${item.placeName} was saved as NEW_LEAD and assigned successfully.`, life: 5000 })
    closePinForm()
    loadSavedPlaceIds()
  } catch (caught) {
    pinError.value = crmError(caught)
    toast.add({ severity: 'error', summary: 'Save failed', detail: pinError.value, life: 6000 })
  } finally {
    pinSaving.value = false
  }
}

function customerToResult(c: CustomerMarker): PlaceResult {
  return {
    customerId: c.customerId,
    googlePlaceId: c.googlePlaceId,
    name: c.name,
    address: c.address,
    category: 'Existing Customer',
    distance: 0,
    rating: 0,
    userRatingCount: 0,
    businessStatus: 'OPERATIONAL',
    latitude: c.latitude,
    longitude: c.longitude,
    phone: '',
    website: '',
    googleMapsUrl: '',
    markerCategory: 'customer',
    markerColor: '#16a34a',
    markerIcon: '',
    placeTypes: [],
    isCustomer: true,
  }
}

function initializeMap() {
  if (!mapElement.value || map) return
  map = L.map(mapElement.value, { zoomControl: true, preferCanvas: true }).setView([-6.2, 106.8], 12)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
  }).addTo(map)
  const center = L.marker([latitude.value, longitude.value], {
    draggable: true,
    icon: centerMarkerIcon(),
    keyboard: true,
    zIndexOffset: 900,
    title: 'Drag to move the search center',
  })
    .bindTooltip(`${latitude.value.toFixed(5)}, ${longitude.value.toFixed(5)}`, { direction: 'top', offset: [0, -20] })
    .on('dragstart', () => { draggingCenter = true })
    .on('drag', () => {
      const pos = center.getLatLng()
      latitude.value = pos.lat
      longitude.value = pos.lng
      searchCircle?.setLatLng(pos)
      center.setTooltipContent(`${pos.lat.toFixed(5)}, ${pos.lng.toFixed(5)}`)
    })
    .on('dragend', () => { draggingCenter = false })
    .addTo(map)
  centerMarker = center
  map.on('zoomstart', () => { zooming = true })
  map.on('zoomend', () => { zooming = false })
  map.on('move', () => {
    if (draggingCenter || zooming || !map) return
    const c = map.getCenter()
    latitude.value = c.lat
    longitude.value = c.lng
    centerMarker?.setLatLng([c.lat, c.lng])
    searchCircle?.setLatLng([c.lat, c.lng])
    centerMarker?.setTooltipContent(`${c.lat.toFixed(5)}, ${c.lng.toFixed(5)}`)
  })
  const initial = map.getCenter()
  latitude.value = initial.lat
  longitude.value = initial.lng
  centerMarker.setLatLng([initial.lat, initial.lng])
  drawSearchArea()
  customerLayer = L.layerGroup().addTo(map)
}

function renderCustomerMarkers() {
  if (!map || !customerLayer) return
  customerLayer.clearLayers()
  customerMarkerMap.clear()
  const bounds: L.LatLngExpression[] = []
  for (const item of customerMarkers.value) {
    if (item.latitude === null || item.longitude === null) continue
    const position: L.LatLngExpression = [item.latitude, item.longitude]
    const marker = L.marker(position, { icon: customerMarkerIcon(selected.value?.customerId === item.customerId), keyboard: true, title: item.name })
      .bindTooltip(item.name, { direction: 'top', offset: [0, -42] })
      .on('click', () => selectCustomer(item))
      .addTo(customerLayer)
    customerMarkerMap.set(item.customerId, marker)
    bounds.push(position)
  }
  if (!results.value.length && bounds.length) {
    map.fitBounds(L.latLngBounds(bounds), { padding: [48, 48], maxZoom: 14 })
  }
}

async function loadCustomerMarkers() {
  customersLoading.value = true
  try {
    customerMarkers.value = await crmApi.getCustomerMarkers()
  } catch (caught) {
    toast.add({ severity: 'warn', summary: 'Customers unavailable', detail: crmError(caught), life: 5000 })
  } finally {
    customersLoading.value = false
    renderCustomerMarkers()
  }
}

function drawSearchArea() {
  if (!map) return
  const latlng: L.LatLngExpression = [latitude.value, longitude.value]
  if (searchCircle) {
    searchCircle.setLatLng(latlng)
    searchCircle.setRadius(radius.value)
  } else {
    searchCircle = L.circle(latlng, {
      radius: radius.value,
      color: '#e63946',
      weight: 1,
      fillColor: '#ef4e5d',
      fillOpacity: 0.06,
    }).addTo(map)
  }
  centerMarker?.setLatLng(latlng)
  centerMarker?.setTooltipContent(`${latitude.value.toFixed(5)}, ${longitude.value.toFixed(5)}`)
}

function renderMarkers() {
  if (!map) return
  markers.forEach((marker) => marker.remove())
  markers.clear()
  for (const item of filteredResults.value) {
    if (item.isCustomer) continue
    if (item.latitude === null || item.longitude === null) continue
    const marker = L.marker([item.latitude, item.longitude], { icon: markerIcon(item, selected.value?.googlePlaceId === item.googlePlaceId), keyboard: true, title: item.name })
      .bindTooltip(item.name, { direction: 'top', offset: [0, -34] })
      .on('click', () => selectResult(item, true))
      .addTo(map)
    markers.set(item.googlePlaceId, marker)
  }
  drawSearchArea()
  nextTick(() => map?.invalidateSize())
}

function closeResults() { results.value = []; filteredResults.value = []; resultSearch.value = ''; detailOpen.value = false; placeDetails.value = null; queried.value = false; nextTick(() => map?.invalidateSize()) }

async function loadSavedPlaceIds() {
  try {
    const items = await crmApi.getPipeline()
    savedPlaceIds.value = new Set(items.map(p => p.googlePlaceId).filter(Boolean))
  } catch {
    savedPlaceIds.value = new Set()
  }
}

async function selectResult(item: PlaceResult, focusMap = true) {
  selected.value = item
  placeDetails.value = null
  placeDetailsError.value = ''
  menuImages.value = []
  menuImagesError.value = ''
  detailOpen.value = true
  if (item.googlePlaceId) {
    placeDetailsLoading.value = true
    try {
      placeDetails.value = await crmApi.getPlaceDetails(item.googlePlaceId)
    } catch (caught) {
      placeDetailsError.value = crmError(caught)
    } finally {
      placeDetailsLoading.value = false
    }
  }
  if (focusMap && map && item.latitude !== null && item.longitude !== null) {
    map.flyTo([item.latitude, item.longitude], Math.max(map.getZoom(), 16), { duration: 0.55 })
    const cid = item.customerId
    if (cid) customerMarkerMap.get(cid)?.openTooltip()
    else markers.get(item.googlePlaceId)?.openTooltip()
  }
  nextTick(() => {
    if (!item.googlePlaceId) return
    const el = resultsScroll.value?.querySelector(`[data-place-id="${item.googlePlaceId}"]`)
    el?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  })
}

function selectCustomer(item: CustomerMarker) {
  selectResult(customerToResult(item), true)
}

watch(selected, (current, previous) => {
  if (previous) {
    if (previous.customerId) customerMarkerMap.get(previous.customerId)?.setIcon(customerMarkerIcon(false))
    else markers.get(previous.googlePlaceId)?.setIcon(markerIcon(previous, false))
  }
  if (current) {
    if (current.customerId) customerMarkerMap.get(current.customerId)?.setIcon(customerMarkerIcon(true))
    else markers.get(current.googlePlaceId)?.setIcon(markerIcon(current, true))
  }
})

watch(radius, drawSearchArea)

watch([latitude, longitude], () => {
  if (draggingCenter || !map) return
  drawSearchArea()
})

async function search() {
  error.value = ''
  success.value = ''
  loading.value = true
  try {
    results.value = await crmApi.searchPlaces({ keyword: keyword.value, categories: categories.value.join(','), radius: radius.value, latitude: latitude.value, longitude: longitude.value })
    queried.value = true
    selected.value = null
    placeDetails.value = null
    detailOpen.value = false
    await nextTick()
    renderMarkers()
  } catch (caught) {
    error.value = crmError(caught)
    toast.add({ severity: 'error', summary: 'Search failed', detail: error.value, life: 6000 })
  } finally {
    loading.value = false
  }
}

function useGPS() {
  error.value = ''
  if (!navigator.geolocation) { error.value = 'Geolocation is not available in this browser.'; geoResolved.value = true; return }
  navigator.geolocation.getCurrentPosition((position) => {
    latitude.value = position.coords.latitude
    longitude.value = position.coords.longitude
    geoResolved.value = true
    drawSearchArea()
    map?.flyTo([latitude.value, longitude.value], 15)
  }, () => {
    error.value = 'Location permission was denied. Please enable location access and try again.'
    geoResolved.value = true
  })
}

async function save() {
  if (!selected.value || !selectedPlaceCategory.value || !salesExecutiveId.value) {
    toast.add({ severity: 'warn', summary: 'Missing information', detail: 'The selected place must have a Category and Sales Executive before saving.', life: 4000 })
    return
  }
  if (selected.value.isCustomer) {
    toast.add({ severity: 'warn', summary: 'Existing customer', detail: 'This place is already an existing customer and cannot be assigned to sales.', life: 4000 })
    return
  }
  error.value = ''
  success.value = ''
  saving.value = true
  try {
    const place = { ...selected.value, category: selectedPlaceCategory.value }
    const item = await crmApi.saveProspect(place, selectedPlaceCategory.value, salesExecutiveId.value)
    success.value = `${item.placeName} saved as NEW_LEAD and assigned successfully.`
    toast.add({ severity: 'success', summary: 'Prospect saved', detail: `${item.placeName} was saved as NEW_LEAD and assigned successfully.`, life: 5000 })
    detailOpen.value = false
    loadSavedPlaceIds()
  } catch (caught) {
    error.value = crmError(caught)
    toast.add({ severity: 'error', summary: 'Save failed', detail: error.value, life: 6000 })
  } finally {
    saving.value = false
  }
}

function crmError(err: unknown) {
  const candidate = err as { response?: { data?: { error?: { message?: string } } }; message?: string }
  return candidate.response?.data?.error?.message ?? candidate.message ?? 'Prospect Finder request failed.'
}

function parkingLabel(key: string) {
  const labels: Record<string, string> = { freeStreetParking:'Free Street', paidStreetParking:'Paid Street', freeParkingLot:'Free Lot', paidParkingLot:'Paid Lot', valetParking:'Valet', garageParking:'Garage' }
  return labels[key] || key
}

function paymentLabel(key: string) {
  const labels: Record<string, string> = { cashOnly:'Cash Only', creditCardOnly:'Credit Card', debitCardOnly:'Debit Card', nfcOnly:'NFC' }
  return labels[key] || key
}

function accessibilityLabel(key: string) {
  const labels: Record<string, string> = { wheelchairAccessibleEntrance:'Entrance', wheelchairAccessibleParking:'Parking', wheelchairAccessibleRestroom:'Restroom', wheelchairAccessibleSeating:'Seating' }
  return labels[key] || key
}

function optionActive(details: PlaceDetails | null, key: string): boolean {
  if (!details) return false
  const parking = details.parkingOptions as Record<string, boolean> | null
  const payment = details.paymentOptions as Record<string, boolean> | null
  const access = details.accessibilityOptions as Record<string, boolean> | null
  return !!(parking?.[key] || payment?.[key] || access?.[key])
}

onMounted(async () => {
  await nextTick()
  initializeMap()
  useGPS()
  loadCustomerMarkers()
  loadSavedPlaceIds()
  try {
    sales.value = await crmApi.getSalesExecutives()
    salesExecutiveId.value = sales.value[0]?.id ?? ''
  } catch (caught) {
    error.value = crmError(caught)
  }
})

onBeforeUnmount(() => {
  revokePlacePhotoObjectUrls()
  map?.remove()
  map = null
  centerMarker = null
  customPinMarker = null
  customerLayer = null
  markers.clear()
  customerMarkerMap.clear()
})
</script>

<template>
  <section class="finder-page">
    <header class="finder-page-header">
      <div class="finder-heading-left">
        <Button
          icon="pi pi-arrow-left"
          severity="secondary"
          text
          rounded
          class="finder-back"
          @click="$router.back()"
          title="Back"
        />
        <div class="finder-page-title">
          <span class="finder-eyebrow">Prospect Management</span>
          <h1>Prospect Finder</h1>
          <p>Discover nearby businesses, review place details, and assign qualified prospects.</p>
        </div>
      </div>

      <div class="finder-page-stats">
        <div class="finder-stat">
          <span>Location</span>
          <strong>{{ geoResolved ? 'Ready' : 'Detecting' }}</strong>
        </div>
        <div class="finder-stat">
          <span>Radius</span>
          <strong>{{ (radius / 1000).toFixed(1) }} km</strong>
        </div>
        <div class="finder-stat">
          <span>Categories</span>
          <strong>{{ categories.length }}</strong>
        </div>
        <div class="finder-stat">
          <span>Results</span>
          <strong>{{ results.length }}</strong>
        </div>
      </div>
    </header>

    <div class="finder-desktop-shell">
      <aside class="finder-left-panel">
        <div class="finder-panel-header">
          <div class="finder-panel-title">
            <i class="pi pi-compass" />
            <div>
              <h1>Prospect Query</h1>
              <span>Discover &amp; save qualified prospects</span>
            </div>
          </div>
          <Button label="Create Pin" icon="pi pi-map-marker" severity="secondary" outlined class="create-pin-button" @click="openPinForm" />
        </div>

        <div class="finder-filter-scroll">
          <div class="filter-section">
            <label class="field finder-keyword-field">
              <span>Keyword</span>
              <div class="keyword-input-wrap">
                <i class="pi pi-search keyword-icon" />
                <InputText v-model="keyword" placeholder="Search cafe, office, laundry..." @keyup.enter="search" />
              </div>
            </label>
          </div>

          <div class="filter-section">
            <div class="filter-section-header">
              <span class="filter-section-title">Categories</span>
              <span class="category-count">{{ categories.length }} selected</span>
            </div>
            <div class="category-tools">
              <button type="button" class="category-tool" @click="selectAllCategories"><i class="pi pi-check" /> Select All</button>
              <button type="button" class="category-tool clear" @click="categories = []"><i class="pi pi-times" /> Clear All</button>
            </div>
            <div class="category-grid">
              <label v-for="option in categoryOptions" :key="option.key" class="category-chip" :class="{ active: categories.includes(option.key) }">
                <Checkbox v-model="categories" :input-id="option.key" :value="option.key" />
                <span class="category-chip-icon">{{ option.icon }}</span>
                <span class="category-chip-label">{{ option.label }}</span>
              </label>
            </div>
          </div>

          <div class="filter-section">
            <div class="radius-header">
              <span class="filter-section-title">Search Radius</span>
              <span class="radius-value">{{ (radius / 1000).toFixed(0) }} km</span>
            </div>
            <Slider v-model="radius" :min="1000" :max="25000" :step="500" class="finder-slider" />
            <div class="radius-range-labels">
              <span>1 km</span>
              <span>25 km</span>
            </div>
          </div>

          <div class="filter-section">
            <p class="filter-section-title">Search Location Anchor</p>
            <Button label="Use Current GPS" icon="pi pi-crosshairs" severity="secondary" outlined fluid class="gps-button" @click="useGPS" />
            <div class="coordinate-grid">
              <label class="field"><span>LATITUDE</span><input v-model.number="latitude" type="number" step="0.000001" /></label>
              <label class="field"><span>LONGITUDE</span><input v-model.number="longitude" type="number" step="0.000001" /></label>
            </div>
          </div>

          <div class="filter-section">
            <p class="filter-section-title">Minimum Rating Threshold</p>
            <div class="segment-row">
              <button v-for="opt in ratingOptions" :key="opt.value" type="button" class="segment-chip" :class="{ active: minRating === opt.value }" @click="minRating = opt.value">{{ opt.label }}</button>
            </div>
          </div>

          <div class="filter-section">
            <p class="filter-section-title">Leads saved state</p>
            <div class="segment-row">
              <button v-for="opt in savedFilterOptions" :key="opt.value" type="button" class="segment-chip" :class="{ active: savedFilter === opt.value }" @click="savedFilter = opt.value">{{ opt.label }}</button>
            </div>
          </div>

          <div class="filter-section">
            <p class="filter-section-title">Menu availability</p>
            <div class="segment-row">
              <button v-for="opt in menuFilterOptions" :key="opt.value" type="button" class="segment-chip" :class="{ active: menuFilter === opt.value }" @click="menuFilter = opt.value">{{ opt.label }}</button>
            </div>
          </div>

          <div class="filter-actions">
            <Button :label="!geoResolved ? 'Detecting location...' : 'PROSES CARI PROSPEK'" icon="pi pi-search" fluid :loading="loading || !geoResolved" :disabled="!categories.length || !geoResolved" @click="search" />
          </div>

          <div v-if="queried" class="query-results-footer">
            <div class="query-results-row">
              <span class="query-results-label">Query Results:</span>
              <strong>{{ filteredResults.length }} items</strong>
            </div>
            <span class="query-source"><i class="pi pi-cloud" /> GOOGLE API</span>
            <span class="query-anchor">Map Anchor: {{ latitude.toFixed(6) }}, {{ longitude.toFixed(6) }}</span>
          </div>
        </div>

      </aside>

      <div class="finder-map-stage" :class="{ 'has-results': results.length }">
        <div ref="mapElement" class="leaflet-map" role="region" aria-label="OpenStreetMap with Google Places prospect markers" />

        

        <div v-if="loading" class="map-loading-state">
          <span class="map-loading-spinner" />
          <strong>Finding nearby prospects...</strong>
        </div>

        <div v-if="results.length" class="finder-floating-results">
          <div class="floating-results-header">
            <div class="floating-results-title">
              <strong>{{ filteredResults.length }} result{{ filteredResults.length !== 1 ? 's' : '' }}</strong>
              <span v-if="results.length && resultSearch" class="results-filter-note">of {{ results.length }}</span>
            </div>
            <button class="floating-results-close" @click="closeResults" title="Close results"><i class="pi pi-times" /></button>
          </div>
          <div class="floating-results-search">
            <i class="pi pi-search" />
            <input v-model="resultSearch" placeholder="Filter results..." />
          </div>
          <div ref="resultsScroll" class="floating-results-list">
            <button
              v-for="item in filteredResults"
              :key="item.googlePlaceId"
              :data-place-id="item.googlePlaceId"
              class="result-card"
              :class="{ selected: selected?.googlePlaceId === item.googlePlaceId }"
              @click="selectResult(item, true)"
            >
              <span class="result-marker" :class="{ 'is-customer': item.isCustomer }" :style="item.isCustomer ? undefined : { background: item.markerColor }">
                <b v-if="item.isCustomer">Y</b>
                <i v-else :class="item.markerIcon" />
              </span>
              <div class="result-info">
                <div class="result-name-row">
                  <strong>{{ item.name }}</strong>
                  <Tag v-if="item.rating" :value="`★ ${item.rating}`" severity="info" class="result-rating-tag" />
                  <Tag v-if="item.isCustomer" value="Existing Customer" severity="success" class="result-customer-tag" />
                </div>
                <span class="result-category">{{ item.category }}</span>
                <span class="result-address">{{ item.address }}</span>
                <div class="result-meta-row">
                  <span v-if="item.distance" class="result-distance"><i class="pi pi-map-marker" /> {{ Math.round(item.distance) }} m</span>
                  <Tag v-if="item.businessStatus" :value="item.businessStatus === 'OPERATIONAL' ? 'Open' : item.businessStatus" :severity="item.businessStatus === 'OPERATIONAL' ? 'success' : 'warn'" class="result-status-tag" />
                  <Tag v-if="item.hasMenuPhotos" value="Menu Ready ✓" severity="success" class="result-menu-tag" />
                  <Tag v-else-if="likelyHasMenu(item)" value="Likely Has Menu" severity="warn" class="result-menu-tag" />
                  <Tag v-else value="Menu Not Ready" severity="secondary" class="result-menu-tag" />
                </div>
              </div>
              <i class="pi pi-chevron-right result-chevron" />
            </button>
          </div>
          <div v-if="!filteredResults.length" class="floating-results-empty">
            <i class="pi pi-filter" />
            <span>No matching results</span>
          </div>
        </div>

        <div class="map-source-badge">
          <i class="pi pi-shield" />
          <div>
            <strong>Private Places search</strong>
            <span>Google Places stays server-side. Map tiles use OpenStreetMap.</span>
          </div>
        </div>

        <div v-if="customerMarkers.length || customersLoading" class="map-customer-badge" :class="{ 'is-loading': customersLoading }">
          <span class="map-customer-dot"><b>Y</b></span>
          <div>
            <strong>{{ customersLoading ? 'Loading existing customers...' : `${customerMarkers.length} existing customer${customerMarkers.length !== 1 ? 's' : ''}` }}</strong>
            <span>Existing customers are always visible and cannot be assigned to sales.</span>
          </div>
        </div>
      </div>
    </div>

    <Dialog v-model:visible="detailOpen" modal header="Place Details" :style="{ width: '520px' }" :closable="true" :breakpoints="{ '576px': '95vw' }">
      <div v-if="placeDetailsLoading" class="dialog-loading"><div class="loading-pulse" /><span>Loading full details...</span></div>
      <div v-else-if="selected" class="detail-dialog">
        <div class="detail-hero-bar">
          <span class="detail-hero" :class="{ 'is-customer': selected.isCustomer }" :style="selected.isCustomer ? undefined : { background: selected.markerColor }">
            <b v-if="selected.isCustomer">Y</b>
            <i v-else :class="selected.markerIcon" />
          </span>
          <div class="detail-hero-info">
            <h2>{{ placeDetails?.placeName || selected.name }}</h2>
            <div class="detail-hero-meta">
              <span>{{ placeDetails?.placeCategory || selected.category }}</span>
              <Tag v-if="selected.isCustomer" value="Existing Customer" severity="success" />
              <Tag v-else-if="(placeDetails?.rating || selected.rating)!" :value="`★ ${(placeDetails?.rating || selected.rating)!.toFixed(1)}`" severity="info" />
              <Tag v-if="placeDetails?.userRatingCount || selected.userRatingCount" :value="`${placeDetails?.userRatingCount || selected.userRatingCount} reviews`" severity="secondary" />
              <Tag v-if="placeDetails?.priceLevel" :value="placeDetails.priceLevel" severity="contrast" />
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.editorialSummary" class="detail-editorial">
          <i class="pi pi-quote-left" /> {{ placeDetails.editorialSummary }}
        </div>

        <div class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-info-circle" /> Basic Info</h3>
          <div class="detail-info-grid">
            <div class="detail-info-item">
              <i class="pi pi-map-marker" />
              <div>
                <span class="detail-info-label">Address</span>
                <span class="detail-info-value">{{ placeDetails?.formattedAddress || selected.address }}</span>
              </div>
            </div>
            <div v-if="placeDetails?.phoneNumber || selected.phone" class="detail-info-item">
              <i class="pi pi-phone" />
              <div>
                <span class="detail-info-label">Phone</span>
                <span class="detail-info-value">{{ placeDetails?.phoneNumber || selected.phone }}</span>
              </div>
            </div>
            <div v-if="placeDetails?.internationalPhone" class="detail-info-item">
              <i class="pi pi-phone" />
              <div>
                <span class="detail-info-label">International</span>
                <span class="detail-info-value">{{ placeDetails.internationalPhone }}</span>
              </div>
            </div>
            <div v-if="placeDetails?.websiteUrl || selected.website" class="detail-info-item">
              <i class="pi pi-globe" />
              <div>
                <span class="detail-info-label">Website</span>
                <a :href="placeDetails?.websiteUrl || selected.website" target="_blank" rel="noreferrer" class="detail-info-link">Open website →</a>
              </div>
            </div>
            <div v-if="placeDetails?.googleMapsUrl || selected.googleMapsUrl" class="detail-info-item">
              <i class="pi pi-external-link" />
              <div>
                <span class="detail-info-label">Google Maps</span>
                <a :href="placeDetails?.googleMapsUrl || selected.googleMapsUrl" target="_blank" rel="noreferrer" class="detail-info-link">View listing →</a>
              </div>
            </div>
            <div class="detail-info-item">
              <i class="pi pi-tag" />
              <div>
                <span class="detail-info-label">Place Types</span>
                <span class="detail-info-value detail-types">{{ placeDetails?.placeTypes?.join(', ') || selected.placeTypes?.join(', ') || selected.markerCategory }}</span>
              </div>
            </div>
            <div class="detail-info-item">
              <i class="pi pi-info-circle" />
              <div>
                <span class="detail-info-label">Status</span>
                <Tag :value="(placeDetails?.businessStatus || selected.businessStatus || 'UNKNOWN')" :severity="(placeDetails?.businessStatus || selected.businessStatus) === 'OPERATIONAL' ? 'success' : 'warn'" />
              </div>
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.openingHours" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-clock" /> Opening Hours</h3>
          <div class="detail-hours-grid">
            <div class="detail-hours-badge" :class="{ 'is-open': placeDetails.openingHours.openNow }">
              <span class="hours-dot" />
              {{ placeDetails.openingHours.openNow ? 'Open now' : 'Closed' }}
            </div>
            <div v-for="day in placeDetails.openingHours.weekdays" :key="day" class="detail-hours-day">{{ day }}</div>
          </div>
        </div>

        <div v-if="placeDetails" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-cog" /> Services &amp; Options</h3>
          <div class="detail-options-grid">
            <div v-if="placeDetails.delivery !== undefined" class="detail-option-chip" :class="{ active: placeDetails.delivery }">
              <i class="pi pi-truck" /> <span>Delivery</span>
            </div>
            <div v-if="placeDetails.dineIn !== undefined" class="detail-option-chip" :class="{ active: placeDetails.dineIn }">
              <i class="pi pi-building" /> <span>Dine In</span>
            </div>
            <div v-if="placeDetails.takeout !== undefined" class="detail-option-chip" :class="{ active: placeDetails.takeout }">
              <i class="pi pi-box" /> <span>Takeout</span>
            </div>
            <div v-if="placeDetails.curbsidePickup !== undefined" class="detail-option-chip" :class="{ active: placeDetails.curbsidePickup }">
              <i class="pi pi-car" /> <span>Curbside Pickup</span>
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.parkingOptions" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-car" /> Parking</h3>
          <div class="detail-options-grid">
            <div v-for="key in ['freeStreetParking','paidStreetParking','freeParkingLot','paidParkingLot','valetParking','garageParking']" :key="key" class="detail-option-chip" :class="{ active: optionActive(placeDetails, key) }">
              <i class="pi pi-check-circle" /> <span>{{ parkingLabel(key) }}</span>
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.paymentOptions" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-credit-card" /> Payment</h3>
          <div class="detail-options-grid">
            <div v-for="key in ['cashOnly','creditCardOnly','debitCardOnly','nfcOnly']" :key="key" class="detail-option-chip" :class="{ active: optionActive(placeDetails, key) }">
              <i class="pi pi-check-circle" /> <span>{{ paymentLabel(key) }}</span>
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.accessibilityOptions" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-universal-access" /> Accessibility</h3>
          <div class="detail-options-grid">
            <div v-for="key in ['wheelchairAccessibleEntrance','wheelchairAccessibleParking','wheelchairAccessibleRestroom','wheelchairAccessibleSeating']" :key="key" class="detail-option-chip" :class="{ active: optionActive(placeDetails, key) }">
              <i class="pi pi-check-circle" /> <span>Wheelchair {{ accessibilityLabel(key) }}</span>
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.reviews?.length" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-comments" /> Reviews ({{ placeDetails.reviews.length }})</h3>
          <div class="detail-reviews-list">
            <div v-for="review in placeDetails.reviews.slice(0, 5)" :key="review.authorName + review.time" class="detail-review">
              <div class="review-author">
                <img v-if="review.authorPhoto" :src="review.authorPhoto" alt="" class="review-author-photo" />
                <span class="review-author-name">{{ review.authorName }}</span>
                <Tag :value="`★ ${review.rating.toFixed(1)}`" severity="info" class="review-rating" />
                <small class="review-time">{{ review.time }}</small>
              </div>
              <p v-if="review.text" class="review-text">{{ review.text }}</p>
            </div>
          </div>
        </div>

        <div v-if="placeDetails" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-book" /> Menu Photos</h3>
          <div v-if="menuPhotos.length" class="detail-photos-row">
            <template v-for="photo in menuPhotos" :key="photo.name">
              <button v-if="canShowPlacePhoto(photo)" type="button" class="detail-photo-tile detail-photo-button" @click="openPhotoPreview(resolvedPlacePhotoUrl(photo), 'Menu photo', photo.attribution)">
                <img :src="resolvedPlacePhotoUrl(photo)" alt="Menu photo" class="detail-photo" loading="lazy" />
                <span class="detail-photo-badge"><i class="pi pi-book" /> Menu</span>
              </button>
            </template>
          </div>
          <div v-else-if="menuImagesLoading" class="detail-photo-empty"><i class="pi pi-spin pi-spinner" /> Mencari foto menu harga...</div>
          <div v-else-if="menuImages.length" class="menu-images">
            <p class="menu-images-note"><i class="pi pi-image" /> Foto menu dari hasil pencarian Google (maksimal 3)</p>
            <div class="detail-photos-row">
              <button v-for="image in menuImages" :key="image.imageUrl" type="button" class="detail-photo-tile detail-photo-button" @click="openPhotoPreview(image.imageUrl, image.title || 'Menu price photo', image.sourceSite)">
                <img :src="image.imageUrl" :alt="image.title || 'Menu price photo'" class="detail-photo" loading="lazy" @error="($event.target as HTMLImageElement).style.display='none'" />
                <span class="detail-photo-badge"><i class="pi pi-book" /> Menu</span>
              </button>
            </div>
          </div>
          <Message v-else-if="menuImagesError" severity="warn" :closable="false">{{ menuImagesError }}</Message>
          <Button v-else-if="likelyHasMenu({ placeTypes: placeDetails.placeTypes } as PlaceResult)" label="Load menu photos" icon="pi pi-download" severity="secondary" outlined size="small" @click="loadMenuImages" />
          <div v-else class="detail-photo-empty"><i class="pi pi-book" /> Menu photos are not loaded automatically</div>
        </div>

        <div v-if="placeDetails" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-images" /> Photos</h3>
          <div v-if="visiblePhotos.length" class="detail-photos-row">
            <template v-for="(photo, index) in visiblePhotos" :key="photo.name">
              <button v-if="canShowPlacePhoto(photo)" type="button" class="detail-photo-tile detail-photo-button" @click="openPhotoPreview(resolvedPlacePhotoUrl(photo), 'Place photo', photo.attribution)">
                <img :src="resolvedPlacePhotoUrl(photo)" alt="Place photo" class="detail-photo" loading="lazy" />
                <span class="detail-photo-badge"><i class="pi pi-images" /> {{ index === 0 ? 'Hero' : 'Photo' }}</span>
              </button>
            </template>
          </div>
          <div v-else class="detail-photo-empty"><i class="pi pi-images" /> No photos available</div>
          <Button v-if="hasMorePhotos" label="Load one more photo" icon="pi pi-plus" severity="secondary" outlined size="small" @click="loadNextPlacePhoto" />
        </div>

        <div v-if="placeDetailsError" class="detail-section">
          <Message severity="warn" :closable="false">{{ placeDetailsError }}</Message>
        </div>

        <div class="detail-assignment">
          <h3>Assignment</h3>
          <Message v-if="selected?.isCustomer" severity="success" :closable="false" class="assignment-warning">
            <strong>Existing customer</strong> — this place has already been converted to a customer and can no longer be assigned to sales.
          </Message>
          <div v-else class="detail-assignment-fields">
            <label class="field"><span>Category</span><InputText :model-value="selectedPlaceCategory" disabled fluid /></label>
            <label class="field"><span>Assign Sales Executive</span><Select v-model="salesExecutiveId" :options="sales" option-label="fullName" option-value="id" placeholder="Select Sales Executive" fluid /></label>
            <Message v-if="selectedSalesCount > 0" severity="warn" :closable="false" class="assignment-warning">
              {{ sales.find(s => s.id === salesExecutiveId)?.fullName }} already has <strong>{{ selectedSalesCount }}</strong> active prospect{{ selectedSalesCount !== 1 ? 's' : '' }} assigned.
            </Message>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="detail-dialog-footer">
          <Button label="Cancel" severity="secondary" text @click="detailOpen = false" />
          <Button v-if="selected?.isCustomer" label="Existing Customer" icon="pi pi-check" severity="success" disabled />
          <Button v-else label="Save as Prospect" icon="pi pi-save" :loading="saving" :disabled="!salesExecutiveId || !selectedPlaceCategory" @click="save" />
        </div>
      </template>
    </Dialog>

    <Dialog v-model:visible="previewOpen" modal header="Photo Preview" :style="{ width: 'min(960px, 95vw)' }" :breakpoints="{ '576px': '96vw' }" @hide="previewPhoto = null">
      <div v-if="previewPhoto" class="finder-photo-preview">
        <img :src="previewPhoto.url" :alt="previewPhoto.alt" />
        <small v-if="previewPhoto.attribution">Photo: {{ previewPhoto.attribution }}</small>
      </div>
    </Dialog>

    <Dialog v-model:visible="pinOpen" :header="'Create Pin'" :style="{ width: '430px' }" :modal="true" :dismissable-mask="false" class="pin-dialog" @hide="closePinForm">
      <div class="pin-form">
        <Message v-if="pinError" severity="error" :closable="false" class="pin-error">{{ pinError }}</Message>

        <div class="pin-map-note">
          <i class="pi pi-map-marker" />
          <span>Drag the purple pin on the map to position it, or move it below by coordinates.</span>
        </div>

        <label class="field">
          <span>Pin Name</span>
          <InputText v-model="pinName" placeholder="e.g. Kopi Nusantara HQ" fluid />
        </label>

        <div class="pin-coords">
          <label class="field">
            <span>Latitude</span>
            <InputNumber v-model="pinLat" :min="-90" :max="90" :min-fraction-digits="6" :max-fraction-digits="6" fluid />
          </label>
          <label class="field">
            <span>Longitude</span>
            <InputNumber v-model="pinLng" :min="-180" :max="180" :min-fraction-digits="6" :max-fraction-digits="6" fluid />
          </label>
        </div>

        <div class="pin-assignment">
          <label class="field"><span>Category</span><InputText :model-value="pinCategory" disabled fluid /></label>
          <label class="field"><span>Assign Sales Executive</span><Select v-model="salesExecutiveId" :options="sales" option-label="fullName" option-value="id" placeholder="Select Sales Executive" fluid /></label>
          <Message v-if="selectedSalesCount > 0" severity="warn" :closable="false" class="assignment-warning">
            {{ sales.find(s => s.id === salesExecutiveId)?.fullName }} already has <strong>{{ selectedSalesCount }}</strong> active prospect{{ selectedSalesCount !== 1 ? 's' : '' }} assigned.
          </Message>
        </div>
      </div>

      <template #footer>
        <div class="pin-dialog-footer">
          <Button label="Cancel" severity="secondary" text @click="closePinForm" />
          <Button label="Save Pin" icon="pi pi-save" :loading="pinSaving" :disabled="!pinName || !salesExecutiveId" @click="savePin" />
        </div>
      </template>
    </Dialog>

    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
  </section>
</template>

<style scoped>
/* ════════════════════════════════════════════════════════════════
   PROSPECT FINDER — Workspace Layout (modernized visual pass)
   ════════════════════════════════════════════════════════════════ */

.finder-page {
  box-sizing: border-box;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  gap: 0;
  width: calc(100% + 3rem);
  min-width: 0;
  height: calc(100dvh - 4rem);
  min-height: 0;
  margin: -1.5rem;
  padding: 0;
  overflow: hidden;
  background: #f8fafc;
}

.finder-page-header {
  z-index: 5;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.8rem;
  min-height: 64px;
  padding: 0.52rem 0.85rem;
  border: 0;
  border-bottom: 1px solid #e5eaf0;
  border-radius: 0;
  background: #ffffff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.finder-heading-left {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.65rem;
}

.finder-back {
  flex: 0 0 auto;
}

.finder-page-title {
  min-width: 0;
}

.finder-eyebrow {
  color: #64748b;
  font-size: 0.58rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.finder-page-title h1 {
  margin: 0.08rem 0 0;
  color: #0f172a;
  font-size: 1.08rem;
  line-height: 1.2;
}

.finder-page-title p {
  margin: 0.12rem 0 0;
  overflow: hidden;
  color: #94a3b8;
  font-size: 0.68rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.finder-page-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(76px, 1fr));
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 9px;
  background: #f8fafc;
}

.finder-stat {
  display: grid;
  gap: 0.04rem;
  min-width: 76px;
  padding: 0.45rem 0.65rem;
  border-right: 1px solid #e5eaf0;
  text-align: center;
}

.finder-stat:last-child { border-right: 0; }
.finder-stat span { color: #94a3b8; font-size: 0.55rem; font-weight: 800; text-transform: uppercase; }
.finder-stat strong { color: #0f172a; font-size: 0.78rem; }

/* ── Desktop Shell ───────────────────────────────────────────── */
.finder-desktop-shell {
  min-height: 0;
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  overflow: hidden;
  background: var(--surface-card);
  border: 0;
  border-radius: 0;
  box-shadow: none;
}

/* ── Left Panel ──────────────────────────────────────────────── */
.finder-left-panel {
  min-height: 0;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  background: var(--surface-card);
  border-right: 1px solid var(--border-light);
}

.finder-panel-header {
  padding: 0.65rem 0.85rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.6rem;
  border-bottom: 1px solid var(--border-light);
  background: linear-gradient(150deg, var(--brand-blue-50) 0%, #ffffff 70%);
}

.create-pin-button {
  flex: 0 0 auto;
}

.create-pin-button :deep(.p-button) {
  padding: 0.45rem 0.7rem;
  font-size: 0.62rem;
  font-weight: 700;
  border-radius: 0.6rem;
}

.finder-panel-title {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  min-width: 0;
}

.finder-panel-title > i {
  width: 2.15rem;
  height: 2.15rem;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(135deg, var(--brand-blue) 0%, #d62839 100%);
  border-radius: 0.7rem;
  font-size: 0.85rem;
  flex-shrink: 0;
  box-shadow: 0 4px 12px -2px rgba(230, 57, 70, 0.45);
}

.finder-panel-title h1 {
  margin: 0;
  font-size: 1.02rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text-primary);
  line-height: 1.2;
}

.finder-panel-title span {
  display: block;
  margin-top: 0.15rem;
  color: var(--text-muted);
  font-size: 0.63rem;
  font-weight: 500;
}

.finder-filter-scroll {
  padding: 0.35rem 0.75rem 0.55rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.finder-filter-scroll::-webkit-scrollbar { width: 4px; }
.finder-filter-scroll::-webkit-scrollbar-track { background: transparent; }
.finder-filter-scroll::-webkit-scrollbar-thumb { background: var(--border-default); border-radius: 999px; }
.finder-filter-scroll::-webkit-scrollbar-thumb:hover { background: var(--text-faint); }

/* ── Filter Sections ─────────────────────────────────────────── */
.filter-section {
  padding: 0.45rem 0;
  border-bottom: 1px solid #eef1f5;
}

.filter-section:last-of-type { border-bottom: 0; }

.filter-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.filter-section-title {
  margin: 0 0 0.5rem;
  color: var(--text-primary);
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.filter-section-header .filter-section-title { margin-bottom: 0; }

.category-count {
  padding: 0.12rem 0.5rem;
  color: var(--brand-blue);
  background: var(--brand-blue-50);
  border: 1px solid var(--brand-blue-100);
  border-radius: 999px;
  font-size: 0.55rem;
  font-weight: 700;
}

/* Keyword */
.finder-keyword-field { gap: 0.32rem; }
.finder-keyword-field > span { color: var(--text-muted); font-size: 0.65rem; font-weight: 700; }

.keyword-input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.keyword-icon {
  position: absolute;
  left: 0.7rem;
  color: var(--text-faint);
  font-size: 0.7rem;
  pointer-events: none;
}

.keyword-input-wrap :deep(.p-inputtext) {
  padding-left: 1.9rem;
  border-radius: 0.6rem;
  border-color: var(--border-default);
  font-size: 0.78rem;
  padding-top: 0.52rem;
  padding-bottom: 0.52rem;
  transition: border-color 160ms ease, box-shadow 160ms ease;
}

.keyword-input-wrap :deep(.p-inputtext:focus) {
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 4px rgba(230, 57, 70, 0.10);
}

/* Categories */
.category-tools {
  display: flex;
  gap: 0.4rem;
  margin: 0 0 0.45rem;
}

.category-tool {
  display: inline-flex;
  align-items: center;
  gap: 0.28rem;
  padding: 0.24rem 0.55rem;
  border: 1px solid var(--border-default);
  border-radius: 999px;
  background: var(--surface-subtle);
  color: var(--brand-blue);
  font-size: 0.56rem;
  font-weight: 700;
  cursor: pointer;
  transition: background 160ms ease, border-color 160ms ease, color 160ms ease;
}

.category-tool:hover { background: var(--brand-blue-50); border-color: var(--brand-blue); }

.category-tool.clear { color: var(--text-muted); }
.category-tool.clear:hover { color: var(--brand-blue); }

.category-tool i { font-size: 0.5rem; }

.category-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.3rem;
}

.category-chip {
  display: flex;
  gap: 0.3rem;
  align-items: center;
  min-width: 0;
  padding: 0.32rem 0.4rem;
  border-radius: 0.55rem;
  background: var(--surface-subtle);
  border: 1px solid transparent;
  cursor: pointer;
  transition: background 160ms ease, border-color 160ms ease, transform 160ms ease;
}

.category-chip:hover { background: var(--surface-hover); border-color: var(--border-default); transform: translateY(-1px); }

.category-chip.active {
  background: linear-gradient(135deg, var(--brand-blue-50) 0%, #fff5f5 100%);
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 1px rgba(230, 57, 70, 0.14), 0 2px 6px -2px rgba(230, 57, 70, 0.25);
}

.category-chip-icon {
  font-size: 0.72rem;
  flex-shrink: 0;
  filter: grayscale(0.15);
}

.category-chip-label {
  min-width: 0;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  color: var(--text-secondary);
  font-size: 0.56rem;
  font-weight: 550;
  line-height: 1.3;
  transition: color 160ms ease;
}

.category-chip.active .category-chip-label { color: var(--brand-blue); font-weight: 700; }

/* Radius */
.radius-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.45rem;
}

.radius-value {
  padding: 0.16rem 0.55rem;
  color: var(--brand-blue);
  background: linear-gradient(135deg, var(--brand-blue-50), #fff5f5);
  border: 1px solid var(--brand-blue-100);
  border-radius: 999px;
  font-size: 0.65rem;
  font-weight: 800;
}

.radius-range-labels {
  display: flex;
  justify-content: space-between;
  margin-top: 0.25rem;
  color: var(--text-faint);
  font-size: 0.52rem;
}

/* Coordinates */
.coordinate-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
  margin-top: 0.45rem;
}

.coordinate-grid .field { gap: 0.22rem; }
.coordinate-grid .field > span {
  color: var(--text-muted);
  font-size: 0.52rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.coordinate-grid input {
  width: 100%;
  padding: 0.42rem 0.6rem;
  border: 1px solid var(--border-default);
  border-radius: 0.55rem;
  font-size: 0.7rem;
  color: var(--text-primary);
  background: var(--surface-card);
  transition: border-color 160ms ease, box-shadow 160ms ease;
}

.coordinate-grid input:focus {
  outline: none;
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 4px rgba(230, 57, 70, 0.10);
}

.gps-button {
  width: 100%;
}

.gps-button :deep(.p-button) {
  padding: 0.5rem;
  font-size: 0.68rem;
  font-weight: 700;
  border-radius: 0.6rem;
}

/* Segmented controls */
.segment-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
}

.segment-chip {
  padding: 0.3rem 0.55rem;
  border: 1px solid var(--border-default);
  border-radius: 999px;
  background: var(--surface-subtle);
  color: var(--text-secondary);
  font-size: 0.6rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 160ms ease, border-color 160ms ease, color 160ms ease, box-shadow 160ms ease;
}

.segment-chip:hover { background: var(--surface-hover); border-color: var(--border-default); }

.segment-chip.active {
  background: linear-gradient(135deg, var(--brand-blue) 0%, #b13a48 100%);
  border-color: var(--brand-blue);
  color: #fff;
  box-shadow: 0 2px 8px -2px rgba(230, 57, 70, 0.45);
}

/* Filter Actions */
.filter-actions {
  margin-top: 0.55rem;
}

.filter-actions :deep(.p-button) {
  padding: 0.62rem;
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.03em;
  border-radius: 0.6rem;
  transition: transform 160ms ease, box-shadow 160ms ease, filter 160ms ease;
}

.filter-actions :deep(.p-button:not(:disabled):hover) {
  transform: translateY(-1px);
  filter: brightness(1.03);
}

/* Query Results Footer */
.query-results-footer {
  margin-top: 0.6rem;
  display: grid;
  gap: 0.3rem;
  padding: 0.6rem 0.7rem;
  border: 1px solid var(--border-light);
  border-radius: 0.7rem;
  background: linear-gradient(135deg, var(--surface-subtle) 0%, #ffffff 100%);
}

.query-results-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.4rem;
}

.query-results-label {
  color: var(--text-muted);
  font-size: 0.6rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.query-results-row strong {
  color: var(--brand-blue);
  font-size: 0.82rem;
  font-weight: 800;
}

.query-source,
.query-anchor {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  color: var(--text-faint);
  font-size: 0.55rem;
  font-weight: 500;
}

.query-source i { color: #16a34a; font-size: 0.55rem; }
.query-anchor { font-family: ui-monospace, 'Cascadia Code', Consolas, monospace; }

/* ── Floating Results Panel ──────────────────────────────────── */
.finder-floating-results {
  position: absolute;
  z-index: 600;
  top: 0.75rem;
  right: 0.75rem;
  width: 340px;
  max-height: calc(100% - 1.5rem);
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(221, 229, 239, 0.9);
  border-radius: 1rem;
  box-shadow: 0 8px 32px rgba(16, 24, 40, 0.14), 0 2px 8px rgba(16, 24, 40, 0.06);
  backdrop-filter: blur(16px) saturate(1.4);
  -webkit-backdrop-filter: blur(16px) saturate(1.4);
  overflow: hidden;
  animation: float-results-in 0.25s ease both;
}

@keyframes float-results-in {
  from { opacity: 0; transform: translateX(12px) scale(0.98); }
  to { opacity: 1; transform: translateX(0) scale(1); }
}

.floating-results-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 0.85rem;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.floating-results-title {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.floating-results-title strong {
  font-size: 0.72rem;
  font-weight: 800;
  color: var(--text-primary);
}

.floating-results-title .results-filter-note {
  color: var(--text-muted);
  font-size: 0.6rem;
  font-weight: 500;
}

.floating-results-close {
  width: 24px;
  height: 24px;
  display: grid;
  place-items: center;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-faint);
  cursor: pointer;
  font-size: 0.6rem;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.floating-results-close:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.floating-results-search {
  position: relative;
  padding: 0.5rem 0.85rem;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.floating-results-search i {
  position: absolute;
  left: 1.15rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-faint);
  font-size: 0.65rem;
  pointer-events: none;
}

.floating-results-search input {
  width: 100%;
  padding: 0.38rem 0.6rem 0.38rem 1.55rem;
  border: 1px solid var(--border-default);
  border-radius: 0.55rem;
  background: var(--surface-subtle);
  font-size: 0.68rem;
  color: var(--text-primary);
  outline: none;
  transition: border-color 160ms ease, box-shadow 160ms ease;
}

.floating-results-search input:focus {
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 4px rgba(230, 57, 70, 0.10);
}

.floating-results-search input::placeholder { color: var(--text-faint); }

.floating-results-list {
  flex: 1;
  min-height: 0;
  padding: 0.5rem;
  overflow-y: auto;
  display: grid;
  gap: 0.35rem;
  align-content: start;
}

.floating-results-list::-webkit-scrollbar { width: 4px; }
.floating-results-list::-webkit-scrollbar-track { background: transparent; }
.floating-results-list::-webkit-scrollbar-thumb { background: var(--border-default); border-radius: 999px; }
.floating-results-list::-webkit-scrollbar-thumb:hover { background: var(--text-faint); }

.floating-results-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 1.5rem;
  color: var(--text-faint);
}

.floating-results-empty i { font-size: 1.2rem; }
.floating-results-empty span { font-size: 0.72rem; }

.result-card {
  width: 100%;
  padding: 0.65rem;
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 0.6rem;
  align-items: center;
  text-align: left;
  color: var(--text-primary);
  background: var(--surface-card);
  border: 1px solid #eef1f5;
  border-radius: 0.85rem;
  cursor: pointer;
  transition: border-color 180ms ease, background 180ms ease, box-shadow 180ms ease, transform 180ms ease;
}

.result-card:hover {
  border-color: var(--brand-blue-100);
  background: #f8faff;
  box-shadow: 0 4px 14px -4px rgba(230, 57, 70, 0.16);
  transform: translateY(-1px);
}

.result-card.selected {
  border-color: var(--brand-blue);
  background: linear-gradient(135deg, #f5f8ff 0%, #eef3ff 100%);
  box-shadow: 0 0 0 2px rgba(230, 57, 70, 0.12), 0 6px 16px -4px rgba(230, 57, 70, 0.18);
}

.result-marker {
  width: 1.95rem;
  height: 1.95rem;
  display: grid;
  place-items: center;
  color: #fff;
  border-radius: 50%;
  font-size: 0.6rem;
  flex-shrink: 0;
  box-shadow: 0 3px 8px rgba(0, 0, 0, 0.16);
}

.result-marker.is-customer {
  background: linear-gradient(135deg, #22c55e 0%, #15803d 100%);
  box-shadow: 0 3px 8px rgba(21, 128, 61, 0.35);
}

.result-marker.is-customer b {
  font-family: 'Segoe UI', system-ui, sans-serif;
  font-size: 0.95rem;
  font-weight: 800;
  font-style: italic;
  line-height: 1;
  letter-spacing: -0.04em;
}

.result-customer-tag { transform: scale(0.85); transform-origin: left; }

.result-info {
  min-width: 0;
  display: grid;
  gap: 0.12rem;
}

.result-name-row {
  display: flex;
  align-items: center;
  gap: 0.38rem;
}

.result-name-row strong {
  font-size: 0.72rem;
  font-weight: 700;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-rating-tag { transform: scale(0.85); transform-origin: left; }

.result-category {
  color: var(--brand-blue);
  font-size: 0.55rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.result-address {
  color: var(--text-faint);
  font-size: 0.58rem;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-meta-row {
  display: flex;
  align-items: center;
  gap: 0.38rem;
  margin-top: 0.12rem;
}

.result-distance {
  display: flex;
  align-items: center;
  gap: 0.22rem;
  color: var(--text-muted);
  font-size: 0.55rem;
  font-weight: 500;
}

.result-distance i { font-size: 0.5rem; }

.result-status-tag { transform: scale(0.8); transform-origin: left; }
.result-menu-tag { transform: scale(0.8); transform-origin: left; }

.result-chevron {
  color: var(--text-faint);
  font-size: 0.6rem;
  flex-shrink: 0;
  transition: color 160ms ease, transform 160ms ease;
}

.result-card:hover .result-chevron,
.result-card.selected .result-chevron { color: var(--brand-blue); transform: translateX(2px); }

/* ── Map Stage ───────────────────────────────────────────────── */
.finder-map-stage {
  position: relative;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  background: #e8eef5;
}

.leaflet-map {
  position: absolute;
  inset: 0;
  z-index: 1;
  width: 100%;
  height: 100%;
}

.leaflet-map :deep(.leaflet-control-zoom) {
  border: 0;
  box-shadow: 0 6px 20px rgba(30, 54, 84, 0.16);
  border-radius: 0.7rem !important;
  overflow: hidden;
}

.leaflet-map :deep(.leaflet-control-zoom a) { color: #26344b; }

.leaflet-map :deep(.leaflet-control-attribution) {
  color: #617087;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 0.4rem 0 0 0;
  font-size: 9px;
}

/* ── Map Source Badge ────────────────────────────────────────── */
.map-source-badge {
  position: absolute;
  z-index: 500;
  left: 0.8rem;
  bottom: 0.8rem;
  max-width: 260px;
  padding: 0.55rem 0.75rem;
  display: flex;
  gap: 0.5rem;
  align-items: flex-start;
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(221, 229, 239, 0.9);
  border-radius: 0.7rem;
  box-shadow: 0 6px 20px rgba(30, 54, 84, 0.12);
  backdrop-filter: blur(14px) saturate(1.4);
  -webkit-backdrop-filter: blur(14px) saturate(1.4);
}

.map-source-badge > i {
  margin-top: 0.02rem;
  color: #16a34a;
  font-size: 0.75rem;
  filter: drop-shadow(0 1px 1px rgba(22, 163, 74, 0.25));
}
.map-source-badge div { display: grid; gap: 0.12rem; }
.map-source-badge strong { font-size: 0.58rem; font-weight: 700; }
.map-source-badge span { color: #718096; font-size: 0.52rem; line-height: 1.5; }

/* ── Existing Customer Badge ─────────────────────────────────── */
.map-customer-badge {
  position: absolute;
  z-index: 500;
  left: 0.8rem;
  bottom: 4.6rem;
  max-width: 280px;
  padding: 0.55rem 0.75rem;
  display: flex;
  gap: 0.55rem;
  align-items: flex-start;
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(187, 247, 208, 0.9);
  border-radius: 0.7rem;
  box-shadow: 0 6px 20px rgba(21, 128, 61, 0.14);
  backdrop-filter: blur(14px) saturate(1.4);
  -webkit-backdrop-filter: blur(14px) saturate(1.4);
}

.map-customer-badge.is-loading { opacity: 0.7; }

.map-customer-dot {
  width: 1.7rem;
  height: 1.7rem;
  margin-top: 0.02rem;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(135deg, #22c55e 0%, #15803d 100%);
  border-radius: 50%;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.22), 0 3px 8px rgba(21, 128, 61, 0.35);
}

.map-customer-dot b {
  font-family: 'Segoe UI', system-ui, sans-serif;
  font-size: 0.95rem;
  font-weight: 800;
  font-style: italic;
  line-height: 1;
  letter-spacing: -0.04em;
}

.map-customer-badge div { display: grid; gap: 0.1rem; }
.map-customer-badge strong { font-size: 0.6rem; font-weight: 700; color: var(--text-primary); }
.map-customer-badge span { color: #718096; font-size: 0.52rem; line-height: 1.5; }

.map-empty-state,
.map-loading-state {
  position: absolute;
  z-index: 450;
  left: 50%;
  top: 50%;
  display: flex;
  width: min(340px, calc(100% - 2rem));
  transform: translate(-50%, -50%);
  flex-direction: column;
  align-items: center;
  gap: 0.4rem;
  padding: 1.1rem 1.25rem;
  border: 1px solid rgba(226, 232, 240, 0.95);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.93);
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.12);
  text-align: center;
  backdrop-filter: blur(12px);
}

.map-empty-icon {
  display: grid;
  width: 42px;
  height: 42px;
  place-content: center;
  border-radius: 12px;
  background: #fff0f1;
  color: #e63946;
}

.map-empty-state strong,
.map-loading-state strong { color: #0f172a; font-size: 0.82rem; }
.map-empty-state p { margin: 0; color: #64748b; font-size: 0.68rem; line-height: 1.45; }

.map-loading-spinner {
  width: 28px;
  height: 28px;
  border: 3px solid #ffd9dc;
  border-top-color: #e63946;
  border-radius: 50%;
  animation: finder-spin 0.75s linear infinite;
}

/* ── Detail Dialog ───────────────────────────────────────────── */
.detail-dialog {
  display: grid;
  gap: 1rem;
}

.dialog-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 2rem;
  color: var(--text-muted);
  font-size: 0.75rem;
}

.loading-pulse {
  width: 28px;
  height: 28px;
  border: 3px solid var(--brand-blue-100);
  border-top-color: var(--brand-blue);
  border-radius: 50%;
  animation: finder-spin 0.75s linear infinite;
}

@keyframes finder-spin { to { transform: rotate(360deg); } }

.detail-hero-bar {
  display: flex;
  gap: 0.8rem;
  align-items: flex-start;
}

.detail-hero {
  width: 2.85rem;
  height: 2.85rem;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  color: #fff;
  border-radius: 0.85rem;
  font-size: 1rem;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.16);
}

.detail-hero.is-customer {
  background: linear-gradient(135deg, #22c55e 0%, #15803d 100%);
  border-radius: 50%;
  box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.2), 0 6px 16px rgba(21, 128, 61, 0.35);
}

.detail-hero.is-customer b {
  font-family: 'Segoe UI', system-ui, sans-serif;
  font-size: 1.35rem;
  font-weight: 800;
  font-style: italic;
  line-height: 1;
  letter-spacing: -0.04em;
}

.detail-hero-info h2 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  line-height: 1.3;
  color: var(--text-primary);
}

.detail-hero-meta {
  display: flex;
  align-items: center;
  gap: 0.38rem;
  margin-top: 0.32rem;
  flex-wrap: wrap;
}

.detail-hero-meta > span {
  color: var(--text-muted);
  font-size: 0.7rem;
  font-weight: 500;
}

.detail-editorial {
  display: flex;
  gap: 0.5rem;
  padding: 0.65rem 0.75rem;
  background: var(--surface-subtle);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 0.72rem;
  font-style: italic;
  line-height: 1.5;
}

.detail-editorial i {
  color: var(--brand-blue);
  font-size: 0.65rem;
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.detail-section {
  display: grid;
  gap: 0.65rem;
}

.detail-section-title {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.detail-section-title i {
  font-size: 0.6rem;
  color: var(--brand-blue);
}

.detail-info-grid {
  display: grid;
  gap: 0.6rem;
}

.detail-info-item {
  display: flex;
  gap: 0.7rem;
  align-items: flex-start;
}

.detail-info-item > i {
  margin-top: 0.12rem;
  color: var(--brand-blue);
  font-size: 0.72rem;
  width: 1rem;
  text-align: center;
  flex-shrink: 0;
}

.detail-info-item > div {
  display: grid;
  gap: 0.08rem;
}

.detail-info-label {
  color: var(--text-muted);
  font-size: 0.55rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.detail-info-value {
  color: var(--text-primary);
  font-size: 0.78rem;
  line-height: 1.5;
}

.detail-info-value.detail-types {
  color: var(--text-secondary);
  font-size: 0.7rem;
}

.detail-info-link {
  color: var(--brand-blue);
  font-size: 0.75rem;
  font-weight: 600;
  text-decoration: none;
  transition: opacity 160ms ease;
}

.detail-info-link:hover { opacity: 0.72; }

/* Hours */
.detail-hours-grid {
  display: grid;
  gap: 0.3rem;
}

.detail-hours-badge {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  font-size: 0.65rem;
  font-weight: 700;
  width: fit-content;
  background: #fef2f2;
  color: #dc2626;
}

.detail-hours-badge.is-open {
  background: #f0fdf4;
  color: #16a34a;
}

.hours-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.detail-hours-day {
  padding: 0.2rem 0.65rem;
  color: var(--text-secondary);
  font-size: 0.62rem;
  line-height: 1.5;
  border-bottom: 1px solid var(--border-light);
}

.detail-hours-day:last-child { border-bottom: 0; }

/* Options grid */
.detail-options-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.detail-option-chip {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.3rem 0.55rem;
  border-radius: 999px;
  font-size: 0.6rem;
  font-weight: 500;
  color: var(--text-faint);
  background: var(--surface-subtle);
  border: 1px solid var(--border-light);
}

.detail-option-chip.active {
  color: #16a34a;
  background: #f0fdf4;
  border-color: #bbf7d0;
}

.detail-option-chip i { font-size: 0.55rem; }

/* Reviews */
.detail-reviews-list {
  display: grid;
  gap: 0.65rem;
}

.detail-review {
  padding: 0.65rem;
  background: var(--surface-subtle);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
}

.review-author {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.review-author-photo {
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  object-fit: cover;
}

.review-author-name {
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-primary);
}

.review-rating { transform: scale(0.8); transform-origin: left; }

.review-time {
  color: var(--text-faint);
  font-size: 0.55rem;
}

.review-text {
  margin: 0.4rem 0 0;
  color: var(--text-secondary);
  font-size: 0.65rem;
  line-height: 1.5;
}

/* Photos */
.detail-photos-row {
  display: flex;
  gap: 0.4rem;
  overflow-x: auto;
  padding-bottom: 0.25rem;
}

.detail-photo-tile {
  position: relative;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  width: 86px;
}

.detail-photo-button {
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: zoom-in;
}

.detail-photo-button:hover .detail-photo,
.detail-photo-button:focus-visible .detail-photo {
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.15);
}

.finder-photo-preview {
  display: grid;
  justify-items: center;
  gap: 0.65rem;
}

.finder-photo-preview img {
  display: block;
  max-width: 100%;
  max-height: 75vh;
  object-fit: contain;
  border-radius: var(--radius-md);
}

.finder-photo-preview small { color: var(--text-muted); }

.detail-photo {
  position: relative;
  width: 86px;
  height: 86px;
  border-radius: var(--radius-sm);
  object-fit: cover;
  border: 1px solid var(--border-light);
}

.detail-photo-badge {
  position: absolute;
  left: 0.35rem;
  bottom: 0.35rem;
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  padding: 0.1rem 0.4rem;
  border-radius: 9999px;
  font-size: 0.52rem;
  font-weight: 700;
  color: #fff;
  background: rgba(15, 23, 42, 0.6);
}

.detail-photo-badge i { font-size: 0.45rem; }

.detail-photo-empty {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.6rem 0.75rem;
  color: var(--text-muted);
  font-size: 0.66rem;
  background: var(--surface-subtle);
  border-radius: var(--radius-sm);
}

.detail-photo-empty i { color: #cbd5e1; }

.menu-images {
  display: grid;
  gap: 0.5rem;
}

.menu-images-note {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin: 0;
  color: var(--text-muted);
  font-size: 0.66rem;
}

.menu-images-note i { color: var(--brand-blue); }

.menu-image-tile {
  text-decoration: none;
}

.menu-images-error {
  font-size: 0.68rem;
}

.detail-photo::-webkit-scrollbar { height: 3px; }
.detail-photo::-webkit-scrollbar-thumb { background: var(--border-default); border-radius: 999px; }

.detail-assignment {
  padding-top: 0.75rem;
  border-top: 1px solid var(--border-light);
}

.detail-assignment h3 {
  margin: 0 0 0.7rem;
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.detail-assignment-fields {
  display: grid;
  gap: 0.7rem;
}

.detail-assignment-fields .field {
  display: grid;
  gap: 0.32rem;
}

.detail-assignment-fields .field > span {
  color: var(--text-muted);
  font-size: 0.65rem;
  font-weight: 700;
}

.detail-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.55rem;
}

.assignment-warning {
  font-size: 0.72rem;
  margin-top: 0.18rem;
  border-radius: 0.6rem;
}

/* ── Create Pin dialog ── */
.pin-form {
  display: grid;
  gap: 0.8rem;
}

.pin-error {
  font-size: 0.72rem;
}

.pin-map-note {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.6rem 0.75rem;
  background: #eef2ff;
  border: 1px solid #e0e7ff;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 0.68rem;
  line-height: 1.5;
}

.pin-map-note i {
  color: var(--brand-blue);
  margin-top: 0.08rem;
}

.pin-form .field {
  display: grid;
  gap: 0.32rem;
}

.pin-form .field > span {
  color: var(--text-muted);
  font-size: 0.65rem;
  font-weight: 700;
}

.pin-coords {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.7rem;
}

.pin-assignment {
  display: grid;
  gap: 0.7rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--border-light);
}

.pin-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.55rem;
}

/* ── Side-by-side results (results column, map never covered) ── */
@media (min-width: 901px) {
  .finder-map-stage.has-results {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 340px;
  }

  .finder-map-stage.has-results .leaflet-map {
    position: relative;
    inset: auto;
    width: auto;
    height: auto;
    min-width: 0;
    min-height: 0;
  }

  .finder-map-stage.has-results .finder-floating-results {
    position: relative;
    top: auto;
    right: auto;
    width: 340px;
    height: 100%;
    max-height: none;
    border: none;
    border-left: 1px solid #e5eaf0;
    border-radius: 0;
    box-shadow: none;
    animation: none;
  }
}

@media (max-width: 1180px) and (min-width: 901px) {
  .finder-map-stage.has-results {
    grid-template-columns: minmax(0, 1fr) 300px;
  }

  .finder-map-stage.has-results .finder-floating-results {
    width: 300px;
  }
}

/* ── Responsive ──────────────────────────────────────────────── */
@media (max-width: 1180px) {
  .finder-page-header { grid-template-columns: 1fr; }
  .finder-page-stats { width: 100%; }
  .finder-desktop-shell { grid-template-columns: 310px minmax(0, 1fr); }
  .category-grid { grid-template-columns: 1fr 1fr; }
  .finder-floating-results { width: 300px; }
}

@media (max-width: 900px) {
  .finder-page {
    height: auto;
    min-height: 100vh;
    overflow: visible;
    padding: 0.75rem;
  }
  .finder-desktop-shell {
    min-height: 780px;
    grid-template-columns: 1fr;
    grid-template-rows: auto minmax(500px, 1fr);
    overflow: visible;
  }
  .finder-left-panel { border-right: 0; border-bottom: 1px solid var(--border-light); }
  .finder-filter-scroll { overflow: visible; }
  .finder-floating-results {
    top: 0.5rem;
    right: 0.5rem;
    width: calc(100% - 1rem);
    max-height: calc(100% - 1rem);
  }
}

@media (max-width: 640px) {
  .finder-page-header { align-items: stretch; }
  .finder-heading-left { align-items: flex-start; }
  .finder-page-title p { white-space: normal; }
  .finder-page-stats { grid-template-columns: repeat(2, 1fr); }
  .finder-stat:nth-child(2) { border-right: 0; }
  .finder-stat:nth-child(-n + 2) { border-bottom: 1px solid #e5eaf0; }
  .finder-panel-title h1 { font-size: 0.9rem; }
  .map-source-badge { max-width: calc(100% - 1.6rem); }
  .category-grid { grid-template-columns: repeat(2, 1fr); }
  .coordinate-grid, .filter-actions { grid-template-columns: 1fr; }
  .finder-floating-results {
    top: 0;
    right: 0;
    width: 100%;
    max-height: 100%;
    border-radius: 0;
    border: none;
  }
}

@supports not (height: 100dvh) {
  .finder-page {
    height: calc(100vh - 4rem);
  }
}

@media (max-height: 760px) and (min-width: 901px) {
  .finder-page-header {
    min-height: 56px;
    padding-block: 0.38rem;
  }

  .finder-page-title p,
  .finder-panel-title span {
    display: none;
  }

  .finder-panel-header {
    padding: 0.45rem 0.75rem;
  }

  .finder-filter-scroll {
    padding-top: 0.2rem;
  }

  .filter-section {
    padding: 0.32rem 0;
  }
}

</style>

<style>
.finder-leaflet-icon-host { border: 0; background: transparent; }

.finder-center-icon-host {
  border: 0;
  background: transparent;
}

.finder-center-icon-host.leaflet-marker-draggable,
.finder-center-marker {
  cursor: grab;
}

.finder-center-icon-host.leaflet-marker-draggable:active,
.finder-center-marker:active { cursor: grabbing; }

.finder-center-marker {
  position: relative;
  z-index: 1;
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  border: 3px solid #fff;
  border-radius: 50%;
  box-shadow: 0 0 0 1px rgba(29, 78, 216, 0.16), 0 6px 18px rgba(29, 78, 216, 0.45);
  cursor: grab;
  transition: width 160ms ease, height 160ms ease, box-shadow 160ms ease;
}

.finder-center-marker:hover { box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.18), 0 6px 18px rgba(29, 78, 216, 0.5); }
.finder-center-marker i { font-size: 1rem; text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2); }

.finder-center-ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 2px solid rgba(59, 130, 246, 0.55);
  animation: finder-ring-pulse 1.8s ease-out infinite;
  pointer-events: none;
}

@keyframes finder-ring-pulse {
  0% { transform: scale(1); opacity: 0.7; }
  100% { transform: scale(1.85); opacity: 0; }
}

.finder-leaflet-marker {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  color: #fff;
  background: var(--marker-color);
  border: 3px solid #fff;
  border-radius: 50% 50% 50% 0;
  box-shadow: 0 6px 16px rgba(22, 41, 67, 0.32), 0 0 0 1px rgba(22, 41, 67, 0.04);
  transform: rotate(-45deg);
  transition: width 160ms cubic-bezier(0.4, 0, 0.2, 1),
              height 160ms cubic-bezier(0.4, 0, 0.2, 1),
              box-shadow 160ms cubic-bezier(0.4, 0, 0.2, 1);
}

.finder-leaflet-marker i { transform: rotate(45deg); font-size: 0.8rem; }

.finder-leaflet-marker.is-customer {
  background: linear-gradient(135deg, #22c55e 0%, #15803d 100%);
  border-color: #ffffff;
  border-radius: 50%;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.25), 0 6px 16px rgba(21, 128, 61, 0.4);
}

.finder-leaflet-marker.is-custom-pin {
  background: linear-gradient(135deg, #a78bfa 0%, #7c3aed 100%);
  border-color: #ffffff;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.25), 0 6px 16px rgba(124, 58, 237, 0.45);
  cursor: grab;
}

.finder-leaflet-marker.is-custom-pin:active { cursor: grabbing; }

.finder-leaflet-marker.is-custom-pin.is-selected {
  box-shadow: 0 0 0 6px rgba(124, 58, 237, 0.22), 0 8px 22px rgba(124, 58, 237, 0.5);
}

.finder-leaflet-marker.is-customer b {
  transform: rotate(45deg);
  font-family: 'Segoe UI', system-ui, sans-serif;
  font-size: 1.15rem;
  font-weight: 800;
  font-style: italic;
  line-height: 1;
  letter-spacing: -0.04em;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.25);
}

.finder-leaflet-marker.is-selected {
  width: 44px;
  height: 44px;
  box-shadow: 0 0 0 6px rgba(230, 57, 70, 0.16),
              0 8px 22px rgba(22, 41, 67, 0.38);
}

.finder-leaflet-marker.is-customer.is-selected {
  width: 46px;
  height: 46px;
  box-shadow: 0 0 0 6px rgba(34, 197, 94, 0.2), 0 8px 22px rgba(21, 128, 61, 0.42);
}

.finder-leaflet-marker.is-selected i { font-size: 1rem; }
.finder-leaflet-marker.is-customer.is-selected b { font-size: 1.4rem; }
</style>
