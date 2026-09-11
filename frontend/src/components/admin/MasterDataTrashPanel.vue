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

/* White Trash surface: keep the section clean and neutral, with red reserved
   for the trash icon/status accent. */
.panel-stack{
  padding:1rem;
  border:1px solid #e5eaf0;
  border-radius:16px;
  background:#fff;
}
.trash-header,
.trash-card,
.trash-card .data-table{background:#fff}
.trash-card .data-table th{background:#fff}
.trash-card .card-head :deep(.p-tag){
  border:1px solid #fecaca;
  background:#fff;
  color:#dc2626;
}
.trash-card .count-badge{border:1px solid #e2e8f0;background:#fff;color:#64748b}

/* Reference Trash layout: equal cards, inset table shell, and quiet neutral surface. */
.panel-stack{gap:.875rem;padding:0;border:0;border-radius:0;background:transparent;box-shadow:none}
.trash-header{min-height:76px;padding:.875rem 1rem;border:1px solid #e2e8f0;border-radius:18px;background:#fff;box-shadow:0 10px 24px rgba(15,23,42,.05)}
.trash-header h2{font-size:1rem}
.trash-header p{max-width:860px;margin-top:.35rem;color:#64748b;font-size:.72rem;line-height:1.5}
.trash-header :deep(.p-button){min-width:8.9rem;justify-content:center;border-color:#d9e2ec;color:#475569}
.trash-grid{grid-template-columns:repeat(2,minmax(0,1fr));gap:1.1rem;align-items:stretch}
.trash-card{min-width:0;min-height:368px;gap:0;padding:0;overflow:hidden;border:1px solid #e2e8f0;border-radius:20px;background:#fff;box-shadow:0 14px 34px rgba(15,23,42,.06)}
.trash-card .card-head{min-height:2.4rem;align-items:center;gap:.5rem;padding:1rem 1.1rem 0}
.trash-card .card-head{min-width:0;flex-wrap:wrap}
.trash-card .card-head h3{min-width:0;flex:1 1 auto;font-size:1rem;font-weight:700;letter-spacing:-.02em;line-height:1.3;overflow-wrap:break-word;word-break:normal}
.trash-card .count-badge{min-height:1.5rem;padding:.2rem .6rem;border-color:#e2e8f0;background:#f8fafc;font-size:.62rem;font-weight:600}
.trash-card .card-head .p-tag{margin-left:auto;flex:none}
.trash-card .card-desc{min-height:2.8rem;margin:0;padding:.35rem 1.1rem 1rem;border-bottom:1px solid #eef2f7;color:#64748b;font-size:.72rem;line-height:1.5;overflow-wrap:break-word;word-break:normal}
.trash-card .data-table{width:calc(100% - 2.2rem);min-width:0;margin:.85rem 1.1rem 1.1rem;border:0;border-collapse:separate;border-spacing:0 .38rem;table-layout:fixed;background:transparent}
.trash-card .data-table th{height:38px;padding:.5rem .75rem;border-top:1px solid #d9e2ec;border-bottom:1px solid #d9e2ec;background:#f8fafc;font-size:.58rem;line-height:1.25;white-space:normal;overflow-wrap:anywhere}
.trash-card .data-table th:first-child{border-left:1px solid #d9e2ec;border-radius:12px 0 0 12px}
.trash-card .data-table th:last-child{border-right:1px solid #d9e2ec;border-radius:0 12px 12px 0}
.trash-card .data-table td{height:54px;min-width:0;padding:.62rem .75rem;border-top:1px solid #e2e8f0;border-bottom:1px solid #e2e8f0;background:#fff;vertical-align:middle;overflow-wrap:break-word;word-break:normal;line-height:1.4;transition:background .15s ease,border-color .15s ease}
.trash-card .data-table td:first-child{border-left:1px solid #e2e8f0;border-radius:12px 0 0 12px}
.trash-card .data-table td:last-child{border-right:1px solid #e2e8f0;border-radius:0 12px 12px 0}
.trash-card .data-table tbody tr:hover td{border-color:#d7e8dd;background:#fbfefc}
.trash-card .data-table td:first-child .cell-primary{font-weight:700}
.trash-card .data-table th:nth-child(1),.trash-card .data-table td:nth-child(1){width:29%}
.trash-card .data-table th:nth-child(2),.trash-card .data-table td:nth-child(2){width:46%}
.trash-card .data-table th:nth-child(3),.trash-card .data-table td:nth-child(3){width:25%}
.trash-card .data-table .cell-primary,.trash-card .data-table .cell-text{display:block;min-width:0;max-width:100%;line-height:1.4;overflow-wrap:break-word;word-break:normal;white-space:normal}
.trash-card .data-table .td-action{white-space:nowrap;text-align:right}
.trash-card .data-table .td-action :deep(.p-button){min-width:5.4rem;height:2rem;justify-content:center;padding:.4rem .7rem;border:1px solid #bbf7d0;border-radius:999px;background:#f0fdf4;color:#166534;font-size:.7rem;font-weight:600;box-shadow:none}
.trash-card .data-table .td-action :deep(.p-button:hover){background:#ecfdf5;color:#166534}
.trash-card .empty-box{min-height:260px;margin:.75rem 1rem 1rem;padding:2rem 1rem;border:1px dashed #d9e2ec;border-radius:20px;background:#f8fafc}
.trash-card .empty-box .state-icon-wrap{background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.08)}
.trash-card .empty-box span{max-width:420px;line-height:1.5}

@media (max-width: 760px){
  .trash-grid{grid-template-columns:minmax(0,1fr)}
  .trash-card .data-table{width:calc(100% - 1.5rem);margin-right:.75rem;margin-left:.75rem}
  .trash-card .data-table td,.trash-card .data-table th{padding-right:.55rem;padding-left:.55rem}
}
</style>
