import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const salesDetail = () => readFile(new URL('../src/views/Sales/Prospect/ProspectDetailView.vue', import.meta.url), 'utf8')
const adminDetail = () => readFile(new URL('../src/views/Admin/Prospect/ProspectReviewView.vue', import.meta.url), 'utf8')
const aiSummary = () => readFile(new URL('../src/components/prospect-ai/AISummaryCard.vue', import.meta.url), 'utf8')
const aiMenu = () => readFile(new URL('../src/components/prospect-ai/AIMenuProfilingCard.vue', import.meta.url), 'utf8')
const tanyaAi = () => readFile(new URL('../src/components/prospect-ai/TanyaAICard.vue', import.meta.url), 'utf8')
const aiExpandedModal = () => readFile(new URL('../src/components/prospect-ai/AIExpandedModal.vue', import.meta.url), 'utf8')

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

test('Admin AI visibility uses the same permission keys enforced by backend routes', async () => {
  const admin = await adminDetail()
  assert.match(admin, /auth\.hasPermission\('view_ai_summary'\)/)
  assert.match(admin, /auth\.hasPermission\('view_ai_menu_profiling'\)/)
  assert.match(admin, /auth\.hasPermission\('use_prospect_ai_chat'\)/)
  assert.doesNotMatch(admin, /isAdminContext\.value \|\| auth\.hasPermission\('view_ai_/)
})

test('Admin and Sales AI Expand reuse the portrait modal without triggering AI calls', async () => {
  const modal = await aiExpandedModal()
  const sales = await salesDetail()
  const admin = await adminDetail()

  assert.match(modal, /width: clamp\(380px, 32vw, 460px\)/)
  assert.match(modal, /height: min\(82dvh, 780px\)/)
  assert.match(modal, /position: fixed/)
  assert.match(modal, /overflow-y: auto/)
  assert.match(modal, /role="dialog"/)
  assert.match(modal, /aria-modal="true"/)
  assert.match(modal, /aria-label="Tutup"/)
  assert.match(modal, /@media \(max-width: 767px\)/)
  assert.match(modal, /width: calc\(100vw - 24px\)/)
  assert.match(modal, /prefers-reduced-motion/)
  assert.doesNotMatch(modal, /askProspectAI|generateProspectSummary|findProspectMenu|profileProspectMenu/)

  for (const source of [sales, admin]) {
    assert.match(source, /import AIExpandedModal/)
    assert.match(source, /<AIExpandedModal/)
    assert.match(source, /expandedPanel === 'summary'/)
    assert.match(source, /expandedPanel === 'chat'/)
    assert.match(source, /@close="closeExpandedPanel"/)
    assert.match(source, /<AISummaryCard v-if="expandedPanel === 'summary'"/)
    assert.match(source, /<TanyaAICard v-else-if="expandedPanel === 'chat'"/)
    const open = source.match(/function openExpandedPanel[\s\S]*?\n}/)?.[0] || ''
    const close = source.match(/function closeExpandedPanel[\s\S]*?\n}/)?.[0] || ''
    assert.match(open, /expandedPanel\.value = panel/)
    assert.match(close, /expandedPanel\.value = null/)
    assert.doesNotMatch(`${open}\n${close}`, /askProspectAI|generateProspectSummary|findProspectMenu|profileProspectMenu/)
  }
})

test('portrait Tanya AI preserves quick actions and the anchored input interaction', async () => {
  const modal = await aiExpandedModal()
  const chat = await tanyaAi()
  assert.match(modal, /ai-modal-panel--chat/)
  assert.match(modal, /tanya-card-expanded \.tanya-head/)
  assert.match(chat, /Ringkas prospek/)
  assert.match(chat, /Cari peluang yoghurt/)
  assert.match(chat, /Buat pitch visit/)
  assert.match(chat, /class="tanya-chips"/)
  assert.match(chat, /class="tanya-input-row"/)
  assert.match(chat, /@submit\.prevent="submit"/)
})

test('collapsed Tanya AI uses a responsive latest-response preview without horizontal scrolling', async () => {
  const chat = await tanyaAi()
  assert.match(chat, /class="tanya-preview"/)
  assert.match(chat, /Respons AI terbaru/)
  assert.match(chat, /Permintaan terakhir/)
  assert.match(chat, /\.tanya-preview \{ width: 100%; min-width: 0;/)
  assert.match(chat, /-webkit-line-clamp: 5/)
  assert.match(chat, /\.tanya-card:not\(\.tanya-card-expanded\) \.tanya-chips \{ flex-wrap: wrap; overflow-x: visible;/)
  assert.match(chat, /tanya-card:not\(\.tanya-card-expanded\) \.tanya-input-row \{ flex-wrap: wrap;/)
  assert.doesNotMatch(chat, /\.tanya-preview[\s\S]*?position:\s*absolute/)
})

test('Discussion retains its independent landscape expanded layout', async () => {
  const modal = await aiExpandedModal()
  assert.match(modal, /\.ai-modal-panel--discussion \{ width: min\(1080px, calc\(100vw - 80px\)\); height: min\(760px, 78dvh\)/)
  assert.match(modal, /\.ai-modal-panel--discussion \.ai-modal-content :deep\(\.pc-wrap\)/)
  assert.doesNotMatch(modal, /ai-modal-panel--discussion[^\n]*tanya-card/)
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
  const api = await readFile(new URL('../src/api/crm.ts', import.meta.url), 'utf8')

  assert.match(source, /MENU_DATA_NOT_AVAILABLE/)
  assert.match(source, /Analisis menu belum dibuat/)
  assert.equal(source.includes('props.placeDetails?.photos'), false)
  assert.equal(source.includes('photo.isMenu'), false)
  assert.equal(source.includes('Menu photo'), false)
  assert.match(source, /Analisis Menu AI/)
  assert.match(source, /Analisis peluang menu/)
  assert.match(source, /Peluang Menu/)
  assert.match(source, /Kecocokan Yoghurt/)
  assert.match(source, /Tingkat Keyakinan/)
  assert.match(source, /Peluang Utama/)
  assert.match(source, /Langkah Sales Berikutnya/)
  assert.match(source, /LOW: 'RENDAH'/)
  assert.match(source, /MEDIUM: 'SEDANG'/)
  assert.match(source, /HIGH: 'TINGGI'/)
  assert.match(source, /menuResult \? 'Analisis Ulang' : 'Analisis Menu'/)
  assert.match(api, /menu-profile`, \{ force \}, \{ timeout: 60000 \}/)
})

test('menu discovery presentation is deterministic, localized, and locally collapsible', async () => {
  const source = await aiMenu()
  assert.match(source, /Intl\.NumberFormat\('id-ID'/)
  assert.match(source, /replace\(\/\\s\/g, ''\)/)
  assert.match(source, /exact_branch: 'Outlet sesuai'/)
  assert.match(source, /likely_same_branch: 'Kemungkinan outlet sama'/)
  assert.match(source, /brand_only: 'Menu tingkat brand'/)
  assert.match(source, /uncertain: 'Belum terverifikasi'/)
  assert.match(source, /const previewLimit = 5/)
  assert.match(source, /Menu representatif/)
  assert.match(source, /items\.slice\(0, previewLimit\)/)
  assert.match(source, /toggleCategory\(category\.name\)/)
  assert.equal((source.match(/findProspectMenu\(props\.prospectId\)/g) || []).length, 1)
  assert.equal((source.match(/profileProspectMenu\(props\.prospectId/g) || []).length, 1)
  assert.doesNotMatch(source, /item\.branchMatch\.split/)
  assert.doesNotMatch(source, /IDR \$\{/)
})

test('Find Menu summary renders discovery-backed responsive stat cards without extra calls', async () => {
  const source = await aiMenu()
  assert.match(source, /class="menu-stats"/)
  assert.match(source, /class="menu-stat"><strong>\{\{ menuFinding\.categories\.length \}\}<\/strong><span>Kategori/)
  assert.match(source, /class="menu-stat"><strong>\{\{ menuItemCount \}\}<\/strong><span>Menu representatif/)
  assert.match(source, /class="menu-stat"><strong>\{\{ menuFinding\.sources\.length \}\}<\/strong><span>Sumber/)
  assert.match(source, /grid-template-columns: repeat\(3, minmax\(0, 1fr\)\)/)
  assert.match(source, /container-type: inline-size/)
  assert.match(source, /@container \(max-width: 520px\)/)
  assert.match(source, /@container \(max-width: 360px\)/)
  assert.match(source, /@media \(max-width: 767px\)/)
  assert.match(source, /showMenu = !showMenu/)
  assert.match(source, /menuResult \? 'Analisis Ulang' : 'Analisis Menu'/)
  assert.equal((source.match(/findProspectMenu\(props\.prospectId\)/g) || []).length, 1)
  assert.equal((source.match(/profileProspectMenu\(props\.prospectId/g) || []).length, 1)
  assert.doesNotMatch(source, /sidebar-collapsed|sidebar-expanded/)
})

test('menu profiling metrics and states are Indonesian and remain visible while loading', async () => {
  const source = await aiMenu()
  assert.match(source, /LOW: 'RENDAH'/)
  assert.match(source, /MEDIUM: 'SEDANG'/)
  assert.match(source, /HIGH: 'TINGGI'/)
  assert.match(source, /UNKNOWN: 'BELUM DIKETAHUI'/)
  assert.match(source, /Menganalisis peluang menu/)
  assert.match(source, /menuFinding\?\.status === 'FOUND'/)
  assert.match(source, /Menu yang ditemukan tetap dapat dilihat selama analisis/)
  assert.match(source, /Pencarian menu belum dapat diselesaikan/)
  assert.match(source, /Layanan pencarian menu sedang sibuk/)
  assert.match(source, /@media\(max-width:767px\)/)
  assert.match(source, /min-width:0/)
})

test('saved menu profiling remains display-only until an explicit analysis click', async () => {
  const source = await aiMenu()
  assert.doesNotMatch(source, /onMounted/)
  assert.match(source, /storedMenu\.value\?\.profiling/)
  assert.match(source, /@click="analyzeMenu"/)
  assert.match(source, /profileProspectMenu\(props\.prospectId, Boolean\(menuResult\.value\)\)/)
  assert.equal((source.match(/profileProspectMenu\(props\.prospectId/g) || []).length, 1)
})

test('Find Menu is explicit and does not run on mount', async () => {
  const source = await aiMenu()
  const api = await readFile(new URL('../src/api/crm.ts', import.meta.url), 'utf8')
  assert.match(source, /@click="findMenu"/)
  assert.match(source, /findProspectMenu\(props\.prospectId\)/)
  assert.doesNotMatch(source, /onMounted/)
  assert.match(api, /\/ai\/prospects\/\$\{id\}\/find-menu/)
  assert.match(api, /\/find-menu`, undefined, \{ timeout: 90000 \}/)
  assert.equal((source.match(/findProspectMenu\(props\.prospectId\)/g) || []).length, 1)
  assert.match(source, /Pencarian menu membutuhkan waktu lebih lama dari biasanya/)
  assert.match(source, /MENU_SOURCE_NOT_AVAILABLE/)
})

test('Sales Detail lazily generates a missing summary once and exposes explicit retry', async () => {
  const sales = await salesDetail()
  const summary = await aiSummary()
  const api = await readFile(new URL('../src/api/crm.ts', import.meta.url), 'utf8')
  const client = await readFile(new URL('../src/api/client.ts', import.meta.url), 'utf8')
  assert.match(sales, /summaryGenerationAttempted/)
  assert.match(sales, /!analysisLoadFailed\.value && !hasPersistedSummary\(analysisData\).*ensureSummary\(\)/)
  assert.match(sales, /summaryLoading\.value \|\| hasPersistedSummary/)
  assert.match(api, /\/ai\/prospects\/\$\{id\}\/summary/)
  assert.match(api, /\/summary`, undefined, \{ timeout: 90000 \}/)
  assert.match(client, /timeout: 30000/)
  assert.equal((sales.match(/generateProspectSummary\(/g) || []).length, 1)
  assert.match(sales, /if \(canViewAISummary\.value && !analysisLoadFailed\.value && !hasPersistedSummary\(analysisData\)\) void ensureSummary\(\)/)
  assert.doesNotMatch(sales, /setInterval\([^)]*ensureSummary|watch\([^)]*ensureSummary/)
  assert.match(sales, /AI Summary membutuhkan waktu lebih lama dari biasanya/)
  assert.match(summary, /Membuat ringkasan prospect/)
  assert.match(summary, /AI Summary belum dapat dibuat/)
  assert.match(summary, /\$emit\('retry'\)/)
})

test('Admin and Sales consume the same persisted AI contract without role-triggered menu regeneration', async () => {
  const sales = await salesDetail()
  const admin = await adminDetail()
  for (const source of [sales, admin]) {
    assert.match(source, /getProspectInitialAnalysis\(prospectId\)/)
    assert.match(source, /:analysis="initialAnalysis"/)
    assert.match(source, /!analysisLoadFailed\.value && !hasPersistedSummary\(analysisData\)/)
    assert.match(source, /Data AI prospect belum dapat dimuat/)
    assert.equal((source.match(/generateProspectSummary\(review\.value\.prospect\.id\)/g) || []).length, 1)
    assert.doesNotMatch(source, /findProspectMenu|profileProspectMenu/)
  }
})

test('Tanya AI loads shared prospect history with author metadata and posts only on explicit submit', async () => {
  const source = await tanyaAi()
  assert.match(source, /onMounted\(loadHistory\)/)
  assert.match(source, /getProspectAIChatHistory\(props\.prospectId\)/)
  assert.match(source, /@submit\.prevent="submit"/)
  assert.equal((source.match(/askProspectAI\(props\.prospectId/g) || []).length, 1)
  assert.match(source, /item\.authorName/)
  assert.match(source, /item\.authorRole/)
  assert.match(source, /timestampLabel\(item\.createdAt\)/)
})

test('Sales Prospect Detail removes the legacy Google menu gallery without affecting regular photos or AI menu', async () => {
  const sales = await salesDetail()
  const admin = await adminDetail()

  const galleryPhotos = String.raw`:photos="(?:placeDetails\.photos|\(placeDetails\?\.photos \?\? \[\]\))"`
  assert.match(sales, new RegExp(`${galleryPhotos}[^>]*section="photos"`))
  assert.doesNotMatch(sales, /<PlacePhotoGallery[^>]*section="menu"/)
  assert.doesNotMatch(sales, /class="dcard dcard-menu"/)
  assert.doesNotMatch(sales, /No menu photos yet/)
  assert.match(sales, /<AIMenuProfilingCard/)
  assert.match(sales, /:analysis="initialAnalysis"/)
  assert.match(sales, /:place-details="placeDetails"/)

  // Admin is outside this cleanup and must retain its existing photo behavior.
  assert.match(admin, new RegExp(`${galleryPhotos}[^>]*section="photos"`))
  assert.match(admin, new RegExp(`${galleryPhotos}[^>]*section="menu"`))
})
