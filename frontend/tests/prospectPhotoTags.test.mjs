import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const GALLERY_URL = new URL('../src/components/PlacePhotoGallery.vue', import.meta.url)
const API_URL = new URL('../src/api/crm.ts', import.meta.url)
const TYPES_URL = new URL('../src/types/crm.ts', import.meta.url)

test('menu matching uses the stable Google photo resource name, not array index', async () => {
  const sfc = await readFile(GALLERY_URL, 'utf8')
  const menuPhotos = sfc.match(/const menuPhotos[\s\S]*?\n}/)?.[0]
  const regularPhotos = sfc.match(/const regularPhotos[\s\S]*?\n}/)?.[0]
  assert.ok(menuPhotos, 'menuPhotos computed must be defined')
  assert.ok(regularPhotos, 'regularPhotos computed must be defined')
  assert.match(menuPhotos, /tags\.value\[photo\.name\] === 'MENU'/)
  assert.match(regularPhotos, /tags\.value\[photo\.name\] !== 'MENU'/)
  assert.doesNotMatch(sfc, /photoIndex/)
  assert.doesNotMatch(sfc, /savingIndex/)
  assert.doesNotMatch(sfc, /item\.index/)
})

test('applyTags keys tags by photoName', async () => {
  const sfc = await readFile(GALLERY_URL, 'utf8')
  const applyTags = sfc.match(/function applyTags[\s\S]*?\n}/)?.[0]
  assert.ok(applyTags, 'applyTags must be defined')
  assert.match(applyTags, /photoName: string/)
  assert.match(applyTags, /map\[t\.photoName\] = t\.category/)
})

test('tagging sends photo.name as photoName through the API', async () => {
  const sfc = await readFile(GALLERY_URL, 'utf8')
  const setCategory = sfc.match(/async function setCategory[\s\S]*?\n}/)?.[0]
  assert.ok(setCategory, 'setCategory must be defined')
  assert.match(setCategory, /setProspectPhotoTag\(props\.prospectId!, photo\.name, category, props\.role\)/)

  const api = await readFile(API_URL, 'utf8')
  assert.match(api, /export async function setProspectPhotoTag\(prospectId: string, photoName: string, category: PhotoCategory, role: UserRole\)/)
  assert.match(api, /\{ photoName, category \}/)
})

test('photo tags type exposes photoName instead of photoIndex', async () => {
  const types = await readFile(TYPES_URL, 'utf8')
  const tagType = types.match(/export interface ProspectPhotoTag[\s\S]*?\n}/)?.[0]
  assert.ok(tagType, 'ProspectPhotoTag must be defined')
  assert.match(tagType, /photoName: string/)
  assert.doesNotMatch(tagType, /photoIndex/)
})

test('reordering Google photos does not change name-keyed matching', async () => {
  const sfc = await readFile(GALLERY_URL, 'utf8')
  assert.match(sfc, /tags\.value\[photo\.name\]/)
  assert.doesNotMatch(sfc, /tags\.value\[index\]/)
  assert.doesNotMatch(sfc, /tags\.value\[i\]/)
})

test('authenticated blob photo flow is preserved', async () => {
  const api = await readFile(API_URL, 'utf8')
  assert.ok(api.includes("path.replace(/^\\/api\\/v1(?=\\/)/, '')"))
  assert.ok(api.includes("api.get(apiClientPath(photoUrl), { responseType: 'blob' })"))
  assert.equal(api.includes("api.get(photoUrl, { responseType: 'blob' })"), false)
})
