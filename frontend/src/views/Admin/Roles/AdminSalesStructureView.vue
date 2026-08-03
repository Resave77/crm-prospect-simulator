<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import Drawer from 'primevue/drawer'
import Skeleton from 'primevue/skeleton'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import * as adminApi from '../../../api/admin'
import { useAdminStore } from '../../../stores/admin'
import type { AdminUserListItem, SalesRoleLevel, SalesStructureItem } from '../../../types/admin'

const store = useAdminStore()
const toast = useToast()
const error = ref('')
const dialogVisible = ref(false)
const formError = ref('')

const USER_PAGE_SIZE = 50
type ViewMode = 'table' | 'tree'
interface TreeNode {
  item: SalesStructureItem
  children: TreeNode[]
  match: boolean
}
interface HierarchyRow {
  item: SalesStructureItem
  depth: number
  hasChildren: boolean
  isExpanded: boolean
  match: boolean
}

const LEVEL_ORDER: SalesRoleLevel[] = [1, 2, 3, 4]
const LEVEL_LABEL: Record<SalesRoleLevel, string> = {
  1: 'Top-level leaders',
  2: 'Area managers',
  3: 'Supervisors',
  4: 'Individual members',
}
const LEVEL_FALLBACK_DESCRIPTION: Record<SalesRoleLevel, string> = {
  1: 'Top-level sales leader who oversees all descendant sales teams.',
  2: 'Regional or area manager who oversees Levels 3 and 4 in their own team.',
  3: 'Supervisor who oversees Level 4 sales members in their own team.',
  4: 'Individual sales member who sees and manages their own activity.',
}

const activeTab = ref<'assigned' | 'unassigned' | 'all'>('assigned')
const viewMode = ref<ViewMode>('tree')
const showBanner = ref(true)
const selectedRows = ref<SalesStructureItem[]>([])
const selectedHierarchyRows = ref<HierarchyRow[]>([])
const expandedNodeIds = ref<Set<string>>(new Set())
const selectedNode = ref<SalesStructureItem | null>(null)
const drawerVisible = ref(false)

const search = ref('')
const levelFilter = ref<SalesRoleLevel | ''>('')
const statusFilter = ref<'ACTIVE' | 'ENDED' | ''>('')
const systemRoleFilter = ref('')
const orgRoleFilter = ref('')
const parentFilter = ref('')

const salesUsers = ref<AdminUserListItem[]>([])
const salesUsersTotal = ref(0)
const salesUsersPage = ref(0)
const salesUsersLoading = ref(false)
const userSearchInput = ref('')
const userSearch = ref('')
let userSearchTimeout: ReturnType<typeof setTimeout> | null = null

const pickerUsers = ref<AdminUserListItem[]>([])
const pickerTotal = ref(0)
const pickerPage = ref(0)
const pickerSearch = ref('')
const pickerLoading = ref(false)
const dialogStructure = ref<SalesStructureItem[]>([])
const dialogStructureLoading = ref(false)
let pickerSearchTimeout: ReturnType<typeof setTimeout> | null = null
let structureToken = 0
const pendingAssignUser = ref<AdminUserListItem | null>(null)

const form = reactive<{ user: AdminUserListItem | null; salesRoleId: string; parentUserId: string | null; effectiveMonth: string }>({
  user: null,
  salesRoleId: '',
  parentUserId: null,
  effectiveMonth: store.selectedEffectiveMonth,
})

const effectiveDate = computed(() => `${store.selectedEffectiveMonth}-01`)
const dialogEffectiveDate = computed(() => `${form.effectiveMonth}-01`)
const activeRoleOptions = computed(() => store.salesRoles.filter((role) => role.isActive).map((role) => ({
  label: `${role.name} — Level ${role.level}`,
  value: role.id,
  level: role.level as SalesRoleLevel,
  description: role.description || LEVEL_FALLBACK_DESCRIPTION[role.level as SalesRoleLevel],
})))
const selectedRole = computed(() => store.salesRoles.find((role) => role.id === form.salesRoleId) ?? null)
const selectedLevel = computed(() => selectedRole.value?.level ?? null)
const requiredParentLevel = computed<SalesRoleLevel | null>(() => (selectedLevel.value && selectedLevel.value > 1 ? (selectedLevel.value - 1) as SalesRoleLevel : null))
const dialogAssignedUserIds = computed(() => new Set(dialogStructure.value.map((item) => item.userId)))
const parentOptions = computed(() => {
  if (!requiredParentLevel.value) return []
  return dialogStructure.value
    .filter((item) => !item.effectiveTo && item.salesRole.level === requiredParentLevel.value && item.userId !== form.user?.id)
    .map((item) => ({ label: item.salesName, value: item.userId, level: item.salesRole.level }))
})
const needsParent = computed(() => Boolean(selectedLevel.value && selectedLevel.value > 1))
const parentUnavailable = computed(() => needsParent.value && parentOptions.value.length === 0)
const pickerUserOptions = computed(() => pickerUsers.value.map((user) => {
  const already = dialogAssignedUserIds.value.has(user.id)
  return {
    label: `${user.fullName} — ${user.employeeId || user.email}`,
    value: user,
    disabled: already,
    alreadyAssigned: already,
  }
}))
const hasMorePickerUsers = computed(() => pickerUsers.value.length < pickerTotal.value)
const canSubmit = computed(() => {
  if (!form.user || !form.salesRoleId || !form.effectiveMonth) return false
  if (dialogAssignedUserIds.value.has(form.user.id)) return false
  if (selectedLevel.value === 1) return true
  if (!selectedLevel.value) return false
  return Boolean(form.parentUserId && form.parentUserId !== form.user.id && !parentUnavailable.value)
})
const previewParent = computed(() => {
  if (selectedLevel.value === 1 || !form.parentUserId) return null
  return dialogStructure.value.find((item) => item.userId === form.parentUserId) ?? null
})
const previewParentLabel = computed(() => {
  if (selectedLevel.value === 1) return 'No parent — top-level role'
  const parent = previewParent.value
  return parent ? `${parent.salesName} — Level ${parent.salesRole.level}` : '—'
})
const previewMonthLabel = computed(() => formatMonth(form.effectiveMonth))

const levelFilterOptions = [{ label: 'All Levels', value: '' }, ...LEVEL_ORDER.map((level) => ({ label: `Level ${level}`, value: level }))]
const statusOptions = [
  { label: 'All Status', value: '' },
  { label: 'Active', value: 'ACTIVE' },
  { label: 'Ended', value: 'ENDED' },
]
const systemRoleOptions = [
  { label: 'All System Roles', value: '' },
  { label: 'Super Admin', value: 'SUPER_ADMIN' },
  { label: 'Administrator', value: 'ADMINISTRATOR' },
  { label: 'Sales Manager', value: 'SALES_MANAGER' },
  { label: 'Sales Executive', value: 'SALES_EXECUTIVE' },
]
const orgRoleFilterOptions = computed(() => store.salesRoles.filter((role) => role.isActive).map((role) => ({ label: `${role.name} — Level ${role.level}`, value: role.id })))
const parentFilterOptions = computed(() => {
  const seen = new Set<string>()
  const options: { label: string; value: string }[] = []
  for (const item of store.salesStructure) {
    if (item.parentUserId && item.parentName && !seen.has(item.parentUserId)) {
      seen.add(item.parentUserId)
      options.push({ label: item.parentName, value: item.parentUserId })
    }
  }
  return options
})
const hasFilters = computed(() => Boolean(
  search.value.trim() || levelFilter.value || statusFilter.value || systemRoleFilter.value || orgRoleFilter.value || parentFilter.value,
))
const filteredStructure = computed(() => {
  const query = search.value.trim().toLowerCase()
  return store.salesStructure.filter((item) => {
    const status = item.effectiveTo ? 'ENDED' : 'ACTIVE'
    const matchesSearch = !query || [item.salesName, item.parentName ?? '', item.salesRole.name, roleLabel(item.systemRole)]
      .some((value) => value.toLowerCase().includes(query))
    const matchesLevel = !levelFilter.value || item.salesRole.level === levelFilter.value
    const matchesStatus = !statusFilter.value || status === statusFilter.value
    const matchesSystemRole = !systemRoleFilter.value || item.systemRole === systemRoleFilter.value
    const matchesOrgRole = !orgRoleFilter.value || item.salesRole.id === orgRoleFilter.value
    const matchesParent = !parentFilter.value || item.parentUserId === parentFilter.value
    return matchesSearch && matchesLevel && matchesStatus && matchesSystemRole && matchesOrgRole && matchesParent
  })
})

const activeStructure = computed(() => store.salesStructure.filter((item) => !item.effectiveTo))
const assignedThisMonth = computed(() => activeStructure.value.length)
const assignedUserIds = computed(() => new Set(store.salesStructure.map((item) => item.userId)))
const totalActiveSales = computed(() => salesUsers.value.length)
const unassignedUsers = computed(() => salesUsers.value.filter((user) => !assignedUserIds.value.has(user.id)))
const unassignedCount = computed(() => unassignedUsers.value.length)
const levelCounts = computed(() => LEVEL_ORDER.map((level) => ({
  level,
  count: activeStructure.value.filter((item) => item.salesRole.level === level).length,
})))
const usersTruncated = computed(() => salesUsersTotal.value > salesUsers.value.length)
const hasMoreUsers = computed(() => salesUsers.value.length < salesUsersTotal.value)
const currentMonthLabel = computed(() => formatMonth(store.selectedEffectiveMonth))

const treeSource = computed(() => activeStructure.value.filter((item) => {
  const matchesLevel = !levelFilter.value || item.salesRole.level === levelFilter.value
  const matchesSystemRole = !systemRoleFilter.value || item.systemRole === systemRoleFilter.value
  const matchesOrgRole = !orgRoleFilter.value || item.salesRole.id === orgRoleFilter.value
  const matchesParent = !parentFilter.value || item.parentUserId === parentFilter.value
  return matchesLevel && matchesSystemRole && matchesOrgRole && matchesParent
}))
const treeRoots = computed<TreeNode[]>(() => buildTree(treeSource.value, search.value.trim().toLowerCase()))
const flattenedTreeItems = computed(() => flattenTree(treeRoots.value))
const visibleTreeRows = computed(() => flattenVisibleTree(treeRoots.value))
const selectedNodeChildren = computed(() => selectedNode.value ? activeStructure.value.filter((item) => item.parentUserId === selectedNode.value?.userId) : [])

const structureLevelsPresent = computed(() => {
  const present = new Set<number>()
  for (const item of activeStructure.value) present.add(item.salesRole.level)
  return present
})
const nextGuidance = computed(() => {
  const present = structureLevelsPresent.value
  if (!present.has(1)) return { title: 'Start by assigning a Level 1 sales leader.', button: 'Assign Level 1', level: 1 as SalesRoleLevel }
  if (!present.has(2)) return { title: 'Level 1 exists. Assign a Level 2 manager under a Level 1 leader.', button: 'Assign Level 2', level: 2 as SalesRoleLevel }
  if (!present.has(3)) return { title: 'Levels 1 and 2 exist. Assign a Level 3 supervisor under a Level 2 manager.', button: 'Assign Level 3', level: 3 as SalesRoleLevel }
  if (!present.has(4)) return { title: 'Levels 1–3 exist. Assign Level 4 sales members under a Level 3 supervisor.', button: 'Assign Level 4', level: 4 as SalesRoleLevel }
  return null
})

const searchInput = computed({
  get: () => (viewMode.value === 'tree' || activeTab.value === 'assigned' ? search.value : userSearchInput.value),
  set: (val: string) => {
    if (viewMode.value === 'tree' || activeTab.value === 'assigned') search.value = val
    else onUserSearch(val)
  },
})

watch(viewMode, (mode) => {
  if (typeof localStorage !== 'undefined') localStorage.setItem('adminSalesStructureViewMode', mode)
  if (mode === 'tree') expandMatches()
})

watch([treeRoots, search], () => {
  if (viewMode.value === 'tree') expandMatches()
})

function firstDay(month: string) {
  return `${month}-01`
}

// Dedicated Position master remains deferred.
function roleLabel(role: string) {
  if (role === 'SUPER_ADMIN') return 'Super Admin'
  if (role === 'SALES_MANAGER') return 'Sales Manager'
  if (role === 'SALES_EXECUTIVE') return 'Sales Executive'
  return 'Administrator'
}

function positionLabel(item: SalesStructureItem) {
  const source = `${item.salesName} ${item.salesRole.name}`.toLowerCase()
  const region = inferRegion(item)
  if (item.salesRole.level === 1) return 'Director of Sales & Marketing'
  if (source.includes('collector')) return region === '-' ? 'Collector' : `Collector ${region}`
  if (source.includes('merchand')) return region === '-' ? 'Merchandiser' : `Merchandiser ${region}`
  if (source.includes('billing')) return region === '-' ? 'Billing Coordinator' : `Billing Coordinator ${region}`
  if (source.includes('key account')) return region === '-' ? 'Key Account Executive' : `Key Account Executive ${region}`
  if (item.salesRole.level === 2) return region === '-' ? 'Regional Manager' : `Regional Manager ${region}`
  if (item.salesRole.level === 3) return region === '-' ? 'Sales Supervisor' : `Supervisor ${region} Raya`
  if (item.salesRole.level === 4) return region === '-' ? 'Sales Executive' : `Sales Executive ${region}`
  return roleLabel(item.systemRole)
}

function reportsToLabel(item: SalesStructureItem) {
  return item.salesRole.level === 1 ? '-' : item.parentName || '-'
}

function employeeIdLabel(item: SalesStructureItem) {
  return (item as SalesStructureItem & { employeeId?: string }).employeeId || 'Use Account List'
}

function roleSeverity(role: string) {
  if (role === 'SUPER_ADMIN') return 'info'
  if (role === 'SALES_MANAGER') return 'info'
  if (role === 'SALES_EXECUTIVE') return 'success'
  return 'warn'
}

function formatMonth(month: string) {
  if (!month) return ''
  const date = new Date(`${month}-01T00:00:00`)
  if (Number.isNaN(date.getTime())) return month
  return new Intl.DateTimeFormat('en', { month: 'long', year: 'numeric' }).format(date)
}

function statusFor(item: SalesStructureItem) {
  return item.effectiveTo ? 'Ended' : 'Active'
}

function levelClass(level: SalesRoleLevel) {
  return `level-${level}`
}

function levelColorClass(level: SalesRoleLevel) {
  return `tree-level-${level}`
}

function initials(name: string) {
  return name.split(/\s+/).filter(Boolean).slice(0, 2).map((part) => part[0]?.toUpperCase()).join('') || 'SA'
}

function inferRegion(item: SalesStructureItem) {
  const source = `${item.salesName} ${item.parentName ?? ''}`
  const regions = ['Jakarta', 'Bandung', 'Surabaya', 'Semarang', 'Medan', 'Makassar', 'Malang']
  return regions.find((region) => source.toLowerCase().includes(region.toLowerCase())) ?? '-'
}

function nodeMatches(item: SalesStructureItem, query: string) {
  if (!query) return false
  return [
    item.salesName,
    item.parentName ?? '',
    item.salesRole.name,
    String(item.salesRole.level),
    `Level ${item.salesRole.level}`,
    roleLabel(item.systemRole),
    positionLabel(item),
    String((item as SalesStructureItem & { employeeId?: string }).employeeId ?? ''),
  ].some((value) => value.toLowerCase().includes(query))
}

function buildTree(items: SalesStructureItem[], query: string): TreeNode[] {
  const byParent = new Map<string, SalesStructureItem[]>()
  const ids = new Set(items.map((item) => item.userId))
  for (const item of items) {
    const parentKey = item.parentUserId && ids.has(item.parentUserId) ? item.parentUserId : 'root'
    byParent.set(parentKey, [...(byParent.get(parentKey) ?? []), item])
  }
  const build = (parentId: string): TreeNode[] => (byParent.get(parentId) ?? [])
    .sort((a, b) => a.salesRole.level - b.salesRole.level || a.salesName.localeCompare(b.salesName))
    .map((item) => {
      const children = build(item.userId)
      const ownMatch = nodeMatches(item, query)
      return { item, children, match: ownMatch || children.some((child) => child.match) }
    })
    .filter((node) => !query || node.match)
  return build('root')
}

function flattenTree(nodes: TreeNode[]): SalesStructureItem[] {
  return nodes.flatMap((node) => [node.item, ...flattenTree(node.children)])
}

function flattenVisibleTree(nodes: TreeNode[], depth = 0): HierarchyRow[] {
  return nodes.flatMap((node) => {
    const row = {
      item: node.item,
      depth,
      hasChildren: node.children.length > 0,
      isExpanded: isExpanded(node.item.userId),
      match: node.match,
    }
    if (!row.hasChildren || !row.isExpanded) return [row]
    return [row, ...flattenVisibleTree(node.children, depth + 1)]
  })
}

function expandAll() {
  expandedNodeIds.value = new Set(flattenedTreeItems.value.map((item) => item.userId))
}

function collapseAll() {
  expandedNodeIds.value = new Set()
}

function expandMatches() {
  if (!search.value.trim()) return
  expandedNodeIds.value = new Set(flattenedTreeItems.value.map((item) => item.userId))
}

function toggleNode(userId: string) {
  const next = new Set(expandedNodeIds.value)
  if (next.has(userId)) next.delete(userId)
  else next.add(userId)
  expandedNodeIds.value = next
}

function isExpanded(userId: string) {
  return expandedNodeIds.value.has(userId)
}

function openNode(item: SalesStructureItem) {
  selectedNode.value = item
  drawerVisible.value = true
}

function loadStructure() {
  error.value = ''
  selectedRows.value = []
  store.fetchSalesStructure(effectiveDate.value).catch((e) => { error.value = store.errorMessage(e) })
}

async function fetchSalesUsers(page = 1, append = false) {
  if (salesUsersLoading.value) return
  salesUsersLoading.value = true
  try {
    const result = await adminApi.getUsers({ page, limit: USER_PAGE_SIZE, search: userSearch.value, role: '', status: 'ACTIVE' })
    const salesOnly = result.data.filter((user) => user.role === 'SUPER_ADMIN' || user.role === 'SALES_MANAGER' || user.role === 'SALES_EXECUTIVE')
    salesUsers.value = append ? [...salesUsers.value, ...salesOnly] : salesOnly
    salesUsersTotal.value = result.total
    salesUsersPage.value = page
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    salesUsersLoading.value = false
  }
}

function onUserSearch(value: string) {
  userSearchInput.value = value
  if (userSearchTimeout) clearTimeout(userSearchTimeout)
  userSearchTimeout = setTimeout(() => {
    userSearch.value = value.trim()
    void fetchSalesUsers(1, false)
  }, 300)
}

function loadMoreUsers() {
  if (salesUsers.value.length >= salesUsersTotal.value) return
  void fetchSalesUsers(salesUsersPage.value + 1, true)
}

async function loadPickerUsers(reset = false) {
  if (pickerLoading.value) return
  pickerLoading.value = true
  try {
    const page = reset ? 1 : pickerPage.value + 1
    const result = await adminApi.getUsers({ page, limit: USER_PAGE_SIZE, search: pickerSearch.value, role: '', status: 'ACTIVE' })
    const salesOnly = result.data.filter((user) => user.role === 'SUPER_ADMIN' || user.role === 'SALES_MANAGER' || user.role === 'SALES_EXECUTIVE')
    pickerUsers.value = reset ? salesOnly : [...pickerUsers.value, ...salesOnly]
    pickerTotal.value = result.total
    pickerPage.value = page
    if (pendingAssignUser.value) {
      const target = pendingAssignUser.value
      pendingAssignUser.value = null
      if (!pickerUsers.value.some((user) => user.id === target.id)) pickerUsers.value.unshift(target)
      form.user = target
    }
  } catch (e) {
    formError.value = store.errorMessage(e)
  } finally {
    pickerLoading.value = false
  }
}

function onPickerFilter(e: { value?: unknown }) {
  const value = String(e.value ?? '')
  if (pickerSearchTimeout) clearTimeout(pickerSearchTimeout)
  pickerSearchTimeout = setTimeout(() => {
    pickerSearch.value = value.trim()
    void loadPickerUsers(true)
  }, 300)
}

async function loadDialogStructure() {
  const token = ++structureToken
  dialogStructureLoading.value = true
  try {
    const result = await adminApi.getSalesStructure(`${form.effectiveMonth}-01`)
    if (token !== structureToken) return
    dialogStructure.value = result
  } catch (e) {
    if (token !== structureToken) return
    formError.value = store.errorMessage(e)
  } finally {
    if (token === structureToken) dialogStructureLoading.value = false
  }
}

function openAssign(level?: SalesRoleLevel) {
  form.user = null
  form.salesRoleId = ''
  form.parentUserId = null
  form.effectiveMonth = store.selectedEffectiveMonth
  formError.value = ''
  if (level) {
    const match = store.salesRoles.find((role) => role.isActive && role.level === level)
    if (match) form.salesRoleId = match.id
  }
  dialogVisible.value = true
  void loadPickerUsers(true)
  void loadDialogStructure()
}

function assignUnassignedUser(user: AdminUserListItem) {
  openAssign()
  pendingAssignUser.value = user
}

async function submitAssignment() {
  if (!canSubmit.value || !selectedRole.value || !form.user) return
  formError.value = ''
  try {
    await store.createSalesAssignment({
      userId: form.user.id,
      salesRoleId: form.salesRoleId,
      parentUserId: selectedRole.value.level === 1 ? null : form.parentUserId,
      effectiveFrom: firstDay(form.effectiveMonth),
    })
    if (form.effectiveMonth !== store.selectedEffectiveMonth) {
      void store.fetchSalesStructure(effectiveDate.value)
    }
    toast.add({ severity: 'success', summary: 'Assignment Created', detail: 'Sales structure has been refreshed.', life: 3000 })
    dialogVisible.value = false
  } catch (e) {
    formError.value = store.errorMessage(e)
  }
}

function clearFilters() {
  search.value = ''
  levelFilter.value = ''
  statusFilter.value = ''
  systemRoleFilter.value = ''
  orgRoleFilter.value = ''
  parentFilter.value = ''
  userSearch.value = ''
  userSearchInput.value = ''
  void fetchSalesUsers(1, false)
}

function openLevelFilter(level: SalesRoleLevel) {
  levelFilter.value = level
  activeTab.value = 'assigned'
}

watch(() => store.selectedEffectiveMonth, loadStructure)
watch(() => form.salesRoleId, () => { form.parentUserId = null })
watch(() => form.effectiveMonth, () => {
  form.parentUserId = null
  if (dialogVisible.value) void loadDialogStructure()
})

onMounted(async () => {
  const savedMode = typeof localStorage !== 'undefined' ? localStorage.getItem('adminSalesStructureViewMode') : null
  if (savedMode === 'tree' || savedMode === 'table') viewMode.value = savedMode
  error.value = ''
  const results = await Promise.allSettled([store.fetchSalesRoles(), store.fetchSalesStructure(effectiveDate.value), fetchSalesUsers(1)])
  const failed = results.find((result) => result.status === 'rejected')
  if (failed && failed.status === 'rejected') error.value = store.errorMessage(failed.reason)
})
</script>

<template>
  <section class="admin-page">
    <header class="page-heading">
      <div class="page-title-wrapper">
        <span class="eyebrow">Sales Organization</span>
        <h1>Sales Structure</h1>
        <p class="muted">Manage monthly sales team assignments in a flat organizational table.</p>
      </div>
      <Button label="Assign Sales" icon="pi pi-plus" size="small" @click="openAssign()" />
    </header>

    <Message v-if="error" severity="error">{{ error }}</Message>

    <div v-if="showBanner" class="banner-panel">
      <div class="banner-icon"><i class="pi pi-sitemap" /></div>
      <div class="banner-content">
        <strong>Build the hierarchy from top to bottom</strong>
        <ol class="banner-steps">
          <li>Assign Level 1</li>
          <li>Assign Level 2 under Level 1</li>
          <li>Assign Level 3 under Level 2</li>
          <li>Assign Level 4 under Level 3</li>
        </ol>
      </div>
      <Button icon="pi pi-times" text rounded size="small" aria-label="Dismiss banner" @click="showBanner = false" />
    </div>

    <div class="summary-grid">
      <button class="summary-card clickable" type="button" @click="activeTab = 'all'">
        <span>Total Sales</span>
        <strong>{{ totalActiveSales }}</strong>
        <small v-if="usersTruncated">First {{ totalActiveSales }} of {{ salesUsersTotal }} loaded — use search to narrow.</small>
      </button>
      <button class="summary-card clickable" type="button" @click="activeTab = 'assigned'">
        <span>Assigned</span>
        <strong>{{ assignedThisMonth }}</strong>
        <small>{{ formatMonth(store.selectedEffectiveMonth) }}</small>
      </button>
      <button class="summary-card clickable" type="button" @click="activeTab = 'unassigned'">
        <span>Unassigned Users</span>
        <strong>{{ unassignedCount }}</strong>
        <small>Active sales accounts not yet assigned</small>
      </button>
      <button v-for="item in levelCounts" :key="item.level" class="summary-card clickable" type="button" @click="openLevelFilter(item.level)">
        <span>Level {{ item.level }}</span>
        <strong>{{ item.count }}</strong>
        <small>{{ LEVEL_LABEL[item.level] }}</small>
      </button>
    </div>

    <div class="view-switcher" aria-label="Sales structure view mode">
      <button class="view-btn" :class="{ active: viewMode === 'table' }" type="button" @click="viewMode = 'table'">
        <i class="pi pi-table" /> Table
      </button>
      <button class="view-btn" :class="{ active: viewMode === 'tree' }" type="button" @click="viewMode = 'tree'">
        <i class="pi pi-sitemap" /> Hierarchy Table
      </button>
    </div>

    <div v-if="viewMode === 'table'" class="tab-bar">
      <button class="tab-btn" :class="{ active: activeTab === 'assigned' }" type="button" @click="activeTab = 'assigned'">Assigned <span class="tab-count">{{ assignedThisMonth }}</span></button>
      <button class="tab-btn" :class="{ active: activeTab === 'unassigned' }" type="button" @click="activeTab = 'unassigned'">Unassigned <span class="tab-count">{{ unassignedCount }}</span></button>
      <button class="tab-btn" :class="{ active: activeTab === 'all' }" type="button" @click="activeTab = 'all'">All Sales Users <span class="tab-count">{{ totalActiveSales }}</span></button>
    </div>

    <div class="toolbar-panel">
      <div class="search-field">
        <i class="pi pi-search" />
        <input
          v-model="searchInput"
          type="search"
          :placeholder="viewMode === 'tree' ? 'Search tree by employee, region, position, role, or manager...' : activeTab === 'assigned' ? 'Search sales, reports to, position, or role...' : 'Search active users by name, email, or employee ID...'"
        />
      </div>
      <div class="filter-field">
        <label>Effective Month</label>
        <input v-model="store.selectedEffectiveMonth" type="month" />
      </div>
      <Button label="Reset" icon="pi pi-replay" severity="secondary" text size="small" @click="clearFilters" />
    </div>

    <div v-if="viewMode === 'tree' || activeTab === 'assigned'" class="filter-panel">
      <div class="filter-field">
        <label>Role Level</label>
        <Select v-model="levelFilter" :options="levelFilterOptions" optionLabel="label" optionValue="value" />
      </div>
      <div class="filter-field">
        <label>System Role</label>
        <Select v-model="systemRoleFilter" :options="systemRoleOptions" optionLabel="label" optionValue="value" />
      </div>
      <div class="filter-field">
        <label>Organizational Role</label>
        <Select v-model="orgRoleFilter" :options="orgRoleFilterOptions" optionLabel="label" optionValue="value" filter placeholder="All roles" />
      </div>
      <div class="filter-field">
        <label>Parent</label>
        <Select v-model="parentFilter" :options="parentFilterOptions" optionLabel="label" optionValue="value" filter placeholder="All parents" />
      </div>
      <div class="filter-field">
        <label>Status</label>
        <Select v-model="statusFilter" :options="statusOptions" optionLabel="label" optionValue="value" />
      </div>
    </div>

    <div v-if="viewMode === 'tree'" class="tree-panel">
      <div class="tree-toolbar">
        <div>
          <strong>Hierarchy Table</strong>
          <span>{{ assignedThisMonth }} assigned sales for {{ currentMonthLabel }}</span>
        </div>
        <div class="tree-actions">
          <Button label="Expand All" icon="pi pi-plus-circle" severity="secondary" text size="small" @click="expandAll" />
          <Button label="Collapse All" icon="pi pi-minus-circle" severity="secondary" text size="small" @click="collapseAll" />
        </div>
      </div>

      <div v-if="store.salesStructureLoading && !store.salesStructure.length" class="skeleton-area">
        <Skeleton v-for="n in 6" :key="n" class="skeleton-row" />
      </div>
      <div v-else-if="!treeRoots.length" class="empty-state">
        <div class="empty-icon"><i class="pi pi-sitemap" /></div>
        <strong>No assignments for this month</strong>
        <span>{{ hasFilters ? 'Adjust your search or filters to view the organization.' : 'Assign the first leader to begin the hierarchy.' }}</span>
        <Button :label="hasFilters ? 'Reset Filters' : 'Assign First Leader'" :icon="hasFilters ? 'pi pi-replay' : 'pi pi-plus'" size="small" @click="hasFilters ? clearFilters() : openAssign(1)" />
      </div>
      <DataTable
        v-else
        v-model:selection="selectedHierarchyRows"
        :value="visibleTreeRows"
        :loading="store.salesStructureLoading"
        dataKey="item.assignmentId"
        scrollable
        class="hierarchy-table"
      >
        <Column selectionMode="multiple" headerStyle="width: 44px" />
        <Column header="Role Level" :style="{ width: '110px' }">
          <template #body="{ data }">
            <span class="hier-level" :class="levelClass(data.item.salesRole.level)">{{ data.item.salesRole.level }}</span>
          </template>
        </Column>
        <Column header="Sales Name" :style="{ minWidth: '300px' }">
          <template #body="{ data }">
            <div class="hier-name-cell" :style="{ '--depth': data.depth }">
              <span v-if="data.depth" class="hier-connector" />
              <button
                v-if="data.hasChildren"
                class="hier-expander"
                type="button"
                :title="data.isExpanded ? 'Collapse row' : 'Expand row'"
                @click.stop="toggleNode(data.item.userId)"
              >
                <i :class="data.isExpanded ? 'pi pi-chevron-down' : 'pi pi-chevron-right'" />
              </button>
              <span v-else class="hier-expander-spacer" />
              <button class="hier-name-button" type="button" @click="openNode(data.item)">
                <strong class="sales-name" :class="levelClass(data.item.salesRole.level)">{{ data.item.salesName }}</strong>
                <small>{{ employeeIdLabel(data.item) }}</small>
              </button>
            </div>
          </template>
        </Column>
        <Column header="Reports To" :style="{ minWidth: '180px' }">
          <template #body="{ data }">{{ reportsToLabel(data.item) }}</template>
        </Column>
        <Column header="Position" :style="{ minWidth: '220px' }">
          <template #body="{ data }">{{ positionLabel(data.item) }}</template>
        </Column>
        <Column header="Role Name" :style="{ minWidth: '210px' }">
          <template #body="{ data }"><span class="role-name-label">{{ data.item.salesRole.name }}</span></template>
        </Column>
        <Column header="Status" :style="{ width: '110px' }">
          <template #body="{ data }"><span class="compact-status" :class="{ inactive: data.item.effectiveTo }">{{ statusFor(data.item) }}</span></template>
        </Column>
        <Column header="Actions" :style="{ width: '90px' }">
          <template #body="{ data }">
            <button class="more-action" type="button" title="More actions" @click="openNode(data.item)">
              <i class="pi pi-ellipsis-v" />
            </button>
          </template>
        </Column>
      </DataTable>
    </div>

    <div v-else class="table-panel">
      <div v-if="store.salesStructureLoading && !store.salesStructure.length" class="skeleton-area">
        <Skeleton v-for="n in 8" :key="n" class="skeleton-row" />
      </div>

      <DataTable
        v-else-if="activeTab === 'assigned'"
        v-model:selection="selectedRows"
        :value="filteredStructure"
        :loading="store.salesStructureLoading"
        dataKey="assignmentId"
        scrollable
        paginator
        :rows="20"
        :rowsPerPageOptions="[20, 50, 100]"
        sortMode="multiple"
        paginatorTemplate="CurrentPageReport FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink RowsPerPageDropdown"
        currentPageReportTemplate="Showing {first} to {last} of {totalRecords} assignments"
      >
        <template #empty>
          <div class="empty-state">
            <div v-if="hasFilters" class="empty-body">
              <div class="empty-icon"><i class="pi pi-filter" /></div>
              <strong>No assignments match your filters</strong>
              <span>Adjust your filters to view assignments.</span>
              <Button label="Reset Filters" icon="pi pi-replay" severity="secondary" text size="small" @click="clearFilters" />
            </div>
            <template v-else>
              <div class="empty-icon"><i class="pi pi-sitemap" /></div>
              <template v-if="nextGuidance">
                <strong>{{ nextGuidance.title }}</strong>
                <span>Build the hierarchy from the top down to unlock higher levels.</span>
                <Button :label="nextGuidance.button" icon="pi pi-plus" size="small" @click="openAssign(nextGuidance.level)" />
              </template>
              <template v-else>
                <strong>No assignments for this month</strong>
                <span>Create an assignment for the selected month.</span>
                <Button label="Assign First Leader" icon="pi pi-plus" size="small" @click="openAssign(1)" />
              </template>
            </template>
          </div>
        </template>
        <Column selectionMode="multiple" headerStyle="width: 3rem" />
        <Column field="salesRole.level" header="Role Level" sortable :style="{ width: '130px' }">
          <template #body="{ data }"><Tag :value="`Level ${data.salesRole.level}`" severity="info" /></template>
        </Column>
        <Column field="salesName" header="Sales Name" sortable :style="{ minWidth: '220px' }">
          <template #body="{ data }"><span class="sales-name" :class="levelClass(data.salesRole.level)">{{ data.salesName }}</span></template>
        </Column>
        <Column field="parentName" header="Reports To" sortable :style="{ minWidth: '180px' }">
          <template #body="{ data }">{{ data.parentName || '-' }}</template>
        </Column>
        <Column field="systemRole" header="Position" sortable :style="{ minWidth: '160px' }">
          <template #body="{ data }">{{ positionLabel(data) }}</template>
        </Column>
        <Column field="salesRole.name" header="Role Name" sortable :style="{ minWidth: '190px' }">
          <template #body="{ data }">{{ data.salesRole.name }}</template>
        </Column>
        <Column header="Status" sortable sortField="effectiveTo" :style="{ width: '120px' }">
          <template #body="{ data }"><Tag :value="statusFor(data)" :severity="data.effectiveTo ? 'secondary' : 'success'" /></template>
        </Column>
        <Column header="Actions" :style="{ width: '120px' }">
          <template #body>
            <Button icon="pi pi-ellipsis-h" text rounded size="small" disabled title="Assignment movement is deferred" />
          </template>
        </Column>
      </DataTable>

      <DataTable
        v-else
        :value="activeTab === 'unassigned' ? unassignedUsers : salesUsers"
        :loading="salesUsersLoading"
        dataKey="id"
        scrollable
      >
        <template #empty>
          <div class="empty-state">
            <div class="empty-icon"><i class="pi pi-users" /></div>
            <strong v-if="activeTab === 'unassigned'">All active sales users are assigned for this month.</strong>
            <strong v-else>No active sales users found.</strong>
            <span v-if="userSearch || activeTab === 'all'">Adjust your search to view results.</span>
          </div>
        </template>
        <Column header="Employee ID" :style="{ width: '160px' }">
          <template #body="{ data }"><code class="code-tag">{{ data.employeeId || '—' }}</code></template>
        </Column>
        <Column field="fullName" header="Sales Name" sortable :style="{ minWidth: '220px' }">
          <template #body="{ data }"><span class="sales-name">{{ data.fullName }}</span></template>
        </Column>
        <Column field="email" header="Email" sortable :style="{ minWidth: '220px' }">
          <template #body="{ data }"><span class="cell-text">{{ data.email }}</span></template>
        </Column>
        <Column field="role" header="System Role" sortable :style="{ width: '160px' }">
          <template #body="{ data }"><Tag :value="roleLabel(data.role)" :severity="roleSeverity(data.role)" /></template>
        </Column>
        <Column field="status" header="Status" sortable :style="{ width: '110px' }">
          <template #body="{ data }"><Tag :value="data.status" :severity="data.status === 'ACTIVE' ? 'success' : 'secondary'" /></template>
        </Column>
        <Column v-if="activeTab === 'all'" header="Assignment" :style="{ width: '140px' }">
          <template #body="{ data }">
            <Tag :value="assignedUserIds.has(data.id) ? 'Assigned' : 'Unassigned'" :severity="assignedUserIds.has(data.id) ? 'info' : 'secondary'" />
          </template>
        </Column>
        <Column header="Action" :style="{ width: '130px' }">
          <template #body="{ data }">
            <Button v-if="!assignedUserIds.has(data.id)" label="Assign" icon="pi pi-user-plus" text rounded size="small" @click="assignUnassignedUser(data)" />
            <span v-else class="cell-hint">Assigned</span>
          </template>
        </Column>
      </DataTable>

      <div v-if="activeTab !== 'assigned' && hasMoreUsers" class="load-more-row">
        <span class="cell-hint">Showing {{ salesUsers.length }} of {{ salesUsersTotal }} active accounts.</span>
        <Button label="Load more" icon="pi pi-chevron-down" severity="secondary" text size="small" :loading="salesUsersLoading" @click="loadMoreUsers" />
      </div>
    </div>

    <Drawer v-model:visible="drawerVisible" position="right" header="Assignment Detail" class="detail-drawer">
      <div v-if="selectedNode" class="drawer-content">
        <div class="drawer-person">
          <span class="drawer-avatar">{{ initials(selectedNode.salesName) }}</span>
          <div>
            <strong>{{ selectedNode.salesName }}</strong>
            <span>{{ positionLabel(selectedNode) }}</span>
          </div>
        </div>
        <div class="detail-grid">
          <div><span>Employee ID</span><strong>Use Account List</strong></div>
          <div><span>Organizational Role</span><strong>{{ selectedNode.salesRole.name }}</strong></div>
          <div><span>Role Level</span><strong>Level {{ selectedNode.salesRole.level }}</strong></div>
          <div><span>Reports To</span><strong>{{ reportsToLabel(selectedNode) }}</strong></div>
          <div><span>Direct Reports</span><strong>{{ selectedNodeChildren.length }}</strong></div>
          <div><span>Effective Month</span><strong>{{ currentMonthLabel }}</strong></div>
          <div><span>Status</span><strong>{{ statusFor(selectedNode) }}</strong></div>
        </div>
        <div class="drawer-actions">
          <Button label="Move Assignment" icon="pi pi-arrow-right-arrow-left" disabled />
          <Button label="View History" icon="pi pi-history" severity="secondary" disabled />
        </div>
      </div>
    </Drawer>

    <Dialog v-model:visible="dialogVisible" header="Assign Sales to Team" modal :style="{ width: 'min(620px, 94vw)' }">
      <div class="dialog-form">
        <Message v-if="formError" severity="error">{{ formError }}</Message>

        <div class="assign-section">
          <div class="assign-section-title"><span class="assign-step">1</span>Select Sales User</div>
          <Select
            v-model="form.user"
            :options="pickerUserOptions"
            optionLabel="label"
            :optionDisabled="(opt) => opt.disabled"
            :loading="pickerLoading"
            filter
            filterPlaceholder="Search by name, employee ID, or email"
            placeholder="Search active sales user"
            @filter="onPickerFilter"
          >
            <template #option="{ option }">
              <div class="user-option" :class="{ 'is-disabled': option.disabled }">
                <div class="user-option-main">
                  <span class="user-option-name">{{ option.user.fullName }}</span>
                  <Tag :value="roleLabel(option.user.role)" :severity="roleSeverity(option.user.role)" size="small" />
                </div>
                <div class="user-option-sub">
                  <span v-if="option.user.employeeId">{{ option.user.employeeId }}</span>
                  <span>{{ option.user.email }}</span>
                </div>
                <span v-if="option.alreadyAssigned" class="already-chip"><i class="pi pi-check-circle" /> Already assigned this month</span>
              </div>
            </template>
            <template #footer>
              <div class="picker-footer">
                <span v-if="pickerTotal" class="cell-hint">{{ pickerUsers.length }} of {{ pickerTotal }} active accounts</span>
                <Button v-if="hasMorePickerUsers" label="Load more" icon="pi pi-chevron-down" text size="small" :loading="pickerLoading" @click="loadPickerUsers(false)" />
              </div>
            </template>
            <template #emptyfilter>
              <div class="picker-empty">No matching active sales users.</div>
            </template>
          </Select>
          <span class="section-note">Only active sales accounts can be assigned. Users already assigned in the selected month are disabled.</span>
        </div>

        <div class="assign-section">
          <div class="assign-section-title"><span class="assign-step">2</span>Select Organizational Role</div>
          <Select v-model="form.salesRoleId" :options="activeRoleOptions" optionLabel="label" optionValue="value" filter placeholder="Select an active role">
            <template #option="{ option }">
              <div class="role-option">
                <span class="role-option-name">{{ option.label }}</span>
                <span class="role-option-desc">{{ option.description }}</span>
              </div>
            </template>
          </Select>
          <span class="section-note">Only active roles can be assigned.</span>
        </div>

        <div class="assign-section">
          <div class="assign-section-title"><span class="assign-step">3</span>Select Reporting Parent</div>
          <div v-if="selectedLevel === 1" class="parent-info">
            <i class="pi pi-info-circle" />
            <span>Level 1 is the top-level role and does not require an approver/parent.</span>
          </div>
          <template v-else-if="selectedLevel">
            <Select
              v-model="form.parentUserId"
              :options="parentOptions"
              optionLabel="label"
              optionValue="value"
              filter
              :loading="dialogStructureLoading"
              :placeholder="`Select parent from Level ${requiredParentLevel}...`"
            />
            <Message v-if="parentUnavailable" severity="warn">No eligible Level {{ requiredParentLevel }} parent exists for {{ previewMonthLabel || form.effectiveMonth }}. Assign the required upper level first.</Message>
            <span class="section-note">A Level {{ selectedLevel }} user must report to an active Level {{ requiredParentLevel }} user in the same month.</span>
          </template>
          <div v-else class="parent-info muted">
            <i class="pi pi-info-circle" />
            <span>Select an organizational role first to see eligible parents.</span>
          </div>
        </div>

        <div class="assign-section">
          <div class="assign-section-title"><span class="assign-step">4</span>Effective Month</div>
          <div class="form-field">
            <input v-model="form.effectiveMonth" type="month" />
          </div>
          <span class="section-note">Changes apply from the first day of that month ({{ dialogEffectiveDate }}).</span>
        </div>

        <div class="assign-section">
          <div class="assign-section-title"><span class="assign-step">5</span>Assignment Preview</div>
          <div class="preview-card">
            <div class="preview-row"><span>Sales</span><strong>{{ form.user?.fullName || '—' }}<template v-if="form.user?.employeeId"><span class="preview-sub"> · {{ form.user.employeeId }}</span></template></strong></div>
            <div class="preview-row"><span>Organizational Role</span><strong>{{ selectedRole ? `${selectedRole.name} — Level ${selectedRole.level}` : '—' }}</strong></div>
            <div class="preview-row"><span>Reports To</span><strong>{{ previewParentLabel }}</strong></div>
            <div class="preview-row"><span>Effective From</span><strong>{{ previewMonthLabel || '—' }}</strong></div>
          </div>
        </div>
      </div>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="dialogVisible = false" :disabled="store.savingSalesAssignment" />
        <Button label="Assign" icon="pi pi-check" :loading="store.savingSalesAssignment" :disabled="!canSubmit || store.savingSalesAssignment" @click="submitAssignment" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.admin-page { display: flex; flex-direction: column; gap: 1.25rem; padding: 1.75rem 2rem; min-height: 100vh; background: #f7f9fb; }
.page-heading { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 1rem; }
.page-title-wrapper .eyebrow { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; color: #0b7766; }
h1 { margin: 0.2rem 0 0.2rem; font-size: 1.6rem; color: #0f172a; }
.muted { margin: 0; color: #7c8798; font-size: 0.85rem; }

.banner-panel { display: flex; gap: 0.9rem; align-items: flex-start; padding: 0.9rem 1.1rem; background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 12px; }
.banner-icon { width: 34px; height: 34px; flex: 0 0 auto; display: grid; place-content: center; border-radius: 10px; background: #dbeafe; color: #2563eb; }
.banner-content { display: grid; gap: 0.35rem; flex: 1; min-width: 0; }
.banner-content strong { color: #1e40af; font-size: 0.9rem; }
.banner-steps { margin: 0; padding-left: 1.1rem; display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 0.3rem 0.8rem; color: #1e40af; font-size: 0.78rem; line-height: 1.45; }
.banner-steps li::marker { color: #2563eb; }

.summary-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(135px, 1fr)); gap: 0.65rem; }
.summary-card, .filter-panel, .table-panel, .toolbar-panel, .tree-panel { background: #fff; border: 1px solid #edf1f6; border-radius: 12px; box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03); }
.summary-card { padding: 0.72rem 0.8rem; display: grid; gap: 0.12rem; text-align: left; font: inherit; cursor: pointer; transition: border-color 140ms ease, background 140ms ease; }
.summary-card:hover { border-color: #bfdbfe; background: #fbfdff; }
.summary-card span { color: #64748b; font-size: 0.68rem; font-weight: 750; text-transform: uppercase; }
.summary-card strong { color: #0f172a; font-size: 1.1rem; }
.summary-card small { color: #94a3b8; font-size: 0.68rem; line-height: 1.3; }

.view-switcher { display: inline-flex; width: fit-content; padding: 0.25rem; gap: 0.2rem; background: #ffffff; border: 1px solid #edf1f6; border-radius: 12px; box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03); }
.view-btn { display: inline-flex; align-items: center; gap: 0.45rem; padding: 0.55rem 0.85rem; border: 0; border-radius: 9px; background: transparent; color: #64748b; font: inherit; font-size: 0.82rem; font-weight: 750; cursor: pointer; }
.view-btn:hover { background: #f8fafc; color: #0f172a; }
.view-btn.active { background: #eaf5f2; color: #0b7766; }

.tab-bar { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.tab-btn { display: inline-flex; align-items: center; gap: 0.45rem; padding: 0.55rem 1rem; border: 1px solid #edf1f6; border-radius: 10px; background: #fff; color: #64748b; font: inherit; font-size: 0.82rem; font-weight: 650; cursor: pointer; transition: background 140ms ease, color 140ms ease, border-color 140ms ease; }
.tab-btn:hover { border-color: #cbd5e1; color: #0f172a; }
.tab-btn.active { background: #0b7766; border-color: #0b7766; color: #fff; }
.tab-count { font-size: 0.7rem; padding: 0.05rem 0.45rem; border-radius: 999px; background: rgba(11, 119, 102, 0.12); color: #0b7766; }
.tab-btn.active .tab-count { background: rgba(255, 255, 255, 0.22); color: #fff; }

.toolbar-panel { display: grid; grid-template-columns: minmax(240px, 1fr) 180px auto; gap: 0.75rem; align-items: end; padding: 1rem; }
.filter-panel { display: grid; grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); gap: 0.75rem; align-items: end; padding: 1rem; }
.search-field { display: flex; align-items: center; gap: 0.6rem; border: 1px solid #dbe3ee; border-radius: 8px; padding: 0 0.75rem; min-height: 2.55rem; background: #ffffff; }
.search-field i { color: #94a3b8; }
.search-field input { flex: 1; border: 0; outline: 0; background: transparent; color: #0f172a; font: inherit; min-width: 0; }
.filter-field, .form-field { display: grid; gap: 0.35rem; }
.filter-field label, .form-field label { font-size: 0.72rem; font-weight: 700; text-transform: uppercase; color: #64748b; }
input[type='month'] { border: 1px solid #dbe3ee; border-radius: 6px; padding: 0.62rem; font: inherit; width: 100%; background: #ffffff; color: #0f172a; }

.table-panel { overflow-x: auto; }
.table-panel :deep(.p-datatable), .table-panel :deep(.p-datatable-table), .table-panel :deep(.p-datatable-table-container) { background: #ffffff; color: #0f172a; }
.table-panel :deep(.p-datatable-thead > tr > th) { background: #f8fafc; color: #475569; font-size: 0.68rem; text-transform: uppercase; border-color: #edf1f6; }
.table-panel :deep(.p-datatable-tbody > tr), .table-panel :deep(.p-datatable-tbody > tr > td) { background: #ffffff; color: #1e293b; border-color: #f1f4f8; }
.table-panel :deep(.p-datatable-tbody > tr > td) { white-space: nowrap; }
.table-panel :deep(.p-datatable-tbody > tr:hover > td) { background: #f8fafc; }
.table-panel :deep(.p-paginator) { background: #ffffff; border-color: #edf1f6; color: #475569; }
.table-panel :deep(.p-paginator-current) { color: #64748b; }
.table-panel :deep(.p-datatable-loading-overlay) { background: rgba(255, 255, 255, 0.72); }
.tree-panel { overflow-x: auto; }
.tree-toolbar { display: flex; justify-content: space-between; gap: 1rem; align-items: center; padding: 0.75rem 0.9rem; border-bottom: 1px solid #edf1f6; }
.tree-toolbar > div:first-child { display: grid; gap: 0.15rem; }
.tree-toolbar strong { color: #0f172a; font-size: 0.92rem; }
.tree-toolbar span { color: #94a3b8; font-size: 0.76rem; }
.tree-actions { display: flex; flex-wrap: wrap; gap: 0.35rem; justify-content: flex-end; }
.hierarchy-table { min-width: 1120px; }
.hierarchy-table :deep(.p-datatable-table),
.hierarchy-table :deep(.p-datatable-table-container) { background: #ffffff; }
.hierarchy-table :deep(.p-datatable-thead > tr > th) { background: #f3f5f8; color: #475569; font-size: 0.66rem; font-weight: 800; text-transform: uppercase; letter-spacing: 0.04em; border-color: #e5eaf0; padding: 0.5rem 0.7rem; white-space: nowrap; }
.hierarchy-table :deep(.p-datatable-tbody > tr > td) { height: 48px; background: #ffffff; color: #1e293b; border-color: #edf1f6; padding: 0.42rem 0.7rem; font-size: 0.8rem; vertical-align: middle; }
.hierarchy-table :deep(.p-datatable-tbody > tr:hover > td) { background: #f8fbff; }
.hier-level { font-weight: 850; font-size: 0.82rem; }
.hier-level.level-1, .sales-name.level-1 { color: #1d4ed8; }
.hier-level.level-2, .sales-name.level-2 { color: #047857; }
.hier-level.level-3, .sales-name.level-3 { color: #9a3412; }
.hier-level.level-4, .sales-name.level-4 { color: #a16207; }
.hier-name-cell { position: relative; display: flex; align-items: center; gap: 0.28rem; padding-left: calc(var(--depth) * 1.45rem); min-height: 34px; }
.hier-connector { position: absolute; left: calc((var(--depth) * 1.45rem) - 0.75rem); top: 0; bottom: 0; width: 0.8rem; border-left: 1px solid #dbe3ee; border-bottom: 1px solid #dbe3ee; transform: translateY(-50%); }
.hier-expander, .hier-expander-spacer { width: 1.45rem; height: 1.45rem; flex: 0 0 auto; }
.hier-expander { display: grid; place-content: center; border: 1px solid #dbe3ee; border-radius: 6px; background: #ffffff; color: #64748b; cursor: pointer; }
.hier-expander:hover { background: #f1f5f9; color: #0f172a; }
.hier-expander i { font-size: 0.7rem; }
.hier-name-button { display: grid; gap: 0.02rem; min-width: 0; border: 0; background: transparent; padding: 0; text-align: left; font: inherit; cursor: pointer; }
.hier-name-button:hover .sales-name { text-decoration: underline; text-underline-offset: 2px; }
.hier-name-button small { color: #94a3b8; font-size: 0.68rem; }
.role-name-label { display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; max-width: 220px; color: #334155; line-height: 1.3; }
.compact-status { display: inline-flex; align-items: center; gap: 0.32rem; color: #047857; font-size: 0.72rem; font-weight: 800; }
.compact-status::before { content: ''; width: 0.42rem; height: 0.42rem; border-radius: 999px; background: #10b981; }
.compact-status.inactive { color: #64748b; }
.compact-status.inactive::before { background: #94a3b8; }
.more-action { width: 2rem; height: 2rem; display: grid; place-content: center; border: 1px solid #e2e8f0; border-radius: 999px; background: #ffffff; color: #64748b; cursor: pointer; }
.more-action:hover { background: #f8fafc; color: #0f172a; }
.tree-list { min-width: 720px; padding: 0.45rem 0.85rem 0.8rem; display: grid; gap: 0; }
.tree-branch { display: grid; gap: 0; }
.tree-row { display: flex; align-items: stretch; gap: 0.35rem; position: relative; }
.tree-row::before { content: ''; width: 0.75rem; height: 1px; background: #dbe3ee; position: absolute; left: 1.75rem; top: 50%; }
.tree-children { margin-left: 1.45rem; padding-left: 1.05rem; border-left: 1px solid #dbe3ee; display: grid; gap: 0; }
.node-toggle, .node-toggle-spacer { width: 1.25rem; height: 1.25rem; margin-top: 0.76rem; flex: 0 0 auto; z-index: 1; }
.node-toggle { display: grid; place-content: center; border: 1px solid #dbe3ee; border-radius: 6px; background: #ffffff; color: #64748b; cursor: pointer; }
.node-toggle:hover { background: #f8fafc; color: #0f172a; }
.org-node { flex: 1; min-width: 0; display: grid; grid-template-columns: auto minmax(190px, 1fr) minmax(340px, auto); gap: 0.7rem; align-items: center; padding: 0.48rem 0.7rem; border: 0; border-bottom: 1px solid #edf1f6; border-left: 3px solid #dbe3ee; background: #ffffff; text-align: left; font: inherit; cursor: pointer; transition: background 140ms ease, border-color 140ms ease; }
.org-node:hover { background: #f8fafc; }
.org-node.matched { background: #fff7ed; }
.tree-level-1 { border-left-color: #1e3a8a; }
.tree-level-2 { border-left-color: #4f9f77; }
.tree-level-3 { border-left-color: #b45309; }
.tree-level-4 { border-left-color: #d6a11d; }
.node-avatar { width: 1.9rem; height: 1.9rem; display: grid; place-content: center; border-radius: 999px; background: #f1f5f9; color: #475569; font-size: 0.68rem; font-weight: 850; }
.node-main { display: grid; gap: 0.08rem; min-width: 0; }
.node-main strong { font-size: 0.82rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.tree-level-1 .node-main strong { color: #1d4ed8; }
.tree-level-2 .node-main strong { color: #047857; }
.tree-level-3 .node-main strong { color: #9a3412; }
.tree-level-4 .node-main strong { color: #a16207; }
.node-main span { color: #64748b; font-size: 0.72rem; }
.node-meta { display: grid; grid-template-columns: minmax(120px, 1.1fr) minmax(120px, 1fr) auto auto; align-items: center; gap: 0.7rem; color: #334155; font-size: 0.74rem; }
.node-meta > span:not(.level-indicator):not(.status-dot) { display: grid; gap: 0.06rem; min-width: 0; }
.node-meta em { color: #94a3b8; font-style: normal; font-size: 0.62rem; font-weight: 800; text-transform: uppercase; }
.level-indicator { display: inline-flex; justify-content: center; min-width: 4rem; padding: 0.14rem 0.45rem; border-radius: 999px; background: #f8fafc; color: #475569; border: 1px solid #e2e8f0; font-size: 0.68rem; font-weight: 800; }
.status-dot { display: inline-flex; align-items: center; gap: 0.3rem; color: #047857; font-size: 0.7rem; font-weight: 800; }
.status-dot::before { content: ''; width: 0.42rem; height: 0.42rem; border-radius: 999px; background: #10b981; }
.status-dot.inactive { color: #64748b; }
.status-dot.inactive::before { background: #94a3b8; }
.sales-name { font-weight: 750; }
.sales-name.level-1 { color: #2563eb; }
.sales-name.level-2 { color: #059669; }
.sales-name.level-3 { color: #ea580c; }
.sales-name.level-4 { color: #b45309; }
.skeleton-area { padding: 0.6rem 1rem 1rem; }
.skeleton-row { height: 3rem; margin-top: 0.65rem; border-radius: 10px; }
.empty-state { min-height: 240px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.45rem; padding: 2rem; text-align: center; color: #8492a6; }
.empty-body { display: flex; flex-direction: column; align-items: center; gap: 0.45rem; }
.empty-icon { width: 52px; height: 52px; display: grid; place-content: center; border-radius: 14px; background: #f8fafc; color: #94a3b8; }
.empty-state strong { color: #0f172a; }
.code-tag { display: inline-block; font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace; font-size: 0.76rem; font-weight: 600; padding: 0.18rem 0.55rem; border-radius: 6px; background: #f1f5f9; color: #475569; }
.cell-text { color: #475569; }
.cell-hint { color: #94a3b8; font-size: 0.75rem; }
.load-more-row { display: flex; align-items: center; justify-content: center; gap: 1rem; padding: 0.9rem 1rem; border-top: 1px solid #f1f4f8; }
.load-more-row .cell-hint { margin: 0; }

.dialog-form { display: grid; gap: 1.1rem; max-height: 68vh; overflow-y: auto; padding-right: 0.25rem; }
.assign-section { display: grid; gap: 0.55rem; }
.assign-section-title { display: flex; align-items: center; gap: 0.55rem; font-weight: 750; color: #0f172a; font-size: 0.86rem; }
.assign-step { width: 22px; height: 22px; display: grid; place-content: center; border-radius: 999px; background: #0b7766; color: #fff; font-size: 0.7rem; font-weight: 700; }
.section-note { color: #8492a6; font-size: 0.75rem; }
.parent-info { display: flex; gap: 0.5rem; align-items: center; padding: 0.6rem 0.75rem; border: 1px solid #bfdbfe; border-radius: 8px; background: #eff6ff; color: #1e40af; font-size: 0.8rem; line-height: 1.4; }
.parent-info.muted { border-color: #e2e8f0; background: #f8fafc; color: #64748b; }
.user-option { display: grid; gap: 0.15rem; padding: 0.25rem 0; }
.user-option.is-disabled { opacity: 0.55; }
.user-option-main { display: flex; align-items: center; gap: 0.5rem; }
.user-option-name { font-weight: 650; color: #0f172a; font-size: 0.84rem; }
.user-option-sub { display: flex; gap: 0.7rem; color: #94a3b8; font-size: 0.74rem; }
.already-chip { display: inline-flex; align-items: center; gap: 0.3rem; color: #b45309; font-size: 0.7rem; font-weight: 700; }
.picker-footer { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; padding: 0.5rem 0.75rem; border-top: 1px solid #edf1f6; }
.picker-empty { padding: 0.75rem; text-align: center; color: #94a3b8; font-size: 0.8rem; }
.role-option { display: grid; gap: 0.15rem; padding: 0.2rem 0; }
.role-option-name { font-weight: 650; color: #0f172a; font-size: 0.84rem; }
.role-option-desc { color: #94a3b8; font-size: 0.74rem; }
.preview-card { display: grid; gap: 0.55rem; padding: 0.85rem 1rem; border: 1px solid #e2e8f0; border-radius: 10px; background: #f8fafc; }
.preview-row { display: flex; justify-content: space-between; gap: 1rem; align-items: baseline; }
.preview-row span { color: #64748b; font-size: 0.72rem; font-weight: 700; text-transform: uppercase; }
.preview-row strong { color: #0f172a; font-size: 0.84rem; font-weight: 700; text-align: right; }
.preview-sub { color: #94a3b8; font-weight: 500; text-transform: none; }
.detail-drawer :deep(.p-drawer) { background: #ffffff; color: #0f172a; width: min(420px, 94vw); }
.detail-drawer :deep(.p-drawer-header), .detail-drawer :deep(.p-drawer-content) { background: #ffffff; color: #0f172a; }
.detail-drawer :deep(.p-drawer-header) { border-bottom: 1px solid #edf1f6; padding: 0.9rem 1rem; }
.drawer-content { display: grid; gap: 1rem; }
.drawer-person { display: flex; align-items: center; gap: 0.75rem; padding: 0.75rem 0; border-bottom: 1px solid #edf1f6; }
.drawer-avatar { width: 2.45rem; height: 2.45rem; display: grid; place-content: center; border-radius: 999px; background: #eff6ff; color: #1d4ed8; font-weight: 850; }
.drawer-person div { display: grid; gap: 0.15rem; }
.drawer-person strong { color: #0f172a; }
.drawer-person span { color: #64748b; font-size: 0.78rem; }
.detail-grid { display: grid; gap: 0; border: 1px solid #edf1f6; border-radius: 10px; overflow: hidden; }
.detail-grid > div { display: grid; grid-template-columns: 130px 1fr; gap: 0.65rem; align-items: center; padding: 0.65rem 0.75rem; border-bottom: 1px solid #f1f4f8; }
.detail-grid > div:last-child { border-bottom: 0; }
.detail-grid span { color: #64748b; font-size: 0.68rem; font-weight: 750; text-transform: uppercase; }
.detail-grid strong { color: #0f172a; font-size: 0.86rem; }
.drawer-actions { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem; }

:deep(.p-inputtext), :deep(.p-select), :deep(.p-textarea) { background: #ffffff; color: #0f172a; border-color: #dbe3ee; }
:deep(.p-select-label), :deep(.p-select-dropdown), :deep(.p-inputtext::placeholder) { color: #64748b; }
:deep(.p-dialog), :deep(.p-dialog-header), :deep(.p-dialog-content), :deep(.p-dialog-footer) { background: #ffffff; color: #0f172a; }
:deep(.p-select-overlay), :deep(.p-select-list), :deep(.p-select-option) { background: #ffffff; color: #0f172a; }
:deep(.p-select-option.p-select-option-selected), :deep(.p-select-option:hover) { background: #f1f5f9; color: #0f172a; }
@media (max-width: 1100px) { .toolbar-panel { grid-template-columns: 1fr 1fr; } .search-field { grid-column: 1 / -1; } .banner-steps { grid-template-columns: 1fr 1fr; } }
@media (max-width: 900px) { .tree-list { min-width: 0; } .tree-children { margin-left: 0.85rem; padding-left: 0.75rem; } .org-node { grid-template-columns: auto minmax(0, 1fr); } .node-meta { grid-column: 2; grid-template-columns: 1fr 1fr; gap: 0.4rem 0.65rem; } }
@media (max-width: 768px) { .admin-page { padding: 1rem; } .toolbar-panel { grid-template-columns: 1fr; } .banner-steps { grid-template-columns: 1fr; } .view-switcher { width: 100%; } .view-btn { flex: 1; justify-content: center; } .detail-grid > div { grid-template-columns: 1fr; gap: 0.2rem; } }
</style>
