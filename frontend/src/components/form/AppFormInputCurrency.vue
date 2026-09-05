<script setup lang="ts">
import InputNumber from "primevue/inputnumber";

const props = withDefaults(
  defineProps<{
    class?: string;
    disabled?: boolean;
    error?: string | null;
    label?: string;
    max?: number;
    maxFractionDigits?: number;
    min?: number;
    minFractionDigits?: number;
    modelValue?: number | null | undefined;
    placeholder?: string;
    readonly?: boolean;
    removeBackground?: boolean;
    required?: boolean;
    showClear?: boolean;
    size?: "small" | "large";
    useGrouping?: boolean;
  }>(),
  {
    class: "",
    disabled: false,
    error: null,
    max: undefined,
    maxFractionDigits: 0,
    min: undefined,
    minFractionDigits: 0,
    modelValue: null,
    placeholder: "",
    readonly: false,
    removeBackground: false,
    required: false,
    showClear: false,
    size: "small",
    useGrouping: true,
  }
);

const emit = defineEmits<{
  (e: "update:modelValue", value: number | null): void;
  (e: "input"): void;
}>();

const onInput = (val: number | null | undefined) => {
  emit("update:modelValue", val ?? null);
  emit("input");
};
</script>

<template>
  <div :class="['space-y-1.5', props.class]">
    <label v-if="label" class="block text-sm font-semibold text-slate-700">
      {{ label }}
      <span v-if="required" class="text-red-500">*</span>
    </label>
    <InputNumber
      :model-value="modelValue ?? null"
      mode="currency"
      currency="IDR"
      locale="id-ID"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="readonly"
      :invalid="!!error"
      :min="min"
      :max="max"
      :size="size"
      :use-grouping="useGrouping"
      :min-fraction-digits="minFractionDigits"
      :max-fraction-digits="maxFractionDigits"
      :show-clear="showClear"
      fluid
      input-class="w-full"
      :class="[{ 'is-remove-bg': removeBackground, 'is-readonly': readonly }]"
      @update:model-value="onInput"
    />
    <p v-if="error" class="text-xs text-red-600">{{ error }}</p>
    <slot />
  </div>
</template>

<style scoped>
:deep(.p-inputnumber) {
  width: 100%;
}

/* default: abu-abu, no border */
:deep(.p-inputnumber-input) {
  background-color: #e2e8f0 !important;
  border: none !important;
  box-shadow: none !important;
  color: #0f172a !important;
}

/* removeBackground: transparan + ada border */
:deep(.p-inputnumber.is-remove-bg .p-inputnumber-input) {
  background-color: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

/* hover/focus default */
:deep(.p-inputnumber:hover .p-inputnumber-input) {
  background-color: #cbd5e1 !important;
}
:deep(.p-inputnumber-input:focus) {
  background-color: #cbd5e1 !important;
  outline: 2px solid #3b82f6 !important;
  outline-offset: 0 !important;
}

/* hover/focus removeBackground */
:deep(.p-inputnumber.is-remove-bg:hover .p-inputnumber-input) {
  background-color: transparent !important;
  border-color: #94a3b8 !important;
}
:deep(.p-inputnumber.is-remove-bg .p-inputnumber-input:focus) {
  background-color: transparent !important;
  border-color: transparent !important;
  outline: 2px solid #3b82f6 !important;
  outline-offset: 0 !important;
}

/* invalid */
:deep(.p-inputnumber.p-invalid .p-inputnumber-input) {
  background-color: #fef2f2 !important;
}
:deep(.p-inputnumber.is-remove-bg.p-invalid .p-inputnumber-input) {
  background-color: transparent !important;
  border-color: #ef4444 !important;
}
:deep(.p-inputnumber.p-invalid .p-inputnumber-input:focus) {
  outline: 2px solid #ef4444 !important;
}

/* readonly */
:deep(.p-inputnumber.is-readonly) {
  pointer-events: none;
}

/* disabled */
:deep(.p-inputnumber.p-disabled .p-inputnumber-input) {
  background-color: #94a3b8 !important;
  color: #64748b !important;
  cursor: not-allowed;
  opacity: 1 !important;
}
:deep(.p-inputnumber.is-remove-bg.p-disabled .p-inputnumber-input) {
  background-color: transparent !important;
  border-color: #cbd5e1 !important;
}

:deep(.p-inputnumber-input::placeholder) {
  color: #94a3b8 !important;
  opacity: 1;
}
</style>
