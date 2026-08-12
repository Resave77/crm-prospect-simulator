import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const crmApi = () => readFile(new URL('../src/api/crm.ts', import.meta.url), 'utf8')
const crmTypes = () => readFile(new URL('../src/types/crm.ts', import.meta.url), 'utf8')

test('AI status helper uses the shared authenticated API client', async () => {
  const source = await crmApi()
  assert.ok(source.includes("api.get<ApiEnvelope<AIStatus>>('/ai/status')"))
  assert.equal(source.includes('axios.create'), false)
  assert.equal(source.includes('OPENAI_API_KEY'), false)
  assert.equal(source.includes('api.openai.com'), false)
})

test('AI status type exposes only configuration booleans', async () => {
  const source = await crmTypes()
  assert.match(source, /export interface AIStatus/)
  assert.match(source, /enabled: boolean/)
  assert.match(source, /configured: boolean/)
  assert.match(source, /modelConfigured: boolean/)
  assert.equal(source.includes('apiKey'), false)
  assert.equal(source.includes('OPENAI_API_KEY'), false)
})
