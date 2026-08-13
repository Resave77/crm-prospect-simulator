<script setup lang="ts">
import { computed, ref } from 'vue'
import { findProspectMenu, profileProspectMenu } from '../../api/crm'
import type { PlaceDetails, ProspectInitialAnalysis, ProspectMenuDocument, ProspectMenuFinding, ProspectMenuFindingItem, ProspectMenuProfile } from '../../types/crm'

const props = withDefaults(defineProps<{
  prospectId: string
  placeDetails: PlaceDetails | null
  analysis?: ProspectInitialAnalysis | null
  mode?: 'prospect' | 'customer'
}>(), { mode: 'prospect' })
const customerMode = computed(() => props.mode === 'customer')
const loading = ref(false)
const finding = ref(false)
const error = ref('')
const result = ref<ProspectMenuProfile | null>(null)
const foundResult = ref<ProspectMenuFinding | null>(null)
const showMenu = ref(false)
const expandedCategories = ref<Record<string, boolean>>({})
const previewLimit = 5

const storedMenu = computed(() => (props.analysis?.menu ?? null) as (ProspectMenuProfile & ProspectMenuDocument) | null)
const menuFinding = computed(() => foundResult.value ?? storedMenu.value?.discovery ?? storedMenu.value?.finding ?? null)
const storedProfile = computed(() => storedMenu.value?.profiling ?? storedMenu.value?.profile ?? (storedMenu.value?.menus || storedMenu.value?.state ? storedMenu.value as ProspectMenuProfile : null))
const menuResult = computed(() => result.value ?? storedProfile.value ?? null)
const aggregateProfile = computed(() => menuResult.value?.menuOpportunity ? menuResult.value : null)
const menuItemCount = computed(() => menuFinding.value?.categories.reduce((total, category) => total + category.items.length, 0) ?? 0)
const finderUnavailable = computed(() => menuFinding.value?.status === 'NOT_FOUND' || menuFinding.value?.status === 'MENU_SOURCE_NOT_AVAILABLE')
const branchRank = { uncertain: 0, brand_only: 1, likely_same_branch: 2, exact_branch: 3 } as const
const primaryBranchMatch = computed(() => {
  const matches = menuFinding.value?.sources.map(source => source.branchMatch) ?? []
  return matches.sort((a, b) => branchRank[b] - branchRank[a])[0] ?? 'uncertain'
})
const menuState = computed(() => String(menuResult.value?.state ?? ''))
const menuUnavailable = computed(() => menuState.value === 'MENU_DATA_NOT_AVAILABLE')

type MenuProfileRow = { menuName?: unknown; menu?: unknown; profile?: unknown; yoghurtFit?: unknown; opportunity?: unknown }
const menuRows = computed(() => Array.isArray(menuResult.value?.menus) ? (menuResult.value.menus as MenuProfileRow[]).slice(0, 6) : [])

function menuRowLabel(row: MenuProfileRow) { return String(row.menuName || row.menu || 'Data belum tersedia') }
function localizedLevel(value: unknown) { return ({ LOW: 'RENDAH', MEDIUM: 'SEDANG', HIGH: 'TINGGI', UNKNOWN: 'BELUM DIKETAHUI' } as Record<string, string>)[String(value)] ?? 'BELUM DIKETAHUI' }
function levelDots(value: unknown) { return ({ LOW: 1, MEDIUM: 2, HIGH: 3 } as Record<string, number>)[String(value)] ?? 0 }
function branchLabel(value: string) { return ({ exact_branch: 'Outlet sesuai', likely_same_branch: 'Kemungkinan outlet sama', brand_only: 'Menu tingkat brand', uncertain: 'Belum terverifikasi' } as Record<string, string>)[value] ?? 'Belum terverifikasi' }
function formatPrice(item: ProspectMenuFindingItem) {
  const price = item.prices.find(entry => Number.isFinite(entry.price) && entry.price > 0)?.price
  return price == null ? 'Harga tidak tersedia' : new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(price).replace(/\s/g, '')
}
function visibleItems(categoryName: string, items: ProspectMenuFindingItem[]) { return expandedCategories.value[categoryName] ? items : items.slice(0, previewLimit) }
function toggleCategory(categoryName: string) { expandedCategories.value = { ...expandedCategories.value, [categoryName]: !expandedCategories.value[categoryName] } }
function confidencePercent(value: number) { return `${Math.round(Math.max(0, Math.min(1, value)) * 100)}% cocok` }
function sourceLabel(name: string, url: string) { try { return name || new URL(url).hostname } catch { return name || 'Sumber menu' } }
function finderError(caught: any) {
  if (caught?.code === 'ECONNABORTED' || /timeout/i.test(caught?.message || '')) return 'Pencarian menu membutuhkan waktu lebih lama dari biasanya.'
  const code = caught?.response?.data?.error?.code
  if (code === 'AI_RATE_LIMITED' || code === 'AI_UNAVAILABLE') return 'Layanan pencarian menu sedang sibuk. Silakan coba lagi.'
  return 'Pencarian menu belum dapat diselesaikan.'
}
async function analyzeMenu() {
  if (loading.value || !props.prospectId) return
  loading.value = true; error.value = ''
  try { result.value = await profileProspectMenu(props.prospectId, Boolean(menuResult.value)) }
  catch { error.value = 'Analisis menu belum dapat diselesaikan.' }
  finally { loading.value = false }
}
async function findMenu() {
  if (finding.value || !props.prospectId) return
  finding.value = true; error.value = ''
  try { foundResult.value = await findProspectMenu(props.prospectId) }
  catch (caught: any) { error.value = finderError(caught) }
  finally { finding.value = false }
}
</script>

<template>
  <article class="ai-menu-card">
    <header class="ai-menu-head">
      <div><p class="ai-eyebrow"><i class="pi pi-chart-line" /> {{ customerMode ? 'Peluang Produk' : 'Analisis Menu AI' }}</p><h2>{{ customerMode ? 'Customer product opportunity' : 'Analisis peluang menu' }}</h2></div>
      <button v-if="!customerMode" class="ai-primary-btn" type="button" :disabled="finding" @click="findMenu"><i :class="finding ? 'pi pi-spin pi-spinner' : 'pi pi-search'" /> {{ finding ? 'Mencari...' : menuFinding ? 'Cari Ulang' : 'Cari Menu' }}</button>
      <button v-else-if="menuFinding?.status === 'FOUND'" class="ai-secondary-btn" type="button" @click="showMenu = !showMenu">{{ showMenu ? 'Sembunyikan Menu' : 'Lihat Menu' }}</button>
    </header>

    <div v-if="finding && !customerMode" class="state-card loading-state"><i class="pi pi-spin pi-spinner" /><div><strong>Mencari menu dari sumber online...</strong><span>Data yang sudah tersimpan tidak akan dihapus jika pencarian gagal.</span></div></div>

    <section v-if="menuFinding?.status === 'FOUND' && !customerMode" class="discovery-summary">
      <div class="summary-header">
        <div class="summary-title"><i class="pi pi-check-circle" /><strong>Menu ditemukan</strong></div>
        <div class="ai-menu-actions"><button class="ai-secondary-btn" type="button" @click="showMenu = !showMenu">{{ showMenu ? 'Sembunyikan Menu' : 'Lihat Menu' }}</button><button class="ai-secondary-btn accent" type="button" :disabled="loading" @click="analyzeMenu">{{ loading ? 'Menganalisis...' : menuResult ? 'Analisis Ulang' : 'Analisis Menu' }}</button></div>
      </div>
      <div class="menu-stats" aria-label="Ringkasan menu ditemukan">
        <div class="menu-stat"><strong>{{ menuFinding.categories.length }}</strong><span>Kategori</span></div>
        <div class="menu-stat"><strong>{{ menuItemCount }}</strong><span>Menu representatif</span></div>
        <div class="menu-stat"><strong>{{ menuFinding.sources.length }}</strong><span>Sumber</span></div>
      </div>
      <div class="branch-summary"><i class="pi pi-map-marker" /><span>Kecocokan outlet</span><strong>{{ branchLabel(primaryBranchMatch) }}</strong></div>
    </section>

    <div v-else-if="finderUnavailable && !customerMode" class="state-card"><i class="pi pi-search" /><div><strong>Menu belum ditemukan.</strong><span>Belum ditemukan sumber menu yang dapat diverifikasi.</span></div></div>
    <div v-else-if="!menuFinding && !finding && !customerMode" class="state-card"><i class="pi pi-book" /><div><strong>Menu belum dicari.</strong><span>Cari menu dari sumber online untuk melihat peluang produk.</span></div></div>
    <div v-if="customerMode && !aggregateProfile" class="state-card"><i class="pi pi-sparkles" /><div><strong>Insight produk belum tersedia.</strong><span>Insight produk belum tersedia dari riwayat prospect.</span></div></div>
    <div v-if="error && !customerMode" class="state-card error-state"><i class="pi pi-exclamation-triangle" /><div><strong>{{ error }}</strong><span>Gunakan tombol Cari Menu untuk mencoba kembali.</span></div></div>

    <div v-if="loading" class="state-card loading-state"><i class="pi pi-spin pi-spinner" /><div><strong>Menganalisis peluang menu...</strong><span>Menu yang ditemukan tetap dapat dilihat selama analisis.</span></div></div>

    <section v-if="aggregateProfile" class="profile-dashboard">
      <div class="metric-grid">
        <div v-for="metric in [{ label: 'Peluang Menu', value: aggregateProfile.menuOpportunity }, { label: 'Kecocokan Yoghurt', value: aggregateProfile.yoghurtFit }, { label: 'Tingkat Keyakinan', value: aggregateProfile.confidence }]" :key="metric.label" class="metric-card">
          <span>{{ metric.label }}</span><strong>{{ localizedLevel(metric.value) }}</strong><div class="level-dots" :aria-label="`${levelDots(metric.value)} dari 3`"><i v-for="dot in 3" :key="dot" :class="{ active: dot <= levelDots(metric.value) }" /></div>
        </div>
      </div>
      <section class="insight-card opportunity"><div class="insight-title"><i class="pi pi-star-fill" /> Peluang Utama</div><strong>{{ aggregateProfile.topOpportunity }}</strong></section>
      <section class="insight-card"><div class="insight-title"><i class="pi pi-lightbulb" /> Mengapa ini menarik?</div><p>{{ aggregateProfile.why }}</p></section>
      <section class="insight-card next-action"><div class="insight-title"><i class="pi pi-check-circle" /> Langkah Sales Berikutnya</div><p>{{ aggregateProfile.recommendedAction }}</p></section>
    </section>
    <div v-else-if="menuFinding?.status === 'FOUND' && !loading && !customerMode" class="profiling-empty"><i class="pi pi-sparkles" /><div><strong>Analisis menu belum dibuat.</strong><span>Jalankan analisis setelah data menu tersedia.</span></div></div>
    <div v-if="menuUnavailable && !customerMode" class="state-card"><i class="pi pi-info-circle" /><div><strong>Analisis menu belum dibuat.</strong><span>Jalankan analisis setelah data menu tersedia.</span></div></div>

    <div v-if="showMenu && menuFinding?.status === 'FOUND'" class="menu-discovery">
      <div class="category-grid">
        <section v-for="category in menuFinding.categories" :key="category.name" class="category-card">
          <header><div><i class="pi pi-list" /><strong>{{ category.name }}</strong></div><span>{{ category.items.length }} menu</span></header>
          <div class="item-list">
            <article v-for="item in visibleItems(category.name, category.items)" :key="`${category.name}-${item.name}`" class="menu-item">
              <div class="item-main"><strong>{{ item.name }}</strong><span class="price" :class="{ unavailable: !item.prices.length }">{{ formatPrice(item) }}</span></div>
              <div class="item-meta"><span class="branch-chip">{{ branchLabel(item.branchMatch) }}</span><span class="confidence-chip">{{ confidencePercent(item.confidence) }}</span></div>
            </article>
          </div>
          <button v-if="category.items.length > previewLimit" class="category-toggle" type="button" @click="toggleCategory(category.name)">{{ expandedCategories[category.name] ? 'Sembunyikan' : `Lihat ${category.items.length - previewLimit} menu lainnya` }} <i :class="expandedCategories[category.name] ? 'pi pi-chevron-up' : 'pi pi-arrow-right'" /></button>
        </section>
      </div>
      <section v-if="menuFinding.sources.length" class="source-section"><h3>Sumber Menu</h3><div><a v-for="source in menuFinding.sources" :key="source.url" :href="source.url" target="_blank" rel="noopener noreferrer"><i class="pi pi-external-link" /> {{ sourceLabel(source.name, source.url) }}</a></div></section>
    </div>

    <div v-if="menuRows.length" class="legacy-menu-table" role="table" aria-label="Pratinjau analisis menu AI"><div v-for="row in menuRows" :key="menuRowLabel(row)"><strong>{{ menuRowLabel(row) }}</strong><span>{{ String(row.profile || 'Data belum tersedia') }}</span></div></div>
  </article>
</template>

<style scoped>
.ai-menu-card{display:grid;gap:14px;min-width:0;padding:18px;border:1px solid #eadde0;border-radius:16px;background:#fff;box-shadow:0 8px 24px rgba(73,34,41,.06);color:#342a2d}.ai-menu-head{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}.ai-menu-head h2{margin:3px 0 0;font-size:18px;line-height:1.3}.ai-eyebrow{margin:0;display:flex;align-items:center;gap:6px;color:#d62839;font-size:12px;font-weight:800;text-transform:uppercase;letter-spacing:.06em}.ai-primary-btn,.ai-secondary-btn{min-height:42px;display:inline-flex;align-items:center;justify-content:center;gap:7px;padding:0 14px;border-radius:11px;font-size:13px;font-weight:800;cursor:pointer}.ai-primary-btn{border:0;background:#d62839;color:#fff}.ai-primary-btn:disabled,.ai-secondary-btn:disabled{cursor:not-allowed;opacity:.6}.state-card,.profiling-empty{display:flex;gap:12px;padding:14px;border:1px dashed #dfd0d4;border-radius:12px;background:#fcf9f9}.state-card>i,.profiling-empty>i{width:36px;height:36px;display:grid;place-items:center;flex:0 0 auto;border-radius:10px;background:#fff0f1;color:#d62839}.state-card div,.profiling-empty div{display:grid;gap:3px}.state-card strong,.profiling-empty strong{font-size:14px}.state-card span,.profiling-empty span{color:#6f6266;font-size:13px;line-height:1.5}.error-state{border-color:#efb5bc;background:#fff6f7}.loading-state{border-style:solid}.discovery-summary{display:grid;grid-template-columns:auto 1fr auto;align-items:center;gap:12px;padding:14px;border:1px solid #eccdd2;border-radius:14px;background:linear-gradient(135deg,#fff7f8,#fff)}.summary-title{display:flex;align-items:center;gap:7px;color:#218653;text-transform:uppercase;font-size:13px}.summary-stats{display:flex;flex-wrap:wrap;gap:8px}.summary-stats span{padding:7px 10px;border-radius:9px;background:#fff;border:1px solid #eadde0;color:#65585c;font-size:12px}.summary-stats b{color:#342a2d;font-size:14px}.branch-summary{grid-column:1/3;display:grid;grid-template-columns:auto auto 1fr;align-items:center;gap:7px;font-size:13px}.branch-summary i{color:#d62839}.branch-summary span{color:#74676b}.ai-menu-actions{grid-row:1/3;grid-column:3;display:flex;gap:8px}.ai-secondary-btn{border:1px solid #e3cfd3;background:#fff;color:#b82032}.ai-secondary-btn.accent{background:#fff0f1}.profile-dashboard{display:grid;gap:12px}.metric-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:10px}.metric-card{display:grid;justify-items:center;gap:6px;padding:14px 8px;border:1px solid #eadde0;border-radius:13px;background:#fcf9f9;text-align:center}.metric-card>span{color:#6c5f63;font-size:12px;font-weight:800;text-transform:uppercase}.metric-card>strong{color:#b82032;font-size:16px}.level-dots{display:flex;gap:5px}.level-dots i{width:9px;height:9px;border-radius:50%;background:#ded5d7}.level-dots i.active{background:#d62839}.insight-card{display:grid;gap:7px;padding:16px;border:1px solid #e5d9dc;border-radius:13px;background:#fff}.insight-card.opportunity{border-left:4px solid #d62839;background:#fff8f9}.insight-card.next-action{border-left:4px solid #218653;background:#f5fbf7}.insight-title{display:flex;align-items:center;gap:7px;color:#8e1e2c;font-size:13px;font-weight:800;text-transform:uppercase}.next-action .insight-title{color:#176a40}.insight-card strong,.insight-card p{max-width:76ch;margin:0;color:#342a2d;font-size:14px;line-height:1.6}.menu-discovery{display:grid;gap:14px}.category-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px}.category-card{min-width:0;overflow:hidden;border:1px solid #e6dade;border-radius:14px;background:#fff}.category-card>header{display:flex;align-items:center;justify-content:space-between;gap:10px;padding:13px 14px;border-bottom:1px solid #eee4e6;background:#fcf9f9}.category-card>header div{display:flex;align-items:center;gap:8px;min-width:0}.category-card>header i{color:#d62839}.category-card>header strong{font-size:14px;overflow-wrap:anywhere}.category-card>header span{flex:0 0 auto;color:#6c5f63;font-size:12px;font-weight:700}.item-list{padding:0 14px}.menu-item{display:flex;align-items:flex-start;justify-content:space-between;gap:12px;padding:12px 0;border-bottom:1px solid #f0e8ea}.menu-item:last-child{border-bottom:0}.item-main{display:grid;gap:4px;min-width:0}.item-main>strong{font-size:15px;line-height:1.35;overflow-wrap:anywhere}.price{color:#b82032;font-size:14px;font-weight:800}.price.unavailable{color:#74676b;font-size:12px;font-weight:600}.item-meta{display:flex;justify-content:flex-end;flex-wrap:wrap;gap:5px;max-width:46%}.branch-chip,.confidence-chip{padding:4px 7px;border-radius:999px;font-size:12px;font-weight:700;line-height:1.2}.branch-chip{background:#f3edef;color:#66585c}.confidence-chip{background:#eaf7ef;color:#176a40}.category-toggle{width:100%;min-height:42px;padding:8px 14px;border:0;border-top:1px solid #eee4e6;background:#fff9fa;color:#b82032;font-size:13px;font-weight:800;cursor:pointer}.source-section{display:grid;gap:9px;padding:14px;border:1px solid #e6dade;border-radius:13px;background:#fcf9f9}.source-section h3{margin:0;font-size:13px;text-transform:uppercase}.source-section>div{display:flex;flex-wrap:wrap;gap:8px}.source-section a{display:inline-flex;align-items:center;gap:5px;padding:7px 9px;border:1px solid #e5d7da;border-radius:9px;background:#fff;color:#a51d2e;font-size:12px;font-weight:700;text-decoration:none}.legacy-menu-table{display:grid;gap:8px}.legacy-menu-table>div{display:grid;gap:4px;padding:12px;border:1px solid #eadde0;border-radius:10px}.legacy-menu-table strong{font-size:14px}.legacy-menu-table span{color:#65585c;font-size:13px;line-height:1.5}
@media(max-width:767px){.ai-menu-card{padding:14px}.ai-menu-head{align-items:stretch;flex-direction:column}.ai-primary-btn{width:100%}.discovery-summary{grid-template-columns:1fr}.summary-title,.summary-stats,.branch-summary,.ai-menu-actions{grid-column:1;grid-row:auto}.branch-summary{grid-template-columns:auto auto 1fr}.ai-menu-actions{display:grid;grid-template-columns:1fr 1fr}.metric-grid{grid-template-columns:repeat(3,minmax(0,1fr));gap:6px}.metric-card{padding:11px 4px}.metric-card>span{font-size:10px;line-height:1.25}.metric-card>strong{font-size:12px;overflow-wrap:anywhere}.category-grid{grid-template-columns:1fr}.menu-item{flex-direction:column}.item-meta{justify-content:flex-start;max-width:100%}.source-section a{max-width:100%;overflow-wrap:anywhere}.ai-secondary-btn{padding:0 8px}}
@media(max-width:390px){.metric-grid{grid-template-columns:1fr}.metric-card{grid-template-columns:1fr auto auto;justify-items:start;align-items:center;text-align:left}.branch-summary{grid-template-columns:auto 1fr}.branch-summary strong{grid-column:2}.ai-menu-actions{grid-template-columns:1fr}}

/* Find Menu summary: container-responsive CRM statistics panel. */
.discovery-summary {
  container-type: inline-size;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  align-items: stretch;
  gap: 12px;
  min-width: 0;
  padding: 14px;
  box-sizing: border-box;
}
.summary-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 10px 14px;
  min-width: 0;
}
.summary-title { min-width: 0; }
.menu-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  min-width: 0;
  width: 100%;
}
.menu-stat {
  display: grid;
  align-content: center;
  justify-items: center;
  gap: 2px;
  min-width: 0;
  width: 100%;
  min-height: 62px;
  padding: 9px 8px;
  box-sizing: border-box;
  border: 1px solid #eadde0;
  border-radius: 10px;
  background: #fff;
  text-align: center;
}
.menu-stat strong {
  color: #342a2d;
  font-size: 20px;
  font-weight: 800;
  line-height: 1.1;
}
.menu-stat span {
  min-width: 0;
  color: #6c5f63;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.25;
  overflow-wrap: anywhere;
}
.branch-summary {
  display: flex;
  grid-column: auto;
  align-items: center;
  flex-wrap: wrap;
  gap: 7px 9px;
  min-width: 0;
  padding-top: 1px;
}
.branch-summary strong {
  padding: 5px 9px;
  border-radius: 999px;
  background: #eaf7ef;
  color: #176a40;
  font-size: 12px;
  line-height: 1.2;
}
.ai-menu-actions {
  display: flex;
  grid-column: auto;
  grid-row: auto;
  flex-wrap: wrap;
  gap: 8px;
  min-width: 0;
}

@container (max-width: 520px) {
  .summary-header { align-items: stretch; }
  .summary-title { width: 100%; }
  .ai-menu-actions { width: 100%; }
  .ai-menu-actions .ai-secondary-btn { flex: 1 1 130px; }
}

@container (max-width: 360px) {
  .menu-stats { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .menu-stat:last-child { grid-column: 1 / -1; }
}

@media (max-width: 767px) {
  .menu-stats { grid-template-columns: minmax(0, 1fr); }
  .menu-stat:last-child { grid-column: auto; }
  .ai-menu-actions { grid-template-columns: none; }
}
</style>
