<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useAdminStore } from '../../../stores/admin'
import type { AdminPermission } from '../../../types/admin'

const route = useRoute()
const router = useRouter()
const store = useAdminStore()

const id = computed(() => String(route.params.id))
const role = computed(() => (store.selectedRoleDetail?.id === id.value ? store.selectedRoleDetail : null))

const error = ref('')
const notFound = ref(false)
const loading = ref(true)

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

const permissionCount = computed(() => role.value?.permissionCount ?? role.value?.permissions?.length ?? 0)

const permissionGroups = computed(() => {
  const permissions = (role.value?.permissions ?? []).filter((p) => p.isActive)
  const map = new Map<string, AdminPermission[]>()
  for (const permission of permissions) {
    const list = map.get(permission.groupKey) ?? []
    list.push(permission)
    map.set(permission.groupKey, list)
  }
  return [...map.entries()]
    .map(([key, items]) => ({
      key,
      label: groupLabel(key),
      items: [...items].sort((a, b) => a.sortOrder - b.sortOrder),
      sort: Math.min(...items.map((i) => i.sortOrder)),
    }))
    .sort((a, b) => a.sort - b.sort)
})

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return new Intl.DateTimeFormat('en', { year: 'numeric', month: 'short', day: '2-digit' }).format(date)
}

async function load() {
  loading.value = true
  error.value = ''
  notFound.value = false
  try {
    await store.fetchSalesRoleDetail(id.value)
  } catch (e) {
    notFound.value = true
    error.value = store.errorMessage(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <section class="admin-page">
    <div v-if="notFound" class="state-box">
      <div class="state-icon-wrap"><i class="pi pi-id-card" /></div>
      <strong>Role not found</strong>
      <span class="muted">{{ error || 'The requested role does not exist or was removed.' }}</span>
      <Button label="Back to Role List" icon="pi pi-arrow-left" size="small" @click="router.push('/admin/role-management')" />
    </div>

    <template v-else>
      <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

      <div v-if="loading" class="state-box">
        <i class="pi pi-spin pi-spinner state-icon" />
        <span>Loading role...</span>
      </div>

      <template v-if="role">
        <Button icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.push('/admin/role-management')" title="Back to role list" />

        <header class="page-heading">
          <div class="page-title-wrapper">
            <span class="eyebrow">Role Detail</span>
            <h1>{{ role.name }}</h1>
            <div class="subtitle-row">
              <Tag :value="role.isActive ? 'Active' : 'Inactive'" :severity="role.isActive ? 'success' : 'secondary'" size="small" />
              <span class="level-chip" :class="`level-chip-${role.level}`">Level {{ role.level }}</span>
              <span class="muted">Updated {{ formatDate(role.updatedAt) }}</span>
            </div>
          </div>
          <div class="page-heading-actions">
            <Button label="Edit Role" icon="pi pi-pencil" size="small" @click="router.push(`/admin/role-management/${id}/edit`)" />
          </div>
        </header>

        <div class="detail-layout">
          <!-- LEFT: ROLE INFORMATION -->
          <div class="info-column">
            <div class="panel">
              <div class="panel-header">
                <div class="panel-icon si-teal"><i class="pi pi-id-card" /></div>
                <div>
                  <h3>Role Information</h3>
                  <p>Core details for this sales role.</p>
                </div>
              </div>
              <div class="detail-rows">
                <div class="detail-row">
                  <span class="detail-label">Role Name</span>
                  <span class="detail-value">{{ role.name }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Hierarchy Level</span>
                  <span class="detail-value"><Tag :value="`Level ${role.level}`" :severity="role.level === 1 ? 'info' : role.level === 2 ? 'success' : role.level === 3 ? 'warn' : 'secondary'" size="small" /></span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Description</span>
                  <span class="detail-value">{{ role.description || '—' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Landing Page</span>
                  <span class="detail-value"><code class="route-badge">{{ role.landingPage || '—' }}</code></span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Permissions</span>
                  <span class="detail-value"><strong>{{ permissionCount }}</strong> selected</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Status</span>
                  <span class="detail-value"><Tag :value="role.isActive ? 'Active' : 'Inactive'" :severity="role.isActive ? 'success' : 'secondary'" size="small" /></span>
                </div>
              </div>
            </div>
          </div>

          <!-- RIGHT: PERMISSIONS -->
          <div class="permissions-column">
            <div class="panel">
              <div class="panel-header">
                <div class="panel-icon si-blue"><i class="pi pi-lock" /></div>
                <div>
                  <h3>Permissions</h3>
                  <p>Granted permissions grouped by module.</p>
                </div>
              </div>
              <div v-if="!permissionGroups.length" class="empty-permissions">No permissions granted for this role.</div>
              <div class="perm-groups">
                <div v-for="group in permissionGroups" :key="group.key" class="perm-group">
                  <div class="perm-group-label">{{ group.label }}</div>
                  <div class="perm-group-items">
                    <div v-for="permission in group.items" :key="permission.key" class="perm-item">
                      <i class="pi pi-check-circle" />
                      <span class="perm-item-name">{{ permission.name }}</span>
                      <code class="key-badge">{{ permission.key }}</code>
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
  background: #ffffff;
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
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #0b7766;
}
h1 {
  margin: 0.2rem 0 0.2rem;
  font-size: 1.55rem;
  color: #0f172a;
}
.subtitle-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.muted {
  margin: 0;
  color: #7c8798;
  font-size: 0.82rem;
}
.page-heading-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.detail-layout {
  display: grid;
  grid-template-columns: minmax(320px, 36%) minmax(0, 64%);
  gap: 1.1rem;
  align-items: start;
}
.panel {
  background: #ffffff;
  border: 1px solid #edf1f6;
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
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
  background: #ecfdf5;
  color: #047857;
}
.si-blue {
  background: #fff1f2;
  color: #d14350;
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

.detail-rows {
  display: flex;
  flex-direction: column;
}
.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.6rem 0;
  border-bottom: 1px solid #f4f6f9;
}
.detail-row:last-child {
  border-bottom: 0;
}
.detail-label {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #64748b;
  padding-top: 0.15rem;
}
.detail-value {
  font-size: 0.84rem;
  color: #1e293b;
  text-align: right;
  word-break: break-word;
}

.route-badge {
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.72rem;
  font-weight: 600;
  color: #d14350;
  background: #fff1f2;
  border: 1px solid #ffd9dd;
  border-radius: 6px;
  padding: 0.14rem 0.45rem;
  word-break: break-all;
}
.level-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.16rem 0.5rem;
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 800;
  border: 1px solid transparent;
}
.level-chip-1 {
  background: #fff1f2;
  color: #bb3342;
  border-color: #ffd9dd;
}
.level-chip-2 {
  background: #ecfdf5;
  color: #047857;
  border-color: #d1fae5;
}
.level-chip-3 {
  background: #fff7ed;
  color: #9a3412;
  border-color: #fed7aa;
}
.level-chip-4 {
  background: #f8fafc;
  color: #475569;
  border-color: #e2e8f0;
}

.empty-permissions {
  padding: 1.5rem;
  text-align: center;
  color: #94a3b8;
  font-size: 0.84rem;
}
.perm-groups {
  display: grid;
  gap: 1.1rem;
}
.perm-group-label {
  font-size: 0.72rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #334155;
  margin-bottom: 0.45rem;
}
.perm-group-items {
  display: flex;
  flex-direction: column;
}
.perm-item {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.4rem 0.3rem;
  border-radius: 8px;
}
.perm-item:hover {
  background: #f8fafc;
}
.perm-item > i {
  flex-shrink: 0;
  color: #059669;
  font-size: 0.78rem;
}
.perm-item-name {
  flex: 1;
  font-size: 0.82rem;
  color: #1e293b;
}
.key-badge {
  flex: 0 0 auto;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.66rem;
  font-weight: 600;
  color: #64748b;
  background: #f1f5f9;
  border: 1px solid #e8edf4;
  border-radius: 5px;
  padding: 0.12rem 0.4rem;
  white-space: nowrap;
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
  color: #0b7766;
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

@media (max-width: 1180px) {
  .detail-layout {
    grid-template-columns: 1fr;
  }
}
@media (max-width: 768px) {
  .admin-page {
    padding: 1rem;
  }
  .page-heading {
    flex-direction: column;
  }
}
</style>
