<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
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
import { fallbackCategories, fallbackSegments } from '../../utils/masterDataFallback'

type MasterTab = 'segment' | 'category'
type MasterView = 'main' | 'trash'

const activeTab = ref<MasterTab>('segment')
const route = useRoute()
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
const categoryExpanded = ref(false)

watch(() => route.query.action, (action) => {
  if (action === 'add-segment') openAddSegment()
  if (action === 'add-category') openAddCategory()
  if (action === 'trash') view.value = 'trash'
}, { immediate: true })

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

// ---------- Read-only detail dialog ----------
const detailDialogVisible = ref(false)
const detailType = ref<MasterTab>('segment')
const detailSegment = ref<MasterDataSegment | null>(null)
const detailCategory = ref<MasterDataCategory | null>(null)

// ---------- Inline display order ----------
const orderType = ref<MasterTab | null>(null)
const orderDraftIds = ref<string[]>([])
const draggingOrderId = ref<string | null>(null)

// ---------- Delete dialogs ----------
const deleteDialogVisible = ref(false)
const deleting = ref(false)
const deleteTarget = ref<{ type: MasterTab; id: string; name: string; extra?: string } | null>(null)

const segmentOptions = computed(() =>
  segments.value.map((s) => ({ label: s.name, value: s.id }))
)

const activeSegmentCount = computed(() => segments.value.filter((s) => s.status === 'ACTIVE').length)
const activeCategoryCount = computed(() => categories.value.filter((c) => c.status === 'ACTIVE').length)
const visibleCategories = computed(() => categoryExpanded.value ? categories.value : categories.value.slice(0, 3))

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

function placeTypeTokens(value: string) {
  return value.split(',').map((item) => item.trim()).filter(Boolean)
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
    segments.value = segmentResult?.length ? segmentResult : fallbackSegments
    categories.value = categoryResult?.length ? categoryResult : fallbackCategories
    categoryExpanded.value = false
  } catch (e) {
    // Keep the management panel usable when the optional master-data API is unavailable.
    segments.value = fallbackSegments
    categories.value = fallbackCategories
    error.value = ''
    console.warn('Master data could not be loaded:', e)
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

function openSegmentDetail(segment: MasterDataSegment) {
  detailType.value = 'segment'
  detailSegment.value = segment
  detailCategory.value = null
  detailDialogVisible.value = true
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

function openCategoryDetail(category: MasterDataCategory) {
  detailType.value = 'category'
  detailCategory.value = category
  detailSegment.value = null
  detailDialogVisible.value = true
}

function beginOrder(type: MasterTab) {
  if (orderType.value) return
  const source = type === 'segment' ? segments.value : categories.value
  if (source.length <= 1) return
  orderType.value = type
  orderDraftIds.value = source.map((item) => item.id)
}

function cancelOrder() {
  orderType.value = null
  orderDraftIds.value = []
  draggingOrderId.value = null
}

function startOrderDrag(type: MasterTab, id: string, event: DragEvent) {
  if (orderType.value !== type) return
  draggingOrderId.value = id
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', id)
  }
}

function moveDraftRow(type: MasterTab, targetId: string) {
  if (orderType.value !== type || !draggingOrderId.value || draggingOrderId.value === targetId) return
  const nextIds = [...orderDraftIds.value]
  const fromIndex = nextIds.indexOf(draggingOrderId.value)
  const toIndex = nextIds.indexOf(targetId)
  if (fromIndex < 0 || toIndex < 0) return
  const [movedId] = nextIds.splice(fromIndex, 1)
  nextIds.splice(toIndex, 0, movedId)
  orderDraftIds.value = nextIds
}

function endOrderDrag() {
  draggingOrderId.value = null
}

function saveOrder() {
  if (!orderType.value) return
  const itemById = new Map((orderType.value === 'segment' ? segments.value : categories.value).map((item) => [item.id, item]))
  const orderedItems = orderDraftIds.value.map((id) => itemById.get(id)).filter(Boolean) as (MasterDataSegment | MasterDataCategory)[]
  if (orderType.value === 'segment') segments.value = orderedItems as MasterDataSegment[]
  else categories.value = orderedItems as MasterDataCategory[]
  cancelOrder()
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
    <div class="master-reference-layout">
      <div class="reference-page-header">
        <div>
          <span class="reference-eyebrow">Master Data</span>
          <p>Manage B2B &amp; B2C segments and prospect categories.</p>
        </div>
        <Button label="Trash" icon="pi pi-trash" severity="secondary" outlined size="small" @click="view = 'trash'" />
      </div>

      <div v-if="loading" class="state-box master-loading">
        <i class="pi pi-spin pi-spinner state-icon" />
        <span>Loading master data...</span>
      </div>

      <div v-else class="master-data-grid">
        <section class="master-data-card">
          <div class="master-card-header">
            <div class="master-card-title">
              <div class="master-card-title-line">
                <h3>Customer Site Segment</h3>
                <span class="items-badge">{{ segments.length }} items</span>
              </div>
              <p>Customer segment used for customer grouping and reporting.</p>
              <div v-if="orderType === 'segment'" class="order-mode-hint"><i class="pi pi-info-circle" />Drag the handle to reorder rows.</div>
            </div>
            <div :class="['master-card-actions', { 'order-actions-active': orderType === 'segment' }]">
              <template v-if="orderType === 'segment'">
                <Button label="Cancel" severity="secondary" outlined size="small" @click="cancelOrder" />
                <Button label="Save Order" icon="pi pi-check" severity="danger" size="small" @click="saveOrder" />
              </template>
              <template v-else>
                <Button label="Edit Order" icon="pi pi-pencil" severity="secondary" outlined size="small" :disabled="segments.length <= 1 || !!orderType" :title="orderType ? 'Finish the current reorder first' : segments.length <= 1 ? 'Need at least two rows to reorder' : 'Edit row order'" @click="beginOrder('segment')" />
                <Button label="Create Segment" icon="pi pi-plus" severity="secondary" outlined size="small" :disabled="!!orderType" @click="openAddSegment" />
              </template>
            </div>
          </div>
          <div class="master-card-body">
            <div v-if="!segments.length" class="state-box">
              <div class="state-icon-wrap"><i class="pi pi-sitemap" /></div>
              <strong>No segments found</strong>
              <span class="muted">Create your first segment to start classifying customers.</span>
            </div>
            <div v-else class="table-scroll card-table-scroll">
              <table :class="['data-table', 'segment-table', { 'is-reordering': orderType === 'segment' }]">
                <thead>
                  <tr><th v-if="orderType === 'segment'" class="order-handle-heading" /><th>Name</th><th>Description</th><th>Total Category</th><th v-if="orderType !== 'segment'" class="th-action">Action</th></tr>
                </thead>
                <tbody>
                  <tr v-for="segment in (orderType === 'segment' ? orderDraftIds.map((id) => segments.find((item) => item.id === id)).filter(Boolean) as MasterDataSegment[] : segments)" :key="segment.id" :class="{ 'order-row-active': draggingOrderId === segment.id }" @dragenter.prevent="moveDraftRow('segment', segment.id)" @dragover.prevent @drop.prevent="endOrderDrag">
                    <td v-if="orderType === 'segment'" class="order-handle-cell">
                      <button type="button" class="order-drag-handle" draggable="true" :aria-label="`Drag ${segment.name} to reorder`" title="Drag to reorder" @dragstart="startOrderDrag('segment', segment.id, $event)" @dragend="endOrderDrag"><i class="pi pi-bars" /></button>
                    </td>
                    <td><span class="entity-cell"><span class="cell-primary">{{ segment.name }}</span></span></td>
                    <td><span class="cell-text muted-cell">{{ segment.description || '-' }}</span></td>
                    <td><span class="count-badge">{{ segment.categoryCount }} {{ segment.categoryCount === 1 ? 'category' : 'categories' }}</span></td>
                    <td v-if="orderType !== 'segment'" class="td-action"><div class="row-actions"><Button icon="pi pi-eye" text rounded size="small" class="act-view" title="View detail" @click="openSegmentDetail(segment)" /><Button icon="pi pi-pencil" text rounded size="small" class="act-edit" title="Edit" @click="openEditSegment(segment)" /><Button icon="pi pi-trash" text rounded size="small" class="act-delete" title="Delete" @click="confirmDelete('segment', segment.id, segment.name)" /></div></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <section class="master-data-card">
          <div class="master-card-header">
            <div class="master-card-title">
              <div class="master-card-title-line">
                <h3>Customer Site Category</h3>
                <span class="items-badge">{{ categories.length }} items</span>
              </div>
              <p>Customer category master used for status and account grouping.</p>
              <div v-if="orderType === 'category'" class="order-mode-hint"><i class="pi pi-info-circle" />Drag the handle to reorder rows.</div>
            </div>
            <div :class="['master-card-actions', { 'order-actions-active': orderType === 'category' }]">
              <template v-if="orderType === 'category'">
                <Button label="Cancel" severity="secondary" outlined size="small" @click="cancelOrder" />
                <Button label="Save Order" icon="pi pi-check" severity="danger" size="small" @click="saveOrder" />
              </template>
              <template v-else>
                <Button label="Edit Order" icon="pi pi-pencil" severity="secondary" outlined size="small" :disabled="categories.length <= 1 || !!orderType" :title="orderType ? 'Finish the current reorder first' : categories.length <= 1 ? 'Need at least two rows to reorder' : 'Edit row order'" @click="beginOrder('category')" />
                <Button label="Create Category" icon="pi pi-plus" severity="secondary" outlined size="small" :disabled="!!orderType" @click="openAddCategory" />
              </template>
            </div>
          </div>
          <div class="master-card-body">
            <div v-if="!categories.length" class="state-box">
              <div class="state-icon-wrap"><i class="pi pi-tags" /></div>
              <strong>No categories found</strong>
              <span class="muted">Create a category and assign it to a segment.</span>
            </div>
            <div v-else class="table-scroll card-table-scroll">
              <table :class="['data-table', 'category-table', { 'is-reordering': orderType === 'category' }]">
                <thead>
                  <tr><th v-if="orderType === 'category'" class="order-handle-heading" /><th>Name</th><th>Segment</th><th>Description</th><th>Google Place Types</th><th v-if="orderType !== 'category'" class="th-action">Action</th></tr>
                </thead>
                <tbody>
                  <tr v-for="category in (orderType === 'category' ? orderDraftIds.map((id) => categories.find((item) => item.id === id)).filter(Boolean) as MasterDataCategory[] : visibleCategories)" :key="category.id" :class="{ 'order-row-active': draggingOrderId === category.id }" @dragenter.prevent="moveDraftRow('category', category.id)" @dragover.prevent @drop.prevent="endOrderDrag">
                    <td v-if="orderType === 'category'" class="order-handle-cell">
                      <button type="button" class="order-drag-handle" draggable="true" :aria-label="`Drag ${category.name} to reorder`" title="Drag to reorder" @dragstart="startOrderDrag('category', category.id, $event)" @dragend="endOrderDrag"><i class="pi pi-bars" /></button>
                    </td>
                    <td><span class="entity-cell"><span class="cell-primary">{{ category.name }}</span></span></td>
                    <td><Tag :value="category.segmentName" severity="info" /></td>
                    <td><span class="cell-text muted-cell">{{ category.description || '-' }}</span></td>
                    <td>
                      <div v-if="category.placeApi" class="place-api-list">
                        <span v-for="placeType in placeTypeTokens(category.placeApi)" :key="placeType" class="place-api-pill">{{ placeType }}</span>
                      </div>
                      <span v-else class="muted-cell">-</span>
                    </td>
                    <td v-if="orderType !== 'category'" class="td-action"><div class="row-actions"><Button icon="pi pi-eye" text rounded size="small" class="act-view" title="View detail" @click="openCategoryDetail(category)" /><Button icon="pi pi-pencil" text rounded size="small" class="act-edit" title="Edit" @click="openEditCategory(category)" /><Button icon="pi pi-trash" text rounded size="small" class="act-delete" title="Delete" @click="confirmDelete('category', category.id, category.name)" /></div></td>
                  </tr>
                </tbody>
              </table>
              <button v-if="categories.length > 3 && orderType !== 'category'" type="button" class="table-expand-button" @click="categoryExpanded = !categoryExpanded">
                <span>{{ categoryExpanded ? 'Show Less' : 'Show All' }}</span>
                <i :class="categoryExpanded ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
              </button>
            </div>
          </div>
        </section>
      </div>
    </div>
    <div class="legacy-master-layout">
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
        <table v-else class="data-table segment-table">
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
        <table v-else class="data-table category-table">
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
    </div>
    </template>

    <!-- READ-ONLY DETAIL DIALOG -->
    <Dialog v-model:visible="detailDialogVisible" modal :header="detailType === 'segment' ? 'Segment Detail' : 'Category Detail'" :style="{ width: '520px' }">
      <div v-if="detailType === 'segment' && detailSegment" class="detail-stack">
        <div class="detail-field"><span>Name</span><strong>{{ detailSegment.name }}</strong></div>
        <div class="detail-field"><span>Description</span><p>{{ detailSegment.description || '-' }}</p></div>
        <div class="detail-field"><span>Total Category</span><strong>{{ detailSegment.categoryCount }} {{ detailSegment.categoryCount === 1 ? 'category' : 'categories' }}</strong></div>
        <div class="detail-field"><span>Status</span><Tag :value="statusLabel(detailSegment.status)" :severity="statusSeverity(detailSegment.status)" /></div>
      </div>
      <div v-else-if="detailType === 'category' && detailCategory" class="detail-stack">
        <div class="detail-field"><span>Category</span><strong>{{ detailCategory.name }}</strong></div>
        <div class="detail-field"><span>Segment</span><Tag :value="detailCategory.segmentName" severity="info" /></div>
        <div class="detail-field"><span>Description</span><p>{{ detailCategory.description || '-' }}</p></div>
        <div class="detail-field"><span>Google Place Types</span>
          <div v-if="detailCategory.placeApi" class="place-api-list detail-place-list">
            <span v-for="placeType in placeTypeTokens(detailCategory.placeApi)" :key="placeType" class="place-api-pill">{{ placeType }}</span>
          </div>
          <p v-else>-</p>
        </div>
        <div class="detail-field"><span>Status</span><Tag :value="statusLabel(detailCategory.status)" :severity="statusSeverity(detailCategory.status)" /></div>
      </div>
      <template #footer>
        <Button label="Close" severity="secondary" text @click="detailDialogVisible = false" />
      </template>
    </Dialog>

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
    <Dialog v-model:visible="categoryDialogVisible" modal class="category-dialog" :header="categoryEditingId ? 'Edit Category' : 'Add Category'" :style="{ width: '440px' }">
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
          <Textarea v-model="categoryForm.placeApi" class="category-place-api-input" placeholder="e.g. hotel, lodging" rows="3" fluid autoResize />
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
.panel-stack{display:flex;min-width:0;flex-direction:column;gap:1rem}
.md-message{margin:0}
.master-panel{display:flex;flex-direction:column;gap:0;padding:0;border:0;border-radius:0;background:transparent;box-shadow:none}
.master-header{display:grid;grid-template-columns:minmax(220px,1fr) auto auto;align-items:center;gap:1rem;padding:1rem 1.1rem;min-height:76px;border:1px solid #e5ebf2;border-radius:16px 16px 0 0;background:#fff}
.header-actions{display:flex;align-items:center;gap:.4rem}
.master-header h2{margin:0;color:#0f172a;font-size:.95rem;letter-spacing:-.01em}
.master-header p{margin:.1rem 0 0;color:#94a3b8;font-size:.68rem}
.master-tabs{display:flex;gap:.25rem;padding:.2rem;border:1px solid #e5ebf2;border-radius:10px;background:#f8fafc}
.master-tab{display:inline-flex;align-items:center;gap:.35rem;padding:.42rem .72rem;border:0;border-radius:8px;background:transparent;color:#64748b;font-size:.68rem;font-weight:700;cursor:pointer;transition:all .15s ease}
.master-tab i{font-size:.66rem}
.master-tab strong{padding:.05rem .38rem;border-radius:999px;background:#e2e8f0;color:#475569;font-size:.58rem}
.master-tab:hover{color:#0f172a}
.master-tab.active{background:#fff;color:#0f172a;box-shadow:0 1px 3px rgba(15,23,42,.08)}
.master-tab.active strong{background:#e8f3fb;color:#25617e}
.master-filters{display:flex;flex-wrap:wrap;align-items:flex-end;gap:.6rem;padding:.8rem 1.1rem;border-right:1px solid #e5ebf2;border-left:1px solid #e5ebf2;background:#fff}
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
.table-scroll{overflow-x:auto;padding:.7rem 1rem 1rem;border:1px solid #e5ebf2;border-top:0;border-radius:0 0 16px 16px;background:#fff}
.data-table{width:100%;min-width:700px;border-collapse:separate;border-spacing:0 .45rem;font-size:.72rem}
.data-table th{height:38px;padding:.5rem .75rem;border-top:1px solid #dfe7ef;border-bottom:1px solid #dfe7ef;background:#f8fafc;color:#64748b;font-size:.58rem;font-weight:800;text-align:left;text-transform:uppercase;letter-spacing:.06em;white-space:nowrap}
.data-table th:first-child{border-left:1px solid #dfe7ef;border-radius:12px 0 0 12px}
.data-table th:last-child{border-right:1px solid #dfe7ef;border-radius:0 12px 12px 0}
.data-table td{height:56px;padding:.6rem .75rem;border-top:1px solid #e5ebf2;border-bottom:1px solid #e5ebf2;background:#fff;color:#334155;vertical-align:middle;transition:background .12s ease}
.data-table td:first-child{border-left:1px solid #e5ebf2;border-radius:12px 0 0 12px}
.data-table td:last-child{border-right:1px solid #e5ebf2;border-radius:0 12px 12px 0}
.data-table tbody tr:hover td{background:#fbfdff}
.segment-table th:nth-child(1),.segment-table td:nth-child(1){width:25%}
.segment-table th:nth-child(2),.segment-table td:nth-child(2){width:30%}
.segment-table th:nth-child(3),.segment-table td:nth-child(3){width:18%}
.segment-table th:nth-child(4),.segment-table td:nth-child(4){width:13%}
.segment-table th:nth-child(5),.segment-table td:nth-child(5){width:14%}
.category-table th:nth-child(1),.category-table td:nth-child(1){width:17%}
.category-table th:nth-child(2),.category-table td:nth-child(2){width:13%}
.category-table th:nth-child(3),.category-table td:nth-child(3){width:22%}
.category-table th:nth-child(4),.category-table td:nth-child(4){width:25%}
.category-table th:nth-child(5),.category-table td:nth-child(5){width:11%}
.category-table th:nth-child(6),.category-table td:nth-child(6){width:12%}
.data-table td:nth-child(2){max-width:260px}
.data-table td:nth-child(3){max-width:230px}
.entity-cell{display:inline-flex;align-items:center;gap:.5rem;min-width:0}
.entity-icon{display:inline-grid;place-items:center;width:1.8rem;height:1.8rem;border-radius:.6rem;background:#f1f5f9;color:#64748b;font-size:.85rem;line-height:1;flex:none}
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
.row-actions{display:inline-flex;align-items:center;justify-content:flex-end;gap:.15rem;padding:0;border:0;background:transparent}
.row-actions :deep(.p-button){width:2rem;height:2rem;padding:0;border:0;border-radius:8px;background:transparent;box-shadow:none}
.row-actions :deep(.p-button:hover){background:#f1f5f9}
.act-view{color:#64748b}
.act-view:hover{color:#2563eb}
.act-edit{color:#6366f1}
.act-delete{color:#ef4444}
.master-footer{display:flex;flex-wrap:wrap;gap:.5rem;padding:.8rem 0 0;color:#94a3b8;font-size:.64rem}
.footer-chip{display:inline-flex;align-items:center;gap:.4rem;padding:.3rem .7rem;border:1px solid #e5eaf0;border-radius:999px;background:#f8fafc;color:#94a3b8}
.footer-chip i{font-size:.68rem;color:#cbd5e1}
.footer-chip strong{color:#0f172a}
.form-stack{display:grid;gap:.85rem;padding-top:.2rem}
.detail-stack{display:grid;gap:.65rem}
.detail-field{display:grid;grid-template-columns:150px minmax(0,1fr);align-items:start;gap:.75rem;padding:.7rem .8rem;border:1px solid #e2e8f0;border-radius:10px;background:#f8fafc}
.detail-field>span{color:#64748b;font-size:.68rem;font-weight:700}
.detail-field>strong,.detail-field>p{min-width:0;margin:0;color:#0f172a;font-size:.74rem;line-height:1.45;overflow-wrap:anywhere;word-break:break-word}
.detail-place-list{align-items:flex-start}
.form-field{display:grid;gap:.32rem}
.form-field span{color:#475569;font-size:.68rem;font-weight:700}
.form-field span em{color:#ef4444;font-style:normal}
.field-error{color:#ef4444;font-size:.62rem}
.delete-text{margin:0;color:#475569;font-size:.78rem;line-height:1.5}
.category-dialog :deep(.p-dialog){max-width:calc(100vw - 2rem);overflow:hidden}
.category-dialog :deep(.p-dialog-header){padding:1rem 1.2rem;border-bottom:1px solid #eef2f7}
.category-dialog :deep(.p-dialog-title){min-width:0;color:#0f172a;font-size:.95rem;line-height:1.3;overflow-wrap:anywhere}
.category-dialog :deep(.p-dialog-content){min-width:0;overflow-x:hidden;padding:1rem 1.2rem 1.15rem}
.category-dialog .form-stack{min-width:0;gap:.72rem;padding-top:0}
.category-dialog .form-field{min-width:0;gap:.3rem}
.category-dialog .form-field>span{min-width:0;line-height:1.35;overflow-wrap:anywhere}
.category-dialog .form-field .field-hint{display:block;min-width:0;line-height:1.4;overflow-wrap:anywhere;word-break:break-word}
.category-dialog :deep(.p-inputtext),.category-dialog :deep(.p-textarea),.category-dialog :deep(.p-select){box-sizing:border-box;width:100%;max-width:100%;min-width:0}
.category-dialog .category-place-api-input :deep(textarea),.category-dialog :deep(.category-place-api-input){line-height:1.45;overflow-wrap:anywhere;white-space:pre-wrap}
.category-dialog :deep(.p-select-label){min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.category-dialog :deep(.p-dialog-footer){display:flex;flex-wrap:wrap;justify-content:flex-end;gap:.5rem;padding:.85rem 1.2rem 1rem;border-top:1px solid #eef2f7}
.category-dialog :deep(.p-dialog-footer .p-button){max-width:100%;white-space:normal}
@media (max-width: 720px){
  .master-header{grid-template-columns:1fr;align-items:stretch}
  .master-tabs{justify-content:center}
  .header-actions{justify-content:flex-end}
}

/* Reference layout: two equal cards with a shared, quiet page surface. */
.legacy-master-layout{display:none}
.master-reference-layout{display:flex;min-width:0;flex-direction:column;gap:.875rem}
.reference-page-header{display:flex;align-items:center;justify-content:space-between;gap:1rem;padding:.875rem 1rem;border:1px solid #e2e8f0;border-radius:16px;background:#fff;box-shadow:0 10px 24px rgba(15,23,42,.05)}
.reference-eyebrow{display:block;color:#94a3b8;font-size:.68rem;font-weight:700;letter-spacing:.08em;text-transform:uppercase}
.reference-page-header p{margin:.3rem 0 0;color:#64748b;font-size:.72rem;line-height:1.45}
.master-data-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:1rem;min-width:0}
.master-data-card{display:flex;min-width:0;min-height:368px;flex-direction:column;overflow:hidden;border:1px solid #e2e8f0;border-radius:20px;background:#fff;box-shadow:0 14px 34px rgba(15,23,42,.06)}
.master-card-header{display:flex;align-items:flex-start;justify-content:space-between;gap:.75rem;padding:1rem 1.1rem;border-bottom:1px solid #eef2f7}
.master-card-title{min-width:0}
.master-card-title-line{display:flex;flex-wrap:wrap;align-items:center;gap:.5rem}
.master-card-title h3{margin:0;color:#0f172a;font-size:1rem;font-weight:700;letter-spacing:-.02em}
.master-card-title p{margin:.25rem 0 0;color:#64748b;font-size:.72rem;line-height:1.5}
.items-badge{display:inline-flex;align-items:center;min-height:1.5rem;padding:.2rem .6rem;border:1px solid #e2e8f0;border-radius:999px;background:#f8fafc;color:#475569;font-size:.62rem;font-weight:600;white-space:nowrap}
.master-card-actions{display:flex;flex-wrap:wrap;align-items:center;justify-content:flex-end;gap:.45rem;flex:none}
.order-mode-hint{display:inline-flex;max-width:100%;align-items:center;gap:.4rem;margin-top:.55rem;padding:.3rem .55rem;border:1px solid #fecaca;border-radius:999px;background:#fff5f5;color:#b91c1c;font-size:.62rem;font-weight:600;line-height:1.3}
.order-mode-hint i{font-size:.68rem}
.master-card-header :deep(.p-button){height:2rem;flex:none;border-radius:10px;padding:.4rem .7rem;font-size:.7rem;font-weight:600}
.master-card-body{display:flex;min-width:0;flex:1;flex-direction:column;padding:.75rem 1rem 1rem}
.card-table-scroll{overflow-x:auto;padding:0;border:0;border-radius:0;background:transparent}
.card-table-scroll .data-table{min-width:620px;border-spacing:0 .45rem;table-layout:fixed}
.card-table-scroll .category-table{min-width:690px}
.card-table-scroll .data-table th{height:38px;padding:.5rem .75rem;border-top:1px solid #d9e2ec;border-bottom:1px solid #d9e2ec;background:#f8fafc}
.card-table-scroll .data-table th:first-child{border-left:1px solid #d9e2ec;border-radius:12px 0 0 12px}
.card-table-scroll .data-table th:last-child{border-right:1px solid #d9e2ec;border-radius:0 12px 12px 0}
.card-table-scroll .data-table td{height:54px;padding:.55rem .75rem;border-top:1px solid #e2e8f0;border-bottom:1px solid #e2e8f0;background:#fff}
.card-table-scroll .data-table td:first-child{border-left:1px solid #e2e8f0;border-radius:12px 0 0 12px}
.card-table-scroll .data-table td:last-child{border-right:1px solid #e2e8f0;border-radius:0 12px 12px 0}
.card-table-scroll .data-table tbody tr:hover td{background:#fbfdff}
.card-table-scroll .segment-table th:nth-child(1),.card-table-scroll .segment-table td:nth-child(1){width:24%}
.card-table-scroll .segment-table th:nth-child(2),.card-table-scroll .segment-table td:nth-child(2){width:29%}
.card-table-scroll .segment-table th:nth-child(3),.card-table-scroll .segment-table td:nth-child(3){width:19%}
.card-table-scroll .segment-table th:nth-child(4),.card-table-scroll .segment-table td:nth-child(4){width:14%}
.card-table-scroll .segment-table th:nth-child(5),.card-table-scroll .segment-table td:nth-child(5){width:14%}
.card-table-scroll .category-table th:nth-child(1),.card-table-scroll .category-table td:nth-child(1){width:18%}
.card-table-scroll .category-table th:nth-child(2),.card-table-scroll .category-table td:nth-child(2){width:13%}
.card-table-scroll .category-table th:nth-child(3),.card-table-scroll .category-table td:nth-child(3){width:21%}
.card-table-scroll .category-table th:nth-child(4),.card-table-scroll .category-table td:nth-child(4){width:25%}
.card-table-scroll .category-table th:nth-child(5),.card-table-scroll .category-table td:nth-child(5){width:11%}
.card-table-scroll .category-table th:nth-child(6),.card-table-scroll .category-table td:nth-child(6){width:12%}
.master-loading{min-height:368px;border:1px solid #e2e8f0;border-radius:20px;background:#fff}
.place-api-list{display:flex;max-width:100%;flex-wrap:wrap;align-items:center;gap:.25rem}
.place-api-pill{display:inline-flex;max-width:100%;align-items:center;padding:.18rem .38rem;border:1px solid #e2e8f0;border-radius:6px;background:#f8fafc;color:#475569;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:.56rem;font-weight:500;line-height:1.25;overflow-wrap:anywhere;white-space:normal}
.card-table-scroll .category-table td:nth-child(4){vertical-align:top}
.card-table-scroll .category-table{width:100%;min-width:660px;table-layout:fixed}
.card-table-scroll .category-table th{white-space:normal;line-height:1.25;overflow-wrap:anywhere}
.card-table-scroll .category-table td{min-width:0;overflow-wrap:anywhere;word-break:break-word}
.card-table-scroll .category-table th:nth-child(1),.card-table-scroll .category-table td:nth-child(1){width:22%}
.card-table-scroll .category-table th:nth-child(2),.card-table-scroll .category-table td:nth-child(2){width:12%}
.card-table-scroll .category-table th:nth-child(3),.card-table-scroll .category-table td:nth-child(3){width:20%}
.card-table-scroll .category-table th:nth-child(4),.card-table-scroll .category-table td:nth-child(4){width:18%}
.card-table-scroll .category-table th:nth-child(5),.card-table-scroll .category-table td:nth-child(5){width:12%}
.card-table-scroll .category-table th:nth-child(6),.card-table-scroll .category-table td:nth-child(6){width:16%}
.card-table-scroll .category-table .entity-cell{display:flex;width:100%;min-width:0}
.card-table-scroll .category-table .entity-cell .entity-icon{flex:0 0 1.8rem}
.card-table-scroll .category-table .entity-cell .cell-primary{flex:1 1 auto;max-width:calc(100% - 2.2rem)}
.card-table-scroll .category-table .cell-primary,.card-table-scroll .category-table .cell-text{min-width:0;overflow-wrap:anywhere;word-break:break-word}
.card-table-scroll .category-table .cell-primary,.card-table-scroll .category-table .cell-text{display:block;max-width:100%;line-height:1.4}
.card-table-scroll .category-table td:nth-child(1),.card-table-scroll .category-table td:nth-child(2),.card-table-scroll .category-table td:nth-child(3),.card-table-scroll .category-table td:nth-child(4){vertical-align:top}
.card-table-scroll .category-table td:nth-child(1) .entity-cell{align-items:flex-start}
.card-table-scroll .category-table td:nth-child(2) :deep(.p-tag){display:inline-flex;min-width:0}
.card-table-scroll .category-table :deep(.p-tag){max-width:100%;white-space:normal;overflow-wrap:anywhere;word-break:break-word;text-align:center;line-height:1.25}
.card-table-scroll .category-table .row-actions{flex-wrap:nowrap;max-width:100%;white-space:nowrap}
.card-table-scroll .category-table .row-actions .p-button{flex:0 0 auto}
.card-table-scroll .category-table td:first-child{overflow:hidden}
.card-table-scroll .category-table td:first-child .entity-cell{align-items:flex-start;gap:.4rem;max-width:100%;overflow:hidden}
.card-table-scroll .category-table td:first-child .cell-primary{display:block;flex:1 1 0;max-width:100%;min-width:0;overflow-wrap:anywhere;word-break:break-word;white-space:normal;line-height:1.35}
.card-table-scroll .category-table td:first-child .entity-icon{margin-top:.05rem}
.card-table-scroll .category-table .row-actions :deep(.p-button){width:1.75rem;height:1.75rem}
.card-table-scroll .category-table td{overflow:hidden}
.card-table-scroll .category-table .place-api-list{min-width:0;align-content:flex-start}
.card-table-scroll .category-table .place-api-pill{min-width:0}
.card-table-scroll .category-table th:nth-child(1),.card-table-scroll .category-table td:nth-child(1){width:20%}
.card-table-scroll .category-table th:nth-child(2),.card-table-scroll .category-table td:nth-child(2){width:13%}
.card-table-scroll .category-table th:nth-child(3),.card-table-scroll .category-table td:nth-child(3){width:17%}
.card-table-scroll .category-table th:nth-child(4),.card-table-scroll .category-table td:nth-child(4){width:24%}
.card-table-scroll .category-table th:nth-child(5),.card-table-scroll .category-table td:nth-child(5){width:12%}
.card-table-scroll .category-table th:nth-child(6),.card-table-scroll .category-table td:nth-child(6){width:16%}
.card-table-scroll .category-table td{padding-right:.55rem;padding-left:.55rem}
.card-table-scroll .category-table .place-api-list{gap:.16rem;row-gap:.16rem}
.card-table-scroll .category-table .place-api-pill{padding:.12rem .27rem;font-size:.5rem;line-height:1.12}
.card-table-scroll .category-table th:nth-child(1),.card-table-scroll .category-table td:nth-child(1){width:22%}
.card-table-scroll .category-table th:nth-child(2),.card-table-scroll .category-table td:nth-child(2){width:12%}
.card-table-scroll .category-table th:nth-child(3),.card-table-scroll .category-table td:nth-child(3){width:16%}
.card-table-scroll .category-table th:nth-child(4),.card-table-scroll .category-table td:nth-child(4){width:22%}
.card-table-scroll .category-table th:nth-child(5),.card-table-scroll .category-table td:nth-child(5){width:12%}
.card-table-scroll .category-table th:nth-child(6),.card-table-scroll .category-table td:nth-child(6){width:16%}
.card-table-scroll .category-table td:first-child .cell-primary{overflow-wrap:break-word;word-break:normal}
.card-table-scroll .category-table th:nth-child(2){white-space:nowrap;word-break:normal;overflow-wrap:normal}
.card-table-scroll .category-table th,.card-table-scroll .category-table td{box-sizing:border-box}
.card-table-scroll .category-table td>*{max-width:100%}
.card-table-scroll .category-table .entity-icon{flex:0 0 1.8rem}
.card-table-scroll .category-table .cell-primary{flex:1 1 auto}
.card-table-scroll .category-table .place-api-list{width:100%;overflow-wrap:anywhere}
.card-table-scroll .category-table th:nth-child(1),.card-table-scroll .category-table td:nth-child(1){width:26%}
.card-table-scroll .category-table th:nth-child(2),.card-table-scroll .category-table td:nth-child(2){width:11%}
.card-table-scroll .category-table th:nth-child(3),.card-table-scroll .category-table td:nth-child(3){width:15%}
.card-table-scroll .category-table th:nth-child(4),.card-table-scroll .category-table td:nth-child(4){width:20%}
.card-table-scroll .category-table th:nth-child(5),.card-table-scroll .category-table td:nth-child(5){width:12%}
.card-table-scroll .category-table th:nth-child(6),.card-table-scroll .category-table td:nth-child(6){width:16%}
.card-table-scroll .category-table td:first-child .cell-primary{max-width:none;overflow-wrap:break-word;word-break:normal}
.card-table-scroll .segment-table{width:100%;min-width:100%;table-layout:fixed}
.card-table-scroll .segment-table th{white-space:normal;line-height:1.25;overflow-wrap:anywhere}
.card-table-scroll .segment-table td{min-width:0;overflow-wrap:anywhere;word-break:break-word}
.card-table-scroll .segment-table th:nth-child(1),.card-table-scroll .segment-table td:nth-child(1){width:22%}
.card-table-scroll .segment-table th:nth-child(2),.card-table-scroll .segment-table td:nth-child(2){width:28%}
.card-table-scroll .segment-table th:nth-child(3),.card-table-scroll .segment-table td:nth-child(3){width:18%}
.card-table-scroll .segment-table th:nth-child(4),.card-table-scroll .segment-table td:nth-child(4){width:14%}
.card-table-scroll .segment-table th:nth-child(5),.card-table-scroll .segment-table td:nth-child(5){width:18%}
.card-table-scroll .segment-table .entity-cell{display:flex;width:100%;min-width:0}
.card-table-scroll .segment-table .cell-primary,.card-table-scroll .segment-table .cell-text{min-width:0;overflow-wrap:anywhere;word-break:break-word}
.card-table-scroll .segment-table .count-badge{display:inline-flex;max-width:100%;align-items:center;justify-content:center;text-align:center;white-space:normal;overflow-wrap:anywhere;line-height:1.25}
.card-table-scroll .segment-table :deep(.p-tag){max-width:100%;white-space:normal;overflow-wrap:anywhere;word-break:break-word;text-align:center;line-height:1.25}
.card-table-scroll .segment-table .row-actions{flex-wrap:wrap}
.card-table-scroll .segment-table .row-actions :deep(.p-button){width:1.6rem;height:1.6rem}

/* Keep both master tables inside their cards without forcing a horizontal overflow. */
.master-data-card{height:100%}
.card-table-scroll{min-width:0;overflow-x:hidden}
.card-table-scroll .data-table,
.card-table-scroll .category-table,
.card-table-scroll .segment-table{width:100%;min-width:0;table-layout:fixed}
.card-table-scroll .data-table th,
.card-table-scroll .data-table td{box-sizing:border-box;min-width:0}
.card-table-scroll .data-table .cell-primary,
.card-table-scroll .data-table .cell-text{overflow-wrap:break-word;word-break:normal;white-space:normal}
.card-table-scroll .category-table th:nth-child(1),.card-table-scroll .category-table td:nth-child(1){width:24%}
.card-table-scroll .category-table th:nth-child(2),.card-table-scroll .category-table td:nth-child(2){width:11%}
.card-table-scroll .category-table th:nth-child(3),.card-table-scroll .category-table td:nth-child(3){width:18%}
.card-table-scroll .category-table th:nth-child(4),.card-table-scroll .category-table td:nth-child(4){width:20%}
.card-table-scroll .category-table th:nth-child(5),.card-table-scroll .category-table td:nth-child(5){width:12%}
.card-table-scroll .category-table th:nth-child(6),.card-table-scroll .category-table td:nth-child(6){width:16%}
.card-table-scroll .category-table td:first-child .cell-primary{overflow-wrap:break-word;word-break:normal}
.card-table-scroll .category-table .row-actions{display:inline-flex;flex-wrap:nowrap;max-width:100%;gap:.15rem;justify-content:flex-end;white-space:nowrap}
.card-table-scroll .category-table .row-actions :deep(.p-button){width:2rem;height:2rem;flex:0 0 auto}
.card-table-scroll .category-table td:nth-child(5) :deep(.p-tag){display:inline-flex;min-width:4rem;justify-content:center;padding:.28rem .65rem;border-radius:999px;font-size:.68rem;font-weight:600;white-space:nowrap}
.card-table-scroll .category-table .place-api-list{gap:.2rem;row-gap:.2rem}
.card-table-scroll .category-table .place-api-pill{font-size:.5rem;line-height:1.2;overflow-wrap:break-word;word-break:normal}
.card-table-scroll .category-table td{overflow:visible}
.card-table-scroll .category-table td:nth-child(1) .entity-cell{display:flex;align-items:flex-start;min-width:0;max-width:100%;overflow:visible}
.card-table-scroll .category-table td:nth-child(1) .cell-primary{display:block;min-width:0;max-width:100%;overflow-wrap:break-word;word-break:normal;white-space:normal;line-height:1.35}
.card-table-scroll .category-table td:nth-child(2) :deep(.p-tag){max-width:100%;overflow-wrap:break-word;word-break:normal;white-space:normal}
.card-table-scroll .category-table td:nth-child(3) .cell-text{display:block;min-width:0;max-width:100%;overflow-wrap:break-word;word-break:normal;white-space:normal;line-height:1.4}
.card-table-scroll .category-table td:nth-child(4) .place-api-list{display:flex;min-width:0;max-width:100%;align-items:flex-start;align-content:flex-start;column-gap:.14rem;row-gap:.14rem;overflow:visible}
.card-table-scroll .category-table td:nth-child(4) .place-api-pill{display:inline-flex;min-width:0;max-width:100%;align-items:center;padding:.13rem .29rem;border-radius:7px;overflow-wrap:anywhere;word-break:normal;white-space:normal;line-height:1.18}
.card-table-scroll .category-table td:nth-child(5),.card-table-scroll .category-table td:nth-child(6){vertical-align:top}
.card-table-scroll .category-table th:nth-child(1),.card-table-scroll .category-table td:nth-child(1){width:22%}
.card-table-scroll .category-table th:nth-child(2),.card-table-scroll .category-table td:nth-child(2){width:11%}
.card-table-scroll .category-table th:nth-child(3),.card-table-scroll .category-table td:nth-child(3){width:17%}
.card-table-scroll .category-table th:nth-child(4),.card-table-scroll .category-table td:nth-child(4){width:23%}
.card-table-scroll .category-table th:nth-child(5),.card-table-scroll .category-table td:nth-child(5){width:11%}
.card-table-scroll .category-table th:nth-child(6),.card-table-scroll .category-table td:nth-child(6){width:16%}
.card-table-scroll .category-table td:nth-child(1),.card-table-scroll .category-table td:nth-child(2),.card-table-scroll .category-table td:nth-child(3),.card-table-scroll .category-table td:nth-child(4){vertical-align:top}
.card-table-scroll .category-table td{padding-right:.48rem;padding-left:.48rem}
.card-table-scroll .category-table{border-spacing:0 .32rem}
.card-table-scroll .category-table th,.card-table-scroll .category-table td{padding-right:.4rem;padding-left:.4rem}
.card-table-scroll .category-table th:nth-child(1),.card-table-scroll .category-table td:nth-child(1){width:20%}
.card-table-scroll .category-table th:nth-child(2),.card-table-scroll .category-table td:nth-child(2){width:11%}
.card-table-scroll .category-table th:nth-child(3),.card-table-scroll .category-table td:nth-child(3){width:16%}
.card-table-scroll .category-table th:nth-child(4),.card-table-scroll .category-table td:nth-child(4){width:26%}
.card-table-scroll .category-table th:nth-child(5),.card-table-scroll .category-table td:nth-child(5){width:11%}
.card-table-scroll .category-table th:nth-child(6),.card-table-scroll .category-table td:nth-child(6){width:16%}
.card-table-scroll .category-table td:nth-child(4){padding-right:.35rem;padding-left:.35rem}
.card-table-scroll .category-table td:nth-child(4) .place-api-list{column-gap:.12rem;row-gap:.12rem}
.card-table-scroll .category-table td:nth-child(4) .place-api-pill{padding:.12rem .27rem;font-size:.49rem}
.card-table-scroll .category-table th:nth-child(5),.card-table-scroll .category-table td:nth-child(5){text-align:center}
.card-table-scroll .category-table th:nth-child(6),.card-table-scroll .category-table td:nth-child(6){text-align:center}
.card-table-scroll .category-table td:nth-child(6) .row-actions{justify-content:center}
.table-expand-button{display:flex;align-items:center;justify-content:center;gap:.35rem;width:max-content;max-width:100%;margin:.45rem auto .1rem;padding:.38rem .78rem;border:1px solid #d9e2ec;border-radius:999px;background:#f8fafc;color:#64748b;font-size:.68rem;font-weight:600;line-height:1.2;cursor:pointer;transition:background .15s ease,border-color .15s ease,color .15s ease}
.table-expand-button:hover{border-color:#cbd5e1;background:#fff;color:#334155}
.table-expand-button i{font-size:.62rem}
.card-table-scroll .category-table{border-spacing:0 .24rem}
.card-table-scroll .category-table th,.card-table-scroll .category-table td{padding-top:.36rem;padding-bottom:.36rem}
.card-table-scroll .category-table td:nth-child(4) .place-api-list{column-gap:.08rem;row-gap:.08rem}
.card-table-scroll .category-table td:nth-child(4) .place-api-pill{padding:.1rem .23rem;font-size:.47rem;line-height:1.1}
.master-data-grid{align-items:stretch}
.master-data-grid{grid-auto-rows:auto}
.master-data-card{height:auto;min-height:368px;box-sizing:border-box;align-self:stretch}
.master-card-body{min-height:0;overflow:visible}
.card-table-scroll{overflow:visible}
.card-table-scroll{display:flex;min-height:0;flex:1 1 auto;flex-direction:column}
.card-table-scroll .data-table{flex:0 0 auto}
.card-table-scroll .segment-table th:nth-child(4),.card-table-scroll .segment-table td:nth-child(4){text-align:center}
.card-table-scroll .segment-table th:nth-child(5),.card-table-scroll .segment-table td:nth-child(5){text-align:center}
.card-table-scroll .segment-table td:nth-child(5) .row-actions{justify-content:center}
.card-table-scroll .segment-table th:nth-child(1),.card-table-scroll .segment-table td:nth-child(1){width:23%}
.card-table-scroll .segment-table th:nth-child(2),.card-table-scroll .segment-table td:nth-child(2){width:29%}
.card-table-scroll .segment-table th:nth-child(3),.card-table-scroll .segment-table td:nth-child(3){width:18%}
.card-table-scroll .segment-table th:nth-child(4),.card-table-scroll .segment-table td:nth-child(4){width:14%}
.card-table-scroll .segment-table th:nth-child(5),.card-table-scroll .segment-table td:nth-child(5){width:16%}
.card-table-scroll .segment-table .row-actions{display:flex;flex-wrap:nowrap;gap:.05rem;justify-content:flex-end}
.order-handle-heading{width:7%;padding-right:.25rem!important;padding-left:.25rem!important}
.order-handle-cell{width:7%;padding-right:.25rem!important;padding-left:.25rem!important;text-align:center!important;vertical-align:middle!important}
.order-drag-handle{display:inline-grid;place-items:center;width:1.5rem;height:1.5rem;padding:0;border:1px solid #e2e8f0;border-radius:6px;background:#fff;color:#94a3b8;cursor:grab;transition:background .15s ease,border-color .15s ease,color .15s ease}
.order-drag-handle:hover{border-color:#cbd5e1;background:#f8fafc;color:#475569}
.order-drag-handle:active{cursor:grabbing}
.order-drag-handle i{font-size:.7rem}
.order-row-active td{background:#fff7f7!important}
.card-table-scroll .segment-table.is-reordering th:nth-child(1),.card-table-scroll .segment-table.is-reordering td:nth-child(1){width:7%}
.card-table-scroll .segment-table.is-reordering th:nth-child(2),.card-table-scroll .segment-table.is-reordering td:nth-child(2){width:20%}
.card-table-scroll .segment-table.is-reordering th:nth-child(3),.card-table-scroll .segment-table.is-reordering td:nth-child(3){width:26%}
.card-table-scroll .segment-table.is-reordering th:nth-child(4),.card-table-scroll .segment-table.is-reordering td:nth-child(4){width:17%}
.card-table-scroll .segment-table.is-reordering th:nth-child(5),.card-table-scroll .segment-table.is-reordering td:nth-child(5){width:13%}
.card-table-scroll .segment-table.is-reordering th:nth-child(6),.card-table-scroll .segment-table.is-reordering td:nth-child(6){width:17%}
.card-table-scroll .category-table.is-reordering th:nth-child(1),.card-table-scroll .category-table.is-reordering td:nth-child(1){width:6%}
.card-table-scroll .category-table.is-reordering th:nth-child(2),.card-table-scroll .category-table.is-reordering td:nth-child(2){width:18%}
.card-table-scroll .category-table.is-reordering th:nth-child(3),.card-table-scroll .category-table.is-reordering td:nth-child(3){width:10%}
.card-table-scroll .category-table.is-reordering th:nth-child(4),.card-table-scroll .category-table.is-reordering td:nth-child(4){width:15%}
.card-table-scroll .category-table.is-reordering th:nth-child(5),.card-table-scroll .category-table.is-reordering td:nth-child(5){width:22%}
.card-table-scroll .category-table.is-reordering th:nth-child(6),.card-table-scroll .category-table.is-reordering td:nth-child(6){width:11%}
.card-table-scroll .category-table.is-reordering th:nth-child(7),.card-table-scroll .category-table.is-reordering td:nth-child(7){width:18%}
.card-table-scroll .category-table.is-reordering th,.card-table-scroll .category-table.is-reordering td{padding-right:.34rem;padding-left:.34rem}
.card-table-scroll .category-table.is-reordering td:nth-child(2){vertical-align:top}
.card-table-scroll .category-table.is-reordering td:nth-child(3){vertical-align:top;text-align:left}
.card-table-scroll .category-table.is-reordering td:nth-child(5){vertical-align:top}
.card-table-scroll .category-table.is-reordering td:nth-child(6),.card-table-scroll .category-table.is-reordering td:nth-child(7){vertical-align:middle}
.card-table-scroll .category-table.is-reordering td:nth-child(2) .entity-cell{display:flex;align-items:flex-start;gap:.35rem;min-width:0;max-width:100%}
.card-table-scroll .category-table.is-reordering td:nth-child(2) .cell-primary{min-width:0;max-width:100%;overflow-wrap:anywhere;word-break:normal;line-height:1.35}
.card-table-scroll .category-table.is-reordering td:nth-child(3) :deep(.p-tag){display:inline-flex;max-width:100%;white-space:nowrap}
.card-table-scroll .category-table.is-reordering td:nth-child(5) .place-api-list{column-gap:.1rem;row-gap:.1rem}
.card-table-scroll .category-table.is-reordering td:nth-child(5) .place-api-pill{padding:.11rem .22rem;font-size:.46rem;line-height:1.12}
.card-table-scroll .category-table.is-reordering td:nth-child(6) :deep(.p-tag){min-width:3.7rem;padding:.26rem .4rem;font-size:.65rem}
.card-table-scroll .category-table.is-reordering td:nth-child(7) .row-actions{display:flex;align-items:center;justify-content:center;gap:0}
.card-table-scroll .category-table.is-reordering td:nth-child(7) .row-actions :deep(.p-button){width:1.75rem;height:1.75rem}
.card-table-scroll .segment-table.is-reordering th:nth-child(1),.card-table-scroll .segment-table.is-reordering td:nth-child(1){width:7%}
.card-table-scroll .segment-table.is-reordering th:nth-child(2),.card-table-scroll .segment-table.is-reordering td:nth-child(2){width:22%}
.card-table-scroll .segment-table.is-reordering th:nth-child(3),.card-table-scroll .segment-table.is-reordering td:nth-child(3){width:31%}
.card-table-scroll .segment-table.is-reordering th:nth-child(4),.card-table-scroll .segment-table.is-reordering td:nth-child(4){width:20%}
.card-table-scroll .segment-table.is-reordering th:nth-child(5),.card-table-scroll .segment-table.is-reordering td:nth-child(5){width:20%}
.card-table-scroll .category-table.is-reordering th:nth-child(1),.card-table-scroll .category-table.is-reordering td:nth-child(1){width:6%}
.card-table-scroll .category-table.is-reordering th:nth-child(2),.card-table-scroll .category-table.is-reordering td:nth-child(2){width:22%}
.card-table-scroll .category-table.is-reordering th:nth-child(3),.card-table-scroll .category-table.is-reordering td:nth-child(3){width:11%}
.card-table-scroll .category-table.is-reordering th:nth-child(4),.card-table-scroll .category-table.is-reordering td:nth-child(4){width:18%}
.card-table-scroll .category-table.is-reordering th:nth-child(5),.card-table-scroll .category-table.is-reordering td:nth-child(5){width:31%}
.card-table-scroll .category-table.is-reordering th:nth-child(6),.card-table-scroll .category-table.is-reordering td:nth-child(6){width:12%}
.card-table-scroll .category-table.is-reordering td:nth-child(6){text-align:center}
.card-table-scroll .segment-table.is-reordering th:nth-child(1),.card-table-scroll .segment-table.is-reordering td:nth-child(1){width:7%}
.card-table-scroll .segment-table.is-reordering th:nth-child(2),.card-table-scroll .segment-table.is-reordering td:nth-child(2){width:24%}
.card-table-scroll .segment-table.is-reordering th:nth-child(3),.card-table-scroll .segment-table.is-reordering td:nth-child(3){width:43%}
.card-table-scroll .segment-table.is-reordering th:nth-child(4),.card-table-scroll .segment-table.is-reordering td:nth-child(4){width:26%}
.card-table-scroll .category-table.is-reordering th:nth-child(1),.card-table-scroll .category-table.is-reordering td:nth-child(1){width:6%}
.card-table-scroll .category-table.is-reordering th:nth-child(2),.card-table-scroll .category-table.is-reordering td:nth-child(2){width:23%}
.card-table-scroll .category-table.is-reordering th:nth-child(3),.card-table-scroll .category-table.is-reordering td:nth-child(3){width:11%}
.card-table-scroll .category-table.is-reordering th:nth-child(4),.card-table-scroll .category-table.is-reordering td:nth-child(4){width:20%}
.card-table-scroll .category-table.is-reordering th:nth-child(5),.card-table-scroll .category-table.is-reordering td:nth-child(5){width:40%}
.card-table-scroll .segment-table:not(.is-reordering) th:nth-child(1),.card-table-scroll .segment-table:not(.is-reordering) td:nth-child(1){width:24%}
.card-table-scroll .segment-table:not(.is-reordering) th:nth-child(2),.card-table-scroll .segment-table:not(.is-reordering) td:nth-child(2){width:38%}
.card-table-scroll .segment-table:not(.is-reordering) th:nth-child(3),.card-table-scroll .segment-table:not(.is-reordering) td:nth-child(3){width:22%}
.card-table-scroll .segment-table:not(.is-reordering) th:nth-child(4),.card-table-scroll .segment-table:not(.is-reordering) td:nth-child(4){width:16%}
.card-table-scroll .category-table:not(.is-reordering) th:nth-child(1),.card-table-scroll .category-table:not(.is-reordering) td:nth-child(1){width:23%}
.card-table-scroll .category-table:not(.is-reordering) th:nth-child(2),.card-table-scroll .category-table:not(.is-reordering) td:nth-child(2){width:12%}
.card-table-scroll .category-table:not(.is-reordering) th:nth-child(3),.card-table-scroll .category-table:not(.is-reordering) td:nth-child(3){width:17%}
.card-table-scroll .category-table:not(.is-reordering) th:nth-child(4),.card-table-scroll .category-table:not(.is-reordering) td:nth-child(4){width:30%}
.card-table-scroll .category-table:not(.is-reordering) th:nth-child(5),.card-table-scroll .category-table:not(.is-reordering) td:nth-child(5){width:18%}
.master-data-card{height:100%}
.master-data-grid{align-items:stretch}
.master-data-card{align-self:stretch}
.master-card-body{min-height:285px}
.card-table-scroll{height:100%}
.card-table-scroll .table-expand-button{margin-top:auto}
.master-card-actions.order-actions-active{flex-wrap:nowrap;white-space:nowrap}

@media (max-width: 980px) and (min-width: 761px){
  .master-data-grid{grid-template-columns:minmax(0,1fr)}
}

@media (max-width: 760px){
  .reference-page-header{align-items:flex-start;flex-direction:column}
  .reference-page-header :deep(.p-button){align-self:flex-end}
  .master-data-grid{grid-template-columns:minmax(0,1fr)}
  .master-data-grid{grid-auto-rows:auto}
  .master-data-card{height:auto;min-height:368px}
  .master-card-body{overflow:visible}
  .master-card-header{align-items:stretch;flex-direction:column}
  .master-card-actions{width:100%;justify-content:flex-end}
  .detail-field{grid-template-columns:1fr;gap:.3rem}
}
@media (max-width: 480px){
  .category-dialog :deep(.p-dialog-header),.category-dialog :deep(.p-dialog-content),.category-dialog :deep(.p-dialog-footer){padding-right:1rem;padding-left:1rem}
  .category-dialog :deep(.p-dialog-footer .p-button){flex:1 1 100%;justify-content:center}
}
</style>
