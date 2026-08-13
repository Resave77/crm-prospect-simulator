import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const salesDetail = () => readFile(new URL('../src/views/Sales/Prospect/ProspectDetailView.vue', import.meta.url), 'utf8')
const adminDetail = () => readFile(new URL('../src/views/Admin/Prospect/ProspectReviewView.vue', import.meta.url), 'utf8')
const aiSummary = () => readFile(new URL('../src/components/prospect-ai/AISummaryCard.vue', import.meta.url), 'utf8')
const aiMenu = () => readFile(new URL('../src/components/prospect-ai/AIMenuProfilingCard.vue', import.meta.url), 'utf8')
const tanyaAi = () => readFile(new URL('../src/components/prospect-ai/TanyaAICard.vue', import.meta.url), 'utf8')

test('Prospect Detail pages include AI prep components behind expected permissions', async () => {
  const sales = await salesDetail()
  const admin = await adminDetail()

  for (const source of [sales, admin]) {
    assert.match(source, /AISummaryCard/)
    assert.match(source, /AIMenuProfilingCard/)
    assert.match(source, /TanyaAICard/)
    assert.match(source, /view_ai_summary/)
    assert.match(source, /view_ai_menu_profiling/)
    assert.match(source, /use_prospect_ai_chat/)
  }
})

test('AI prep components keep OpenAI access backend-only and activate Tanya only on submit', async () => {
  const combined = [await aiSummary(), await aiMenu(), await tanyaAi()].join('\n')

  assert.equal(combined.includes('api.openai.com'), false)
  assert.equal(combined.includes('OPENAI_API_KEY'), false)
  assert.equal(combined.includes('getAIStatus'), false)
  assert.equal(combined.includes('GenerateText'), false)
  assert.equal(combined.includes('askProspectAI'), true)
  assert.equal(combined.includes('@submit.prevent="submit"'), true)
  assert.equal(combined.includes('@keydown.enter.exact.prevent'), true)
  assert.equal(combined.includes('onMounted'), true)
  assert.match(combined, /getProspectAIChatHistory/)
})

test('AI menu profiling waits for real menu data instead of Google photos', async () => {
  const source = await aiMenu()

  assert.match(source, /MENU_DATA_NOT_AVAILABLE/)
  assert.match(source, /Menu data not available yet/)
  assert.equal(source.includes('props.placeDetails?.photos'), false)
  assert.equal(source.includes('photo.isMenu'), false)
  assert.equal(source.includes('Menu photo'), false)
})

test('Prospect Detail keeps secure photo gallery flow for photos and menu', async () => {
  const sales = await salesDetail()
  const admin = await adminDetail()

  const galleryPhotos = String.raw`:photos="(?:placeDetails\.photos|\(placeDetails\?\.photos \?\? \[\]\))"`
  for (const source of [sales, admin]) {
    assert.match(source, new RegExp(`${galleryPhotos}[^>]*section="photos"`))
    assert.match(source, new RegExp(`${galleryPhotos}[^>]*section="menu"`))
  }
})
