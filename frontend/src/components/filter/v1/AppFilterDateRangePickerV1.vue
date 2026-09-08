<script setup lang="ts">
import DatePicker from "primevue/datepicker";
import { computed } from "vue";

import {
  filterSizeClasses,
  useFilterSize,
  type FilterSize,
} from "./composables/useFilterSize";

type FilterDateRangePresetKey =
  | "today"
  | "yesterday"
  | "last-7-days"
  | "last-30-days"
  | "custom";

type FilterDateRangePresetOption = {
  key: FilterDateRangePresetKey;
  label: string;
};

interface Props {
  dateFormat?: string;
  disabled?: boolean;
  inputClass?: string;
  manualInput?: boolean;
  maxDate?: Date;
  minDate?: Date;
  modelValue?: string | null;
  placeholder?: string;
  presets?: FilterDateRangePresetOption[];
  showClearButton?: boolean;
  size?: FilterSize | null;
}

const DEFAULT_PRESETS: FilterDateRangePresetOption[] = [
  { key: "today", label: "Today" },
  { key: "yesterday", label: "Yesterday" },
  { key: "last-7-days", label: "Last 7 Days" },
  { key: "last-30-days", label: "Last 30 Days" },
  { key: "custom", label: "Custom Range" },
];

const props = withDefaults(defineProps<Props>(), {
  dateFormat: "yy-mm-dd",
  disabled: false,
  inputClass: "",
  manualInput: false,
  modelValue: "",
  placeholder: "",
  showClearButton: true,
  size: null,
});

const { actualSize } = useFilterSize({ size: props.size });

const emit = defineEmits<{
  (e: "clear-click"): void;
  (e: "preset-select", value: FilterDateRangePresetKey): void;
  (e: "update:modelValue", value: string): void;
}>();

const placeholderValue = computed(() => props.placeholder || "Select date");
const resolvedPresets = computed(() => props.presets ?? DEFAULT_PRESETS);
const showButtonBar = computed(
  () => props.showClearButton || resolvedPresets.value.length > 0
);

const parseLocalDateString = (value: string): Date | null => {
  if (!value) return null;

  const parts = value.split("-").map(Number);
  const year = parts[0];
  const month = parts[1];
  const day = parts[2];

  if (
    year === undefined ||
    month === undefined ||
    day === undefined ||
    Number.isNaN(year) ||
    Number.isNaN(month) ||
    Number.isNaN(day) ||
    year < 1000
  ) {
    return null;
  }

  return new Date(year, month - 1, day);
};

const formatDateLocal = (date: Date) => {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
};

const createDateWithOffset = (offsetDays = 0) => {
  const date = new Date();
  date.setHours(0, 0, 0, 0);
  date.setDate(date.getDate() + offsetDays);
  return date;
};

const getPresetValue = (
  preset: Exclude<FilterDateRangePresetKey, "custom">
) => {
  if (preset === "today") {
    return formatDateLocal(createDateWithOffset(0));
  }

  if (preset === "yesterday") {
    return formatDateLocal(createDateWithOffset(-1));
  }

  if (preset === "last-7-days") {
    return `${formatDateLocal(createDateWithOffset(-6))},${formatDateLocal(
      createDateWithOffset(0)
    )}`;
  }

  return `${formatDateLocal(createDateWithOffset(-29))},${formatDateLocal(
    createDateWithOffset(0)
  )}`;
};

const activePreset = computed<FilterDateRangePresetKey | null>(() => {
  const currentValue = props.modelValue?.trim() ?? "";

  if (!currentValue) return null;
  if (currentValue === getPresetValue("today")) return "today";
  if (currentValue === getPresetValue("yesterday")) return "yesterday";
  if (currentValue === getPresetValue("last-7-days")) return "last-7-days";
  if (currentValue === getPresetValue("last-30-days")) return "last-30-days";

  return "custom";
});

const dateValue = computed(() => {
  const rawValue = props.modelValue?.trim() ?? "";
  if (!rawValue) return null;

  const parsedDates = rawValue
    .split(",")
    .filter(Boolean)
    .map((value) => parseLocalDateString(value))
    .filter((value): value is Date => value instanceof Date);

  return parsedDates.length > 0 ? parsedDates : null;
});

const updateModelValue = (value: string) => {
  emit("update:modelValue", value);
};

const onDateChange = (
  value: Date | Date[] | (Date | null)[] | null | undefined
) => {
  if (!value) {
    updateModelValue("");
    return;
  }

  if (Array.isArray(value)) {
    const [start, end] = value;

    if (!(start instanceof Date) || Number.isNaN(start.getTime())) {
      updateModelValue("");
      return;
    }

    if (!(end instanceof Date) || Number.isNaN(end.getTime())) {
      updateModelValue(formatDateLocal(start));
      return;
    }

    updateModelValue(`${formatDateLocal(start)},${formatDateLocal(end)}`);
    return;
  }

  if (value instanceof Date && !Number.isNaN(value.getTime())) {
    updateModelValue(formatDateLocal(value));
    return;
  }

  updateModelValue("");
};

const clearValue = () => {
  updateModelValue("");
  emit("clear-click");
};

const applyPreset = (preset: FilterDateRangePresetKey) => {
  emit("preset-select", preset);
  if (preset === "custom") return;

  updateModelValue(getPresetValue(preset));
};

const getPresetButtonClass = (preset: FilterDateRangePresetKey) => {
  const isActive = activePreset.value === preset;

  return isActive
    ? "bg-blue-600 text-white hover:bg-blue-700"
    : "bg-[#f1f5f9] text-[#475569] hover:bg-[#e2e8f0]";
};
</script>

<template>
  <div
    class="filter-date flex items-center border border-[#d9e2ec] bg-white transition"
    :class="[
      filterSizeClasses[actualSize].wrapper,
      { 'bg-[#f1f5f9] opacity-60 pointer-events-none': disabled },
      inputClass,
    ]"
  >
    <DatePicker
      :model-value="dateValue"
      class="w-full"
      selection-mode="range"
      :manual-input="manualInput"
      :placeholder="placeholderValue"
      :date-format="dateFormat"
      :disabled="disabled"
      :min-date="minDate"
      :max-date="maxDate"
      :input-class="['filter-date-input', filterSizeClasses[actualSize].input]"
      :show-button-bar="showButtonBar"
      :pt="{
        panel: { class: 'max-w-[92vw]' },
      }"
      @update:model-value="onDateChange"
    >
      <template v-if="showButtonBar" #buttonbar>
        <div class="flex w-full max-w-full min-w-0 flex-col gap-[8px]">
          <div
            v-if="resolvedPresets.length > 0"
            class="grid w-full min-w-0 grid-cols-2 gap-[8px]"
          >
            <button
              v-for="preset in resolvedPresets"
              :key="preset.key"
              type="button"
              class="w-full rounded-[10px] px-[10px] py-[8px] text-center text-[12px] leading-tight font-medium whitespace-normal break-words transition-colors"
              :class="getPresetButtonClass(preset.key)"
              @click="applyPreset(preset.key)"
            >
              {{ preset.label }}
            </button>
          </div>
          <button
            v-if="showClearButton"
            type="button"
            class="w-full rounded-[10px] px-[10px] py-[8px] text-[12px] font-medium text-[#64748b] transition-colors hover:bg-[#f1f5f9] hover:text-[#334155]"
            @click="clearValue"
          >
            Clear
          </button>
        </div>
      </template>
    </DatePicker>
  </div>
</template>
<style scoped>
.filter-date {
  border-radius: 12px;
  box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.04);
  overflow: hidden;
}

.filter-date:focus-within {
  border-color: #94a3b8;
  box-shadow: 0 0 0 2px #dbeafe;
}

.filter-date:hover:not(:focus-within) {
  border-color: #cbd5e1;
}

:deep(.p-datepicker-input),
:deep(.p-datepicker) {
  min-width: 0 !important;
  border: none !important;
  border-radius: 0 !important;
  background: transparent !important;
  box-shadow: none !important;
  width: 100% !important;
}

/* ── PERBAIKAN: Padding & Font Size ── */
:deep(.p-datepicker-input) {
  /* Memberi jarak 12px/16px di kiri agar tidak mepet */
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
  padding: 0 12px !important;
}

:deep(.p-datepicker-input::placeholder) {
  color: #94a3b8 !important;
  /* Menghapus font-size global agar ukuran font placeholder */
  /* otomatis mengikuti class size dari `filterSizeClasses` */
}

:deep(.p-datepicker.p-disabled) {
  background: transparent !important;
  color: #64748b !important;
  opacity: 1 !important;
}

/* ── Date Panel ─────────────────────────────────── */
:deep(.p-datepicker-panel) {
  border-radius: 16px !important;
  box-shadow:
    0 8px 24px rgba(15, 23, 42, 0.12),
    0 2px 8px rgba(15, 23, 42, 0.06) !important;
  border: 1px solid #e2e8f0 !important;
  max-width: 92vw !important;
}

/* ── Today Button Bar ───────────────────────────── */
:deep(.p-datepicker-panel) .p-datepicker-buttonbar {
  padding: 12px !important;
  border-top: 1px solid #f1f5f9 !important;
}

/* ── PERBAIKAN: Set Ukuran Font & Padding ── */
:deep(.p-datepicker-input) {
  padding: 0 12px !important;
  /* Mengatur ukuran font input teks utama (misal: 14px) */
  font-size: 14px !important;
}

:deep(.p-datepicker-input::placeholder) {
  color: #94a3b8 !important;
  /* Mengatur ukuran font placeholder agar lebih kecil (misal: 12px / text-xs) */
  font-size: 13px !important;
  font-weight: 500 !important;
}
</style>
