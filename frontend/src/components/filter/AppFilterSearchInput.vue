<script setup lang="ts">
import FloatLabel from "primevue/floatlabel";
import InputText from "primevue/inputtext";
import { computed, ref, watch } from "vue";

type LabelPosition = "none" | "top" | "float";

interface Props {
  disabled?: boolean;
  inputClass?: string;
  label?: string;
  labelPosition?: LabelPosition;
  modelValue: string;
  placeholder?: string;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  inputClass: "",
  label: "",
  labelPosition: "none",
  placeholder: "",
});

const inputId = computed(
  () => `search-${Math.random().toString(36).slice(2, 9)}`
);

const placeholderValue = computed(() => props.placeholder || "Search");

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();

// Local value for immediate UI feedback
const localValue = ref(props.modelValue);

const onInput = (value: string | undefined) => {
  localValue.value = value || "";
  emit("update:modelValue", localValue.value);
};

// Sync localValue when modelValue changes externally
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
  <!-- Label Position: Top -->
  <div v-if="labelPosition === 'top'" class="app-filter-search-wrapper">
    <label
      v-if="label"
      :for="inputId"
      class="mb-0.5 block text-xs font-medium text-slate-500"
    >
      {{ label }}
    </label>
    <div
      class="app-search-input flex items-center gap-2 !bg-slate-100 !border-none !shadow-none text-slate-600 transition hover:!bg-slate-200 focus-within:!bg-slate-200 focus-within:!outline focus-within:!outline-2 focus-within:!outline-blue-500"
      :class="{ 'opacity-50 pointer-events-none': disabled }"
    >
      <span class="pi pi-search text-slate-400" />
      <InputText
        :id="inputId"
        :model-value="localValue"
        :placeholder="placeholderValue"
        :disabled="disabled"
        class="w-full border-none bg-transparent p-0 focus:ring-0"
        :class="inputClass"
        @update:model-value="onInput"
      />
    </div>
  </div>

  <!-- Label Position: Float -->
  <FloatLabel v-else-if="labelPosition === 'float'" variant="on">
    <div
      class="app-search-input flex items-center gap-2 overflow-hidden !bg-slate-100 !border-none !shadow-none text-slate-600 transition hover:!bg-slate-200 focus-within:!bg-slate-200 focus-within:!outline focus-within:!outline-2 focus-within:!outline-blue-500"
      :class="{ 'opacity-50 pointer-events-none': disabled }"
    >
      <span class="pi pi-search text-slate-400" />
      <InputText
        :id="inputId"
        :model-value="localValue"
        :placeholder="placeholderValue"
        :disabled="disabled"
        class="w-full border-none bg-transparent p-0 focus:ring-0"
        :class="inputClass"
        @update:model-value="onInput"
      />
    </div>
    <label :for="inputId" class="text-xs font-medium text-slate-500">
      {{ label }}
    </label>
  </FloatLabel>

  <!-- Label Position: None (default) -->
  <div
    v-else
    class="app-search-input flex items-center gap-2 overflow-hidden !bg-slate-100 !border-none !shadow-none text-slate-600 transition hover:!bg-slate-200 focus-within:!bg-slate-200 focus-within:!outline focus-within:!outline-2 focus-within:!outline-blue-500"
    :class="{ 'opacity-50 pointer-events-none': disabled }"
  >
    <span class="pi pi-search text-slate-400" />
    <InputText
      :model-value="localValue"
      :placeholder="placeholderValue"
      :disabled="disabled"
      class="w-full border-none bg-transparent p-0 focus:ring-0"
      :class="inputClass"
      @update:model-value="onInput"
    />
  </div>
</template>

<style scoped>
.app-search-input {
  min-height: 2rem;
  height: 2rem;
  padding: 0 0.625rem;
  border-radius: 6px;
  font-size: var(--font-select);
}

.app-search-input .pi-search {
  font-size: 0.875rem;
}

.app-search-input :deep(.p-inputtext) {
  font-size: var(--font-select);
  padding: 0;
  border: none;
  outline: none;
  box-shadow: none;
  background: transparent;
  height: auto;
  line-height: 1.5;
}

.app-search-input :deep(.p-inputtext:focus),
.app-search-input :deep(.p-inputtext:enabled:focus) {
  border: none;
  outline: none;
  box-shadow: none;
}

.app-search-input :deep(.p-inputtext::placeholder) {
  color: #64748b !important;
}
</style>
