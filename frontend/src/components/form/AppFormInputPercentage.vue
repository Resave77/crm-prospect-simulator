<script setup lang="ts">
import InputNumber from "primevue/inputnumber";

import { toNumberOrNull } from "@/lib/utils/formatters";

const props = withDefaults(
  defineProps<{
    class?: string;
    disabled?: boolean;
    error?: string | null;
    label?: string;

    max?: number;

    maxFractionDigits?: number;
    /** default 0..100 */
    min?: number;
    minFractionDigits?: number;
    /** nilai yang disimpan: 0..100 (bukan 0..1) */
    modelValue?: number | null | undefined;

    placeholder?: string;
    readonly?: boolean;

    removeBackground?: boolean;
    required?: boolean;

    /** default: true -> tampil "12%" */
    showSuffix?: boolean;

    /** default: false (biar ga ada grouping 1,000) */
    useGrouping?: boolean;
  }>(),
  {
    class: "",
    disabled: false,
    error: null,
    max: 100,
    maxFractionDigits: 2,
    min: 0,
    minFractionDigits: 0,
    modelValue: null,
    placeholder: "",
    readonly: false,
    removeBackground: false,
    required: false,
    showSuffix: true,
    useGrouping: false,
  }
);

const emit = defineEmits<{
  (e: "update:modelValue", value: number | null): void;
  (e: "input"): void;
}>();

const onInput = (val: number | null | undefined) => {
  emit("update:modelValue", val ?? 0);
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
      mode="decimal"
      locale="id-ID"
      :min="min"
      :max="max"
      :min-fraction-digits="minFractionDigits"
      :max-fraction-digits="maxFractionDigits"
      :use-grouping="useGrouping"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="readonly"
      :invalid="!!error"
      :suffix="showSuffix ? '%' : undefined"
      input-class="w-full no-spinner"
      :class="[{ 'is-remove-bg': removeBackground, 'is-readonly': readonly }]"
      fluid
      @input="(e) => emit('update:modelValue', toNumberOrNull(e.value))"
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

/* removeBackground: transparan */
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
}
:deep(.p-inputnumber.is-remove-bg .p-inputnumber-input:focus) {
  background-color: transparent !important;
  outline: 2px solid #3b82f6 !important;
  outline-offset: 0 !important;
}

/* invalid */
:deep(.p-inputnumber.p-invalid .p-inputnumber-input) {
  background-color: #fef2f2 !important;
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

:deep(.p-inputnumber-input::placeholder) {
  color: #94a3b8 !important;
  opacity: 1;
}

/* HILANGIN SPINNER/ARROW (Chrome/Safari/Edge) */
:deep(input.no-spinner::-webkit-outer-spin-button),
:deep(input.no-spinner::-webkit-inner-spin-button) {
  -webkit-appearance: none;
  margin: 0;
}
</style>
