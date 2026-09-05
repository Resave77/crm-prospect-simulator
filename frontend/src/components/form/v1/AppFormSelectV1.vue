<script setup lang="ts">
import Select from "primevue/select";
import { computed, ref } from "vue";

import type { DialogMobileBottomBarAction } from "@/composables/useDialogMobileBottomBar";

import AppDialog from "@/components/base/AppDialog.vue";
import isMobile from "@/composables/useIsMobile";

type SelectValue = string | number | null;

const props = withDefaults(
  defineProps<{
    badgeLabel?: string;
    disabled?: boolean;
    error?: string | null;
    helper?: string;
    label?: string;
    mobileModal?: boolean;
    mode?: "single" | "multiple";
    modelValue: any;
    options: { label: string; value: SelectValue }[];
    placeholder?: string;
    readonly?: boolean;
    required?: boolean;
  }>(),
  {
    badgeLabel: "",
    disabled: false,
    error: null,
    helper: "",
    mobileModal: true,
    mode: "single",
    placeholder: "",
    readonly: false,
    required: false,
  }
);

const emit = defineEmits<{
  (event: "update:modelValue", value: any): void;
  (event: "change"): void;
}>();

const onChange = (value: any) => {
  emit("update:modelValue", value);
  emit("change");
};

const dialogVisible = ref(false);
const localSearch = ref("");

const openMobileDialog = () => {
  localSearch.value = "";
  dialogVisible.value = true;
};

const isSelected = (val: SelectValue) => {
  if (props.modelValue == null) return false;
  if (Array.isArray(props.modelValue)) {
    return props.modelValue.includes(val);
  }
  return props.modelValue === val;
};

const toggleOption = (optVal: SelectValue) => {
  if (props.mode === "multiple") {
    const current = Array.isArray(props.modelValue)
      ? [...props.modelValue]
      : [];
    const index = current.indexOf(optVal);
    if (index > -1) {
      current.splice(index, 1);
    } else {
      current.push(optVal);
    }
    onChange(current);
  } else {
    onChange(optVal);
    dialogVisible.value = false;
  }
};

const filteredOptions = computed(() => {
  if (!localSearch.value.trim()) return props.options;
  const q = localSearch.value.trim().toLowerCase();
  return props.options.filter((opt) => opt.label.toLowerCase().includes(q));
});

const displayLabel = computed(() => {
  if (
    props.modelValue == null ||
    (Array.isArray(props.modelValue) && props.modelValue.length === 0)
  ) {
    return props.placeholder || "Select";
  }

  if (props.mode === "multiple" && Array.isArray(props.modelValue)) {
    const labels = props.modelValue
      .map((val) => props.options.find((opt) => opt.value === val)?.label)
      .filter(Boolean);
    return labels.join(", ");
  }

  const found = props.options.find((opt) => opt.value === props.modelValue);
  return found ? found.label : String(props.modelValue);
});

const dialogBottomBarActions = computed<DialogMobileBottomBarAction[]>(() => {
  if (props.mode === "multiple") {
    return [
      {
        icon: "pi-filter-slash",
        key: "clear",
        label: "Clear All",
        onClick: () => {
          onChange([]);
          dialogVisible.value = false;
        },
        type: "button",
      },
      {
        icon: "pi-check",
        key: "apply",
        label: "Apply",
        onClick: () => {
          dialogVisible.value = false;
        },
        type: "button",
      },
    ];
  } else {
    return [
      {
        icon: "pi-filter-slash",
        key: "clear",
        label: "Clear Selection",
        onClick: () => {
          onChange(null);
          dialogVisible.value = false;
        },
        type: "button",
      },
      {
        icon: "pi-times",
        key: "close",
        label: "Close",
        onClick: () => {
          dialogVisible.value = false;
        },
        type: "button",
      },
    ];
  }
});
</script>

<template>
  <div class="flex flex-col gap-[8px]">
    <label
      v-if="label"
      class="flex flex-wrap items-center gap-[6px] text-[12px] font-semibold text-[#334155]"
    >
      <span>{{ label }}</span>
      <span v-if="required" class="text-[#dc2626]">*</span>
      <span
        v-if="badgeLabel"
        class="rounded-full border border-[#e2e8f0] bg-[#f8fafc] px-[8px] py-[2px] text-[11px] font-semibold text-[#64748b]"
      >
        {{ badgeLabel }}
      </span>
    </label>

    <template v-if="isMobile && mobileModal">
      <div
        class="filter-select flex w-full items-center border border-[#d9e2ec] bg-white transition cursor-pointer px-3 justify-between"
        :class="[
          error ? 'border-red-300' : 'border-[#d9e2ec]',
          disabled || readonly
            ? 'bg-[#f1f5f9] opacity-60 pointer-events-none'
            : '',
        ]"
        @click.stop="openMobileDialog"
      >
        <span
          class="truncate text-[12px]"
          :class="modelValue != null ? 'text-[#0f172a]' : 'text-[#94a3b8]'"
        >
          {{ displayLabel }}
        </span>
        <span
          class="pi pi-chevron-down text-slate-400 text-[12px] shrink-0 ml-2"
        />
      </div>

      <AppDialog
        :model-value="dialogVisible"
        :title="label || placeholder || 'Select'"
        :mobile-actions="dialogBottomBarActions"
        @update:model-value="dialogVisible = $event"
      >
        <div class="flex flex-col h-full min-h-0 gap-3">
          <div v-if="options.length > 5" class="relative w-full shrink-0">
            <span
              class="pi pi-search absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400 text-sm"
            />
            <input
              v-model="localSearch"
              type="text"
              placeholder="Search..."
              class="w-full min-w-0 h-11 overflow-hidden text-ellipsis whitespace-nowrap pl-10 pr-4 text-[14px] border border-slate-200 rounded-xl bg-slate-50 focus:bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-100 outline-none transition"
            />
          </div>

          <div
            class="flex-1 overflow-y-auto min-h-0 py-2 -mx-4 px-4 flex flex-col gap-1"
          >
            <div
              v-for="option in filteredOptions"
              :key="String(option.value)"
              class="flex items-center gap-3 py-3 px-3 rounded-xl cursor-pointer hover:bg-slate-50 transition active:bg-slate-100"
              :class="{
                'bg-rose-50/50 hover:bg-rose-50/70': isSelected(option.value),
              }"
              @click.stop="toggleOption(option.value)"
            >
              <div
                class="w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0 transition"
                :class="
                  isSelected(option.value)
                    ? 'border-red-600 bg-red-600'
                    : 'border-slate-300 bg-white'
                "
              >
                <div
                  class="w-2 h-2 rounded-full bg-white transition scale-0"
                  :class="{ 'scale-100': isSelected(option.value) }"
                />
              </div>

              <span
                class="text-[14px] font-medium truncate flex-1 min-w-0"
                :class="
                  isSelected(option.value)
                    ? 'text-red-600 font-semibold'
                    : 'text-slate-700'
                "
              >
                {{ option.label }}
              </span>
            </div>

            <div
              v-if="filteredOptions.length === 0"
              class="py-12 text-center text-slate-400 text-sm"
            >
              No options found
            </div>
          </div>
        </div>
      </AppDialog>
    </template>

    <div
      v-else
      class="filter-select flex items-center"
      :class="[
        error ? 'border-red-300' : 'border-[#d9e2ec]',
        disabled ? 'bg-[#f1f5f9] opacity-60 pointer-events-none' : '',
      ]"
    >
      <Select
        :model-value="modelValue"
        :options="options"
        option-label="label"
        option-value="value"
        :placeholder="placeholder"
        :disabled="disabled"
        :invalid="!!error"
        :class="['w-full']"
        @update:model-value="onChange"
      >
        <template #option="{ option, selected }">
          <div
            class="simple-option"
            :class="{ 'simple-option--selected': selected }"
          >
            <div class="simple-option__radio">
              <div class="simple-option__radio-dot" />
            </div>
            <span class="simple-option__label">{{ option.label }}</span>
          </div>
        </template>
      </Select>
    </div>

    <span v-if="error" class="text-[11px] text-red-600">
      {{ error }}
    </span>
    <span v-else-if="helper" class="text-[11px] text-[#64748b]">
      {{ helper }}
    </span>
    <slot v-else />
  </div>
</template>

<style scoped>
.filter-select {
  height: 48px;
  border-radius: 14px;
  border-width: 1px;
  border-style: solid;
  background: #ffffff;
  box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.04);
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease;
  overflow: hidden;
}

.filter-select:focus-within {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.12);
}

.filter-select:has(.p-select.p-invalid) {
  border-color: #fca5a5 !important;
}

.filter-select:has(.p-select.p-invalid):focus-within {
  box-shadow: 0 0 0 2px #fee2e2 !important;
}

.filter-select:not(:has(.p-select.p-invalid)):not(:focus-within):hover {
  border-color: #cbd5e1;
}

:deep(.p-select) {
  min-width: 0 !important;
  border: none !important;
  border-radius: 0 !important;
  background: transparent !important;
  box-shadow: none !important;
  height: 100% !important;
}

:deep(.p-select-label) {
  min-width: 0 !important;
  overflow: hidden !important;
  color: #0f172a !important;
  font-size: 12px !important;
  font-weight: 400 !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
  padding: 0 12px !important;
  display: flex !important;
  align-items: center !important;
  height: 100% !important;
}

:deep(.p-select-label.p-placeholder) {
  color: #94a3b8 !important;
}

:deep(.p-select-dropdown) {
  color: #64748b !important;
  margin-right: 4px !important;
}

:deep(.p-select.p-disabled) {
  background: transparent !important;
}

:deep(.p-select-overlay) {
  border-radius: 14px !important;
  box-shadow:
    0 8px 24px rgba(15, 23, 42, 0.12),
    0 2px 8px rgba(15, 23, 42, 0.06) !important;
  border: 1px solid #e2e8f0 !important;
  margin-top: 4px !important;
}

:deep(.p-select-items) {
  padding: 6px !important;
}

:deep(.p-select-item) {
  margin: 0 !important;
  padding: 0 !important;
  border-radius: 4px !important;
  background: transparent !important;
}

:deep(.p-select-item--selected) {
  background-color: #fff1f2 !important;
}

:deep(.p-select-item--selected):hover {
  background-color: #fff1f2 !important;
}

:deep(.p-select-item):not(.p-select-item--selected):hover {
  background-color: #fee2e2 !important;
}

:deep(.p-select-item--selected):hover .simple-option {
  background-color: #fff1f2 !important;
}

.simple-option {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 6px;
  border-radius: 4px;
  cursor: pointer;
  background: transparent !important;
}

.simple-option--selected {
  background-color: #fff1f2;
}

.simple-option:not(.simple-option--selected):hover {
  background-color: #fff1f2;
}

.simple-option__radio {
  width: 14px;
  height: 14px;
  min-width: 14px;
  border-radius: 50%;
  border: 2px solid #cbd5e1;
  display: flex;
  align-items: center;
  justify-content: center;
  transition:
    border-color 150ms ease,
    background-color 150ms ease,
    transform 150ms ease;
}

.simple-option--selected .simple-option__radio {
  border-color: #dc2626;
  background-color: #dc2626;
}

.simple-option__radio-dot {
  width: 6px;
  height: 6px;
  min-width: 6px;
  border-radius: 50%;
  background-color: transparent;
  transform: scale(0.5);
  transition:
    background-color 150ms ease,
    transform 150ms ease;
}

.simple-option--selected .simple-option__radio-dot {
  background-color: #ffffff;
  transform: scale(1);
}

.simple-option__label {
  font-size: 12px;
  font-weight: 500;
  color: #0f172a;
  line-height: 1.3;
  white-space: nowrap;
}

.simple-option--selected .simple-option__label {
  color: #dc2626;
  font-weight: 600;
}

:deep(.p-select-option-check-icon) {
  display: none !important;
}

:deep(.p-select-option-blank-icon) {
  display: none !important;
}

:deep(.p-select-empty-message) {
  text-align: center !important;
  padding: 12px !important;
  color: #94a3b8 !important;
  font-size: 12px !important;
}
</style>
