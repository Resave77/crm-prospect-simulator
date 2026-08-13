import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const GALLERY_URL = new URL('../src/components/PlacePhotoGallery.vue', import.meta.url)
const API_URL = new URL('../src/api/crm.ts', import.meta.url)
const TYPES_URL = new URL('../src/types/crm.ts', import.meta.url)

test('menu matching uses the original Google photo index', async () => {
  const sfc = await readFile(GALLERY_URL, 'utf8')
  const menuPhotos = sfc.match(/const menuPhotos[\s\S]*?\n}/)?.[0]
  const regularPhotos = sfc.match(/const regularPhotos[\s\S]*?\n}/)?.[0]
  assert.ok(menuPhotos, 'menuPhotos computed must be defined')
  assert.ok(regularPhotos, 'regularPhotos computed must be defined')
  assert.match(menuPhotos, /photoIndexOf\(photo\).*category === 'MENU'/s)
  assert.match(regularPhotos, /photoIndexOf\(photo\).*category !== 'MENU'/s)
  assert.match(sfc, /photoIndexOf\(photo\)/)
})

test('applyTags ignores legacy tags without a resolvable name or photoIndex', async () => {
  const sfc = await readFile(GALLERY_URL, 'utf8')
  const applyTags = sfc.match(/function applyTags[\s\S]*?\n}/)?.[0]
  assert.ok(applyTags, 'applyTags must be defined')
  assert.match(applyTags, /photoIndex\?: number \| null/)
  assert.match(applyTags, /photo_index\?: number \| null/)
  assert.match(applyTags, /photoName\?: string \| null/)
  assert.match(applyTags, /photo_name\?: string \| null/)
  assert.match(applyTags, /Array\.isArray\(payload\.data\)/)
  assert.match(applyTags, /map\[String\(photoIndex\)\]/)
})

test('normalizes enveloped MENU tags against the original index', async () => {
  const sfc = await readFile(GALLERY_URL, 'utf8')
  const applyTags = sfc.match(/function applyTags[\s\S]*?\n}/)?.[0]
  assert.ok(applyTags, 'applyTags must be defined')
  assert.match(applyTags, /category.*toUpperCase\(\)/)
  assert.match(sfc, /tags\.value\[String\(photoIndexOf\(photo\)\)\].*category === 'MENU'/s)
  assert.match(sfc, /photoIndexOf\(photo\), category/)
})

test('tagging sends photo.name as photoName through the API', async () => {
  const sfc = await readFile(GALLERY_URL, 'utf8')
  const setCategory = sfc.match(/async function setCategory[\s\S]*?\n}/)?.[0]
  assert.ok(setCategory, 'setCategory must be defined')
  assert.match(setCategory, /setProspectPhotoTag\(props\.prospectId!, photo\.name, photoIndexOf\(photo\), category, props\.role\)/)

  const api = await readFile(API_URL, 'utf8')
  assert.match(api, /export async function setProspectPhotoTag\(prospectId: string, photoName: string, photoIndex: number, category: PhotoCategory, role: UserRole\)/)
  assert.match(api, /\{ photoName, photoIndex, category \}/)
})

test('photo tags type exposes the additive photoIndex', async () => {
  const types = await readFile(TYPES_URL, 'utf8')
  const tagType = types.match(/export interface ProspectPhotoTag[\s\S]*?\n}/)?.[0]
  assert.ok(tagType, 'ProspectPhotoTag must be defined')
  assert.match(tagType, /photoName: string \| null/)
  assert.match(tagType, /photoIndex: number \| null/)
})

test('filtered photo rendering does not shift original indexes', async () => {
  const sfc = await readFile(GALLERY_URL, 'utf8')
  assert.match(sfc, /photoIndexOf\(photo\)/)
  assert.doesNotMatch(sfc, /regularPhotos.*index/)
})

test('authenticated blob photo flow is preserved', async () => {
  const api = await readFile(API_URL, 'utf8')
  assert.ok(api.includes("path.replace(/^\\/api\\/v1(?=\\/)/, '')"))
  assert.ok(api.includes("api.get(apiClientPath(photoUrl), { responseType: 'blob' })"))
  assert.equal(api.includes("api.get(photoUrl, { responseType: 'blob' })"), false)
})
