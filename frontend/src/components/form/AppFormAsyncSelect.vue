<script setup lang="ts">
import InputText from "primevue/inputtext";
import Select from "primevue/select";
import { computed, reactive, watch } from "vue";

import type {
  ShortListItem,
  ShortListParams,
  ShortListResult,
} from "@/api/general";

type LoaderFn = (params: ShortListParams) => Promise<ShortListResult>;

interface Props {
  disabled?: boolean;
  error?: string | null;
  initialCode?: string | null;
  initialId?: number | null;
  initialLabel?: string | null;
  label?: string;
  loader: LoaderFn;
  modelValue: number | null;
  optionLabel?: string;
  optionValue?: string;
  placeholder?: string;
  readonly?: boolean;
  removeBackground?: boolean;
  required?: boolean;
  showClear?: boolean;
  size?: "small" | "default";
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  error: null,
  initialCode: null,
  initialId: null,
  initialLabel: null,
  optionLabel: "name",
  optionValue: "id",
  placeholder: "Select",
  readonly: false,
  removeBackground: false,
  required: false,
  showClear: true,
  size: "small",
});

const emit = defineEmits<{
  (e: "update:modelValue", value: number | null): void;
  (e: "update:item", value: ShortListItem | null): void;
}>();

interface ShortListState {
  initialized: boolean;
  items: ShortListItem[];
  loading: boolean;
  page: number;
  search: string;
  total: number;
  totalPages: number;
}

const PAGE_SIZE = 25;

const state = reactive<ShortListState>({
  initialized: false,
  items: [],
  loading: false,
  page: 1,
  search: "",
  total: 0,
  totalPages: 1,
});

const options = computed(() => {
  const mapped = state.items.map((item) => {
    const raw = item as any;
    const name = raw[props.optionLabel] ?? raw.name ?? "";
    const code = raw.code;
    const label = code ? `(${code}) ${name}` : name;

    return {
      label,
      value: raw[props.optionValue] ?? raw.id,
    };
  });

  if (initialRaw.value) {
    const exists = mapped.some((o) => o.value === initialRaw.value!.id);
    if (!exists) {
      const { code, id, name } = initialRaw.value;
      const label = code ? `(${code}) ${name}` : name;
      return [{ label, value: id }, ...mapped];
    }
  }

  return mapped;
});

const initialRaw = computed<any | null>(() => {
  const initialValue = props.initialId ?? props.modelValue;
  if (!initialValue || !props.initialLabel) {
    return null;
  }

  return {
    code: props.initialCode ?? null,
    id: initialValue,
    name: props.initialLabel,
  } as any;
});

const rawMap = computed(() => {
  const map = new Map<number, ShortListItem>();

  // 1. inject initial
  if (initialRaw.value) {
    map.set(initialRaw.value.id, initialRaw.value);
  }

  // 2. override dengan hasil loader
  for (const item of state.items) {
    const id = (item as any)[props.optionValue] ?? (item as any).id;
    map.set(id, item);
  }

  return map;
});

const hasMore = computed(() => state.page < state.totalPages);

let filterTimer: number | undefined;

const load = async (opts?: { page?: number; reset?: boolean }) => {
  const targetPage = opts?.page ?? (opts?.reset ? 1 : state.page + 1);
  state.loading = true;
  try {
    const { items, pagination } = await props.loader({
      page: targetPage,
      page_size: PAGE_SIZE,
      search: state.search || undefined,
      sort_direction: "asc",
      sort_field: "name",
    });
    state.page = pagination?.page ?? targetPage;
    state.total = pagination?.total ?? items.length;
    state.totalPages =
      pagination?.total_pages ??
      Math.max(1, Math.ceil(state.total / PAGE_SIZE));
    state.items = opts?.reset ? items : [...state.items, ...items];
    state.initialized = true;
  } catch (err) {
    console.error("Failed to load options", err);
    if (opts?.reset) state.items = [];
  } finally {
    state.loading = false;
  }
};

const onShow = () => {
  if (!state.initialized) {
    load({ reset: true });
  }
};

const onFilter = (event: { value: string }) => {
  const newSearch = event.value?.trim() ?? "";
  if (newSearch === state.search) return;
  state.search = newSearch;
  if (filterTimer) window.clearTimeout(filterTimer);
  filterTimer = window.setTimeout(() => load({ reset: true }), 350);
};

const onLazyLoad = (event: { last: number }) => {
  if (state.loading || !hasMore.value) return;
  const threshold = 5;
  if (event.last >= state.items.length - threshold) {
    load({ page: state.page + 1 });
  }
};

const virtualScrollerOptions = computed(() => ({
  autoSize: true,
  itemSize: 36,
  lazy: true,
  onLazyLoad,
}));

const internalValue = computed({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", val),
});

const readonlyLabel = computed(() => {
  const currentValue = props.modelValue;
  if (currentValue == null) return "";
  const matched = options.value.find((opt) => opt.value === currentValue);
  return matched?.label ?? props.initialLabel ?? "";
});

// Reset state when loader changes
watch(
  () => props.loader,
  () => {
    state.items = [];
    state.page = 1;
    state.total = 0;
    state.totalPages = 1;
    state.search = "";
    state.initialized = false;
  }
);

watch(
  () => props.modelValue,
  (val) => {
    if (val == null) {
      emit("update:item", null);
      return;
    }

    emit("update:item", rawMap.value.get(val) ?? null);
  },
  { immediate: true }
);
</script>

<template>
  <div class="space-y-1.5">
    <label v-if="label" class="block text-sm font-semibold text-slate-700">
      {{ label }}
      <span v-if="required" class="text-red-500">*</span>
    </label>
    <InputText
      v-if="readonly"
      :model-value="readonlyLabel"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="true"
      :invalid="!!error"
      :class="[
        'w-full',
        {
          'is-readonly': readonly,
          'is-remove-bg': removeBackground,
        },
      ]"
    />
    <Select
      v-else
      v-model="internalValue"
      :options="options"
      :virtual-scroller-options="virtualScrollerOptions"
      option-label="label"
      option-value="value"
      :placeholder="placeholder"
      filter
      :size="size"
      :disabled="disabled"
      :class="['w-full', { 'is-readonly': readonly }]"
      :invalid="!!error"
      :show-clear="showClear"
      :loading="state.loading"
      @show="onShow"
      @filter="onFilter"
    />
    <p v-if="error" class="text-xs text-red-600">{{ error }}</p>
    <slot />
  </div>
</template>

<style scoped>
:deep(.p-select) {
  background-color: #e2e8f0 !important;
  border: none !important;
  box-shadow: none !important;
  color: #0f172a !important;
}

:deep(.is-readonly .p-select) {
  pointer-events: none;
}

:deep(.p-select-label) {
  background-color: transparent !important;
}

:deep(.p-select:not(.p-disabled):hover) {
  background-color: #cbd5e1 !important;
}

:deep(.p-select:not(.p-disabled).p-focus) {
  background-color: #cbd5e1 !important;
  outline: 2px solid #3b82f6 !important;
  outline-offset: 0 !important;
}

:deep(.p-select.p-invalid) {
  background-color: #fef2f2 !important;
}

:deep(.p-select.p-invalid.p-focus) {
  outline: 2px solid #ef4444 !important;
}

:deep(.p-select.p-disabled) {
  background-color: #94a3b8 !important;
  color: #64748b !important;
  cursor: not-allowed;
  opacity: 1 !important;
}

:deep(.p-select.p-disabled .p-select-label) {
  color: #64748b !important;
}

:deep(.p-select-label.p-placeholder) {
  color: #94a3b8 !important;
  opacity: 1;
}
</style>

<style scoped>
:deep(.p-inputtext) {
  background-color: #e2e8f0 !important;
  border: none !important;
  box-shadow: none !important;
  color: #0f172a !important;
}

:deep(.is-readonly .p-inputtext) {
  pointer-events: none;
}

:deep(.p-inputtext:hover) {
  background-color: #cbd5e1 !important;
}

:deep(.p-inputtext:focus) {
  background-color: #cbd5e1 !important;
  outline: 2px solid #3b82f6 !important;
  outline-offset: 0 !important;
}

:deep(.p-inputtext.p-invalid) {
  background-color: #fef2f2 !important;
}

:deep(.p-inputtext.p-invalid:focus) {
  outline: 2px solid #ef4444 !important;
}

:deep(.p-inputtext:disabled) {
  background-color: #94a3b8 !important;
  color: #64748b !important;
  cursor: not-allowed;
  opacity: 1 !important;
}
</style>

<style scoped>
:deep(.p-inputtext::placeholder) {
  color: #94a3b8 !important;
  opacity: 1;
}
</style>
<style scoped>
/* hover/focus ketika removeBackground */
:deep(.p-inputtext.is-remove-bg:hover) {
  background-color: transparent !important;
  border-color: #94a3b8 !important;
}
:deep(.p-inputtext.is-remove-bg:focus) {
  background-color: transparent !important;
  border-color: transparent !important;
  outline: none !important;
  outline-offset: 0 !important;
}

/* removeBackground = true: transparan + ada border */
:deep(.p-inputtext.is-remove-bg) {
  background-color: transparent !important;
  border: none !important;
  box-shadow: none !important;
}
</style>
