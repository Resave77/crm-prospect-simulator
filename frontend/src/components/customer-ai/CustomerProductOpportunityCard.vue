<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ProspectMenuFinding, ProspectMenuProfile } from '../../types/crm'

const props = defineProps<{
  discovery?: ProspectMenuFinding | null
  profiling?: ProspectMenuProfile | null
}>()

const showMenu = ref(false)
const menuCount = computed(() => props.discovery?.categories?.reduce((total, category) => total + category.items.length, 0) ?? 0)
const branchRank = { uncertain: 0, brand_only: 1, likely_same_branch: 2, exact_branch: 3 } as const
const branchMatch = computed(() => {
  const matches = [...(props.discovery?.sources?.map(source => source.branchMatch) ?? [])]
  return matches.sort((a, b) => branchRank[b] - branchRank[a])[0] ?? 'uncertain'
})

function level(value: unknown) {
  return ({ LOW: 'RENDAH', MEDIUM: 'SEDANG', HIGH: 'TINGGI', UNKNOWN: 'BELUM DIKETAHUI' } as Record<string, string>)[String(value)] ?? 'BELUM DIKETAHUI'
}
function branchLabel(value: string) {
  return ({ exact_branch: 'Outlet sesuai', likely_same_branch: 'Kemungkinan outlet sama', brand_only: 'Menu tingkat brand', uncertain: 'Belum terverifikasi' } as Record<string, string>)[value] ?? 'Belum terverifikasi'
}
</script>

<template>
  <article class="customer-opportunity-card">
    <header>
      <div><p>Peluang Produk</p><h2>Insight produk tersimpan</h2></div>
      <i class="pi pi-briefcase" />
    </header>

    <template v-if="profiling?.menuOpportunity">
      <div class="metrics">
        <span>Peluang Produk: <strong>{{ level(profiling.menuOpportunity) }}</strong></span>
        <span>Kecocokan Yoghurt: <strong>{{ level(profiling.yoghurtFit) }}</strong></span>
        <span>Keyakinan: <strong>{{ level(profiling.confidence) }}</strong></span>
      </div>
      <dl class="insights">
        <div><dt>Rekomendasi Produk</dt><dd>{{ profiling.topOpportunity || 'Belum tersedia' }}</dd></div>
        <div><dt>Alasan</dt><dd>{{ profiling.why || 'Belum tersedia' }}</dd></div>
        <div><dt>Langkah Sales</dt><dd>{{ profiling.recommendedAction || 'Belum tersedia' }}</dd></div>
      </dl>
    </template>
    <div v-else class="empty"><i class="pi pi-sparkles" /><span>Insight produk belum tersedia dari riwayat prospect.</span></div>

    <section v-if="discovery?.status === 'FOUND'" class="stored-menu">
      <div>
        <p>Menu Tersimpan</p>
        <strong>{{ discovery.categories.length }} kategori · {{ menuCount }} menu · {{ discovery.sources.length }} sumber</strong>
        <span>{{ branchLabel(branchMatch) }}</span>
      </div>
      <button type="button" @click="showMenu = !showMenu">{{ showMenu ? 'Sembunyikan menu' : 'Lihat menu tersimpan' }}</button>
    </section>
    <div v-if="showMenu && discovery?.status === 'FOUND'" class="menu-detail">
      <section v-for="category in discovery.categories" :key="category.name">
        <h3>{{ category.name }}</h3>
        <ul><li v-for="item in category.items" :key="item.name">{{ item.name }}</li></ul>
      </section>
    </div>
  </article>
</template>

<style scoped>
.customer-opportunity-card{display:grid;gap:16px;padding:18px;border:1px solid #e4dadd;border-radius:16px;background:#fff;box-shadow:0 6px 20px rgba(73,34,41,.05);color:#342a2d}.customer-opportunity-card>header{display:flex;align-items:center;justify-content:space-between}.customer-opportunity-card header p,.stored-menu p{margin:0;color:#a51d2e;font-size:11px;font-weight:800;letter-spacing:.08em;text-transform:uppercase}.customer-opportunity-card h2{margin:3px 0 0;font-size:17px}.customer-opportunity-card>header>i{display:grid;place-items:center;width:38px;height:38px;border-radius:11px;background:#fff0f1;color:#d62839}.metrics{display:flex;flex-wrap:wrap;gap:7px}.metrics span{padding:7px 10px;border-radius:999px;background:#f6f1f2;color:#65585c;font-size:11px;font-weight:700;text-transform:uppercase}.metrics strong{color:#a51d2e}.insights{display:grid;margin:0}.insights>div{display:grid;gap:4px;padding:12px 0;border-top:1px solid #eee5e7}.insights dt{color:#74676b;font-size:11px;font-weight:800;text-transform:uppercase}.insights dd{margin:0;font-size:14px;line-height:1.55}.empty{display:flex;align-items:center;gap:9px;padding:12px;border-radius:10px;background:#faf7f8;color:#74676b;font-size:13px}.stored-menu{display:flex;align-items:center;justify-content:space-between;gap:12px;padding-top:14px;border-top:1px solid #e8dfe1}.stored-menu>div{display:grid;gap:3px}.stored-menu strong{font-size:13px}.stored-menu span{color:#218653;font-size:12px;font-weight:700}.stored-menu button{flex:0 0 auto;padding:8px 10px;border:1px solid #dfcdd1;border-radius:9px;background:#fff;color:#a51d2e;font-size:12px;font-weight:800;cursor:pointer}.menu-detail{display:grid;gap:10px;padding:12px;border-radius:11px;background:#faf7f8}.menu-detail section{display:grid;gap:5px}.menu-detail h3{margin:0;font-size:13px}.menu-detail ul{display:flex;flex-wrap:wrap;gap:6px;margin:0;padding:0;list-style:none}.menu-detail li{padding:5px 8px;border-radius:7px;background:#fff;color:#65585c;font-size:12px}@media(max-width:600px){.stored-menu{align-items:flex-start;flex-direction:column}.stored-menu button{width:100%}}
</style>
