<script setup lang="ts">
import MultiSelect from "primevue/multiselect";
import Select from "primevue/select";
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

import {
  filterSizeClasses,
  useFilterSize,
  type FilterSize,
} from "./composables/useFilterSize";

type SelectMode = "single" | "multiple";

interface Props {
  disabled?: boolean;
  inputClass?: string;
  maxSelectedLabels?: number;
  mode?: SelectMode;
  modelValue: string | number | null | undefined | (string | number)[];
  options: {
    label: string;
    subLabel?: string;
    value: string | number | null;
  }[];
  placeholder?: string;
  showToggleAll?: boolean;
  size?: FilterSize | null;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  inputClass: "",
  mode: "single",
  placeholder: "",
  showToggleAll: false,
  size: null,
});

const emit = defineEmits<{
  (e: "update:modelValue", value: typeof props.modelValue): void;
}>();

const { actualSize } = useFilterSize({ size: props.size });

const filterFieldRef = ref<HTMLElement | null>(null);
const filterFieldWidth = ref(0);
let filterFieldResizeObserver: ResizeObserver | null = null;

const RESET_OPTION_VALUE = "__app_filter_select_reset__";

const placeholderValue = computed(() => props.placeholder || "Select");

const resetOption = computed(() => ({
  label: `All ${props.placeholder || "Status"}`,
  subLabel: "Show all available options",
  value: RESET_OPTION_VALUE,
}));

const selectOptions = computed(() => [resetOption.value, ...props.options]);

const isResetOption = (value: string | number | null) =>
  value === RESET_OPTION_VALUE;

const isEmptySelection = (value: typeof props.modelValue) =>
  Array.isArray(value) ? value.length === 0 : value == null || value === "";

const isOptionSelected = (value: string | number | null) => {
  if (isResetOption(value)) {
    return isEmptySelection(props.modelValue);
  }

  if (Array.isArray(props.modelValue)) {
    if (value == null) return false;
    return props.modelValue.includes(value);
  }
  return props.modelValue === value;
};

const usePrimeToggleAll = computed(() => props.showToggleAll && false);

const listPassThrough = computed(() => ({
  list: {
    class: "p-1",
  },
  option: {
    class: "!m-0 !p-0 !rounded-xl !bg-transparent",
  },
  overlay: {
    class:
      "!mt-2 !min-w-[18rem] !max-w-[calc(100vw-2rem)] !overflow-hidden !rounded-2xl !border !border-slate-200 !shadow-[0_18px_40px_rgba(15,23,42,0.12),0_4px_12px_rgba(15,23,42,0.06)]",
  },
  panel: {
    class:
      "!mt-2 !min-w-[18rem] !max-w-[calc(100vw-2rem)] !overflow-hidden !rounded-2xl !border !border-slate-200 !shadow-[0_18px_40px_rgba(15,23,42,0.12),0_4px_12px_rgba(15,23,42,0.06)]",
  },
}));

const multiSelectPassThrough = computed(() => ({
  ...listPassThrough.value,
  pcHeaderCheckbox: {
    root: {
      class: "hidden",
    },
  },
  pcOptionCheckbox: {
    root: {
      class: "hidden",
    },
  },
}));

const resolveModelValue = (val: typeof props.modelValue) => {
  if (Array.isArray(val)) {
    return val.includes(RESET_OPTION_VALUE) ? [] : val;
  }

  return val === RESET_OPTION_VALUE ? null : val;
};

const internalValue = computed<typeof props.modelValue>({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", resolveModelValue(val)),
});

const selectedChipLabels = computed(() => {
  if (!Array.isArray(props.modelValue) || props.modelValue.length === 0) {
    return [];
  }

  return props.modelValue.map(
    (value) =>
      props.options.find((option) => option.value === value)?.label ??
      String(value)
  );
});

const estimateChipWidth = (label: string) =>
  Math.min(Math.max(label.length * 7 + 30, 54), 168);

const shouldCompactSelectedChips = computed(() => {
  const labels = selectedChipLabels.value;
  if (labels.length <= 1 || filterFieldWidth.value === 0) return false;

  const availableWidth = Math.max(filterFieldWidth.value - 48, 80);
  const totalChipWidth =
    labels.reduce((total, label) => total + estimateChipWidth(label), 0) +
    (labels.length - 1) * 5;

  return totalChipWidth > availableWidth;
});

const selectedSummary = computed(() => {
  const labels = selectedChipLabels.value;
  if (labels.length === 0) return null;

  return {
    extraCount: Math.max(labels.length - 1, 0),
    label: labels[0],
  };
});

onMounted(() => {
  const updateFieldWidth = () => {
    filterFieldWidth.value = filterFieldRef.value?.clientWidth ?? 0;
  };

  updateFieldWidth();

  if (filterFieldRef.value && typeof ResizeObserver !== "undefined") {
    filterFieldResizeObserver = new ResizeObserver(updateFieldWidth);
    filterFieldResizeObserver.observe(filterFieldRef.value);
  }
});

onBeforeUnmount(() => {
  filterFieldResizeObserver?.disconnect();
});
</script>

<template>
  <div
    ref="filterFieldRef"
    class="filter-select flex items-center border border-[#d9e2ec] bg-white transition"
    :class="[
      filterSizeClasses[actualSize].wrapper,
      { 'bg-[#f1f5f9] opacity-60 pointer-events-none': disabled },
      inputClass,
    ]"
  >
    <MultiSelect
      v-if="mode === 'multiple'"
      v-model="internalValue"
      :options="selectOptions"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      display="chip"
      :class="['w-full', filterSizeClasses[actualSize].input]"
      :max-selected-labels="maxSelectedLabels"
      :pt="multiSelectPassThrough"
      :show-toggle-all="usePrimeToggleAll"
      :disabled="disabled"
    >
      <template #value="{ placeholder: valuePlaceholder }">
        <span v-if="selectedChipLabels.length === 0" class="text-slate-400">
          {{ valuePlaceholder }}
        </span>
        <span
          v-else-if="shouldCompactSelectedChips && selectedSummary"
          class="filter-select-summary-chip"
        >
          <span class="min-w-0 truncate">{{ selectedSummary.label }}</span>
          <span v-if="selectedSummary.extraCount > 0" class="shrink-0">
            +{{ selectedSummary.extraCount }}
          </span>
        </span>
        <span v-else class="filter-select-chip-list">
          <span
            v-for="(label, index) in selectedChipLabels"
            :key="`${label}-${index}`"
            class="filter-select-summary-chip"
          >
            <span class="min-w-0 truncate">{{ label }}</span>
          </span>
        </span>
      </template>
      <template #option="{ option }">
        <div
          :class="[
            'relative flex w-full min-w-0 items-start gap-2 rounded-xl border px-2.5 py-1 text-left transition-colors',
            isResetOption(option.value) && isOptionSelected(option.value)
              ? 'border-red-200 bg-red-50 text-slate-800'
              : isOptionSelected(option.value)
                ? 'border-blue-200 bg-blue-50 text-slate-800'
                : 'border-transparent text-slate-700 hover:bg-slate-50',
          ]"
        >
          <span
            :class="[
              'mt-0.5 flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded border-2 transition-colors',
              isResetOption(option.value) && isOptionSelected(option.value)
                ? 'border-red-600 bg-red-600'
                : isOptionSelected(option.value)
                  ? 'border-blue-600 bg-blue-600'
                  : 'border-slate-300 bg-white',
            ]"
          >
            <span
              v-if="isOptionSelected(option.value)"
              class="pi pi-check text-[10px] leading-none text-white"
            />
          </span>
          <span class="min-w-0 flex-1">
            <span
              :class="[
                'block truncate text-[13px] leading-5',
                isOptionSelected(option.value)
                  ? 'font-semibold text-slate-800'
                  : 'font-medium text-slate-700',
              ]"
            >
              {{ option.label }}
            </span>
            <span
              v-if="option.subLabel"
              class="block truncate text-[11px] leading-4 text-slate-400"
            >
              {{ option.subLabel }}
            </span>
          </span>
        </div>
      </template>
    </MultiSelect>

    <Select
      v-else
      v-model="internalValue"
      :options="selectOptions"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      :class="['w-full', filterSizeClasses[actualSize].input]"
      :pt="listPassThrough"
      :disabled="disabled"
    >
      <template #value="{ value, placeholder: valuePlaceholder }">
        <span v-if="value" class="filter-select-summary-chip">
          <span class="min-w-0 truncate">
            {{ selectOptions.find((option) => option.value === value)?.label }}
          </span>
        </span>
        <span v-else class="text-slate-400">{{ valuePlaceholder }}</span>
      </template>
      <template #option="{ option }">
        <div
          :class="[
            'relative flex w-full min-w-0 items-start gap-2 rounded-xl border px-2.5 py-1 text-left transition-colors',
            isResetOption(option.value) && isOptionSelected(option.value)
              ? 'border-red-200 bg-red-50 text-slate-800'
              : isOptionSelected(option.value)
                ? 'border-blue-200 bg-blue-50 text-slate-800'
                : 'border-transparent text-slate-700 hover:bg-slate-50',
          ]"
        >
          <span
            :class="[
              'mt-0.5 flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded border-2 transition-colors',
              isResetOption(option.value) && isOptionSelected(option.value)
                ? 'border-red-600 bg-red-600'
                : isOptionSelected(option.value)
                  ? 'border-blue-600 bg-blue-600'
                  : 'border-slate-300 bg-white',
            ]"
          >
            <span
              v-if="isOptionSelected(option.value)"
              class="pi pi-check text-[10px] leading-none text-white"
            />
          </span>
          <span class="min-w-0 flex-1">
            <span
              :class="[
                'block truncate text-[13px] leading-5',
                isOptionSelected(option.value)
                  ? 'font-semibold text-slate-800'
                  : 'font-medium text-slate-700',
              ]"
            >
              {{ option.label }}
            </span>
            <span
              v-if="option.subLabel"
              class="block truncate text-[11px] leading-4 text-slate-400"
            >
              {{ option.subLabel }}
            </span>
          </span>
        </div>
      </template>
    </Select>
  </div>
</template>

<style scoped>
.filter-select {
  border-radius: 12px;
  box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.04);
  overflow: visible;
}

.filter-select:focus-within {
  border-color: #94a3b8;
  box-shadow: 0 0 0 2px #dbeafe;
}

.filter-select:hover:not(:focus-within) {
  border-color: #cbd5e1;
}

:deep(.p-select),
:deep(.p-multiselect) {
  min-width: 0 !important;
  border: none !important;
  border-radius: 0 !important;
  background: transparent !important;
  box-shadow: none !important;
}

:deep(.p-select-label),
:deep(.p-multiselect-label) {
  min-width: 0 !important;
  overflow: hidden !important;
  color: #0f172a !important;
  font-weight: 400 !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
  padding: 0 12px !important;
  display: flex !important;
  align-items: center !important;
  height: 100% !important;
  line-height: 1 !important;
}

:deep(.p-select-label.p-placeholder),
:deep(.p-multiselect-label.p-placeholder) {
  color: #94a3b8 !important;
  font-size: 13px;
}

:deep(.p-multiselect-label) {
  display: flex !important;
  min-width: 0 !important;
  max-width: 100% !important;
  align-items: center !important;
  gap: 5px !important;
  overflow-x: auto !important;
  overflow-y: visible !important;
  padding: 0 10px !important;
  flex-wrap: nowrap !important;
  scrollbar-width: none;
}

:deep(.p-multiselect-label::-webkit-scrollbar) {
  display: none;
}

:deep(.p-multiselect-chip) {
  display: inline-flex !important;
  max-width: 10rem !important;
  min-width: max-content !important;
  flex: 0 0 auto !important;
  align-items: center !important;
  gap: 5px !important;
  border: 1px solid #cbd5e1 !important;
  border-radius: 10px !important;
  background: linear-gradient(180deg, #ffffff 0%, #f1f5f9 100%) !important;
  height: 23px !important;
  padding: 0 7px !important;
  color: #475569 !important;
  font-size: 11px !important;
  font-weight: 700 !important;
  line-height: 1.15 !important;
  transform: none !important;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    0 1px 2px rgba(15, 23, 42, 0.08) !important;
}

:deep(.p-multiselect-chip .p-chip-label),
:deep(.p-multiselect-chip-label) {
  min-width: 0 !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
}

:deep(.p-multiselect-chip .p-chip-remove-icon),
:deep(.p-multiselect-chip-remove-icon) {
  width: 12px !important;
  height: 12px !important;
  min-width: 12px !important;
  border-radius: 999px !important;
  color: #64748b !important;
  opacity: 0.8 !important;
}

:deep(.p-multiselect-chip .p-chip-remove-icon:hover),
:deep(.p-multiselect-chip-remove-icon:hover) {
  background: #e2e8f0 !important;
  opacity: 1 !important;
}

.filter-select-chip-list {
  display: inline-flex;
  max-width: 100%;
  min-width: 0;
  align-items: center;
  gap: 5px;
  overflow: hidden;
}

.filter-select-summary-chip {
  display: inline-flex;
  max-width: 100%;
  min-width: 0;
  height: 23px;
  align-items: center;
  gap: 5px;
  overflow: hidden;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  background: linear-gradient(180deg, #ffffff 0%, #f1f5f9 100%);
  padding: 0 7px;
  color: #475569;
  font-size: 11px;
  font-weight: 700;
  line-height: 1.15;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    0 1px 2px rgba(15, 23, 42, 0.08);
}

:deep(.p-select-dropdown),
:deep(.p-multiselect-dropdown) {
  color: #64748b !important;
  margin-right: 4px !important;
}

:deep(.p-select-overlay),
:deep(.p-select-panel),
:deep(.p-multiselect-overlay),
:deep(.p-multiselect-panel) {
  overflow: hidden !important;
  border: 1px solid #e2e8f0 !important;
  border-radius: 14px !important;
  margin-top: 8px !important;
  box-shadow:
    0 18px 40px rgba(15, 23, 42, 0.12),
    0 4px 12px rgba(15, 23, 42, 0.06) !important;
}

:deep(.p-select-items),
:deep(.p-select-list),
:deep(.p-multiselect-items),
:deep(.p-multiselect-list) {
  padding: 4px !important;
}

:deep(.p-select-item),
:deep(.p-select-option),
:deep(.p-multiselect-item),
:deep(.p-multiselect-option) {
  margin-top: 0 !important;
  padding: 0 !important;
  border: 1px solid transparent !important;
  border-radius: 12px !important;
  background: transparent !important;
  color: #334155 !important;
}

:deep(.p-multiselect-item .p-checkbox),
:deep(.p-multiselect-option .p-checkbox),
:deep(.p-multiselect-item .p-multiselect-checkbox),
:deep(.p-multiselect-option .p-multiselect-checkbox),
:deep(.p-multiselect-item [data-pc-name="checkbox"]),
:deep(.p-multiselect-option [data-pc-name="checkbox"]),
:deep(.p-multiselect-item .p-checkbox-box),
:deep(.p-multiselect-option .p-checkbox-box) {
  display: none !important;
}

:deep(
  .p-select-item:not(.p-highlight):not(.p-select-item--selected):not(
      .p-select-option-selected
    ):hover
),
:deep(
  .p-select-option:not(.p-highlight):not(.p-select-item--selected):not(
      .p-select-option-selected
    ):hover
),
:deep(
  .p-multiselect-item:not(.p-highlight):not(.p-multiselect-item--selected):not(
      .p-multiselect-option-selected
    ):hover
),
:deep(
  .p-multiselect-option:not(.p-highlight):not(
      .p-multiselect-item--selected
    ):not(.p-multiselect-option-selected):hover
) {
  background-color: #f8fafc !important;
  color: #0f172a !important;
}

:deep(.p-select-item.p-highlight),
:deep(.p-select-item.p-select-item--selected),
:deep(.p-select-option.p-highlight),
:deep(.p-select-option.p-select-option-selected),
:deep(.p-multiselect-item.p-highlight),
:deep(.p-multiselect-item.p-multiselect-item--selected),
:deep(.p-multiselect-option.p-highlight),
:deep(.p-multiselect-option.p-multiselect-option-selected) {
  border-color: #bfdbfe !important;
  background-color: #eff6ff !important;
  color: #1e293b !important;
  font-weight: 700 !important;
}

.filter-select-option {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
  padding: 9px 11px;
  border-radius: 12px;
  cursor: pointer;
  transition:
    background-color 150ms ease,
    color 150ms ease;
}

:deep(.p-select-item:hover) .filter-select-option,
:deep(.p-select-option:hover) .filter-select-option,
:deep(.p-multiselect-item:hover) .filter-select-option,
:deep(.p-multiselect-option:hover) .filter-select-option {
  background-color: transparent;
}

:deep(.p-select-item:has(.filter-select-option--active)),
:deep(.p-select-option:has(.filter-select-option--active)),
:deep(.p-multiselect-item:has(.filter-select-option--active)),
:deep(.p-multiselect-option:has(.filter-select-option--active)) {
  border-color: #bfdbfe !important;
  background-color: #eff6ff !important;
}

.filter-select-option--active {
  background-color: transparent;
}

.filter-select-option--reset {
  position: relative;
  margin-bottom: 4px;
  color: #64748b;
}

.filter-select-option--reset::after {
  position: absolute;
  right: -10px;
  bottom: -2px;
  left: -10px;
  height: 1px;
  background: #e2e8f0;
  content: "";
}

.filter-select-option__checkbox {
  display: flex;
  width: 18px;
  height: 18px;
  min-width: 18px;
  align-items: center;
  justify-content: center;
  border: 2px solid #cbd5e1;
  border-radius: 4px;
  background: #ffffff;
  transition:
    border-color 150ms ease,
    background-color 150ms ease;
}

:deep(.p-select-item:hover) .filter-select-option__checkbox,
:deep(.p-select-option:hover) .filter-select-option__checkbox,
:deep(.p-multiselect-item:hover) .filter-select-option__checkbox,
:deep(.p-multiselect-option:hover) .filter-select-option__checkbox {
  border-color: #93c5fd;
}

.filter-select-option__checkbox--active {
  border-color: #2563eb;
  background: #2563eb;
}

.filter-select-option__content {
  min-width: 0;
  flex: 1;
}

.filter-select-option__label {
  display: block;
  overflow: hidden;
  color: #334155;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.filter-select-option--active .filter-select-option__label {
  color: #1e293b;
  font-weight: 700;
}

:deep(.p-select-empty-message),
:deep(.p-multiselect-empty-message) {
  text-align: center !important;
  padding: 20px !important;
  color: #94a3b8 !important;
  font-size: 13px !important;
}

:deep(.p-multiselect-header) {
  padding: 8px 12px !important;
  border-bottom: 1px solid #f1f5f9 !important;
}
</style>
