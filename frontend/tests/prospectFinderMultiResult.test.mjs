import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const source = await readFile(new URL('../src/views/Admin/Prospect/ProspectFinderView.vue', import.meta.url), 'utf8')

test('Prospect Finder renders the complete search collection and only loads details on explicit selection', () => {
  assert.match(source, /results\.value = await crmApi\.searchPlaces\(/)
  assert.match(source, /v-for="item in filteredResults"/)
  assert.match(source, /@click="selectResult\(item, true\)"/)
  assert.doesNotMatch(source, /results\.value\s*=.*\[0\]/)
  assert.equal((source.match(/getPlaceDetails\(/g) || []).length, 1)
  assert.equal((source.match(/getPlacePhotoBlob\(/g) || []).length, 1)
})
