<script setup lang="ts">
import DatePicker from "primevue/datepicker";
import FloatLabel from "primevue/floatlabel";
import { computed } from "vue";

type LabelPosition = "none" | "top" | "float";
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
  label?: string;
  labelPosition?: LabelPosition;
  manualInput?: boolean;
  maxDate?: Date;
  minDate?: Date;
  modelValue?: string | null;
  placeholder?: string;
  presets?: FilterDateRangePresetOption[];
  showClearButton?: boolean;
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
  label: "",
  labelPosition: "none",
  manualInput: false,
  modelValue: "",
  placeholder: "",
  showClearButton: true,
});

const emit = defineEmits<{
  (e: "clear-click"): void;
  (e: "preset-select", value: FilterDateRangePresetKey): void;
  (e: "update:modelValue", value: string): void;
}>();

const inputId = computed(
  () => `date-range-${Math.random().toString(36).slice(2, 9)}`
);

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
    : "bg-slate-100 text-slate-600 hover:bg-slate-200";
};
</script>

<template>
  <div v-if="labelPosition === 'top'" class="app-filter-date-range-wrapper">
    <label
      v-if="label"
      :for="inputId"
      class="mb-0.5 block text-xs font-medium text-slate-500"
    >
      {{ label }}
    </label>
    <DatePicker
      :model-value="dateValue"
      :input-id="inputId"
      class="w-full"
      selection-mode="range"
      :manual-input="manualInput"
      :placeholder="placeholderValue"
      :date-format="dateFormat"
      :disabled="disabled"
      :min-date="minDate"
      :max-date="maxDate"
      :input-class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :show-button-bar="showButtonBar"
      :pt="{
        panel: { class: 'max-w-[92vw]' },
      }"
      @update:model-value="onDateChange"
    >
      <template v-if="showButtonBar" #buttonbar>
        <div class="flex w-full max-w-full min-w-0 flex-col gap-2">
          <div
            v-if="resolvedPresets.length > 0"
            class="grid w-full min-w-0 grid-cols-2 gap-2"
          >
            <button
              v-for="preset in resolvedPresets"
              :key="preset.key"
              type="button"
              class="w-full rounded-md px-2.5 py-1.5 text-center text-xs leading-tight font-medium whitespace-normal break-words transition-colors"
              :class="getPresetButtonClass(preset.key)"
              @click="applyPreset(preset.key)"
            >
              {{ preset.label }}
            </button>
          </div>
          <button
            v-if="showClearButton"
            type="button"
            class="w-full rounded-md px-2.5 py-1.5 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-700"
            @click="clearValue"
          >
            Clear
          </button>
        </div>
      </template>
    </DatePicker>
  </div>

  <FloatLabel v-else-if="labelPosition === 'float'" variant="on">
    <DatePicker
      :model-value="dateValue"
      :input-id="inputId"
      class="w-full"
      selection-mode="range"
      :manual-input="manualInput"
      :placeholder="placeholderValue"
      :date-format="dateFormat"
      :disabled="disabled"
      :min-date="minDate"
      :max-date="maxDate"
      :input-class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :show-button-bar="showButtonBar"
      :pt="{
        panel: { class: 'max-w-[92vw]' },
      }"
      @update:model-value="onDateChange"
    >
      <template v-if="showButtonBar" #buttonbar>
        <div class="flex w-full max-w-full min-w-0 flex-col gap-2">
          <div
            v-if="resolvedPresets.length > 0"
            class="grid w-full min-w-0 grid-cols-2 gap-2"
          >
            <button
              v-for="preset in resolvedPresets"
              :key="preset.key"
              type="button"
              class="w-full rounded-md px-2.5 py-1.5 text-center text-xs leading-tight font-medium whitespace-normal break-words transition-colors"
              :class="getPresetButtonClass(preset.key)"
              @click="applyPreset(preset.key)"
            >
              {{ preset.label }}
            </button>
          </div>
          <button
            v-if="showClearButton"
            type="button"
            class="w-full rounded-md px-2.5 py-1.5 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-700"
            @click="clearValue"
          >
            Clear
          </button>
        </div>
      </template>
    </DatePicker>
    <label :for="inputId" class="text-xs font-medium text-slate-500">
      {{ label }}
    </label>
  </FloatLabel>

  <template v-else>
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
      :input-class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :show-button-bar="showButtonBar"
      :pt="{
        panel: { class: 'max-w-[92vw]' },
      }"
      @update:model-value="onDateChange"
    >
      <template v-if="showButtonBar" #buttonbar>
        <div class="flex w-full max-w-full min-w-0 flex-col gap-2">
          <div
            v-if="resolvedPresets.length > 0"
            class="grid w-full min-w-0 grid-cols-2 gap-2"
          >
            <button
              v-for="preset in resolvedPresets"
              :key="preset.key"
              type="button"
              class="w-full rounded-md px-2.5 py-1.5 text-center text-xs leading-tight font-medium whitespace-normal break-words transition-colors"
              :class="getPresetButtonClass(preset.key)"
              @click="applyPreset(preset.key)"
            >
              {{ preset.label }}
            </button>
          </div>
          <button
            v-if="showClearButton"
            type="button"
            class="w-full rounded-md px-2.5 py-1.5 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-700"
            @click="clearValue"
          >
            Clear
          </button>
        </div>
      </template>
    </DatePicker>
  </template>
</template>

<style scoped>
:deep(.p-datepicker-input) {
  font-size: var(--font-select);
  min-height: 2rem;
  height: 2rem;
}

:deep(.p-inputtext::placeholder) {
  color: #64748b !important;
}

:deep(.p-datepicker-panel) {
  max-width: 92vw;
}
</style>
