<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import {
  listTrashedCategories,
  listTrashedSegments,
  restoreCategory,
  restoreSegment,
  type MasterDataCategory,
  type MasterDataSegment
} from '../../api/masterData'

const emit = defineEmits<{ back: [] }>()

const loading = ref(false)
const error = ref('')
const segments = ref<MasterDataSegment[]>([])
const categories = ref<MasterDataCategory[]>([])
const restoringId = ref('')

function extractError(e: unknown, fallback: string) {
  const anyError = e as { response?: { data?: { error?: { message?: string } } } }
  return anyError?.response?.data?.error?.message || fallback
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [segmentResult, categoryResult] = await Promise.all([
      listTrashedSegments(),
      listTrashedCategories()
    ])
    segments.value = segmentResult
    categories.value = categoryResult
  } catch (e) {
    error.value = extractError(e, 'Failed to load trash.')
  } finally {
    loading.value = false
  }
}

async function restore(type: 'segment' | 'category', item: MasterDataSegment | MasterDataCategory) {
  restoringId.value = `${type}:${item.id}`
  try {
    if (type === 'segment') {
      await restoreSegment(item.id)
      segments.value = segments.value.filter((s) => s.id !== item.id)
    } else {
      await restoreCategory(item.id)
      categories.value = categories.value.filter((c) => c.id !== item.id)
    }
  } catch (e) {
    error.value = extractError(e, 'Failed to restore record.')
  } finally {
    restoringId.value = ''
  }
}

onMounted(load)
</script>

<template>
  <div class="panel-stack">
    <Message v-if="error" severity="error" class="trash-message" @close="error = ''">{{ error }}</Message>

    <!-- TRASH HEADER -->
    <div class="trash-header">
      <div>
        <h2><i class="pi pi-trash" /> Trash</h2>
        <p>Review deleted customer reference data items here and restore them back into the active layout when needed.</p>
      </div>
      <Button label="Back to Master Data" icon="pi pi-arrow-left" severity="secondary" outlined size="small" @click="emit('back')" />
    </div>

    <!-- LOADING -->
    <div v-if="loading" class="state-box">
      <i class="pi pi-spin pi-spinner state-icon" />
      <span>Loading trash...</span>
    </div>

    <div v-else class="trash-grid">
      <!-- ==================== CUSTOMER SITE SEGMENT ==================== -->
      <div class="trash-card">
        <div class="card-head">
          <h3>Customer Site Segment</h3>
          <span class="count-badge">{{ segments.length }} {{ segments.length === 1 ? 'item' : 'items' }}</span>
          <Tag value="Trash" severity="danger" />
        </div>
        <p class="card-desc">Deleted customer site segment items can be restored from here.</p>
        <div v-if="!segments.length" class="empty-box">
          <div class="state-icon-wrap"><i class="pi pi-sitemap" /></div>
          <strong>No deleted customer site segment yet</strong>
          <span>Deleted customer site segment items will appear here and can be restored later.</span>
        </div>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>Segment</th>
              <th>Description</th>
              <th class="th-action">Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="segment in segments" :key="segment.id">
              <td><span class="cell-primary">{{ segment.name }}</span></td>
              <td><span class="cell-text muted-cell">{{ segment.description || '—' }}</span></td>
              <td class="td-action">
                <Button
                  label="Restore"
                  icon="pi pi-replay"
                  severity="success"
                  text
                  size="small"
                  :loading="restoringId === `segment:${segment.id}`"
                  @click="restore('segment', segment)"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- ==================== CUSTOMER SITE CATEGORY ==================== -->
      <div class="trash-card">
        <div class="card-head">
          <h3>Customer Site Category</h3>
          <span class="count-badge">{{ categories.length }} {{ categories.length === 1 ? 'item' : 'items' }}</span>
          <Tag value="Trash" severity="danger" />
        </div>
        <p class="card-desc">Deleted customer site category items can be restored from here.</p>
        <div v-if="!categories.length" class="empty-box">
          <div class="state-icon-wrap"><i class="pi pi-tags" /></div>
          <strong>No deleted customer site category yet</strong>
          <span>Deleted customer site category items will appear here and can be restored later.</span>
        </div>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>Category</th>
              <th>Description</th>
              <th class="th-action">Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="category in categories" :key="category.id">
              <td><span class="cell-primary">{{ category.name }}</span></td>
              <td><span class="cell-text muted-cell">{{ category.description || '—' }}</span></td>
              <td class="td-action">
                <Button
                  label="Restore"
                  icon="pi pi-replay"
                  severity="success"
                  text
                  size="small"
                  :loading="restoringId === `category:${category.id}`"
                  @click="restore('category', category)"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.panel-stack{display:flex;min-width:0;flex-direction:column;gap:.75rem}
.trash-message{margin:0}
.trash-header{display:flex;flex-wrap:wrap;align-items:center;justify-content:space-between;gap:.7rem;padding:.9rem;border:1px solid #e5eaf0;border-radius:12px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.03)}
.trash-header h2{margin:0;color:#0f172a;font-size:.95rem;letter-spacing:-.01em}
.trash-header h2 i{margin-right:.35rem;color:#ef4444;font-size:.8rem}
.trash-header p{margin:.1rem 0 0;color:#94a3b8;font-size:.68rem}
.trash-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:.75rem}
.trash-card{display:flex;flex-direction:column;gap:.55rem;padding:.9rem;border:1px solid #e5eaf0;border-radius:12px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.03)}
.card-head{display:flex;align-items:center;gap:.5rem}
.card-head h3{margin:0;color:#0f172a;font-size:.82rem;letter-spacing:-.01em}
.count-badge{padding:.14rem .5rem;border-radius:999px;background:#f1f5f9;color:#475569;font-size:.6rem;font-weight:700}
.card-head .p-tag{margin-left:auto}
.card-desc{margin:0;color:#94a3b8;font-size:.66rem}
.empty-box{display:flex;flex-direction:column;align-items:center;gap:.3rem;padding:2.2rem 1rem;color:#94a3b8;font-size:.72rem;text-align:center}
.state-icon-wrap{font-size:1.3rem;color:#cbd5e1}
.empty-box strong{color:#475569;font-size:.78rem}
.data-table{width:100%;border-collapse:collapse;font-size:.72rem;border:1px solid #e5eaf0;border-radius:10px}
.data-table th{padding:.55rem .75rem;border-bottom:1px solid #e5eaf0;background:#f8fafc;color:#64748b;font-size:.58rem;font-weight:800;text-align:left;text-transform:uppercase;letter-spacing:.06em;white-space:nowrap}
.data-table td{padding:.55rem .75rem;border-bottom:1px solid #f1f5f9;color:#334155;vertical-align:middle}
.data-table tbody tr:last-child td{border-bottom:0}
.data-table tbody tr:hover td{background:#fafbfd}
.cell-primary{color:#0f172a;font-weight:600}
.cell-text{color:#475569}
.muted-cell{color:#94a3b8}
.th-action,.td-action{text-align:right;white-space:nowrap}
.state-box{display:flex;flex-direction:column;align-items:center;gap:.3rem;padding:2.2rem 1rem;color:#94a3b8;font-size:.72rem;text-align:center}
.state-box .state-icon{font-size:1.3rem;color:#cbd5e1}
@media (max-width: 720px){
  .trash-header{flex-direction:column;align-items:stretch}
}
</style>
