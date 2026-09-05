<script setup lang="ts">
import Select from "primevue/select";
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    class?: string;
    disabled?: boolean;
    error?: string | null;
    label?: string;
    modelValue: string | number | null | undefined;
    options: { label: string; value: string | number | null }[];
    placeholder?: string;
    readonly?: boolean;
    removeBackground?: boolean;
    required?: boolean;
    size?: "small" | "large";
  }>(),
  {
    disabled: false,
    error: null,
    placeholder: "",
    readonly: false,
    removeBackground: false,
    required: false,
    size: "small",
  }
);

const placeholderLabel = computed(() => props.placeholder || "Select");

const emit = defineEmits<{
  (e: "update:modelValue", value: string | number | null): void;
  (e: "change"): void;
}>();

const onChange = (value: string | number | null) => {
  emit("update:modelValue", value);
  emit("change");
};
</script>

<template>
  <div :class="['space-y-1.5', props.class]">
    <label v-if="label" class="block text-sm font-semibold text-slate-700">
      {{ label }}
      <span v-if="required" class="text-red-500">*</span>
    </label>
    <Select
      :model-value="modelValue"
      :options="options"
      option-label="label"
      option-value="value"
      :placeholder="placeholderLabel"
      :disabled="disabled"
      :class="[
        'w-full',
        {
          'is-readonly': readonly,
          'is-remove-bg': removeBackground,
        },
      ]"
      :invalid="!!error"
      :size="size"
      @update:model-value="onChange"
    />
    <p v-if="error" class="text-xs text-red-600">{{ error }}</p>
    <slot />
  </div>
</template>

<style scoped>
/* DEFAULT */
:deep(.p-select) {
  background-color: #e2e8f0 !important;
  border: none !important;
  box-shadow: none !important;
  color: #0f172a !important;
}

/* removeBackground */
:deep(.p-select.is-remove-bg) {
  background-color: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

/* hover / focus default */
:deep(.p-select:not(.p-disabled):hover) {
  background-color: #cbd5e1 !important;
}
:deep(.p-select:not(.p-disabled).p-focus) {
  background-color: #cbd5e1 !important;
  outline: 2px solid #3b82f6 !important;
  outline-offset: 0 !important;
}

/* hover / focus removeBackground */
:deep(.p-select.is-remove-bg:not(.p-disabled):hover) {
  background-color: transparent !important;
  border-color: #94a3b8 !important;
}
:deep(.p-select.is-remove-bg:not(.p-disabled).p-focus) {
  background-color: transparent !important;
  border-color: transparent !important;
  outline: none !important;
  outline-offset: 0 !important;
}

/* invalid */
:deep(.p-select.p-invalid) {
  background-color: #fef2f2 !important;
}
:deep(.p-select.is-remove-bg.p-invalid) {
  background-color: transparent !important;
  border-color: #ef4444 !important;
}
:deep(.p-select.p-invalid.p-focus) {
  outline: 2px solid #ef4444 !important;
}

/* readonly */
:deep(.p-select.is-readonly) {
  pointer-events: none;
}
:deep(.p-select.is-readonly .p-select-dropdown) {
  display: none !important;
}

/* disabled */
:deep(.p-select.p-disabled) {
  background-color: #94a3b8 !important;
  color: #64748b !important;
  cursor: not-allowed;
  opacity: 1 !important;
}
:deep(.p-select.is-remove-bg.p-disabled) {
  background-color: transparent !important;
  border-color: #cbd5e1 !important;
}

:deep(.p-select.p-disabled .p-select-label) {
  color: #64748b !important;
}

/* placeholder */
:deep(.p-select-label) {
  background-color: transparent !important;
}
:deep(.p-select-label.p-placeholder) {
  color: #94a3b8 !important;
  opacity: 1;
}
</style>
