<script setup lang="ts">
import InputText from "primevue/inputtext";
import { computed, ref, watch } from "vue";

import {
  filterSizeClasses,
  useFilterSize,
  type FilterSize,
} from "./composables/useFilterSize";

interface Props {
  disabled?: boolean;
  inputClass?: string;
  modelValue: string;
  placeholder?: string;
  size?: FilterSize | null;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  inputClass: "",
  placeholder: "",
  size: null,
});

const placeholderValue = computed(() => props.placeholder || "Search");

const { actualSize } = useFilterSize({ size: props.size });

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();

const localValue = ref(props.modelValue);

const onInput = (value: string | undefined) => {
  localValue.value = value || "";
  emit("update:modelValue", localValue.value);
};

watch(
  () => props.modelValue,
  (newVal) => {
    if (newVal !== localValue.value) {
      localValue.value = newVal;
    }
  }
);
</script>

<template>
  <div
    class="filter-input flex items-center gap-[8px] overflow-hidden border border-[#d9e2ec] bg-white text-slate-600 transition hover:border-[#cbd5e1] focus-within:border-[#94a3b8] focus-within:shadow-[0_0_0_2px_#dbeafe]"
    :class="[
      filterSizeClasses[actualSize].wrapper,
      { 'opacity-50 pointer-events-none': disabled },
      inputClass,
    ]"
  >
    <span
      class="pi pi-search text-slate-400 shrink-0 ml-[12px]"
      :class="filterSizeClasses[actualSize].icon"
    />
    <InputText
      :model-value="localValue"
      :placeholder="placeholderValue"
      :disabled="disabled"
      :class="[
        'w-full border-none bg-transparent focus:ring-0',
        filterSizeClasses[actualSize].input,
      ]"
      @update:model-value="onInput"
    />
  </div>
</template>

<style scoped>
.filter-input {
  border-radius: 12px;
  box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.04);
  padding: 0 !important;
}

.filter-input :deep(.p-inputtext) {
  min-width: 0 !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
  padding: 0 !important;
  border: none !important;
  outline: none !important;
  box-shadow: none !important;
  background: transparent !important;
  height: auto !important;
  line-height: 1.5 !important;
  border-radius: 0 !important;
}

.filter-input :deep(.p-inputtext:focus),
.filter-input :deep(.p-inputtext:enabled:focus) {
  border: none !important;
  outline: none !important;
  box-shadow: none !important;
}

.filter-input :deep(.p-inputtext::placeholder) {
  color: #94a3b8 !important;
  font-size: 13px;
}

.filter-input :deep(.p-inputtext:disabled) {
  background: transparent !important;
  color: #64748b !important;
}
</style>
