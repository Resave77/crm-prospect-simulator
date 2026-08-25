<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import {
  createCategory,
  createSegment,
  deleteCategory,
  deleteSegment,
  listCategories,
  listSegments,
  updateCategory,
  updateSegment,
  type MasterDataCategory,
  type MasterDataSegment
} from '../../api/masterData'
import MasterDataTrashPanel from './MasterDataTrashPanel.vue'
import { categoryIcon } from '../../utils/categoryIcons'

type MasterTab = 'segment' | 'category'
type MasterView = 'main' | 'trash'

const activeTab = ref<MasterTab>('segment')
const view = ref<MasterView>('main')
const loading = ref(false)
const error = ref('')

const segments = ref<MasterDataSegment[]>([])
const categories = ref<MasterDataCategory[]>([])

const segmentSearch = ref('')
const segmentStatusFilter = ref('')
const categorySearch = ref('')
const categorySegmentFilter = ref('')
const categoryStatusFilter = ref('')

const statusOptions = [
  { label: 'All Status', value: '' },
  { label: 'Active', value: 'ACTIVE' },
  { label: 'Inactive', value: 'INACTIVE' }
]

let searchTimeout: ReturnType<typeof setTimeout> | null = null

// ---------- Segment dialog ----------
const segmentDialogVisible = ref(false)
const segmentSaving = ref(false)
const segmentEditingId = ref<string | null>(null)
const segmentForm = ref({ name: '', description: '', status: 'ACTIVE' as 'ACTIVE' | 'INACTIVE' })
const segmentSubmitted = ref(false)

// ---------- Category dialog ----------
const categoryDialogVisible = ref(false)
const categorySaving = ref(false)
const categoryEditingId = ref<string | null>(null)
const categoryForm = ref({
  segmentId: '',
  name: '',
  description: '',
  placeApi: '',
  status: 'ACTIVE' as 'ACTIVE' | 'INACTIVE'
})
const categorySubmitted = ref(false)

// ---------- Delete dialogs ----------
const deleteDialogVisible = ref(false)
const deleting = ref(false)
const deleteTarget = ref<{ type: MasterTab; id: string; name: string; extra?: string } | null>(null)

const segmentOptions = computed(() =>
  segments.value.map((s) => ({ label: s.name, value: s.id }))
)

const activeSegmentCount = computed(() => segments.value.filter((s) => s.status === 'ACTIVE').length)
const activeCategoryCount = computed(() => categories.value.filter((c) => c.status === 'ACTIVE').length)

function statusSeverity(status: string) {
  return status === 'ACTIVE' ? 'success' : 'secondary'
}

function statusLabel(status: string) {
  return status === 'ACTIVE' ? 'Active' : 'Inactive'
}

function segmentAccent(name: string) {
  const key = name.trim().toLowerCase()
  if (key === 'b2b') return 'seg-b2b'
  if (key === 'b2c') return 'seg-b2c'
  return ''
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [segmentResult, categoryResult] = await Promise.all([
      listSegments({ search: segmentSearch.value.trim(), status: segmentStatusFilter.value }),
      listCategories({
        search: categorySearch.value.trim(),
        segmentId: categorySegmentFilter.value,
        status: categoryStatusFilter.value
      })
    ])
    segments.value = segmentResult
    categories.value = categoryResult
  } catch (e) {
    error.value = extractError(e, 'Failed to load master data.')
  } finally {
    loading.value = false
  }
}

function extractError(e: unknown, fallback: string) {
  const anyError = e as { response?: { data?: { error?: { message?: string } } } }
  return anyError?.response?.data?.error?.message || fallback
}

function onSegmentSearch(value: string) {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    load().catch(() => {})
  }, 350)
  void value
}

function selectTab(tab: MasterTab) {
  activeTab.value = tab
}

function backFromTrash() {
  view.value = 'main'
  load().catch(() => {})
}

// ---------- Segment CRUD ----------
function openAddSegment() {
  segmentEditingId.value = null
  segmentForm.value = { name: '', description: '', status: 'ACTIVE' }
  segmentSubmitted.value = false
  segmentDialogVisible.value = true
}

function openEditSegment(segment: MasterDataSegment) {
  segmentEditingId.value = segment.id
  segmentForm.value = { name: segment.name, description: segment.description, status: segment.status }
  segmentSubmitted.value = false
  segmentDialogVisible.value = true
}

async function saveSegment() {
  segmentSubmitted.value = true
  if (!segmentForm.value.name.trim()) return
  segmentSaving.value = true
  try {
    if (segmentEditingId.value) {
      await updateSegment(segmentEditingId.value, segmentForm.value)
    } else {
      await createSegment(segmentForm.value)
    }
    segmentDialogVisible.value = false
    await load()
  } catch (e) {
    error.value = extractError(e, 'Failed to save segment.')
  } finally {
    segmentSaving.value = false
  }
}

// ---------- Category CRUD ----------
function openAddCategory() {
  categoryEditingId.value = null
  categoryForm.value = { segmentId: categorySegmentFilter.value, name: '', description: '', placeApi: '', status: 'ACTIVE' }
  categorySubmitted.value = false
  categoryDialogVisible.value = true
}

function openEditCategory(category: MasterDataCategory) {
  categoryEditingId.value = category.id
  categoryForm.value = {
    segmentId: category.segmentId,
    name: category.name,
    description: category.description,
    placeApi: category.placeApi || '',
    status: category.status
  }
  categorySubmitted.value = false
  categoryDialogVisible.value = true
}

async function saveCategory() {
  categorySubmitted.value = true
  if (!categoryForm.value.name.trim() || !categoryForm.value.segmentId) return
  categorySaving.value = true
  try {
    if (categoryEditingId.value) {
      await updateCategory(categoryEditingId.value, categoryForm.value)
    } else {
      await createCategory(categoryForm.value)
    }
    categoryDialogVisible.value = false
    await load()
  } catch (e) {
    error.value = extractError(e, 'Failed to save category.')
  } finally {
    categorySaving.value = false
  }
}

// ---------- Delete ----------
function confirmDelete(type: MasterTab, id: string, name: string, extra?: string) {
  deleteTarget.value = { type, id, name, extra }
  deleteDialogVisible.value = true
}

async function executeDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    if (deleteTarget.value.type === 'segment') {
      await deleteSegment(deleteTarget.value.id)
    } else {
      await deleteCategory(deleteTarget.value.id)
    }
    deleteDialogVisible.value = false
    await load()
  } catch (e) {
    error.value = extractError(e, 'Failed to delete record.')
    deleteDialogVisible.value = false
  } finally {
    deleting.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="panel-stack">
    <Message v-if="error" severity="error" class="md-message" @close="error = ''">{{ error }}</Message>

    <!-- ==================== TRASH VIEW ==================== -->
    <MasterDataTrashPanel v-if="view === 'trash'" @back="backFromTrash" />

    <template v-else>
    <div class="master-panel">
      <!-- MASTER DATA HEADER -->
      <div class="master-header">
        <div>
          <h2>Master Data</h2>
          <p>Manage B2B &amp; B2C segments and prospect categories.</p>
        </div>
        <div class="master-tabs">
          <button type="button" :class="['master-tab', { active: activeTab === 'segment' }]" @click="selectTab('segment')">
            <i class="pi pi-sitemap" /><span>Segment</span><strong>{{ segments.length }}</strong>
          </button>
          <button type="button" :class="['master-tab', { active: activeTab === 'category' }]" @click="selectTab('category')">
            <i class="pi pi-tags" /><span>Category</span><strong>{{ categories.length }}</strong>
          </button>
        </div>
        <div class="header-actions">
          <Button label="Trash" icon="pi pi-trash" severity="danger" outlined size="small" @click="view = 'trash'" />
          <Button
            v-if="activeTab === 'segment'"
            label="Add Segment"
            icon="pi pi-plus"
            size="small"
            @click="openAddSegment"
          />
          <Button
            v-else
            label="Add Category"
            icon="pi pi-plus"
            size="small"
            @click="openAddCategory"
          />
        </div>
      </div>

      <!-- FILTERS -->
      <div class="master-filters">
        <div class="search-field">
          <i class="pi pi-search" />
          <input
            v-if="activeTab === 'segment'"
            type="text"
            placeholder="Search segment..."
            :value="segmentSearch"
            @input="segmentSearch = ($event.target as HTMLInputElement).value; onSegmentSearch(segmentSearch)"
          />
          <input
            v-else
            type="text"
            placeholder="Search category..."
            :value="categorySearch"
            @input="categorySearch = ($event.target as HTMLInputElement).value; onSegmentSearch(categorySearch)"
          />
        </div>
        <div v-if="activeTab === 'category'" class="filter-field">
          <label>Segment</label>
          <Select
            v-model="categorySegmentFilter"
            :options="[{ label: 'All Segments', value: '' }, ...segmentOptions]"
            optionLabel="label"
            optionValue="value"
            placeholder="All Segments"
            size="small"
            @change="load"
          />
        </div>
        <div class="filter-field">
          <label>Status</label>
          <Select
            v-if="activeTab === 'segment'"
            v-model="segmentStatusFilter"
            :options="statusOptions"
            optionLabel="label"
            optionValue="value"
            placeholder="All Status"
            size="small"
            @change="load"
          />
          <Select
            v-else
            v-model="categoryStatusFilter"
            :options="statusOptions"
            optionLabel="label"
            optionValue="value"
            placeholder="All Status"
            size="small"
            @change="load"
          />
        </div>
      </div>

      <!-- LOADING -->
      <div v-if="loading" class="state-box">
        <i class="pi pi-spin pi-spinner state-icon" />
        <span>Loading master data...</span>
      </div>

      <!-- ======================== SEGMENT TABLE ======================== -->
      <div v-else-if="activeTab === 'segment'" class="table-scroll">
        <div v-if="!segments.length" class="state-box">
          <div class="state-icon-wrap"><i class="pi pi-sitemap" /></div>
          <strong>No segments found</strong>
          <span class="muted">Create your first segment to start classifying customers.</span>
        </div>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Description</th>
              <th>Total Category</th>
              <th>Status</th>
              <th class="th-action">Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="segment in segments" :key="segment.id">
              <td>
                <span class="entity-cell">
                  <span class="entity-icon" :class="segmentAccent(segment.name)"><i class="pi pi-sitemap" /></span>
                  <span class="cell-primary">{{ segment.name }}</span>
                </span>
              </td>
              <td><span class="cell-text muted-cell">{{ segment.description || '—' }}</span></td>
              <td><span class="count-badge">{{ segment.categoryCount }} {{ segment.categoryCount === 1 ? 'category' : 'categories' }}</span></td>
              <td><Tag :value="statusLabel(segment.status)" :severity="statusSeverity(segment.status)" /></td>
              <td class="td-action">
                <div class="row-actions">
                  <Button icon="pi pi-pencil" text rounded size="small" class="act-edit" title="Edit" @click="openEditSegment(segment)" />
                  <Button icon="pi pi-trash" text rounded size="small" class="act-delete" title="Delete" @click="confirmDelete('segment', segment.id, segment.name, `${segment.categoryCount} category(s)`)" />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- ======================== CATEGORY TABLE ======================== -->
      <div v-else-if="activeTab === 'category'" class="table-scroll">
        <div v-if="!categories.length" class="state-box">
          <div class="state-icon-wrap"><i class="pi pi-tags" /></div>
          <strong>No categories found</strong>
          <span class="muted">Create a category and assign it to a segment.</span>
        </div>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Segment</th>
              <th>Description</th>
              <th>Google Place Types</th>
              <th>Status</th>
              <th class="th-action">Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="category in categories" :key="category.id">
              <td>
                <span class="entity-cell">
                  <span class="entity-icon">{{ categoryIcon(category.name) }}</span>
                  <span class="cell-primary">{{ category.name }}</span>
                </span>
              </td>
              <td><Tag :value="category.segmentName" severity="info" /></td>
              <td><span class="cell-text muted-cell">{{ category.description || '—' }}</span></td>
              <td><code v-if="category.placeApi" class="place-api-cell">{{ category.placeApi }}</code><span v-else class="muted-cell">—</span></td>
              <td><Tag :value="statusLabel(category.status)" :severity="statusSeverity(category.status)" /></td>
              <td class="td-action">
                <div class="row-actions">
                  <Button icon="pi pi-pencil" text rounded size="small" class="act-edit" title="Edit" @click="openEditCategory(category)" />
                  <Button icon="pi pi-trash" text rounded size="small" class="act-delete" title="Delete" @click="confirmDelete('category', category.id, category.name)" />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="master-footer">
        <span class="footer-chip"><i class="pi pi-sitemap" /><strong>{{ activeSegmentCount }}</strong> active segments</span>
        <span class="footer-chip"><i class="pi pi-tags" /><strong>{{ activeCategoryCount }}</strong> active categories</span>
      </div>
    </div>
    </template>

    <!-- ADD / EDIT SEGMENT DIALOG -->
    <Dialog v-model:visible="segmentDialogVisible" modal :header="segmentEditingId ? 'Edit Segment' : 'Add Segment'" :style="{ width: '440px' }">
      <div class="form-stack">
        <label class="form-field">
          <span>Segment Name <em>*</em></span>
          <InputText v-model="segmentForm.name" placeholder="e.g. B2B" fluid :invalid="segmentSubmitted && !segmentForm.name.trim()" maxlength="100" />
          <small v-if="segmentSubmitted && !segmentForm.name.trim()" class="field-error">Segment name is required.</small>
        </label>
        <label class="form-field">
          <span>Description</span>
          <Textarea v-model="segmentForm.description" placeholder="Short description of this segment" rows="3" fluid autoResize />
        </label>
        <div class="form-field">
          <span>Status</span>
          <Select v-model="segmentForm.status" :options="[statusOptions[1], statusOptions[2]]" optionLabel="label" optionValue="value" fluid />
        </div>
      </div>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="segmentDialogVisible = false" :disabled="segmentSaving" />
        <Button :label="segmentEditingId ? 'Update Segment' : 'Save Segment'" icon="pi pi-check" :loading="segmentSaving" @click="saveSegment" />
      </template>
    </Dialog>

    <!-- ADD / EDIT CATEGORY DIALOG -->
    <Dialog v-model:visible="categoryDialogVisible" modal :header="categoryEditingId ? 'Edit Category' : 'Add Category'" :style="{ width: '440px' }">
      <div class="form-stack">
        <div class="form-field">
          <span>Segment <em>*</em></span>
          <Select
            v-model="categoryForm.segmentId"
            :options="segmentOptions"
            optionLabel="label"
            optionValue="value"
            placeholder="Select segment"
            fluid
            :invalid="categorySubmitted && !categoryForm.segmentId"
          />
          <small v-if="categorySubmitted && !categoryForm.segmentId" class="field-error">Segment is required.</small>
        </div>
        <label class="form-field">
          <span>Category Name <em>*</em></span>
          <InputText v-model="categoryForm.name" placeholder="e.g. Resto & Cafe" fluid :invalid="categorySubmitted && !categoryForm.name.trim()" maxlength="100" />
          <small v-if="categorySubmitted && !categoryForm.name.trim()" class="field-error">Category name is required.</small>
        </label>
        <label class="form-field">
          <span>Description</span>
          <Textarea v-model="categoryForm.description" placeholder="Short description of this category" rows="3" fluid autoResize />
        </label>
        <label class="form-field">
          <span>Google Place Types</span>
          <InputText v-model="categoryForm.placeApi" placeholder="e.g. hotel, lodging" fluid />
          <small class="field-hint">Optional — Google Places keywords used for prospect search, separated by commas.</small>
        </label>
        <div class="form-field">
          <span>Status</span>
          <Select v-model="categoryForm.status" :options="[statusOptions[1], statusOptions[2]]" optionLabel="label" optionValue="value" fluid />
        </div>
      </div>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="categoryDialogVisible = false" :disabled="categorySaving" />
        <Button :label="categoryEditingId ? 'Update Category' : 'Save Category'" icon="pi pi-check" :loading="categorySaving" @click="saveCategory" />
      </template>
    </Dialog>

    <!-- DELETE CONFIRMATION DIALOG -->
    <Dialog v-model:visible="deleteDialogVisible" modal :header="deleteTarget?.type === 'segment' ? 'Delete Segment' : 'Delete Category'" :style="{ width: '420px' }">
      <p class="delete-text">
        Are you sure you want to delete <strong>{{ deleteTarget?.name }}</strong>?
        <span v-if="deleteTarget?.extra"> It still has <strong>{{ deleteTarget.extra }}</strong>.</span>
        Deleted items go to <strong>Trash</strong> and can be restored later.
      </p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="deleteDialogVisible = false" :disabled="deleting" />
        <Button label="Delete" severity="danger" icon="pi pi-trash" :loading="deleting" @click="executeDelete" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.panel-stack{display:flex;min-width:0;flex-direction:column;gap:.75rem}
.md-message{margin:0}
.master-panel{display:flex;flex-direction:column;gap:.85rem;padding:.9rem;border:1px solid #e5eaf0;border-radius:12px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.03)}
.master-header{display:flex;flex-wrap:wrap;align-items:center;justify-content:space-between;gap:.7rem}
.header-actions{display:flex;align-items:center;gap:.4rem}
.master-header h2{margin:0;color:#0f172a;font-size:.95rem;letter-spacing:-.01em}
.master-header p{margin:.1rem 0 0;color:#94a3b8;font-size:.68rem}
.master-tabs{display:flex;gap:.3rem;padding:.28rem;border:1px solid #e5eaf0;border-radius:9px;background:#f8fafc}
.master-tab{display:inline-flex;align-items:center;gap:.35rem;padding:.34rem .7rem;border:0;border-radius:7px;background:transparent;color:#64748b;font-size:.68rem;font-weight:700;cursor:pointer;transition:all .15s ease}
.master-tab i{font-size:.66rem}
.master-tab strong{padding:.05rem .38rem;border-radius:999px;background:#e2e8f0;color:#475569;font-size:.58rem}
.master-tab:hover{color:#0f172a}
.master-tab.active{background:#fff;color:#0f172a;box-shadow:inset 0 -2px 0 #ef4444,0 1px 2px rgba(15,23,42,.08)}
.master-tab.active strong{background:#fee2e2;color:#b91c1c}
.master-filters{display:flex;flex-wrap:wrap;align-items:flex-end;gap:.6rem}
.search-field{position:relative;display:flex;flex:1;min-width:220px;align-items:center}
.search-field i{position:absolute;left:.65rem;color:#94a3b8;font-size:.7rem}
.search-field input{width:100%;padding:.42rem .7rem .42rem 1.9rem;border:1px solid #e5eaf0;border-radius:8px;font-size:.72rem;color:#0f172a;background:#fff;outline:none}
.search-field input:focus{border-color:#c7d2fe;box-shadow:0 0 0 2px rgba(99,102,241,.12)}
.filter-field{display:grid;gap:.22rem}
.filter-field label{color:#94a3b8;font-size:.56rem;font-weight:800;text-transform:uppercase;letter-spacing:.06em}
.state-box{display:flex;flex-direction:column;align-items:center;gap:.3rem;padding:2.2rem 1rem;color:#94a3b8;font-size:.72rem;text-align:center}
.state-icon-wrap{display:grid;place-items:center;width:2.8rem;height:2.8rem;border-radius:999px;background:#f8fafc;border:1px solid #eef2f7}
.state-icon-wrap,.state-box .state-icon{font-size:1.1rem;color:#cbd5e1}
.state-box strong{color:#475569;font-size:.78rem}
.table-scroll{overflow-x:auto;border:1px solid #e5eaf0;border-radius:10px;background:#fff}
.data-table{width:100%;border-collapse:collapse;font-size:.72rem}
.data-table th{padding:.55rem .75rem;border-bottom:1px solid #e5eaf0;background:#f8fafc;color:#64748b;font-size:.58rem;font-weight:800;text-align:left;text-transform:uppercase;letter-spacing:.06em;white-space:nowrap}
.data-table td{padding:.55rem .75rem;border-bottom:1px solid #f1f5f9;color:#334155;vertical-align:middle;transition:background .12s ease}
.data-table tbody tr:last-child td{border-bottom:0}
.data-table tbody tr:hover td{background:#f8fafc}
.entity-cell{display:inline-flex;align-items:center;gap:.5rem;min-width:0}
.entity-icon{display:inline-grid;place-items:center;width:1.65rem;height:1.65rem;border-radius:.5rem;background:#f1f5f9;color:#64748b;font-size:.85rem;line-height:1;flex:none}
.entity-icon i{font-size:.7rem}
.seg-b2b{background:#eff6ff;color:#2563eb}
.seg-b2c{background:#fdf2f8;color:#db2777}
.cell-primary{color:#0f172a;font-weight:600}
.cell-text{color:#475569}
.muted-cell{color:#94a3b8}
.count-badge{display:inline-block;padding:.14rem .5rem;border-radius:999px;background:#eff6ff;color:#2563eb;font-size:.62rem;font-weight:700}
.place-api-cell{display:inline-block;max-width:260px;padding:.14rem .45rem;border:1px solid #e2e8f0;border-radius:6px;background:#f8fafc;color:#475569;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:.62rem;white-space:normal;word-break:break-word}
.field-hint{color:#94a3b8;font-size:.62rem;line-height:1.4}
.th-action,.td-action{text-align:right;white-space:nowrap}
.row-actions{display:inline-flex;gap:.15rem}
.act-edit{color:#6366f1}
.act-delete{color:#ef4444}
.master-footer{display:flex;flex-wrap:wrap;gap:.5rem;color:#94a3b8;font-size:.64rem}
.footer-chip{display:inline-flex;align-items:center;gap:.4rem;padding:.3rem .7rem;border:1px solid #e5eaf0;border-radius:999px;background:#f8fafc;color:#94a3b8}
.footer-chip i{font-size:.68rem;color:#cbd5e1}
.footer-chip strong{color:#0f172a}
.form-stack{display:grid;gap:.85rem;padding-top:.2rem}
.form-field{display:grid;gap:.32rem}
.form-field span{color:#475569;font-size:.68rem;font-weight:700}
.form-field span em{color:#ef4444;font-style:normal}
.field-error{color:#ef4444;font-size:.62rem}
.delete-text{margin:0;color:#475569;font-size:.78rem;line-height:1.5}
@media (max-width: 720px){
  .master-header{flex-direction:column;align-items:stretch}
  .master-tabs{justify-content:center}
}
</style>
