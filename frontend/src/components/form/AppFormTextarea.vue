<script setup lang="ts">
import Textarea from "primevue/textarea";

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
    rows?: number;
    size?: "small" | "large";
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
    rows: 3,
    size: "small",
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
    <Textarea
      :model-value="modelValue ?? ''"
      :placeholder="placeholder"
      :rows="rows"
      :maxlength="maxlength"
      :disabled="disabled"
      :readonly="readonly"
      :invalid="!!error"
      :size="size"
      :class="[
        'w-full',
        {
          'is-readonly': readonly,
          'is-remove-bg': removeBackground, // ✅
        },
      ]"
      @update:model-value="onInput"
    />
    <p v-if="error" class="text-xs text-red-600">{{ error }}</p>
    <slot />
  </div>
</template>
<style scoped>
/* DEFAULT */
:deep(textarea.p-textarea) {
  background-color: #e2e8f0 !important;
  border: none !important;
  box-shadow: none !important;
  color: #0f172a !important;
}

:deep(textarea.p-textarea:hover) {
  background-color: #cbd5e1 !important;
}

:deep(textarea.p-textarea:focus) {
  background-color: #cbd5e1 !important;
  outline: none !important;
  outline-offset: 0 !important;
}

:deep(textarea.p-textarea.is-remove-bg) {
  background-color: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

:deep(textarea.p-textarea.is-remove-bg:hover) {
  background-color: transparent !important;
  border-color: transparent !important;
}

:deep(textarea.p-textarea.is-remove-bg:focus) {
  background-color: transparent !important;
  border-color: transparent !important;
  outline: none !important;
  outline-offset: 0 !important;
}

/* invalid */
:deep(textarea.p-textarea.p-invalid) {
  background-color: #fef2f2 !important;
}
:deep(textarea.p-textarea.is-remove-bg.p-invalid) {
  background-color: transparent !important;
  border-color: #ef4444 !important;
}
:deep(textarea.p-textarea.p-invalid:focus) {
  outline: 2px solid #ef4444 !important;
}

/* readonly */
:deep(.is-readonly textarea.p-textarea) {
  pointer-events: none;
}

/* disabled */
:deep(textarea.p-textarea:disabled) {
  background-color: #94a3b8 !important;
  color: #64748b !important;
  cursor: not-allowed;
  opacity: 1 !important;
}
:deep(textarea.p-textarea.is-remove-bg:disabled) {
  background-color: transparent !important;
  border-color: #cbd5e1 !important;
}

/* placeholder */
:deep(textarea.p-textarea::placeholder) {
  color: #94a3b8 !important;
  opacity: 1;
}
</style>
