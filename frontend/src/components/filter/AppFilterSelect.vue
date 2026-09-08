<!-- TODO: masih ada bug, dia hit ke api padahal scroll belum sampai habis -->
<script setup lang="ts">
import FloatLabel from "primevue/floatlabel";
import MultiSelect from "primevue/multiselect";
import Select from "primevue/select";
import { computed } from "vue";

type SelectMode = "single" | "multiple";
type LabelPosition = "none" | "top" | "float";

interface Props {
  disabled?: boolean;
  inputClass?: string;
  label?: string;
  labelPosition?: LabelPosition;
  maxSelectedLabels?: number;
  mode?: SelectMode;
  modelValue:
    | number
    | number[]
    | string
    | string[]
    | boolean
    | boolean[]
    | null;
  options: {
    label: string;
    value: string | number | boolean | null;
  }[];
  placeholder?: string;
  showToggleAll?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  inputClass: "",
  label: "",
  labelPosition: "none",
  maxSelectedLabels: 2,
  mode: "single",
  placeholder: "",
  showToggleAll: false,
});

const inputId = computed(
  () => `select-${Math.random().toString(36).slice(2, 9)}`
);

const placeholderValue = computed(() => props.placeholder || "Select");

const emit = defineEmits<{
  (
    e: "update:modelValue",
    value: number | number[] | string | string[] | boolean | boolean[] | null
  ): void;
}>();

const internalValue = computed({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", val),
});
</script>

<template>
  <!-- Label Position: Top -->
  <div v-if="labelPosition === 'top'" class="app-filter-select-wrapper">
    <label
      v-if="label"
      :for="inputId"
      class="mb-0.5 block text-xs font-medium text-slate-500"
    >
      {{ label }}
    </label>
    <MultiSelect
      v-if="mode === 'multiple'"
      v-model="internalValue"
      :input-id="inputId"
      :options="options"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      filter
      display="chip"
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :max-selected-labels="maxSelectedLabels"
      :show-toggle-all="showToggleAll"
      :disabled="disabled"
    />
    <Select
      v-else
      v-model="internalValue"
      :input-id="inputId"
      :options="options"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      filter
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :disabled="disabled"
    />
  </div>

  <!-- Label Position: Float -->
  <FloatLabel v-else-if="labelPosition === 'float'" variant="on">
    <MultiSelect
      v-if="mode === 'multiple'"
      v-model="internalValue"
      :input-id="inputId"
      :options="options"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      filter
      display="chip"
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :max-selected-labels="maxSelectedLabels"
      :show-toggle-all="showToggleAll"
      :disabled="disabled"
    />
    <Select
      v-else
      v-model="internalValue"
      :input-id="inputId"
      :options="options"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      filter
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :disabled="disabled"
    />
    <label :for="inputId" class="text-xs font-medium text-slate-500">
      {{ label }}
    </label>
  </FloatLabel>

  <!-- Label Position: None (default) -->
  <template v-else>
    <MultiSelect
      v-if="mode === 'multiple'"
      v-model="internalValue"
      :options="options"
      option-label="label"
      option-value="value"
      :placeholder="placeholder"
      filter
      display="chip"
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :max-selected-labels="maxSelectedLabels"
      :show-toggle-all="showToggleAll"
      :disabled="disabled"
    />
    <Select
      v-else
      v-model="internalValue"
      :options="options"
      option-label="label"
      option-value="value"
      :placeholder="placeholder"
      filter
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :disabled="disabled"
    />
  </template>
</template>

<style scoped>
:deep(.p-multiselect),
:deep(.p-select) {
  font-size: var(--font-select);
  min-height: 2rem;
  height: 2rem;
}

:deep(.p-multiselect-label),
:deep(.p-select-label) {
  padding-top: 0.25rem;
  padding-bottom: 0.25rem;
}

:deep(.p-multiselect-header) {
  padding: 0.35rem 0.5rem;
}

:deep(.p-multiselect-filter),
:deep(.p-select-filter) {
  height: 2rem;
  font-size: var(--font-select);
}

:deep(.p-multiselect-items .p-multiselect-item),
:deep(.p-select-items .p-select-item) {
  font-size: calc(var(--font-select) * 0.92);
  padding-top: 0.35rem;
  padding-bottom: 0.35rem;
}

:deep(.p-multiselect-label.p-placeholder),
:deep(.p-select-label.p-placeholder) {
  color: #64748b !important;
}
</style>
