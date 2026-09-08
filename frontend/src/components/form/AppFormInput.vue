<script setup lang="ts">
import InputText from "primevue/inputtext";

withDefaults(
  defineProps<{
    disabled?: boolean;
    error?: string | null;
    label?: string;
    maxlength?: number;
    modelValue?: string | null | undefined;
    placeholder?: string;
    readonly?: boolean;
    removeBackground?: boolean;
    required?: boolean;
    size?: "small" | "large";
    type?: string;
  }>(),
  {
    disabled: false,
    error: null,
    maxlength: undefined,
    modelValue: "",
    placeholder: "",
    readonly: false,
    removeBackground: false,
    required: false,
    size: "small",
    type: "text",
  }
);

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
  (e: "input"): void;
}>();

const onInput = (val: string | undefined) => {
  emit("update:modelValue", val ?? "");
  emit("input");
};
</script>

<template>
  <div class="space-y-1.5">
    <label v-if="label" class="block text-sm font-semibold text-slate-700">
      {{ label }}
      <span v-if="required" class="text-red-500">*</span>
    </label>
    <InputText
      :model-value="modelValue ?? ''"
      :type="type || 'text'"
      :maxlength="maxlength"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="readonly"
      :invalid="!!error"
      :size="size"
      :class="[
        'w-full',
        { 'is-readonly': readonly, 'is-remove-bg': removeBackground },
      ]"
      @update:model-value="onInput"
    />
    <p v-if="error" class="text-xs text-red-600">{{ error }}</p>
    <slot />
  </div>
</template>

<style scoped>
/* default (removeBackground = false): abu-abu, no border */
:deep(.p-inputtext) {
  background-color: #e2e8f0 !important;
  border: none !important;
  box-shadow: none !important;
  color: #0f172a !important;
}

/* removeBackground = true: transparan + ada border */
:deep(.p-inputtext.is-remove-bg) {
  background-color: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

/* hover/focus default */
:deep(.p-inputtext:hover) {
  background-color: #cbd5e1 !important;
}
:deep(.p-inputtext:focus) {
  background-color: #cbd5e1 !important;
  outline: 2px solid #3b82f6 !important;
  outline-offset: 0 !important;
}

/* hover/focus ketika removeBackground */
:deep(.p-inputtext.is-remove-bg:hover) {
  background-color: transparent !important;
  border-color: #94a3b8 !important;
}
:deep(.p-inputtext.is-remove-bg:focus) {
  background-color: transparent !important;
  border-color: transparent !important; /* biar outline yang keliatan */
  outline: none !important;
  outline-offset: 0 !important;
}

/* invalid */
:deep(.p-inputtext.p-invalid) {
  background-color: #fef2f2 !important;
}
:deep(.p-inputtext.is-remove-bg.p-invalid) {
  background-color: transparent !important;
  border-color: #ef4444 !important;
}
:deep(.p-inputtext.p-invalid:focus) {
  outline: 2px solid #ef4444 !important;
}

/* readonly */
:deep(.is-readonly .p-inputtext) {
  pointer-events: none;
}

/* disabled */
:deep(.p-inputtext:disabled) {
  background-color: #94a3b8 !important;
  color: #64748b !important;
  cursor: not-allowed;
  opacity: 1 !important;
}
:deep(.p-inputtext.is-remove-bg:disabled) {
  background-color: transparent !important;
  border-color: #cbd5e1 !important;
}

/* placeholder */
:deep(.p-inputtext::placeholder) {
  color: #94a3b8 !important;
  opacity: 1;
}

/* Chrome, Edge, Safari */
:deep(.p-inputtext[type="number"]::-webkit-inner-spin-button),
:deep(.p-inputtext[type="number"]::-webkit-outer-spin-button) {
  -webkit-appearance: none;
  margin: 0;
}
</style>
