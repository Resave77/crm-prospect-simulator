<script setup lang="ts">
import DatePicker from "primevue/datepicker";
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    class?: string;
    dateFormat?: string;
    disabled?: boolean;
    disabledDates?: Date[];
    disabledDays?: number[];
    error?: string | null;
    hourFormat?: "12" | "24";
    inline?: boolean;
    label?: string;
    maxDate?: Date;
    minDate?: Date;
    modelValue: string | Date | Date[] | (Date | null)[] | null | undefined;
    placeholder?: string;
    readonly?: boolean;
    required?: boolean;
    selectionMode?: "single" | "multiple" | "range";
    showButtonBar?: boolean;
    showIcon?: boolean;
    showTime?: boolean;
    size?: "small" | "large";
  }>(),
  {
    dateFormat: "dd/mm/yy",
    disabled: false,
    error: null,
    hourFormat: "24",
    inline: false,
    placeholder: "",
    readonly: false,
    required: false,
    selectionMode: "single",
    showButtonBar: false,
    showIcon: true,
    showTime: false,
    size: "small",
  }
);

const isEmptyValue = (val: typeof props.modelValue) => {
  if (!val) return true;
  if (Array.isArray(val)) return val.length === 0;
  return false;
};

const effectivePlaceholder = computed(() =>
  isEmptyValue(props.modelValue) ? props.placeholder || "Select date" : ""
);

const normalizeDateString = (value: string): Date | null => {
  if (!value) return null;
  const [p0, p1, p2] = value.split(/[-/]/);
  const n0 = Number(p0);
  const n1 = Number(p1);
  const n2 = Number(p2);
  if (Number.isNaN(n0) || Number.isNaN(n1)) return null;

  // Support both YYYY-MM-DD and DD/MM/YYYY inputs
  const isYearFirst = n0 > 31 || value.includes("-");
  const year = isYearFirst ? n0 : n2;
  const month = n1;
  const day = isYearFirst ? n2 : n0;
  if (Number.isNaN(year) || Number.isNaN(month) || Number.isNaN(day))
    return null;
  return new Date(year, month - 1, day);
};

const formatDateLocal = (value: Date | null): string | null => {
  if (!value) return null;
  const y = value.getFullYear();
  const m = String(value.getMonth() + 1).padStart(2, "0");
  const d = String(value.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
};

// Convert string to Date for DatePicker without timezone shift
const dateValue = computed(() => {
  if (!props.modelValue) return null;
  if (typeof props.modelValue === "string") {
    return normalizeDateString(props.modelValue);
  }
  return props.modelValue;
});

const emit = defineEmits<{
  (
    e: "update:modelValue",
    value: string | Date | Date[] | (Date | null)[] | null | undefined
  ): void;
  (e: "change"): void;
  (e: "date-select", value: Date): void;
  (e: "today-click", value: Date): void;
  (e: "clear-click"): void;
}>();

const onChange = (
  value: Date | Date[] | (Date | null)[] | null | undefined
) => {
  // Handle null/undefined
  if (!value) {
    emit("update:modelValue", "");
    emit("change");
    return;
  }

  // Preserve string payload without timezone shift if parent passed a string
  if (typeof props.modelValue === "string") {
    if (value instanceof Date) {
      emit("update:modelValue", formatDateLocal(value));
    } else if (Array.isArray(value)) {
      const mapped = (value as (Date | null)[]).map((v) =>
        v instanceof Date ? formatDateLocal(v) : v
      );
      emit("update:modelValue", mapped as unknown as any);
    } else {
      emit("update:modelValue", value as any);
    }
  } else {
    emit("update:modelValue", value);
  }
  emit("change");
};

const onDateSelect = (value: Date) => {
  emit("date-select", value);
};

const onTodayClick = (value: Date) => {
  emit("today-click", value);
};

const onClearClick = () => {
  emit("clear-click");
};
</script>
<template>
  <div class="space-y-1.5 w-full min-w-0">
    <label v-if="label" class="block text-sm font-semibold text-slate-700">
      {{ label }} <span v-if="required" class="text-red-500">*</span>
    </label>

    <DatePicker
      :model-value="dateValue"
      :placeholder="effectivePlaceholder"
      :disabled="disabled"
      :readonly="readonly"
      :invalid="!!error"
      :show-icon="showIcon"
      :date-format="dateFormat"
      :show-time="showTime"
      :hour-format="hourFormat"
      :size="size"
      :min-date="minDate"
      :max-date="maxDate"
      :disabled-dates="disabledDates"
      :disabled-days="disabledDays"
      :inline="inline"
      :show-button-bar="showButtonBar"
      :selection-mode="selectionMode"
      :class="['w-full min-w-0', props.class, { 'is-readonly': readonly }]"
      :pt="{
        root: { class: 'w-full min-w-0' },
        input: { class: 'w-full min-w-0' },
        panel: { class: 'max-w-[92vw]' },
      }"
      @update:model-value="onChange"
      @date-select="onDateSelect"
      @today-click="onTodayClick"
      @clear-click="onClearClick"
    />

    <p v-if="error" class="text-xs text-red-600">{{ error }}</p>
    <slot />
  </div>
</template>

<style scoped>
/* icon container tetap seperti punya lu */
:deep(.p-datepicker-input-icon-container) {
  background-color: #e2e8f0 !important;
  border: none !important;
  box-shadow: none !important;
  color: #0f172a !important;
}

:deep(.p-datepicker-input-icon-container:hover) {
  background-color: #cbd5e1 !important;
}

:deep(.p-datepicker-input-icon-container:focus-within) {
  background-color: #cbd5e1 !important;
  outline: 2px solid #3b82f6 !important;
}

/* IMPORTANT: bikin input & root ngikut lebar parent */
:deep(.p-datepicker),
:deep(.p-datepicker-input),
:deep(.p-inputtext) {
  width: 100% !important;
  min-width: 0 !important;
}

/* panel kalender responsif */
:deep(.p-datepicker-panel) {
  max-width: 92vw;
}

/* kecilin placeholder text */
:deep(input::placeholder) {
  font-size: 0.75rem; /* text-xs */
  line-height: 1rem;
  color: #94a3b8; /* slate-400 */
}
</style>
