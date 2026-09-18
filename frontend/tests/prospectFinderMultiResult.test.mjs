import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const source = await readFile(new URL('../src/views/Admin/Prospect/ProspectFinderView.vue', import.meta.url), 'utf8')
const apiSource = await readFile(new URL('../src/api/crm.ts', import.meta.url), 'utf8')

test('Prospect Finder renders the complete search collection and loads details on selection', () => {
  assert.match(source, /results\.value = await crmApi\.searchPlaces\(/)
  assert.match(source, /v-for="item in filteredResults"/)
  assert.match(source, /@click="selectResult\(item, true\)"/)
  assert.doesNotMatch(source, /results\.value\s*=.*\[0\]/)
  assert.equal((source.match(/getPlaceDetails\(/g) || []).length, 1)
  assert.match(source, /await getPlaceDetails\(item\.googlePlaceId\)/)
  assert.match(apiSource, /'\/admin\/prospect-finder\/place-details', \{ params: \{ googlePlaceId \} \}/)
  assert.match(source, /placeDetails\.value = placeResultToDetails\(item\)/)
  assert.equal((source.match(/getPlacePhotoBlob\(/g) || []).length, 1)
  assert.doesNotMatch(source, /categorySelections/)
  assert.doesNotMatch(source, /label="Create Pin"[\s\S]*@click="openPinForm"/)
})

test('Prospect Finder protects map, photo, menu, and coordinate state across actions', () => {
  assert.match(source, /markers\.forEach\(\(marker\) => marker\.remove\(\)\)/)
  assert.match(source, /selected\.value = null/)
  assert.match(source, /revokePlacePhotoObjectUrls\(\)/)
  assert.match(source, /let menuImagesRequestToken = 0/)
  assert.match(source, /requestToken !== menuImagesRequestToken/)
  assert.match(source, /const hasVisiblePhotos = computed/)
  assert.match(source, /Number\.isFinite\(latitude\.value\)/)
  assert.match(source, /Number\.isFinite\(pinLat\.value\)/)
})
