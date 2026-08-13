import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const sales = () => readFile(new URL('../src/views/Sales/Customer/CustomerDetailView.vue', import.meta.url), 'utf8')
const admin = () => readFile(new URL('../src/views/Admin/Customer/CustomerDetailView.vue', import.meta.url), 'utf8')

test('Customer Detail reuses canonical Prospect data with a separate customer presentation', async () => {
  for (const source of [await sales(), await admin()]) {
    assert.match(source, /sourceProspectId/)
    assert.match(source, /getProspectInitialAnalysis/)
    assert.match(source, /AISummaryCard/)
    assert.match(source, /CustomerProductOpportunityCard/)
    assert.doesNotMatch(source, /AIMenuProfilingCard/)
    assert.match(source, /TanyaAICard/)
    assert.match(source, /:discovery="storedDiscovery"/)
    assert.match(source, /:profiling="storedProfiling"/)
    assert.doesNotMatch(source, /CustomerAISummary|CustomerAIMenu|CustomerTanyaAI/)
  }
})

test('Customer AI presentation is read-only while Prospect keeps prospecting defaults', async () => {
  const customerMenuUsages = [await sales(), await admin()]
  for (const source of customerMenuUsages) {
    assert.match(source, /CustomerProductOpportunityCard/)
    assert.match(source, /AISummaryCard[\s\S]*?context="customer"/)
    assert.match(source, /sourceProspectId/)
  }

  const menuCard = await readFile(new URL('../src/components/prospect-ai/AIMenuProfilingCard.vue', import.meta.url), 'utf8')
  const summaryCard = await readFile(new URL('../src/components/prospect-ai/AISummaryCard.vue', import.meta.url), 'utf8')
  assert.match(menuCard, /Cari Ulang/)
  assert.match(menuCard, /Analisis Ulang/)
  assert.match(summaryCard, /Customer Insight/)
  assert.match(summaryCard, /Prospect insight/)
})

test('Customer page load performs reads only and never starts AI generation', async () => {
  for (const source of [await sales(), await admin()]) {
    const mounted = source.match(/onMounted\(async \(\) => \{[\s\S]*?\n\}\)/)?.[0] || ''
    assert.match(mounted, /getProspectInitialAnalysis/)
    assert.doesNotMatch(mounted, /generateProspectSummary|findProspectMenu|profileProspectMenu|askProspectAI/)
  }
})

test('Customer without source Prospect has a truthful no-AI fallback', async () => {
  for (const source of [await sales(), await admin()]) {
    assert.match(source, /!sourceProspectId|v-else/)
    assert.match(source, /Riwayat AI tidak tersedia karena customer ini tidak memiliki relasi prospect sumber/)
  }
})

test('Sales Customer layout is sidebar-safe and preserves mobile actions', async () => {
  const source = await sales()
  assert.match(source, /class="customer-hero"/)
  assert.match(source, /class="customer-content-grid"/)
  assert.match(source, /class="customer-main-column"/)
  assert.match(source, /class="customer-intelligence-column"/)
  assert.match(source, /class="dcard dcard-customer-core"/)
  assert.match(source, /Customer Code/)
  assert.match(source, /Parent Code/)
  assert.match(source, /grid-template-columns: minmax\(0, 1fr\) minmax\(300px, 350px\)/)
  assert.match(source, /@media \(min-width: 1200px\)/)
  assert.match(source, /@media \(max-width: 767px\)/)
  assert.match(source, /class="detail-bottom-bar"/)
  assert.match(source, /Navigate/)
  assert.match(source, /Call/)
  assert.match(source, /Check in/)
  assert.doesNotMatch(source, /width:\s*100vw/)
  const intelligence = source.match(/<aside class="customer-intelligence-column">[\s\S]*?<\/aside>/)?.[0] || ''
  const main = source.match(/<main class="customer-main-column">[\s\S]*?<\/main>/)?.[0] || ''
  assert.match(intelligence, /AISummaryCard/)
  assert.match(intelligence, /customer-business-summary/)
  assert.match(intelligence, /ProspectComments/)
  assert.match(intelligence, /TanyaAICard/)
  assert.match(main, /dcard-customer-core/)
  assert.match(main, /Google Maps Info/)
  assert.match(main, /CustomerProductOpportunityCard/)
  assert.match(main, /dcard-conversion/)
  assert.match(main, /Conversion History/)
  assert.match(main, /CustomerSnapshotCard/)
  assert.doesNotMatch(main, /ProspectComments/)
  assert.ok(main.indexOf('<CustomerSnapshotCard') < main.indexOf('<CustomerProductOpportunityCard'))
  assert.ok(main.indexOf('<CustomerProductOpportunityCard') < main.indexOf('Google Maps Info'))
  assert.ok(main.indexOf('Google Maps Info') < main.indexOf('Conversion History'))
  assert.match(source, /height: clamp\(460px, calc\(100dvh - 210px\), 780px\)/)
  assert.match(source, /overflow-y: auto/)
})

test('Customer opportunity expands persisted menus locally and has no prospecting actions', async () => {
  const component = await readFile(new URL('../src/components/customer-ai/CustomerProductOpportunityCard.vue', import.meta.url), 'utf8')
  assert.match(component, /Lihat menu tersimpan/)
  assert.match(component, /profiling\.topOpportunity/)
  assert.match(component, /discovery\.categories/)
  assert.doesNotMatch(component, /findProspectMenu|profileProspectMenu|askProspectAI|axios|fetch\(/)
  assert.doesNotMatch(component, /Cari Menu|Cari Ulang|Analisis Menu|Analisis Ulang/)
  for (const source of [await sales(), await admin()]) {
    assert.doesNotMatch(source, /Cari Menu|Cari Ulang|Analisis Ulang/)
    const template = source.match(/<template>[\s\S]*<\/template>/)?.[0] || ''
    assert.match(template, /<ProspectComments/)
  }
})

test('Customer Snapshot is deterministic CRM state and Customer Insight is historical', async () => {
  const snapshot = await readFile(new URL('../src/components/customer-ai/CustomerSnapshotCard.vue', import.meta.url), 'utf8')
  const summary = await readFile(new URL('../src/components/prospect-ai/AISummaryCard.vue', import.meta.url), 'utf8')
  assert.match(snapshot, /Customer Aktif/)
  assert.match(snapshot, /customer\.salesExecutiveName/)
  assert.match(snapshot, /customer\.customerCode/)
  assert.match(snapshot, /customer\.convertedAt/)
  assert.match(snapshot, /customer\.updatedAt/)
  assert.doesNotMatch(snapshot, /OpenAI|generate|axios|fetch\(/)
  assert.match(summary, /Customer Insight/)
  assert.match(summary, /Insight AI tersimpan dari fase prospect/)
})

test('Prospect still renders the full profiling workflow', async () => {
  const prospect = await readFile(new URL('../src/views/Sales/Prospect/ProspectDetailView.vue', import.meta.url), 'utf8')
  assert.match(prospect, /<AIMenuProfilingCard/)
  assert.doesNotMatch(prospect, /CustomerProductOpportunityCard/)
})

test('Customer Photos remain normal Google photos without legacy MENU gallery', async () => {
  const source = await sales()
  assert.match(source, /PlacePhotoGallery/)
  assert.match(source, /customer\?\.sourceProspectId/)
  assert.doesNotMatch(source, /MENU Google|menuPhotos|photo\.isMenu/)
})
