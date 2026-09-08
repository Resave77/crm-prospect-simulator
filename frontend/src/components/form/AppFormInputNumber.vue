<script setup lang="ts">
import InputNumber from "primevue/inputnumber";

interface Props {
  disabled?: boolean;
  error?: string | null;
  label?: string;
  max?: number;
  min?: number;
  modelValue?: number | null | undefined; // Tipe data number
  placeholder?: string;
  // Props khusus Currency
  prefix?: string; // misal 'Rp '
  readonly?: boolean;
  removeBackground?: boolean;
  required?: boolean;
  size?: "small" | "large";
}

withDefaults(defineProps<Props>(), {
  disabled: false,
  error: null,
  modelValue: null,
  placeholder: "",
  prefix: "",
  readonly: false,
  removeBackground: false,
  required: false,
  size: "small",
});

const emit = defineEmits<{
  (e: "update:modelValue", value: number | null): void;
  (e: "input", event: any): void;
  (e: "blur", event: any): void;
}>();

const onInput = (event: any) => {
  // InputNumber PrimeVue mengembalikan angka asli di event.value
  emit("update:modelValue", event.value);
  emit("input", event);
};
</script>

<template>
  <div class="space-y-1.5">
    <label v-if="label" class="block text-sm font-semibold text-slate-700">
      {{ label }}
      <span v-if="required" class="text-red-500">*</span>
    </label>

    <InputNumber
      :model-value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="readonly"
      :invalid="!!error"
      :size="size"
      :min="min"
      :max="max"
      mode="decimal"
      locale="id-ID"
      :prefix="prefix"
      :input-class="[
        'w-full text-end',
        { 'is-readonly': readonly, 'is-remove-bg': removeBackground },
      ]"
      class="w-full"
      @input="onInput"
      @blur="$emit('blur', $event)"
    />

    <p v-if="error" class="text-xs text-red-600">{{ error }}</p>
    <slot />
  </div>
</template>

<style scoped>
/* Copy paste semua style CSS dari appFormInput kamu ke sini */
/* Gunakan :deep(.p-inputnumber-input) jika .p-inputtext tidak mempan */

:deep(.p-inputnumber-input) {
  background-color: #e2e8f0 !important;
  border: none !important;
  box-shadow: none !important;
  color: #0f172a !important;
  width: 100%; /* Penting agar memenuhi container */
}

:deep(.p-inputnumber-input.is-remove-bg) {
  background-color: transparent !important;
}

/* Hover/Focus */
:deep(.p-inputnumber-input:focus) {
  background-color: #cbd5e1 !important;
  outline: 2px solid #3b82f6 !important;
  outline-offset: 0 !important;
}

/* Invalid State */
:deep(.p-inputnumber-input.p-invalid) {
  background-color: #fef2f2 !important;
}
</style>
