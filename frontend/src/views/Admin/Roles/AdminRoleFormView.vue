<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Skeleton from 'primevue/skeleton'
import Textarea from 'primevue/textarea'
import Toast from 'primevue/toast'
import { useAdminStore } from '../../../stores/admin'
import type { AdminPermission, SalesRoleLevel } from '../../../types/admin'

interface ExplorerMenu {
  key: string
  permission: AdminPermission
  children: AdminPermission[]
}

interface ExplorerGroup {
  key: string
  label: string
  sort: number
  count: number
  menus: ExplorerMenu[]
  orphans: AdminPermission[]
}

interface LandingOption {
  route: string
  label: string
  name: string
  requiredKey: string
}

const route = useRoute()
const router = useRouter()
const store = useAdminStore()
const toast = useToast()

const isEdit = computed(() => Boolean(route.params.id))
const roleId = computed(() => String(route.params.id ?? ''))

const DESCRIPTION_MAX = 255

const LEVEL_OPTIONS: { label: string; value: SalesRoleLevel }[] = [
  { label: 'Level 2 - Sales Level 1 business role', value: 2 },
  { label: 'Level 3 - Reports to Level 2 and oversees Level 4', value: 3 },
  { label: 'Level 4 - Operational/self-scope role', value: 4 },
]

const RECOMMENDED_PERMISSION_KEYS: Record<2 | 3 | 4, string[]> = {
  2: ['menu_sales_dashboard', 'view_sales_dashboard', 'view_team_dashboard', 'menu_sales_structure', 'view_sales_structure', 'menu_my_prospects', 'view_my_prospects', 'menu_my_customers', 'view_my_customers', 'menu_sales_history', 'view_sales_history', 'view_own_visits', 'menu_profile', 'view_own_profile', 'change_own_password'],
  3: ['menu_sales_dashboard', 'view_sales_dashboard', 'view_team_dashboard', 'menu_my_prospects', 'view_my_prospects', 'view_my_prospect_detail', 'menu_my_customers', 'view_my_customers', 'view_my_customer_detail', 'menu_sales_history', 'view_sales_history', 'view_own_visits', 'menu_profile', 'view_own_profile', 'change_own_password'],
  4: ['menu_sales_dashboard', 'view_sales_dashboard', 'menu_my_prospects', 'view_my_prospects', 'view_my_prospect_detail', 'check_in_prospect', 'check_out_prospect', 'menu_sales_pipeline', 'update_prospect_pipeline', 'update_visit_result', 'view_ai_summary', 'view_ai_menu_profiling', 'use_prospect_ai_chat', 'menu_my_customers', 'view_my_customers', 'view_my_customer_detail', 'check_in_customer', 'check_out_customer', 'menu_sales_history', 'view_sales_history', 'view_own_visits', 'menu_profile', 'view_own_profile', 'change_own_password'],
}

const GROUP_LABELS: Record<string, string> = {
  dashboard: 'Dashboard',
  accounts: 'Accounts',
  roles: 'Roles',
  sales_structure: 'Sales Organization',
  prospects: 'Prospects',
  customers: 'Customers',
  visits: 'Visits',
  reports: 'Reports',
  profile: 'Profile',
}

function groupLabel(key: string) {
  if (GROUP_LABELS[key]) return GROUP_LABELS[key]
  return key.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}

const form = reactive<{ name: string; level: SalesRoleLevel | null; landingPage: string; description: string }>({
  name: '',
  level: null,
  landingPage: '',
  description: '',
})

const selectedKeys = ref(new Set<string>())
const searchQuery = ref('')
const expandedGroups = ref(new Set<string>())
const expandedMenus = ref(new Set<string>())

const loadError = ref('')
const notFound = ref(false)
const loadingRole = ref(false)
const loaded = ref(false)
const submitting = ref(false)
const lastSavedCount = ref<number | null>(null)
const landingWarning = ref('')
let seeded = false

const catalog = computed<AdminPermission[]>(() => (Array.isArray(store.permissions) ? store.permissions : []))
const byKey = computed(() => new Map(catalog.value.map((p) => [p.key, p])))

function childrenOf(menuKey: string) {
  return catalog.value.filter((p) => p.parentKey === menuKey)
}

function ancestorsOf(key: string): string[] {
  const out: string[] = []
  let current = byKey.value.get(key)
  while (current?.parentKey) {
    out.push(current.parentKey)
    current = byKey.value.get(current.parentKey)
  }
  return out
}

function descendantsOf(menuKey: string): AdminPermission[] {
  const out: AdminPermission[] = []
  const queue = [...childrenOf(menuKey)]
  const seen = new Set<string>()
  while (queue.length) {
    const perm = queue.shift()!
    if (seen.has(perm.key)) continue
    seen.add(perm.key)
    out.push(perm)
    for (const child of childrenOf(perm.key)) queue.push(child)
  }
  return out
}

function hasSelectedDescendant(menuKey: string) {
  return descendantsOf(menuKey).some((c) => selectedKeys.value.has(c.key))
}

const groups = computed<ExplorerGroup[]>(() => {
  const map = new Map<string, ExplorerGroup>()
  const ensure = (key: string, sort: number) => {
    let group = map.get(key)
    if (!group) {
      group = { key, label: groupLabel(key), sort, count: 0, menus: [], orphans: [] }
      map.set(key, group)
    }
    return group
  }
  for (const perm of catalog.value) {
    if (perm.nodeType === 'MENU') {
      ensure(perm.groupKey, perm.sortOrder).menus.push({ key: perm.key, permission: perm, children: [] })
    } else if (perm.nodeType === 'ACTION' && !perm.parentKey) {
      ensure(perm.groupKey, perm.sortOrder).orphans.push(perm)
    }
  }
  for (const perm of catalog.value) {
    if (!perm.parentKey) continue
    const parent = byKey.value.get(perm.parentKey)
    if (!parent) {
      ensure(perm.groupKey, perm.sortOrder).orphans.push(perm)
      continue
    }
    const group = map.get(parent.groupKey) ?? ensure(parent.groupKey, parent.sortOrder)
    const menu = group.menus.find((m) => m.key === parent.key)
    if (menu) menu.children.push(perm)
    else group.orphans.push(perm)
  }
  const list = [...map.values()]
  for (const group of list) {
    group.menus.sort((a, b) => a.permission.sortOrder - b.permission.sortOrder)
    group.orphans.sort((a, b) => a.sortOrder - b.sortOrder)
    for (const menu of group.menus) {
      menu.children.sort((a, b) => a.sortOrder - b.sortOrder)
      group.count += menu.children.length
    }
    group.count += group.menus.length + group.orphans.length
  }
  list.sort((a, b) => a.sort - b.sort)
  return list
})

watch(catalog, () => {
  if (seeded || !groups.value.length) return
  seeded = true
  for (const group of groups.value) {
    expandedGroups.value.add(group.key)
    for (const menu of group.menus) expandedMenus.value.add(menu.key)
  }
})

const totalPermissions = computed(() => catalog.value.length)
const selectedCount = computed(() => selectedKeys.value.size)
const allSelected = computed(() => totalPermissions.value > 0 && catalog.value.every((p) => selectedKeys.value.has(p.key)))
const recommendedKeys = computed(() => {
  if (form.level !== 2 && form.level !== 3 && form.level !== 4) return []
  return RECOMMENDED_PERMISSION_KEYS[form.level].filter((key) => byKey.value.has(key))
})
const recommendedSelectedCount = computed(() => recommendedKeys.value.filter((key) => selectedKeys.value.has(key)).length)
const hasRecommendedLevel = computed(() => form.level === 2 || form.level === 3 || form.level === 4)
const recommendedComplete = computed(() => recommendedKeys.value.length > 0 && recommendedSelectedCount.value === recommendedKeys.value.length)

const searchActive = computed(() => searchQuery.value.trim().length > 0)

const visibleKeySet = computed<Set<string>>(() => {
  const query = searchQuery.value.trim().toLowerCase()
  const list = catalog.value
  if (!query) return new Set(list.map((p) => p.key))
  const out = new Set<string>()
  for (const p of list) {
    const hay = [p.name, p.key, p.description ?? '', groupLabel(p.groupKey)].join(' ').toLowerCase()
    if (hay.includes(query)) {
      out.add(p.key)
      for (const ancestor of ancestorsOf(p.key)) out.add(ancestor)
    }
  }
  return out
})

const visibleCount = computed(() => visibleKeySet.value.size)

function isSelected(key: string) {
  return selectedKeys.value.has(key)
}

function toggleAll(checked: boolean) {
  if (checked) {
    for (const p of catalog.value) selectedKeys.value.add(p.key)
  } else {
    selectedKeys.value.clear()
  }
}

function menuState(menuKey: string): 'checked' | 'indeterminate' | 'unchecked' {
  const descendants = descendantsOf(menuKey)
  const ownSelected = selectedKeys.value.has(menuKey)
  if (descendants.length === 0) return ownSelected ? 'checked' : 'unchecked'
  const selectedDescendants = descendants.filter((c) => selectedKeys.value.has(c.key)).length
  if (ownSelected && selectedDescendants === descendants.length) return 'checked'
  if (ownSelected || selectedDescendants > 0) return 'indeterminate'
  return 'unchecked'
}

function visibleChildrenOf(menuKey: string) {
  return childrenOf(menuKey).filter((c) => visibleKeySet.value.has(c.key))
}

function togglePermission(perm: AdminPermission, checked: boolean) {
  if (checked) {
    selectedKeys.value.add(perm.key)
    for (const ancestor of ancestorsOf(perm.key)) selectedKeys.value.add(ancestor)
    return
  }

  selectedKeys.value.delete(perm.key)

  for (const ancestor of ancestorsOf(perm.key)) {
    if (!hasSelectedDescendant(ancestor)) selectedKeys.value.delete(ancestor)
  }
}

function toggleMenu(menu: ExplorerMenu, checked: boolean) {
  const descendants = descendantsOf(menu.key)

  if (checked) {
    selectedKeys.value.add(menu.key)
    for (const child of descendants) selectedKeys.value.add(child.key)
    for (const ancestor of ancestorsOf(menu.key)) selectedKeys.value.add(ancestor)
    return
  }

  selectedKeys.value.delete(menu.key)
  for (const child of descendants) selectedKeys.value.delete(child.key)

  for (const ancestor of ancestorsOf(menu.key)) {
    if (!hasSelectedDescendant(ancestor)) selectedKeys.value.delete(ancestor)
  }
}

function selectVisible() {
  for (const key of visibleKeySet.value) {
    const perm = byKey.value.get(key)
    if (perm && perm.nodeType === 'ACTION') selectedKeys.value.add(perm.key)
  }
  for (const key of [...selectedKeys.value]) {
    for (const ancestor of ancestorsOf(key)) selectedKeys.value.add(ancestor)
  }
}

function clearVisible() {
  const affected = new Set<string>()
  for (const key of [...selectedKeys.value]) {
    const perm = byKey.value.get(key)
    if (perm && perm.nodeType === 'ACTION' && visibleKeySet.value.has(key)) {
      selectedKeys.value.delete(key)
      for (const ancestor of ancestorsOf(key)) affected.add(ancestor)
    }
  }
  for (const ancestor of affected) {
    if (!hasSelectedDescendant(ancestor)) selectedKeys.value.delete(ancestor)
  }
}

function clearAll() {
  selectedKeys.value.clear()
}

function normalizedPermissionSet(keys: string[]) {
  const next = new Set<string>()
  for (const key of keys) {
    if (!byKey.value.has(key)) continue
    next.add(key)
    for (const ancestor of ancestorsOf(key)) next.add(ancestor)
  }
  return next
}

function applyRecommendedPermissions() {
  selectedKeys.value = normalizedPermissionSet(recommendedKeys.value)
}

watch(
  () => form.level,
  () => {
    if (isEdit.value || selectedKeys.value.size > 0 || !hasRecommendedLevel.value) return
    applyRecommendedPermissions()
  },
)

function groupHasVisible(group: ExplorerGroup) {
  return group.menus.some((m) => visibleKeySet.value.has(m.key)) || group.orphans.some((o) => visibleKeySet.value.has(o.key))
}

function menuHasVisible(menu: ExplorerMenu) {
  return menu.children.some((c) => visibleKeySet.value.has(c.key))
}

function isGroupOpen(group: ExplorerGroup) {
  return (searchActive.value && groupHasVisible(group)) || expandedGroups.value.has(group.key)
}

function isMenuOpen(menu: ExplorerMenu) {
  return (searchActive.value && menuHasVisible(menu)) || expandedMenus.value.has(menu.key)
}

function toggleGroup(group: ExplorerGroup) {
  if (expandedGroups.value.has(group.key)) expandedGroups.value.delete(group.key)
  else expandedGroups.value.add(group.key)
}

function toggleMenuOpen(menu: ExplorerMenu) {
  if (expandedMenus.value.has(menu.key)) expandedMenus.value.delete(menu.key)
  else expandedMenus.value.add(menu.key)
}

const landingOptions = computed<LandingOption[]>(() => {
  const seen = new Set<string>()
  const options: LandingOption[] = []
  for (const perm of catalog.value) {
    if (perm.nodeType !== 'ACTION' || !perm.routePath || !perm.routePath.trim()) continue
    if (perm.key === 'view_team_dashboard') continue
    if (seen.has(perm.routePath)) continue
    seen.add(perm.routePath)
    const parent = perm.parentKey ? byKey.value.get(perm.parentKey) : undefined
    const name = parent?.name ?? perm.name
    options.push({ route: perm.routePath, label: `${name} (${perm.routePath})`, name, requiredKey: perm.key })
  }
  options.sort((a, b) => (byKey.value.get(a.requiredKey)?.sortOrder ?? 0) - (byKey.value.get(b.requiredKey)?.sortOrder ?? 0))
  return options
})

const landingRequiredKey = computed(() => landingOptions.value.find((o) => o.route === form.landingPage)?.requiredKey ?? null)
const landingPermitted = computed(() => {
  const requiredKey = landingRequiredKey.value
  return Boolean(form.landingPage) && requiredKey !== null && selectedKeys.value.has(requiredKey)
})
const landingAvailable = computed(() => landingOptions.value.some((o) => o.route === form.landingPage))

function landingOptionDisabled(option: LandingOption) {
  return !selectedKeys.value.has(option.requiredKey)
}

const landingStatus = computed(() => {
  if (!form.landingPage) return { kind: 'missing', text: 'No landing page selected' }
  if (!landingAvailable.value) return { kind: 'blocked', text: 'Landing page is no longer available' }
  if (!landingPermitted.value) return { kind: 'blocked', text: 'Landing page permission not granted' }
  return { kind: 'ok', text: 'Landing page permitted' }
})

watch(selectedKeys, () => {
  const requiredKey = landingRequiredKey.value
  if (form.landingPage && requiredKey && !selectedKeys.value.has(requiredKey)) {
    const option = landingOptions.value.find((o) => o.route === form.landingPage)
    landingWarning.value = option
      ? `The selected landing page requires the "${option.name}" permission. Grant it first or choose another page.`
      : 'The selected landing page is no longer permitted.'
    form.landingPage = ''
  } else {
    landingWarning.value = ''
  }
})

const descriptionCount = computed(() => form.description.length)

const canSubmit = computed(() =>
  Boolean(form.name.trim()) && form.level !== null && LEVEL_OPTIONS.some((option) => option.value === form.level)
    && Boolean(form.landingPage) && landingPermitted.value && !submitting.value && !store.savingSalesRole,
)

async function loadRole() {
  loadingRole.value = true
  loadError.value = ''
  notFound.value = false
  try {
    const role = await store.fetchSalesRoleDetail(roleId.value)
    form.name = role.name
    form.level = role.level
    form.description = role.description ?? ''
    form.landingPage = role.landingPage ?? ''
    selectedKeys.value.clear()
    for (const permission of role.permissions ?? []) selectedKeys.value.add(permission.key)
    lastSavedCount.value = role.permissionCount ?? role.permissions?.length ?? 0
    loaded.value = true
  } catch (e) {
    loadError.value = store.errorMessage(e)
    notFound.value = true
  } finally {
    loadingRole.value = false
  }
}

function goBack() {
  if (isEdit.value) router.push(`/admin/role-management/${roleId.value}`)
  else router.push('/admin/role-management')
}

async function handleSave() {
  if (!canSubmit.value || form.level === null) return
  submitting.value = true
  loadError.value = ''
  try {
    const payload = {
      name: form.name.trim(),
      level: form.level,
      description: form.description.trim(),
      landingPage: form.landingPage.trim(),
      permissionKeys: [...selectedKeys.value],
    }
    if (isEdit.value) {
      await store.updateSalesRole(roleId.value, payload)
      toast.add({ severity: 'success', summary: 'Role Updated', detail: `${payload.name} permissions updated.`, life: 3000 })
      await router.push(`/admin/role-management/${roleId.value}`)
    } else {
      const role = await store.createSalesRole(payload)
      toast.add({ severity: 'success', summary: 'Role Created', detail: `${role.name} has been created.`, life: 3000 })
      await router.push('/admin/role-management')
    }
  } catch (e) {
    loadError.value = store.errorMessage(e)
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await store.fetchPermissions().catch(() => undefined)
  if (isEdit.value) await loadRole()
  else loaded.value = true
})
</script>

<template>
  <section class="admin-page compact-admin-page">
    <Toast position="top-right" />

    <div v-if="notFound && isEdit && !loaded" class="state-box">
      <div class="state-icon-wrap"><i class="pi pi-id-card" /></div>
      <strong>Role could not be loaded</strong>
      <span class="muted">{{ loadError || 'The requested role does not exist or was removed.' }}</span>
      <Button label="Back to Role List" icon="pi pi-arrow-left" size="small" @click="router.push('/admin/role-management')" />
    </div>

    <template v-else>
      <Message v-if="loadError && !notFound" severity="error" :closable="false">{{ loadError }}</Message>

      <div v-if="isEdit && !loaded" class="state-box">
        <i class="pi pi-spin pi-spinner state-icon" />
        <span>Loading role...</span>
      </div>

      <template v-if="!isEdit || loaded">
        <header class="page-heading">
          <div class="compact-heading-main">
            <Button class="header-back" icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.push('/admin/role-management')" title="Back to role list" />
          <div class="page-title-wrapper">
            <span class="eyebrow">{{ isEdit ? 'Edit Role' : 'Create Role' }}</span>
            <h1>{{ isEdit ? (form.name || 'Edit Sales Role') : 'Create Sales Role' }}</h1>
            <p class="muted">{{ isEdit ? 'Update role information and configure its permissions.' : 'Configure a new sales role and grant its permissions.' }}</p>
          </div>
          </div>
          <div class="page-heading-actions">
            <Button label="Cancel" severity="secondary" text size="small" @click="goBack" :disabled="submitting || store.savingSalesRole" />
            <Button :label="isEdit ? 'Save Changes' : 'Create Role'" icon="pi pi-check" size="small" :loading="submitting || store.savingSalesRole" :disabled="!canSubmit" @click="handleSave" />
          </div>
        </header>

        <div class="editor-layout">
          <!-- LEFT: ROLE INFORMATION -->
          <div class="info-column">
            <div class="panel info-panel">
              <div class="panel-header">
                <div class="panel-icon si-teal"><i class="pi pi-id-card" /></div>
                <div>
                  <h3>Role Information</h3>
                  <p>Core details for the sales role.</p>
                </div>
              </div>
              <div class="info-fields">
                <div class="form-field">
                  <label>Role Name <span class="required">*</span></label>
                  <InputText v-model="form.name" placeholder="e.g. Area Sales Manager" />
                </div>
                <div class="form-field">
                  <label>Hierarchy Level <span class="required">*</span></label>
                  <Select v-model="form.level" :options="LEVEL_OPTIONS" optionLabel="label" optionValue="value" placeholder="Select level" />
                </div>
                <div v-if="hasRecommendedLevel" class="recommended-panel">
                  <div class="recommended-copy">
                    <span>Recommended permissions for Level {{ form.level }}</span>
                    <strong>{{ recommendedSelectedCount }} / {{ recommendedKeys.length }} applied</strong>
                    <small v-if="isEdit">Saved custom selections are kept until you apply or reset explicitly.</small>
                    <small v-else>Use the recommended baseline, then adjust permissions manually as needed.</small>
                  </div>
                  <div class="recommended-actions">
                    <Button label="Apply Recommended" icon="pi pi-sparkles" size="small" outlined @click="applyRecommendedPermissions" />
                    <Button label="Reset to Recommended" size="small" severity="secondary" text @click="applyRecommendedPermissions" />
                  </div>
                  <span class="recommended-state" :class="{ ready: recommendedComplete }">
                    <i :class="recommendedComplete ? 'pi pi-check-circle' : 'pi pi-info-circle'" />
                    {{ recommendedComplete ? 'Recommended set applied' : 'Custom or partial selection' }}
                  </span>
                </div>
                <div class="form-field">
                  <label>Initial Open Menu / Landing Page <span class="required">*</span></label>
                  <Select
                    v-model="form.landingPage"
                    :options="landingOptions"
                    optionLabel="label"
                    optionValue="route"
                    :optionDisabled="landingOptionDisabled"
                    filter
                    placeholder="Select landing page"
                  />
                  <div v-if="landingWarning" class="landing-warning">
                    <i class="pi pi-exclamation-triangle" />
                    <span>{{ landingWarning }}</span>
                  </div>
                </div>
                <div class="form-field">
                  <label>Description <span class="optional-label">(optional)</span></label>
                  <Textarea v-model="form.description" :maxlength="DESCRIPTION_MAX" rows="5" autoResize />
                  <div class="char-counter" :class="{ 'char-full': descriptionCount >= DESCRIPTION_MAX }">{{ descriptionCount }} / {{ DESCRIPTION_MAX }}</div>
                </div>
                <div class="department-note">
                  <i class="pi pi-info-circle" />
                  <span>Department master data will be added later.</span>
                </div>
              </div>
            </div>

            <div class="panel counter-panel">
              <div class="counter-row">
                <span>Selected</span>
                <strong>{{ selectedCount }} / {{ totalPermissions }} permissions</strong>
              </div>
              <div class="landing-status" :class="`landing-status-${landingStatus.kind}`">
                <i :class="landingStatus.kind === 'ok' ? 'pi pi-check-circle' : 'pi pi-exclamation-circle'" />
                <span>{{ landingStatus.text }}</span>
              </div>
              <div v-if="isEdit && lastSavedCount !== null" class="last-saved">Last saved: {{ lastSavedCount }} permissions</div>
            </div>
          </div>

          <!-- RIGHT: PERMISSIONS EXPLORER -->
          <div class="explorer-column">
            <div class="panel explorer-panel">
              <div class="explorer-toolbar">
                <div class="toolbar-row">
                  <div class="search-field">
                    <i class="pi pi-search" />
                    <InputText v-model="searchQuery" placeholder="Search permissions..." />
                    <button v-if="searchActive" type="button" class="clear-search" title="Clear search" @click="searchQuery = ''"><i class="pi pi-times" /></button>
                  </div>
                  <div class="toolbar-actions">
                    <Button label="Select Visible" size="small" severity="secondary" outlined @click="selectVisible" />
                    <Button label="Clear Visible" size="small" severity="secondary" text @click="clearVisible" />
                    <Button label="Clear All" size="small" severity="danger" text @click="clearAll" />
                    <div class="select-all">
                      <Checkbox :model-value="allSelected" binary @update:model-value="(checked) => toggleAll(Boolean(checked))" />
                      <button type="button" class="select-all-label" @click="toggleAll(!allSelected)">Select All</button>
                    </div>
                  </div>
                </div>
                <div class="toolbar-counters">
                  <span><strong>{{ visibleCount }}</strong> visible</span>
                  <span><strong>{{ selectedCount }}</strong> selected</span>
                  <span><strong>{{ totalPermissions }}</strong> total</span>
                </div>
              </div>

              <div v-if="store.permissionsLoading && !store.permissions.length" class="explorer-loading">
                <Skeleton v-for="n in 6" :key="n" class="skeleton-row" />
              </div>
              <div v-else-if="!groups.length" class="explorer-empty">No permissions available.</div>

              <div v-else class="explorer-tree">
                <div v-for="group in groups" :key="group.key" class="perm-group">
                  <button type="button" class="perm-group-header" @click="toggleGroup(group)">
                    <i class="pi" :class="isGroupOpen(group) ? 'pi-chevron-down' : 'pi-chevron-right'" />
                    <span class="perm-group-label">{{ group.label }}</span>
                    <span class="perm-group-count">{{ group.count }} permissions</span>
                  </button>
                  <div v-if="isGroupOpen(group)" class="perm-group-body">
                    <template v-for="menu in group.menus" :key="menu.key">
                      <div class="perm-menu">
                        <div class="perm-menu-header">
                          <Checkbox :model-value="menuState(menu.key) === 'checked'" :indeterminate="menuState(menu.key) === 'indeterminate'" binary @update:model-value="(checked) => toggleMenu(menu, Boolean(checked))" />
                          <span class="perm-name perm-menu-name">{{ menu.permission.name }}</span>
                          <code class="key-badge">{{ menu.permission.key }}</code>
                          <button type="button" class="chevron-btn" :title="isMenuOpen(menu) ? 'Collapse' : 'Expand'" @click="toggleMenuOpen(menu)">
                            <i class="pi" :class="isMenuOpen(menu) ? 'pi-chevron-down' : 'pi-chevron-right'" />
                          </button>
                        </div>
                        <div v-if="isMenuOpen(menu)" class="perm-children">
                          <div v-for="child in visibleChildrenOf(menu.key)" :key="child.key" class="perm-row">
                            <Checkbox :model-value="isSelected(child.key)" binary @update:model-value="(checked) => togglePermission(child, Boolean(checked))" />
                            <div class="perm-info">
                              <span class="perm-name">{{ child.name }}</span>
                              <span v-if="child.description" class="perm-desc">{{ child.description }}</span>
                            </div>
                            <code class="key-badge">{{ child.key }}</code>
                          </div>
                          <div v-if="!visibleChildrenOf(menu.key).length" class="perm-row perm-row-empty">No matching permissions</div>
                        </div>
                      </div>
                    </template>
                    <div v-for="orphan in group.orphans" :key="orphan.key" class="perm-row perm-row-orphan">
                      <Checkbox :model-value="isSelected(orphan.key)" binary @update:model-value="(checked) => togglePermission(orphan, Boolean(checked))" />
                      <div class="perm-info">
                        <span class="perm-name">{{ orphan.name }}</span>
                        <span v-if="orphan.description" class="perm-desc">{{ orphan.description }}</span>
                      </div>
                      <code class="key-badge">{{ orphan.key }}</code>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </template>
  </section>
</template>

<style scoped>
.admin-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.4rem 1.6rem;
  min-height: 100vh;
  background: linear-gradient(180deg, #f3f7fd 0%, #f8fafc 48%, #f8fafc 100%);
}
.page-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 1rem;
}
.page-title-wrapper .eyebrow {
  font-size: 0.68rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #2563eb;
}
h1 {
  margin: 0.2rem 0 0.2rem;
  font-size: 1.55rem;
  color: #0f172a;
}
.muted {
  margin: 0;
  color: #7c8798;
  font-size: 0.85rem;
}
.page-heading-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.editor-layout {
  display: grid;
  grid-template-columns: minmax(320px, 360px) minmax(0, 1fr);
  gap: 1.1rem;
  align-items: start;
  min-height: min(760px, calc(100vh - 185px));
}

.panel {
  background: #ffffff;
  border: 1px solid #dbeafe;
  border-radius: 16px;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
}
.info-column {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.info-panel {
  padding: 1.15rem 1.2rem 1.3rem;
}
.panel-header {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding-bottom: 0.9rem;
  margin-bottom: 1rem;
  border-bottom: 1px solid #f0f3f7;
}
.panel-icon {
  width: 36px;
  height: 36px;
  flex: 0 0 auto;
  display: grid;
  place-content: center;
  border-radius: 10px;
  font-size: 0.95rem;
}
.si-teal {
  background: #eff6ff;
  color: #2563eb;
}
.panel-header h3 {
  margin: 0;
  font-size: 0.92rem;
  font-weight: 750;
  color: #0f172a;
}
.panel-header p {
  margin: 0.15rem 0 0;
  font-size: 0.76rem;
  color: #7c8798;
}

.info-fields {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
.form-field label {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #64748b;
}
.required {
  color: #dc2626;
}
.optional-label {
  font-weight: 500;
  text-transform: none;
  color: #94a3b8;
}
.char-counter {
  align-self: flex-end;
  font-size: 0.7rem;
  color: #94a3b8;
}
.char-counter.char-full {
  color: #dc2626;
  font-weight: 700;
}
.landing-warning {
  display: flex;
  align-items: flex-start;
  gap: 0.4rem;
  padding: 0.5rem 0.65rem;
  border: 1px solid #fde68a;
  border-radius: 8px;
  background: #fffbeb;
  color: #92400e;
  font-size: 0.74rem;
  line-height: 1.4;
}
.landing-warning i {
  flex-shrink: 0;
  color: #d97706;
}
.department-note {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.55rem 0.7rem;
  border: 1px dashed #dbe3ee;
  border-radius: 8px;
  background: #f8fafc;
  color: #7c8798;
  font-size: 0.75rem;
}
.department-note i {
  flex-shrink: 0;
  color: #94a3b8;
}

.recommended-panel {
  display: grid;
  gap: 0.65rem;
  padding: 0.8rem;
  border: 1px solid #bfdbfe;
  border-radius: 14px;
  background: #f8fbff;
}
.recommended-copy {
  display: grid;
  gap: 0.14rem;
}
.recommended-copy span {
  color: #2563eb;
  font-size: 0.68rem;
  font-weight: 850;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
.recommended-copy strong {
  color: #0f172a;
  font-size: 0.9rem;
}
.recommended-copy small {
  color: #64748b;
  font-size: 0.74rem;
  line-height: 1.4;
}
.recommended-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}
.recommended-state {
  width: fit-content;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.32rem 0.55rem;
  border: 1px solid #dbeafe;
  border-radius: 999px;
  background: #ffffff;
  color: #64748b;
  font-size: 0.72rem;
  font-weight: 750;
}
.recommended-state.ready {
  color: #15803d;
  border-color: #bbf7d0;
  background: #f0fdf4;
}

.counter-panel {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  padding: 0.9rem 1.1rem;
}
.counter-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.counter-row span {
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #64748b;
}
.counter-row strong {
  font-size: 0.88rem;
  color: #0f172a;
}
.landing-status {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.76rem;
}
.landing-status i {
  flex-shrink: 0;
}
.landing-status-ok {
  color: #059669;
}
.landing-status-missing {
  color: #b45309;
}
.landing-status-blocked {
  color: #dc2626;
}
.last-saved {
  padding-top: 0.55rem;
  border-top: 1px solid #f1f4f8;
  font-size: 0.74rem;
  color: #64748b;
}

.explorer-column {
  min-width: 0;
}
.explorer-panel {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 185px);
  min-height: 620px;
  max-height: calc(100vh - 185px);
  overflow: hidden;
}
.explorer-toolbar {
  flex: 0 0 auto;
  position: sticky;
  top: 0;
  z-index: 5;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  padding: 0.7rem 0.8rem;
  background: #f8fbff;
  border-bottom: 1px solid #dbeafe;
}
.toolbar-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.6rem;
  flex-wrap: wrap;
}
.search-field {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1 1 220px;
  min-height: 2.4rem;
  padding: 0 0.7rem;
  background: #ffffff;
  border: 1px solid #dbe3ee;
  border-radius: 8px;
}
.search-field i {
  color: #94a3b8;
  font-size: 0.8rem;
}
.search-field :deep(.p-inputtext) {
  flex: 1;
  border: 0;
  padding-inline: 0;
  box-shadow: none;
  font-size: 0.82rem;
}
.clear-search {
  border: 0;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  display: grid;
  place-content: center;
  width: 1.4rem;
  height: 1.4rem;
  border-radius: 999px;
}
.clear-search:hover {
  background: #f1f5f9;
  color: #334155;
}
.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  flex-wrap: wrap;
}
.select-all {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.3rem 0.5rem;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #ffffff;
}
.select-all-label {
  border: 0;
  background: transparent;
  padding: 0;
  font: inherit;
  font-size: 0.74rem;
  font-weight: 700;
  color: #334155;
  cursor: pointer;
}
.toolbar-counters {
  display: flex;
  gap: 0.9rem;
  font-size: 0.72rem;
  color: #64748b;
}
.toolbar-counters strong {
  color: #0f172a;
}

.explorer-loading {
  display: grid;
  gap: 0.5rem;
  padding: 0.8rem;
}
.skeleton-row {
  height: 2.2rem;
  border-radius: 8px;
}
.explorer-empty {
  padding: 2rem;
  text-align: center;
  color: #94a3b8;
  font-size: 0.85rem;
}

.explorer-tree {
  flex: 1 1 auto;
  overflow-y: auto;
  padding: 0.4rem 0.5rem 0.7rem;
  min-height: 0;
}
.perm-group-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  border: 0;
  border-radius: 8px;
  padding: 0.5rem 0.55rem;
  background: #eff6ff;
  font: inherit;
  text-align: left;
  cursor: pointer;
  margin-top: 0.25rem;
}
.perm-group-header:first-child {
  margin-top: 0;
}
.perm-group-header:hover {
  background: #dbeafe;
}
.perm-group-header > i {
  color: #94a3b8;
  font-size: 0.72rem;
}
.perm-group-label {
  font-weight: 800;
  font-size: 0.76rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #0f172a;
}
.perm-group-count {
  margin-left: auto;
  font-size: 0.68rem;
  color: #94a3b8;
  white-space: nowrap;
}
.perm-group-body {
  padding: 0.25rem 0 0.35rem 0.35rem;
  border-left: 2px solid #bfdbfe;
  margin-left: 0.95rem;
}

.perm-menu {
  margin-top: 0.2rem;
}
.perm-menu-header {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  min-height: 2.5rem;
  padding: 0.3rem 0.4rem 0.3rem 0.25rem;
  border-radius: 8px;
  background: #ffffff;
}
.perm-menu-header:hover {
  background: #f8fafc;
}
.perm-menu-name {
  font-weight: 850;
  color: #0f172a;
}
.perm-children {
  padding-left: 1.35rem;
  margin-left: 0.7rem;
  border-left: 1px solid #dbeafe;
}
.perm-row {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  min-height: 2.5rem;
  padding: 0.42rem 0.4rem 0.42rem 0.25rem;
  border-radius: 8px;
}
.perm-row:hover {
  background: #f8fafc;
}
.perm-row-orphan {
  margin-left: 0;
}
.perm-info {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 0;
  flex: 1;
}
.perm-name {
  font-size: 0.82rem;
  color: #1e293b;
  line-height: 1.35;
}
.perm-desc {
  font-size: 0.72rem;
  color: #94a3b8;
  line-height: 1.35;
}
.perm-row-empty {
  font-size: 0.74rem;
  color: #b0b8c6;
  font-style: italic;
}
.key-badge {
  flex: 0 1 220px;
  align-self: center;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.66rem;
  font-weight: 600;
  color: #64748b;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
  padding: 0.16rem 0.45rem;
  white-space: normal;
  overflow-wrap: anywhere;
  text-align: right;
}
.chevron-btn {
  flex: 0 0 auto;
  border: 0;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  display: grid;
  place-content: center;
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 6px;
  font-size: 0.68rem;
}
.chevron-btn:hover {
  background: #eef2f6;
  color: #334155;
}

.explorer-panel :deep(.p-checkbox-box) {
  background: #ffffff;
  border-color: #cbd5e1;
}
.explorer-panel :deep(.p-checkbox-checked .p-checkbox-box),
.explorer-panel :deep(.p-checkbox-indeterminate .p-checkbox-box) {
  background: #2563eb;
  border-color: #2563eb;
}
.explorer-panel :deep(.p-checkbox-indeterminate .p-checkbox-box .p-checkbox-icon) {
  color: #ffffff;
}

.state-box {
  min-height: 260px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 2rem;
  text-align: center;
  color: #8492a6;
}
.state-icon {
  font-size: 1.75rem;
  color: #2563eb;
  margin-bottom: 0.25rem;
}
.state-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  background: #f8fafc;
  display: grid;
  place-content: center;
  margin-bottom: 0.35rem;
}
.state-icon-wrap i {
  font-size: 1.4rem;
  color: #94a3b8;
}
.state-box strong {
  color: #0f172a;
  font-size: 0.95rem;
}
.state-box .muted {
  font-size: 0.82rem;
}

:deep(.p-inputtext),
:deep(.p-select),
:deep(.p-textarea) {
  background: #ffffff;
  color: #0f172a;
  border-color: #dbe3ee;
}
:deep(.p-select-label),
:deep(.p-select-dropdown),
:deep(.p-textarea::placeholder),
:deep(.p-inputtext::placeholder) {
  color: #64748b;
}
:deep(.p-select-overlay),
:deep(.p-select-list),
:deep(.p-select-option) {
  background: #ffffff;
  color: #0f172a;
}
:deep(.p-select-option.p-select-option-selected),
:deep(.p-select-option:hover) {
  background: #f1f5f9;
  color: #0f172a;
}

@media (max-width: 1180px) {
  .editor-layout {
    grid-template-columns: 1fr;
  }
  .explorer-panel {
    height: 72vh;
    min-height: 560px;
    max-height: 72vh;
  }
}
@media (max-width: 768px) {
  .admin-page {
    padding: 1rem;
  }
  .page-heading {
    flex-direction: column;
  }
  .toolbar-row {
    flex-direction: column;
    align-items: stretch;
  }
  .explorer-panel {
    height: auto;
    min-height: 0;
    max-height: none;
    overflow: visible;
  }
  .explorer-toolbar {
    position: static;
  }
  .explorer-tree {
    overflow: visible;
  }
  .perm-menu-header,
  .perm-row {
    align-items: flex-start;
  }
  .key-badge {
    max-width: 42%;
  }
}

/* Compact CRM workspace treatment */
.admin-page { padding: 0.9rem 1.25rem 1.5rem; }
.page-heading { margin-bottom: 0.8rem; gap: 0.75rem; }
.page-title-wrapper h1 { font-size: 1.35rem; margin: 0.15rem 0 0; }
.form-layout { gap: 1rem; }
.form-card { padding: 1rem 1.1rem; }
.form-card-header { padding-bottom: 0.65rem; margin-bottom: 0.2rem; }
.permission-explorer { min-height: 0; }
.compact-admin-page > .p-button { margin-bottom: 0.35rem; }
.compact-admin-page .page-heading { padding: 0.7rem 0.85rem; border: 1px solid #e3e9f0; border-radius: 12px; background: #fff; box-shadow: 0 1px 2px rgba(15,23,42,.03); }
.compact-heading-main { display: flex; align-items: center; min-width: 0; gap: 0.35rem; }
.header-back { flex: 0 0 auto; }
</style>
