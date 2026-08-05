<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Drawer from 'primevue/drawer'
import Dialog from 'primevue/dialog'
import * as adminApi from '../../../api/admin'
import AssignSalesDialog from '../../../components/admin/sales-structure/AssignSalesDialog.vue'
import SalesStructureFlatTable from '../../../components/admin/sales-structure/SalesStructureFlatTable.vue'
import SalesStructureHierarchyTable from '../../../components/admin/sales-structure/SalesStructureHierarchyTable.vue'
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
const unassignDialogVisible = ref(false)
const unassignTarget = ref<SalesStructureItem | null>(null)

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

const activeRoleOptions = computed(() =>
  store.salesRoles
    .filter((role) => role.isActive && role.name.trim().toLowerCase() !== 'super admin')
    .map((role) => ({
      label: `${role.name} — Level ${role.level}`,
      value: role.id,
      level: role.level as SalesRoleLevel,
      description:
        role.description ||
        LEVEL_FALLBACK_DESCRIPTION[role.level as SalesRoleLevel],
    })),
)
const selectedRole = computed(() => store.salesRoles.find((role) => role.id === form.salesRoleId) ?? null)
const selectedRoleLabel = computed(() => selectedRole.value ? `${selectedRole.value.name} — Level ${selectedRole.value.level}` : '')
const selectedLevel = computed(() => selectedRole.value?.level ?? null)
const requiredParentLevel = computed<SalesRoleLevel | null>(() => (selectedLevel.value && selectedLevel.value > 1 ? (selectedLevel.value - 1) as SalesRoleLevel : null))
const dialogAssignedUserIds = computed(() => new Set(dialogStructure.value.map((item) => item.userId)))
const parentOptions = computed(() => {
  if (!requiredParentLevel.value) return []
  return dialogStructure.value
    .filter((item) => isAssignmentEffectiveForMonth(item, form.effectiveMonth) && item.salesRole.level === requiredParentLevel.value && item.userId !== form.user?.id)
    .map((item) => ({ label: item.salesName, value: item.userId, level: item.salesRole.level }))
})
const needsParent = computed(() => Boolean(selectedLevel.value && selectedLevel.value > 1))
const parentUnavailable = computed(() => needsParent.value && parentOptions.value.length === 0)
const pickerUserOptions = computed(() => pickerUsers.value.map((user) => {
  const already = dialogAssignedUserIds.value.has(user.id)
  return {
    label: `${user.fullName} — ${user.employeeId || user.email}`,
    value: user,
    user,
    disabled: already,
    alreadyAssigned: already,
  }
}))
const hasMorePickerUsers = computed(() => pickerUsers.value.length < pickerTotal.value)
const canSubmit = computed(() => {
  if (!form.user || !form.salesRoleId || !form.effectiveMonth || !selectedRole.value) return false
  if (dialogAssignedUserIds.value.has(form.user.id)) return false
  if (form.user.role === 'SUPER_ADMIN') return false
  if (selectedLevel.value === 1) return true
  if (!selectedLevel.value) return false

  return Boolean(
    form.parentUserId &&
      form.parentUserId !== form.user.id &&
      !parentUnavailable.value,
  )
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
  { label: 'All Account Roles', value: '' },
  { label: 'Sales Manager', value: 'SALES_MANAGER' },
  { label: 'Sales Executive', value: 'SALES_EXECUTIVE' },
]
const orgRoleFilterOptions = computed(() =>
  store.salesRoles
    .filter((role) => role.isActive && role.name.trim().toLowerCase() !== 'super admin')
    .map((role) => ({ label: `${role.name} — Level ${role.level}`, value: role.id })),
)
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
    const status = isAssignmentEffective(item) ? 'ACTIVE' : 'ENDED'
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

const activeStructure = computed(() => store.salesStructure.filter((item) => isAssignmentEffective(item)))
const assignedThisMonth = computed(() => activeStructure.value.length)
const assignedUserIds = computed(() => new Set(activeStructure.value.map((item) => item.userId)))
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

function isAssignmentEffectiveForMonth(
  item: SalesStructureItem,
  month: string,
): boolean {
  if (!month) return false

  const selectedDate = `${month}-01`
  const effectiveFrom = item.effectiveFrom.slice(0, 10)
  const effectiveTo = item.effectiveTo?.slice(0, 10) ?? null

  return (
    effectiveFrom <= selectedDate &&
    (effectiveTo === null || effectiveTo >= selectedDate)
  )
}

function isAssignmentEffective(
  item: SalesStructureItem,
  month = store.selectedEffectiveMonth,
): boolean {
  return isAssignmentEffectiveForMonth(item, month)
}

function statusFor(item: SalesStructureItem): string {
  return isAssignmentEffective(item) ? 'Active' : 'Ended'
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

function isProtectedAssignment(item: SalesStructureItem): boolean {
  return item.systemRole === 'SUPER_ADMIN'
}

function openMoveAssignment(item: SalesStructureItem) {
  if (isProtectedAssignment(item)) {
    toast.add({
      severity: 'warn',
      summary: 'Protected assignment',
      detail: 'The Super Admin root assignment cannot be moved.',
      life: 3500,
    })
    return
  }

  selectedNode.value = item
  drawerVisible.value = true
  toast.add({
    severity: 'info',
    summary: 'Move Assignment',
    detail: 'Select Move Assignment from the detail panel to continue.',
    life: 3000,
  })
}

function openPromotion(item: SalesStructureItem) {
  if (isProtectedAssignment(item)) {
    toast.add({
      severity: 'warn',
      summary: 'Protected assignment',
      detail: 'The Super Admin root assignment cannot be promoted.',
      life: 3500,
    })
    return
  }

  if (item.salesRole.level <= 1) {
    toast.add({
      severity: 'info',
      summary: 'Highest level',
      detail: 'This assignment is already at the highest hierarchy level.',
      life: 3000,
    })
    return
  }

  selectedNode.value = item
  drawerVisible.value = true
  toast.add({
    severity: 'info',
    summary: 'Promotion',
    detail: `Prepare a monthly move from Level ${item.salesRole.level} to Level ${item.salesRole.level - 1}.`,
    life: 3500,
  })
}

function openDemotion(item: SalesStructureItem) {
  if (isProtectedAssignment(item)) {
    toast.add({
      severity: 'warn',
      summary: 'Protected assignment',
      detail: 'The Super Admin root assignment cannot be demoted.',
      life: 3500,
    })
    return
  }

  if (item.salesRole.level >= 4) {
    toast.add({
      severity: 'info',
      summary: 'Lowest level',
      detail: 'A Level 4 assignment cannot be demoted further.',
      life: 3000,
    })
    return
  }

  selectedNode.value = item
  drawerVisible.value = true
  toast.add({
    severity: 'info',
    summary: 'Demotion',
    detail: `Prepare a monthly move from Level ${item.salesRole.level} to Level ${item.salesRole.level + 1}.`,
    life: 3500,
  })
}

function lastDayOfMonth(month: string) {
  const [year, monthNumber] = month.split('-').map(Number)
  const date = new Date(year, monthNumber, 0)
  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, '0'),
    String(date.getDate()).padStart(2, '0'),
  ].join('-')
}
function nextMonthValue(month: string): string {
  const [year, monthNumber] = month.split('-').map(Number)

  const nextDate = new Date(year, monthNumber, 1)

  return [
    nextDate.getFullYear(),
    String(nextDate.getMonth() + 1).padStart(2, '0'),
  ].join('-')
}

function monthDisplayLabel(month: string): string {
  const [year, monthNumber] = month.split('-').map(Number)

  return new Intl.DateTimeFormat('en-US', {
    month: 'long',
    year: 'numeric',
  }).format(new Date(year, monthNumber - 1, 1))
}
function requestUnassign(item: SalesStructureItem) {
  unassignTarget.value = item
  unassignDialogVisible.value = true
}

function closeUnassignDialog() {
  unassignDialogVisible.value = false
  unassignTarget.value = null
}

async function confirmUnassign() {
  if (!unassignTarget.value || store.endingSalesAssignment) return

  error.value = ''

  const target = unassignTarget.value
  const currentMonth = store.selectedEffectiveMonth
  const effectiveTo = lastDayOfMonth(currentMonth)
  const nextMonth = nextMonthValue(currentMonth)
  const nextEffectiveDate = `${nextMonth}-01`

  try {
    await store.endSalesAssignment(
      target.assignmentId,
      effectiveTo,
      nextEffectiveDate,
    )

    // Setelah assignment bulan berjalan ditutup,
    // pindah ke bulan berikutnya agar user langsung terlihat Unassigned.
    store.selectedEffectiveMonth = nextMonth

    await Promise.all([
      store.fetchSalesStructure(nextEffectiveDate),
      fetchSalesUsers(1, false),
    ])

    activeTab.value = 'unassigned'
    selectedRows.value = []
    selectedHierarchyRows.value = []

    toast.add({
      severity: 'success',
      summary: 'Sales Unassigned',
      detail: `${target.salesName} is unassigned starting ${monthDisplayLabel(nextMonth)}.`,
      life: 4000,
    })

    closeUnassignDialog()
  } catch (e) {
    error.value = store.errorMessage(e)

    toast.add({
      severity: 'error',
      summary: 'Unassign Failed',
      detail: error.value,
      life: 5000,
    })
  }
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
    const salesOnly = result.data.filter((user) => user.role === 'SALES_MANAGER' || user.role === 'SALES_EXECUTIVE')
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
    const salesOnly = result.data.filter((user) => user.role === 'SALES_MANAGER' || user.role === 'SALES_EXECUTIVE')
    pickerUsers.value = reset ? salesOnly : [...pickerUsers.value, ...salesOnly]
    pickerTotal.value = result.total
    pickerPage.value = page
    if (pendingAssignUser.value) {
      const target = pendingAssignUser.value
      pendingAssignUser.value = null
      if (!pickerUsers.value.some((user) => user.id === target.id)) pickerUsers.value.unshift(target)
      form.user = target
      if (target.organizationalRole?.id) form.salesRoleId = target.organizationalRole.id
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
    const match = activeRoleOptions.value.find((role) => role.level === level)
    if (match) form.salesRoleId = match.value
  }

  dialogVisible.value = true
  void loadPickerUsers(true)
  void loadDialogStructure()
}

function assignUnassignedUser(user: AdminUserListItem) {
  // Set the pending user before opening the dialog so the async picker load
  // can reliably restore the selection when its response arrives.
  pendingAssignUser.value = user
  openAssign()
}

function selectAssignUser(user: AdminUserListItem | null) {
  form.user = user
  if (user?.organizationalRole?.id) form.salesRoleId = user.organizationalRole.id
}

async function submitAssignment() {
  if (!canSubmit.value || !selectedRole.value || !form.user) return
  formError.value = ''
  try {
    const assignedMonth = form.effectiveMonth
    await store.createSalesAssignment({
      userId: form.user.id,
      salesRoleId: form.salesRoleId,
      parentUserId: selectedRole.value.level === 1 ? null : form.parentUserId,
      effectiveFrom: firstDay(assignedMonth),
    })

    // Keep the page and all assignment-dependent lists synchronized with the
    // month that was just assigned. The store action may already refresh the
    // structure, but this explicit refresh keeps this view correct regardless
    // of store implementation details.
    if (store.selectedEffectiveMonth !== assignedMonth) {
      store.selectedEffectiveMonth = assignedMonth
    }
    await Promise.all([
      store.fetchSalesStructure(firstDay(assignedMonth)),
      fetchSalesUsers(1, false),
    ])

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
  viewMode.value = 'tree'
  error.value = ''
  const results = await Promise.allSettled([store.fetchSalesRoles(), store.fetchSalesStructure(effectiveDate.value), fetchSalesUsers(1)])
  const failed = results.find((result) => result.status === 'rejected')
  if (failed && failed.status === 'rejected') error.value = store.errorMessage(failed.reason)
})
</script>

<template>
  <section class="admin-page">
    <header class="page-heading compact-heading">
      <div class="page-title-wrapper">
        <span class="eyebrow">Sales Organization</span>
        <h1>Sales Structure</h1>
        <p class="muted">Manage monthly reporting lines and team assignments.</p>
      </div>

      <div class="heading-stats">
        <button type="button" @click="activeTab = 'all'">
          <span>Total</span><strong>{{ totalActiveSales }}</strong>
        </button>
        <button type="button" @click="activeTab = 'assigned'">
          <span>Assigned</span><strong>{{ assignedThisMonth }}</strong>
        </button>
        <button type="button" @click="activeTab = 'unassigned'">
          <span>Unassigned</span><strong>{{ unassignedCount }}</strong>
        </button>
        <button
          v-for="item in levelCounts"
          :key="item.level"
          type="button"
          @click="openLevelFilter(item.level)"
        >
          <span>L{{ item.level }}</span><strong>{{ item.count }}</strong>
        </button>
      </div>

      <Button
        label="Assign Sales"
        icon="pi pi-plus"
        size="small"
        @click="openAssign()"
      />
    </header>

    <Message v-if="error" severity="error">{{ error }}</Message>

    <div class="view-switcher" aria-label="Sales structure view mode">
      <button class="view-btn" :class="{ active: viewMode === 'tree' }" type="button" @click="viewMode = 'tree'">
        <i class="pi pi-sitemap" /> Hierarchy
      </button>
      <button class="view-btn" :class="{ active: viewMode === 'table' }" type="button" @click="viewMode = 'table'">
        <i class="pi pi-table" /> Account Table
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
        <label>Account Role</label>
        <Select v-model="systemRoleFilter" :options="systemRoleOptions" optionLabel="label" optionValue="value" />
      </div>
      <div class="filter-field">
        <label>Organizational Role</label>
        <Select v-model="orgRoleFilter" :options="orgRoleFilterOptions" optionLabel="label" optionValue="value" filter placeholder="All roles" />
      </div>
      <div class="filter-field">
        <label>Reports To</label>
        <Select v-model="parentFilter" :options="parentFilterOptions" optionLabel="label" optionValue="value" filter placeholder="All parents" />
      </div>

    </div>

    <SalesStructureHierarchyTable
      v-if="viewMode === 'tree'"
      v-model:selection="selectedHierarchyRows"
      :rows="visibleTreeRows"
      :loading="store.salesStructureLoading"
      :assigned-count="assignedThisMonth"
      :month-label="currentMonthLabel"
      :has-filters="hasFilters"
      :level-class="levelClass"
      :employee-id-label="employeeIdLabel"
      :reports-to-label="reportsToLabel"
      :position-label="positionLabel"
      :status-for="statusFor"
      :is-protected="isProtectedAssignment"
      :can-end-assignment="false"
      @expand-all="expandAll"
      @collapse-all="collapseAll"
      @toggle-node="toggleNode"
      @open-node="openNode"
      @assign-first="openAssign()"
      @reset-filters="clearFilters"
      @move-assignment="openMoveAssignment"
      @promote-assignment="openPromotion"
      @demote-assignment="openDemotion"
    />

    <SalesStructureFlatTable
      v-else
      v-model:selection="selectedRows"
      :active-tab="activeTab"
      :assigned-rows="filteredStructure"
      :sales-users="salesUsers"
      :unassigned-users="unassignedUsers"
      :sales-structure-loading="store.salesStructureLoading"
      :sales-users-loading="salesUsersLoading"
      :sales-structure-empty="!store.salesStructure.length"
      :assigned-user-ids="assignedUserIds"
      :has-filters="hasFilters"
      :user-search="userSearch"
      :has-more-users="hasMoreUsers"
      :sales-users-total="salesUsersTotal"
      :next-guidance="nextGuidance"
      :level-class="levelClass"
      :position-label="positionLabel"
      :status-for="statusFor"
      :role-label="roleLabel"
      :role-severity="roleSeverity"
      @assign-user="assignUnassignedUser"
      @assign-first="openAssign"
      @load-more="loadMoreUsers"
      @reset-filters="clearFilters"
      @open-detail="openNode"
      @move-assignment="openMoveAssignment"
      @unassign="requestUnassign"
    />
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
          <Button
            label="Move Assignment"
            icon="pi pi-arrow-right-arrow-left"
            :disabled="isProtectedAssignment(selectedNode) || !isAssignmentEffective(selectedNode)"
            @click="openMoveAssignment(selectedNode)"
          />
          <Button label="View History" icon="pi pi-history" severity="secondary" disabled />
        </div>
      </div>
    </Drawer>

    <Dialog
      v-model:visible="unassignDialogVisible"
      modal
      :draggable="false"
      :dismissableMask="!store.endingSalesAssignment"
      header="Unassign Sales"
      class="unassign-dialog"
      :style="{ width: 'min(470px, calc(100vw - 2rem))' }"
      @hide="closeUnassignDialog"
    >
      <div v-if="unassignTarget" class="unassign-content">
        <div class="warning-icon">
          <i class="pi pi-user-minus" />
        </div>
        <div>
          <strong>Unassign {{ unassignTarget.salesName }}?</strong>
          <p>
  The current assignment remains valid through
  <b>{{ lastDayOfMonth(store.selectedEffectiveMonth) }}</b>.
</p>

<p>
  <b>{{ unassignTarget.salesName }}</b> will appear in the
  Unassigned list starting
  <b>{{ monthDisplayLabel(nextMonthValue(store.selectedEffectiveMonth)) }}</b>.
  Assignment history will remain available.
</p>
        </div>
      </div>

      <template #footer>
        <Button
          label="Cancel"
          severity="secondary"
          outlined
          :disabled="store.endingSalesAssignment"
          @click="closeUnassignDialog"
        />
        <Button
          label="Unassign"
          icon="pi pi-user-minus"
          severity="danger"
          :loading="store.endingSalesAssignment"
          @click="confirmUnassign"
        />
      </template>
    </Dialog>

    <AssignSalesDialog
      v-model:visible="dialogVisible"
      :user="form.user"
      :sales-role-id="form.salesRoleId"
      :parent-user-id="form.parentUserId"
      :effective-month="form.effectiveMonth"
      :picker-user-options="pickerUserOptions"
      :role-options="activeRoleOptions"
      :parent-options="parentOptions"
      :picker-users-count="pickerUsers.length"
      :picker-total="pickerTotal"
      :has-more-picker-users="hasMorePickerUsers"
      :picker-loading="pickerLoading"
      :dialog-structure-loading="dialogStructureLoading"
      :saving="store.savingSalesAssignment"
      :error="formError"
      :selected-role-label="selectedRoleLabel"
      :selected-level="selectedLevel"
      :required-parent-level="requiredParentLevel"
      :parent-unavailable="parentUnavailable"
      :preview-parent-label="previewParentLabel"
      :preview-month-label="previewMonthLabel"
      :dialog-effective-date="dialogEffectiveDate"
      :can-submit="canSubmit"
      :role-label="roleLabel"
      :role-severity="roleSeverity"
      @update:user="selectAssignUser"
      @update:sales-role-id="form.salesRoleId = $event"
      @update:parent-user-id="form.parentUserId = $event"
      @update:effective-month="form.effectiveMonth = $event"
      @filter-users="onPickerFilter"
      @load-more-users="loadPickerUsers(false)"
      @submit="submitAssignment"
      @cancel="dialogVisible = false"
    />
  </section>
</template>

<style scoped>
.admin-page { display: flex; flex-direction: column; gap: 1.25rem; width: 100%; min-width: 0; padding: 1.75rem 2rem; min-height: 100vh; overflow-x: hidden; background: #f7f9fb; }
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

.dialog-form { display: grid; gap: 1.1rem; max-height: min(68vh, 720px); overflow-y: auto; overscroll-behavior: contain; padding-right: 0.35rem; }
:deep(.p-dialog) { margin: 1rem; max-height: calc(100vh - 2rem); }
:deep(.p-dialog-content) { overflow: hidden; }
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

/* ===== Presentation-ready Sales Structure redesign ===== */
.admin-page {
  gap: 0;
  padding: 0;
  background: #f8fafc;
}

.page-heading {
  align-items: center;
  padding: 0.9rem 1rem;
  border-bottom: 1px solid #e5eaf0;
  background: #fff;
}

.page-title-wrapper {
  display: grid;
  gap: 0.08rem;
}

.page-title-wrapper .eyebrow {
  color: #64748b;
  font-size: 0.62rem;
  letter-spacing: 0.08em;
}

h1 {
  margin: 0;
  font-size: 1.15rem;
}

.muted {
  font-size: 0.7rem;
  color: #94a3b8;
}

.summary-grid {
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 0;
  border-bottom: 1px solid #e5eaf0;
  background: #fff;
}

.summary-card {
  min-height: 68px;
  padding: 0.7rem 0.85rem;
  border: 0;
  border-right: 1px solid #edf1f6;
  border-radius: 0;
  box-shadow: none;
}

.summary-card:last-child {
  border-right: 0;
}

.summary-card:hover {
  background: #f8fbff;
}

.summary-card span {
  font-size: 0.6rem;
  color: #94a3b8;
}

.summary-card strong {
  font-size: 0.98rem;
}

.summary-card small {
  display: none;
}

.view-switcher {
  margin: 0.8rem 1rem 0;
  border-radius: 9px;
  box-shadow: none;
}

.view-btn {
  padding: 0.45rem 0.72rem;
  font-size: 0.74rem;
}

.view-btn.active {
  background: #eff6ff;
  color: #1d4ed8;
}

.tab-bar {
  margin: 0.6rem 1rem 0;
}

.tab-btn {
  padding: 0.42rem 0.65rem;
  border-radius: 8px;
  font-size: 0.72rem;
}

.tab-btn.active {
  border-color: #bfdbfe;
  background: #eff6ff;
  color: #1d4ed8;
}

.tab-btn.active .tab-count {
  background: #dbeafe;
  color: #1d4ed8;
}

.toolbar-panel {
  grid-template-columns: minmax(280px, 1fr) 180px auto;
  margin: 0.65rem 1rem 0;
  padding: 0.72rem;
  border-radius: 10px;
  box-shadow: none;
}

.filter-panel {
  grid-template-columns: repeat(4, minmax(0, 1fr));
  margin: 0.55rem 1rem 0;
  padding: 0.72rem;
  border-radius: 10px;
  box-shadow: none;
}

.search-field {
  min-height: 40px;
  border-radius: 8px;
}

.filter-field label {
  font-size: 0.61rem;
}

input[type='month'] {
  height: 40px;
  padding: 0 0.65rem;
  border-radius: 8px;
  font-size: 0.74rem;
}

.admin-page > .tree-panel,
.admin-page > .table-panel {
  margin: 0.7rem 1rem 1rem;
  overflow: hidden;
  border-radius: 10px;
  box-shadow: none;
}

.admin-page :deep(.tree-panel) {
  overflow-x: hidden !important;
}

.admin-page :deep(.hierarchy-table) {
  width: 100% !important;
  min-width: 0 !important;
}

.admin-page :deep(.hierarchy-table .p-datatable-table-container),
.admin-page :deep(.table-panel .p-datatable-table-container) {
  overflow-x: hidden !important;
}

.admin-page :deep(.hierarchy-table .p-datatable-table),
.admin-page :deep(.table-panel .p-datatable-table) {
  width: 100% !important;
  min-width: 0 !important;
  table-layout: fixed;
}

.admin-page :deep(.p-datatable-thead > tr > th),
.admin-page :deep(.p-datatable-tbody > tr > td) {
  overflow: hidden;
  padding: 0.55rem 0.62rem;
  text-overflow: ellipsis;
}

.admin-page :deep(.p-datatable-thead > tr > th) {
  background: #f8fafc;
  font-size: 0.62rem;
}

.admin-page :deep(.p-datatable-tbody > tr > td) {
  font-size: 0.73rem;
  white-space: normal;
}

.detail-drawer :deep(.p-drawer) {
  width: min(410px, 94vw);
}

.detail-grid > div {
  grid-template-columns: 120px minmax(0, 1fr);
}

.drawer-actions {
  grid-template-columns: 1fr;
}

@media (max-width: 1100px) {
  .summary-grid {
    grid-template-columns: repeat(4, 1fr);
  }

  .filter-panel {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 760px) {
  .page-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .summary-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .toolbar-panel,
  .filter-panel {
    grid-template-columns: 1fr;
    margin-left: 0.7rem;
    margin-right: 0.7rem;
  }

  .view-switcher,
  .tab-bar {
    margin-left: 0.7rem;
    margin-right: 0.7rem;
  }

  .admin-page > .tree-panel,
  .admin-page > .table-panel {
    margin-left: 0.7rem;
    margin-right: 0.7rem;
  }
}

.compact-heading {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) auto auto;
  align-items: center;
  gap: 1rem;
  padding: 0.85rem 1rem;
  border: 1px solid #e5eaf0;
  border-radius: 12px;
  background: #fff;
}
.heading-stats {
  display: grid;
  grid-template-columns: repeat(7, minmax(58px, auto));
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 9px;
  background: #f8fafc;
}
.heading-stats button {
  display: grid;
  gap: 0.05rem;
  min-width: 64px;
  padding: 0.45rem 0.55rem;
  border: 0;
  border-right: 1px solid #e5eaf0;
  background: transparent;
  text-align: center;
  cursor: pointer;
}
.heading-stats button:last-child {
  border-right: 0;
}
.heading-stats button:hover {
  background: #eff6ff;
}
.heading-stats span {
  color: #94a3b8;
  font-size: 0.56rem;
  font-weight: 800;
  text-transform: uppercase;
}
.heading-stats strong {
  color: #0f172a;
  font-size: 0.82rem;
}
.unassign-dialog :deep(.p-dialog) {
  border-radius: 14px;
}
.unassign-content {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  gap: 0.8rem;
  align-items: start;
}
.warning-icon {
  display: grid;
  width: 44px;
  height: 44px;
  place-content: center;
  border-radius: 12px;
  background: #fef2f2;
  color: #dc2626;
}
.unassign-content strong {
  color: #0f172a;
  font-size: 0.88rem;
}
.unassign-content p {
  margin: 0.3rem 0 0;
  color: #64748b;
  font-size: 0.74rem;
  line-height: 1.55;
}
@media (max-width: 1100px) {
  .compact-heading {
    grid-template-columns: 1fr auto;
  }
  .heading-stats {
    grid-column: 1 / -1;
    grid-row: 2;
  }
}
@media (max-width: 720px) {
  .compact-heading {
    grid-template-columns: 1fr;
  }
  .heading-stats {
    grid-column: auto;
    grid-row: auto;
    grid-template-columns: repeat(4, 1fr);
    width: 100%;
  }
  .compact-heading > .p-button {
    width: 100%;
  }
}
</style>
