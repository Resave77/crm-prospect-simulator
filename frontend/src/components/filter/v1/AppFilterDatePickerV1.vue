<script setup lang="ts">
import DatePicker from "primevue/datepicker";
import { computed } from "vue";

import {
  filterSizeClasses,
  useFilterSize,
  type FilterSize,
} from "./composables/useFilterSize";

const props = withDefaults(
  defineProps<{
    dateFormat?: string;
    disabled?: boolean;
    inputClass?: string;
    maxDate?: Date;
    minDate?: Date;
    modelValue?: string | null;
    placeholder?: string;
    size?: FilterSize | null;
  }>(),
  {
    dateFormat: "yy-mm-dd",
    disabled: false,
    inputClass: "",
    modelValue: "",
    placeholder: "Select date",
    size: null,
  }
);

const emit = defineEmits<{
  (event: "update:modelValue", value: string): void;
}>();

const { actualSize } = useFilterSize({ size: props.size });

const dateValue = computed(() => {
  if (!props.modelValue) return null;
  const [year, month, day] = props.modelValue.split("-").map(Number);
  if (!year || !month || !day) return null;
  return new Date(year, month - 1, day);
});

const formatDateLocal = (value: Date) => {
  const year = value.getFullYear();
  const month = String(value.getMonth() + 1).padStart(2, "0");
  const day = String(value.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
};

const handleUpdate = (
  value: Date | Date[] | (Date | null)[] | null | undefined
) => {
  emit(
    "update:modelValue",
    value instanceof Date ? formatDateLocal(value) : ""
  );
};
</script>

<template>
  <div
    class="filter-date flex items-center border border-[#d9e2ec] bg-white transition"
    :class="[
      filterSizeClasses[actualSize].wrapper,
      { 'pointer-events-none bg-[#f1f5f9] opacity-60': disabled },
      inputClass,
    ]"
  >
    <DatePicker
      :model-value="dateValue"
      class="w-full min-w-0"
      :date-format="dateFormat"
      :disabled="disabled"
      :max-date="maxDate"
      :min-date="minDate"
      :placeholder="placeholder"
      :pt="{
        root: { class: 'w-full min-w-0' },
        pcInputText: { root: { class: 'w-full min-w-0' } },
        panel: { class: 'max-w-[92vw]' },
      }"
      @update:model-value="handleUpdate"
    />
  </div>
</template>

<style scoped>
.filter-date {
  width: 100%;
  min-width: 0;
  overflow: hidden;
  border-radius: 8px;
}

:deep(.p-datepicker),
:deep(.p-inputtext) {
  width: 100% !important;
  min-width: 0 !important;
  height: 100% !important;
  border: 0 !important;
  border-radius: 0 !important;
  background: transparent !important;
  box-shadow: none !important;
}

:deep(.p-inputtext) {
  padding: 0 10px !important;
  color: #334155 !important;
  font-size: inherit !important;
}

:deep(.p-inputtext::placeholder) {
  color: #94a3b8 !important;
}

.filter-date:focus-within {
  border-color: #94a3b8;
  box-shadow: 0 0 0 2px #dbeafe;
}
</style>
