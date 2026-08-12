import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const crmApi = () => readFile(new URL('../src/api/crm.ts', import.meta.url), 'utf8')

test('place photo blob helper strips the api client base path before authenticated fetch', async () => {
  const source = await crmApi()
  const normalized = '/api/v1/places/photo?name=places%2Fabc%2Fphotos%2Fdef'.replace(/^\/api\/v1(?=\/)/, '')
  assert.equal(normalized, '/places/photo?name=places%2Fabc%2Fphotos%2Fdef')
  assert.ok(source.includes("path.replace(/^\\/api\\/v1(?=\\/)/, '')"))
  assert.ok(source.includes("api.get(apiClientPath(photoUrl), { responseType: 'blob' })"))
  assert.equal(source.includes("api.get(photoUrl, { responseType: 'blob' })"), false)
})
