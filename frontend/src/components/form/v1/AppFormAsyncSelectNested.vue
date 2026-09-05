<script setup lang="ts">
import Select from "primevue/select";
import { computed, reactive, ref, watch } from "vue";

import type {
  ShortListItem,
  ShortListParams,
  ShortListResult,
} from "@/api/general";
import type { DialogMobileBottomBarAction } from "@/composables/useDialogMobileBottomBar";

import AppDialog from "@/components/base/AppDialog.vue";
import isMobile from "@/composables/useIsMobile";

type LoaderFn = (params: ShortListParams) => Promise<ShortListResult>;

const props = withDefaults(
  defineProps<{
    badgeLabel?: string;
    disabled?: boolean;
    error?: string | null;
    helper?: string;
    initialCode?: string | null;
    initialId?: number | null;
    initialLabel?: string | null;
    label?: string;
    loader: LoaderFn;
    mobileModal?: boolean;
    modelValue: number | null;
    optionLabel?: string;
    optionSubLabel?: string;
    optionValue?: string;
    placeholder?: string;
    readonly?: boolean;
    required?: boolean;
    showClear?: boolean;
  }>(),
  {
    badgeLabel: "",
    disabled: false,
    error: null,
    helper: "",
    initialCode: null,
    initialId: null,
    initialLabel: null,
    mobileModal: true,
    optionLabel: "name",
    optionSubLabel: "department_name",
    optionValue: "id",
    placeholder: "Select",
    readonly: false,
    required: false,
    showClear: true,
  }
);

const emit = defineEmits<{
  (event: "update:modelValue", value: number | null): void;
  (event: "update:item", value: ShortListItem | null): void;
}>();

const PAGE_SIZE = 25;
let filterTimer: number | undefined;

const state = reactive({
  initialized: false,
  items: [] as ShortListItem[],
  loading: false,
  page: 1,
  search: "",
  total: 0,
  totalPages: 1,
});

const initialRaw = computed<any | null>(() => {
  const initialValue = props.initialId ?? props.modelValue;
  if (!initialValue || !props.initialLabel) return null;

  return {
    code: props.initialCode ?? null,
    id: initialValue,
    name: props.initialLabel,
  };
});

const options = computed(() => {
  const mapped = state.items.map((item) => {
    const raw = item as any;
    const name = raw[props.optionLabel] ?? raw.name ?? "";
    const subLabel = raw[props.optionSubLabel] ?? "";

    return {
      label: name,
      raw,
      subLabel,
      value: raw[props.optionValue] ?? raw.id,
    };
  });

  if (initialRaw.value) {
    const exists = mapped.some(
      (option) => option.value === initialRaw.value!.id
    );
    if (!exists) {
      return [
        {
          label: initialRaw.value.name,
          raw: initialRaw.value,
          subLabel: "",
          value: initialRaw.value.id,
        },
        ...mapped,
      ];
    }
  }

  return mapped;
});

const rawMap = computed(() => {
  const map = new Map<number, ShortListItem>();
  if (initialRaw.value) map.set(initialRaw.value.id, initialRaw.value);

  for (const item of state.items) {
    const id = (item as any)[props.optionValue] ?? (item as any).id;
    map.set(id, item);
  }

  return map;
});

const hasMore = computed(() => state.page < state.totalPages);

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
  if (!state.initialized) load({ reset: true });
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
  if (event.last >= state.items.length - 5) {
    load({ page: state.page + 1 });
  }
};

const onScrollList = (event: Event) => {
  const target = event.target as HTMLElement;
  if (target.scrollTop + target.clientHeight >= target.scrollHeight - 50) {
    if (!state.loading && hasMore.value) {
      load({ page: state.page + 1 });
    }
  }
};

const virtualScrollerOptions = computed(() => ({
  autoSize: true,
  itemSize: 50,
  lazy: true,
  onLazyLoad,
}));

const internalValue = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});

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
  (value) => {
    if (value == null) {
      emit("update:item", null);
      return;
    }

    emit("update:item", rawMap.value.get(value) ?? null);
  },
  { immediate: true }
);
const dialogVisible = ref(false);
const localSearch = ref("");

const openMobileDialog = () => {
  if (!state.initialized) {
    onShow();
  }
  localSearch.value = state.search;
  dialogVisible.value = true;
};

watch(localSearch, (newSearch) => {
  if (filterTimer) {
    window.clearTimeout(filterTimer);
  }
  filterTimer = window.setTimeout(() => {
    state.search = newSearch.trim();
    state.page = 1;
    load({ page: 1 });
  }, 350);
});

const isSelected = (val: number) => {
  return props.modelValue === val;
};

const toggleOption = (optVal: number) => {
  internalValue.value = optVal;
  dialogVisible.value = false;
};

const displayLabel = computed(() => {
  if (props.modelValue == null) {
    return props.placeholder || "Select";
  }
  const found = options.value.find((opt) => opt.value === props.modelValue);
  return found ? found.label : String(props.modelValue);
});

const dialogBottomBarActions = computed<DialogMobileBottomBarAction[]>(() => [
  {
    icon: "pi-filter-slash",
    key: "clear",
    label: "Clear Selection",
    onClick: () => {
      internalValue.value = null;
      dialogVisible.value = false;
    },
    type: "button",
  },
  {
    icon: "pi-times",
    key: "close",
    label: "Close",
    onClick: () => {
      dialogVisible.value = false;
    },
    type: "button",
  },
]);
</script>

<template>
  <div class="flex flex-col gap-[8px]">
    <label
      v-if="label"
      class="flex flex-wrap items-center gap-[6px] text-[12px] font-semibold text-[#334155]"
    >
      <span>{{ label }}</span>
      <span v-if="required" class="text-[#dc2626]">*</span>
      <span
        v-if="badgeLabel"
        class="rounded-full border border-[#e2e8f0] bg-[#f8fafc] px-[8px] py-[2px] text-[11px] font-semibold text-[#64748b]"
      >
        {{ badgeLabel }}
      </span>
    </label>

    <template v-if="isMobile && mobileModal">
      <div
        class="filter-select flex w-full items-center border border-[#d9e2ec] bg-white transition cursor-pointer px-3 justify-between"
        :class="[
          error ? 'border-red-300' : 'border-[#d9e2ec]',
          disabled || readonly
            ? 'bg-[#f1f5f9] opacity-60 pointer-events-none'
            : '',
        ]"
        @click.stop="openMobileDialog"
      >
        <span
          class="truncate text-[12px]"
          :class="modelValue != null ? 'text-[#0f172a]' : 'text-[#94a3b8]'"
        >
          {{ displayLabel }}
        </span>
        <span
          class="pi pi-chevron-down text-slate-400 text-[12px] shrink-0 ml-2"
        />
      </div>

      <AppDialog
        :model-value="dialogVisible"
        :title="label || placeholder || 'Select'"
        :mobile-actions="dialogBottomBarActions"
        @update:model-value="dialogVisible = $event"
      >
        <div class="flex flex-col h-full min-h-0 gap-3">
          <div class="relative w-full shrink-0">
            <span
              class="pi pi-search absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400 text-sm"
            />
            <input
              v-model="localSearch"
              type="text"
              placeholder="Search..."
              class="w-full min-w-0 h-11 overflow-hidden text-ellipsis whitespace-nowrap pl-10 pr-4 text-[14px] border border-slate-200 rounded-xl bg-slate-50 focus:bg-white focus:border-blue-500 focus:ring-2 focus:ring-blue-100 outline-none transition"
            />
          </div>

          <div
            class="flex-1 overflow-y-auto min-h-0 py-2 -mx-4 px-4 flex flex-col gap-1"
            @scroll="onScrollList"
          >
            <div
              v-for="option in options"
              :key="option.value"
              class="flex items-center gap-3 py-3 px-3 rounded-xl cursor-pointer hover:bg-slate-50 transition active:bg-slate-100"
              :class="{
                'bg-rose-50/50 hover:bg-rose-50/70': isSelected(option.value),
              }"
              @click.stop="toggleOption(option.value)"
            >
              <div
                class="w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0 transition"
                :class="
                  isSelected(option.value)
                    ? 'border-red-600 bg-red-600'
                    : 'border-slate-300 bg-white'
                "
              >
                <div
                  class="w-2 h-2 rounded-full bg-white transition scale-0"
                  :class="{ 'scale-100': isSelected(option.value) }"
                />
              </div>

              <div class="flex-1 min-w-0 flex flex-col">
                <span
                  class="text-[14px] font-medium truncate"
                  :class="
                    isSelected(option.value)
                      ? 'text-red-600 font-semibold'
                      : 'text-slate-700'
                  "
                >
                  {{ option.label }}
                </span>
                <span
                  v-if="option.subLabel"
                  class="text-[12px] truncate mt-0.5 text-slate-400"
                >
                  {{ option.subLabel }}
                </span>
              </div>
            </div>

            <div
              v-if="state.loading"
              class="py-4 text-center text-slate-400 text-xs flex items-center justify-center gap-2"
            >
              <span class="pi pi-spinner pi-spin text-sm" />
              Loading items...
            </div>

            <div
              v-else-if="options.length === 0"
              class="py-12 text-center text-slate-400 text-sm"
            >
              No options found
            </div>
          </div>
        </div>
      </AppDialog>
    </template>

    <div
      v-else
      class="filter-select flex items-center"
      :class="[
        error ? 'border-red-300' : 'border-[#d9e2ec]',
        disabled || readonly
          ? 'bg-[#f1f5f9] opacity-60 pointer-events-none'
          : '',
      ]"
    >
      <Select
        v-model="internalValue"
        :options="options"
        :virtual-scroller-options="virtualScrollerOptions"
        option-label="label"
        option-value="value"
        :placeholder="placeholder"
        filter
        :disabled="disabled || readonly"
        :invalid="!!error"
        :show-clear="showClear"
        :loading="state.loading"
        class="w-full"
        @show="onShow"
        @filter="onFilter"
      >
        <template #option="{ option, selected }">
          <div
            class="nested-option"
            :class="{ 'nested-option--selected': selected }"
          >
            <div class="nested-option__radio">
              <div class="nested-option__radio-dot" />
            </div>
            <div class="nested-option__text">
              <span class="nested-option__label">{{ option.label }}</span>
              <span v-if="option.subLabel" class="nested-option__sublabel">{{
                option.subLabel
              }}</span>
            </div>
          </div>
        </template>
      </Select>
    </div>

    <span v-if="error" class="text-[11px] text-red-600">
      {{ error }}
    </span>
    <span v-else-if="helper" class="text-[11px] text-[#64748b]">
      {{ helper }}
    </span>
    <slot v-else />
  </div>
</template>

<style scoped>
.filter-select {
  height: 48px;
  border-radius: 14px;
  border-width: 1px;
  border-style: solid;
  background: #ffffff;
  box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.04);
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease;
  overflow: hidden;
}

.filter-select:focus-within {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.12);
}

.filter-select:has(.p-select.p-invalid) {
  border-color: #fca5a5 !important;
}

.filter-select:has(.p-select.p-invalid):focus-within {
  box-shadow: 0 0 0 2px #fee2e2 !important;
}

.filter-select:not(:has(.p-select.p-invalid)):not(:focus-within):hover {
  border-color: #cbd5e1;
}

/* ---- PrimeVue component overrides ---- */
:deep(.p-select) {
  min-width: 0 !important;
  border: none !important;
  border-radius: 0 !important;
  background: transparent !important;
  box-shadow: none !important;
  height: 100% !important;
}

:deep(.p-select-label) {
  min-width: 0 !important;
  overflow: hidden !important;
  color: #0f172a !important;
  font-size: 12px !important;
  font-weight: 500 !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
  padding: 0 12px !important;
  display: flex !important;
  align-items: center !important;
  height: 100% !important;
}

:deep(.p-select-label.p-placeholder) {
  color: #94a3b8 !important;
}

:deep(.p-select-dropdown) {
  color: #64748b !important;
  margin-right: 4px !important;
}

:deep(.p-select.p-disabled) {
  background: transparent !important;
}

:deep(.p-select-overlay) {
  border-radius: 14px !important;
  box-shadow:
    0 10px 30px rgba(15, 23, 42, 0.1),
    0 2px 8px rgba(15, 23, 42, 0.06) !important;
  border: 1px solid #e2e8f0 !important;
  margin-top: 4px !important;
  padding: 6px !important;
}

/* Filter input area */
:deep(.p-select-filter-container) {
  padding: 8px 10px !important;
  background: #ffffff !important;
  border-bottom: 1px solid #f1f5f9 !important;
  position: relative !important;
  display: flex !important;
  align-items: center !important;
  gap: 8px !important;
}

:deep(.p-select-filter-container)::before {
  content: "";
  display: block;
  width: 15px;
  height: 15px;
  min-width: 15px;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='15' height='15' viewBox='0 0 24 24' fill='none' stroke='%2394a3b8' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Ccircle cx='11' cy='11' r='8'/%3E%3Cline x1='21' y1='21' x2='16.65' y2='16.65'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: center;
  background-size: contain;
  pointer-events: none;
  opacity: 0.7;
}

:deep(.p-select-filter) {
  flex: 1;
  height: 34px !important;
  padding: 0 10px !important;
  border: 1px solid #e2e8f0 !important;
  border-radius: 8px !important;
  background: #f8fafc !important;
  color: #0f172a !important;
  font-size: 13px !important;
  font-weight: 400 !important;
  outline: none !important;
  box-shadow: none !important;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease !important;
}

:deep(.p-select-filter:focus) {
  border-color: #3b82f6 !important;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12) !important;
  background: #ffffff !important;
}

:deep(.p-select-filter::placeholder) {
  color: #94a3b8 !important;
}

:deep(.p-select-filter-icon) {
  display: none !important;
}

/* Items list */
:deep(.p-select-items) {
  padding: 4px !important;
}

/* ---- Nested option item ---- */
:deep(.p-select-item) {
  margin: 0 !important;
  padding: 0 !important;
  border-radius: 8px !important;
  background: transparent !important;
}

:deep(.p-select-item--selected) {
  background-color: #fff1f2 !important;
}

:deep(.p-select-item--selected):hover {
  background-color: #fff1f2 !important;
}

:deep(.p-select-item):not(.p-select-item--selected):hover {
  background-color: #fee2e2 !important;
}

:deep(.p-select-item--selected):hover .nested-option {
  background-color: #fff1f2 !important;
}

.nested-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 6px;
  cursor: pointer;
  background: transparent !important;
}

.nested-option--selected {
  background-color: #fff1f2;
}

.nested-option:not(.nested-option--selected):hover {
  background-color: #fff1f2;
}

.nested-option__radio {
  width: 18px;
  height: 18px;
  min-width: 18px;
  border-radius: 50%;
  border: 2px solid #cbd5e1;
  display: flex;
  align-items: center;
  justify-content: center;
  transition:
    border-color 150ms ease,
    background-color 150ms ease,
    transform 150ms ease;
}

.nested-option--selected .nested-option__radio {
  border-color: #dc2626;
  background-color: #dc2626;
}

.nested-option__radio-dot {
  width: 8px;
  height: 8px;
  min-width: 8px;
  border-radius: 50%;
  background-color: transparent;
  transform: scale(0.5);
  transition:
    background-color 150ms ease,
    transform 150ms ease;
}

.nested-option--selected .nested-option__radio-dot {
  background-color: #ffffff;
  transform: scale(1);
}

.nested-option__text {
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: 1px;
}

.nested-option__label {
  font-size: 13px;
  font-weight: 500;
  color: #0f172a;
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nested-option--selected .nested-option__label {
  color: #dc2626;
  font-weight: 600;
}

.nested-option__sublabel {
  font-size: 11px;
  font-weight: 400;
  color: #64748b;
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nested-option--selected .nested-option__sublabel {
  color: #dc2626;
}

:deep(.p-select-option-check-icon) {
  display: none !important;
}

:deep(.p-select-option-blank-icon) {
  display: none !important;
}

:deep(.p-select-empty-message) {
  text-align: center !important;
  padding: 20px !important;
  color: #94a3b8 !important;
  font-size: 13px !important;
}
</style>
