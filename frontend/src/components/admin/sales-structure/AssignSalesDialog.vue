<script setup lang="ts">
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import type { AdminUserListItem, SalesRoleLevel } from '../../../types/admin'

interface PickerUserOption {
  label: string
  value: AdminUserListItem
  user: AdminUserListItem
  disabled: boolean
  alreadyAssigned: boolean
}

interface RoleOption {
  label: string
  value: string
  level: SalesRoleLevel
  description: string
}

interface ParentOption {
  label: string
  value: string
  level: SalesRoleLevel
}

defineProps<{
  visible: boolean
  user: AdminUserListItem | null
  salesRoleId: string
  parentUserId: string | null
  effectiveMonth: string
  pickerUserOptions: PickerUserOption[]
  roleOptions: RoleOption[]
  parentOptions: ParentOption[]
  pickerUsersCount: number
  pickerTotal: number
  hasMorePickerUsers: boolean
  pickerLoading: boolean
  dialogStructureLoading: boolean
  saving: boolean
  error: string
  selectedRoleLabel: string
  selectedLevel: number | null
  requiredParentLevel: number | null
  parentUnavailable: boolean
  previewParentLabel: string
  previewMonthLabel: string
  dialogEffectiveDate: string
  canSubmit: boolean
  roleLabel: (role: string) => string
  roleSeverity: (role: string) => string
}>()

const emit = defineEmits<{
  'update:visible': [visible: boolean]
  'update:user': [user: AdminUserListItem | null]
  'update:salesRoleId': [salesRoleId: string]
  'update:parentUserId': [parentUserId: string | null]
  'update:effectiveMonth': [effectiveMonth: string]
  'filter-users': [event: { value?: unknown }]
  'load-more-users': []
  submit: []
  cancel: []
}>()
</script>

<template>
  <Dialog
    :visible="visible"
    header="Assign Sales to Team"
    modal
    :draggable="false"
    :closable="!saving"
    :closeOnEscape="!saving"
    :dismissableMask="false"
    :style="{ width: 'min(620px, 94vw)' }"
    @update:visible="emit('update:visible', $event)"
  >
    <div class="dialog-form">
      <Message v-if="error" severity="error">{{ error }}</Message>

      <div class="assign-section">
        <div class="assign-section-title"><span class="assign-step">1</span>Select Sales User</div>
        <Select
          :modelValue="user"
          :options="pickerUserOptions"
          optionLabel="label"
          optionValue="value"
          :optionDisabled="(opt) => opt.disabled"
          :loading="pickerLoading"
          filter
          filterPlaceholder="Search by name, employee ID, or email"
          placeholder="Search active sales user"
          @update:modelValue="emit('update:user', $event)"
          @filter="emit('filter-users', $event)"
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
              <span v-if="pickerTotal" class="cell-hint">{{ pickerUsersCount }} of {{ pickerTotal }} active accounts</span>
              <Button v-if="hasMorePickerUsers" label="Load more" icon="pi pi-chevron-down" text size="small" :loading="pickerLoading" @click="emit('load-more-users')" />
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
        <Select :modelValue="salesRoleId" :options="roleOptions" optionLabel="label" optionValue="value" filter placeholder="Select an active role" @update:modelValue="emit('update:salesRoleId', $event)">
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
            :modelValue="parentUserId"
            :options="parentOptions"
            optionLabel="label"
            optionValue="value"
            filter
            :loading="dialogStructureLoading"
            :placeholder="`Select parent from Level ${requiredParentLevel}...`"
            @update:modelValue="emit('update:parentUserId', $event)"
          />
          <Message v-if="parentUnavailable" severity="warn">No eligible Level {{ requiredParentLevel }} parent exists for {{ previewMonthLabel || effectiveMonth }}. Assign the required upper level first.</Message>
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
          <input :value="effectiveMonth" type="month" @input="emit('update:effectiveMonth', ($event.target as HTMLInputElement).value)" />
        </div>
        <span class="section-note">Changes apply from the first day of that month ({{ dialogEffectiveDate }}).</span>
      </div>

      <div class="assign-section">
        <div class="assign-section-title"><span class="assign-step">5</span>Assignment Preview</div>
        <div class="preview-card">
          <div class="preview-row"><span>Sales</span><strong>{{ user?.fullName || '—' }}<template v-if="user?.employeeId"><span class="preview-sub"> · {{ user.employeeId }}</span></template></strong></div>
          <div class="preview-row"><span>Organizational Role</span><strong>{{ selectedRoleLabel || '—' }}</strong></div>
          <div class="preview-row"><span>Reports To</span><strong>{{ previewParentLabel }}</strong></div>
          <div class="preview-row"><span>Effective From</span><strong>{{ previewMonthLabel || '—' }}</strong></div>
        </div>
      </div>
    </div>
    <template #footer>
      <Button label="Cancel" severity="secondary" text :disabled="saving" @click="emit('cancel')" />
      <Button label="Assign" icon="pi pi-check" :loading="saving" :disabled="!canSubmit || saving" @click="emit('submit')" />
    </template>
  </Dialog>
</template>

<style scoped>
.dialog-form { display: grid; gap: 1.1rem; max-height: min(68vh, 720px); overflow-y: auto; overscroll-behavior: contain; padding-right: 0.35rem; }
:deep(.p-dialog) { margin: 1rem; max-height: calc(100vh - 2rem); }
:deep(.p-dialog-content) { overflow: hidden; }
.assign-section { display: grid; gap: 0.55rem; }
.assign-section-title { display: flex; align-items: center; gap: 0.55rem; font-weight: 750; color: #0f172a; font-size: 0.86rem; }
.assign-step { width: 22px; height: 22px; display: grid; place-content: center; border-radius: 999px; background: #0b7766; color: #fff; font-size: 0.7rem; font-weight: 700; }
.section-note { color: #8492a6; font-size: 0.75rem; }
.parent-info { display: flex; gap: 0.5rem; align-items: center; padding: 0.6rem 0.75rem; border: 1px solid #f4b3ba; border-radius: 8px; background: #fff0f1; color: #d62839; font-size: 0.8rem; line-height: 1.4; }
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
.cell-hint { color: #94a3b8; font-size: 0.75rem; }
.form-field { display: grid; gap: 0.35rem; }
input[type='month'] { border: 1px solid #dbe3ee; border-radius: 6px; padding: 0.62rem; font: inherit; width: 100%; background: #ffffff; color: #0f172a; }
:deep(.p-inputtext), :deep(.p-select), :deep(.p-textarea) { background: #ffffff; color: #0f172a; border-color: #dbe3ee; }
:deep(.p-select-label), :deep(.p-select-dropdown), :deep(.p-inputtext::placeholder) { color: #64748b; }
:deep(.p-dialog), :deep(.p-dialog-header), :deep(.p-dialog-content), :deep(.p-dialog-footer) { background: #ffffff; color: #0f172a; }
:deep(.p-select-overlay), :deep(.p-select-list), :deep(.p-select-option) { background: #ffffff; color: #0f172a; }
:deep(.p-select-option.p-select-option-selected), :deep(.p-select-option:hover) { background: #f1f5f9; color: #0f172a; }
</style>
