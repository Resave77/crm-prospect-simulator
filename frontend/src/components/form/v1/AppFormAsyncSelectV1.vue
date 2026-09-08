<script
  setup
  lang="ts"
  generic="TMode extends 'single' | 'multiple' = 'single'"
>
import MultiSelect from "primevue/multiselect";
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
type SelectMode = "single" | "multiple";
type SelectValue<T extends SelectMode> = T extends "multiple"
  ? number[] | null
  : number | null;
type InitialItem = {
  code?: string | null;
  id: number | string;
  name: string;
};

const props = withDefaults(
  defineProps<{
    badgeLabel?: string;
    disabled?: boolean;
    disabledOptionReasonKey?: string;
    disabledOptionReasonPrefix?: string;
    disabledOptionSelfMatchKey?: string;
    disabledOptionSelfMatchValue?: number | string | null;
    disabledOptionWhenKey?: string;
    error?: string | null;
    helper?: string;
    initialCode?: string | null;
    initialId?: number | null;
    initialItems?: InitialItem[];
    initialLabel?: string | null;
    label?: string;
    loader: LoaderFn;
    maxSelectedLabels?: number;
    mobileModal?: boolean;
    mode?: TMode;
    modelValue: SelectValue<TMode>;
    optionLabel?: string;
    optionValue?: string;
    placeholder?: string;
    readonly?: boolean;
    reloadKey?: number | string | null;
    required?: boolean;
    showClear?: boolean;
    showToggleAll?: boolean;
  }>(),
  {
    badgeLabel: "",
    disabled: false,
    disabledOptionReasonKey: "",
    disabledOptionReasonPrefix: "",
    disabledOptionSelfMatchKey: "",
    disabledOptionSelfMatchValue: null,
    disabledOptionWhenKey: "",
    error: null,
    helper: "",
    initialCode: null,
    initialId: null,
    initialItems: () => [],
    initialLabel: null,
    maxSelectedLabels: 2,
    mobileModal: true,
    optionLabel: "name",
    optionValue: "id",
    placeholder: "Select",
    readonly: false,
    reloadKey: "",
    required: false,
    showClear: true,
    showToggleAll: false,
  }
);

const emit = defineEmits<{
  (event: "update:modelValue", value: SelectValue<TMode>): void;
  (event: "update:item", value: ShortListItem | null): void;
  (event: "items-loaded", items: ShortListItem[]): void;
}>();

const effectiveMode = computed<SelectMode>(() => props.mode ?? "single");

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

const getNestedValue = (source: Record<string, unknown>, path: string) => {
  if (!path) return undefined;

  return path.split(".").reduce<unknown>((acc, key) => {
    if (acc == null || typeof acc !== "object") return undefined;
    return (acc as Record<string, unknown>)[key];
  }, source);
};

const initialRaw = computed<any | null>(() => {
  const initialValue = props.initialId ?? props.modelValue;
  if (Array.isArray(initialValue)) return null;
  if (!initialValue || !props.initialLabel) return null;

  return {
    code: props.initialCode ?? null,
    id: initialValue,
    name: props.initialLabel,
  };
});

const options = computed(() => {
  const mapped = state.items.map((item) => {
    const raw = item as Record<string, unknown>;
    const name = raw[props.optionLabel] ?? raw.name ?? "";
    const code = raw.code;
    const label = code ? `(${code}) ${name}` : name;
    const value = Number(raw[props.optionValue] ?? raw.id);
    const disabledByCondition = props.disabledOptionWhenKey
      ? Boolean(getNestedValue(raw, props.disabledOptionWhenKey))
      : false;
    const selfMatched =
      props.disabledOptionSelfMatchKey &&
      props.disabledOptionSelfMatchValue != null
        ? String(getNestedValue(raw, props.disabledOptionSelfMatchKey)) ===
          String(props.disabledOptionSelfMatchValue)
        : false;
    const disabled = disabledByCondition && !selfMatched;
    const rawDisabledReason = props.disabledOptionReasonKey
      ? getNestedValue(raw, props.disabledOptionReasonKey)
      : undefined;
    const disabledReason =
      rawDisabledReason == null || rawDisabledReason === ""
        ? ""
        : String(rawDisabledReason);

    return {
      disabled,
      disabledReason,
      label,
      value,
    };
  });

  for (const initial of props.initialItems) {
    const value = Number(initial.id);
    if (Number.isNaN(value)) continue;

    const exists = mapped.some((option) => option.value === value);
    if (!exists) {
      const label = initial.code
        ? `(${initial.code}) ${initial.name}`
        : initial.name;
      mapped.unshift({
        disabled: false,
        disabledReason: "",
        label,
        value,
      });
    }
  }

  if (initialRaw.value) {
    const exists = mapped.some(
      (option) => option.value === initialRaw.value!.id
    );
    if (!exists) {
      const { code, id, name } = initialRaw.value;
      const label = code ? `(${code}) ${name}` : name;
      return [
        { disabled: false, disabledReason: "", label, value: id },
        ...mapped,
      ];
    }
  }

  return mapped;
});

const rawMap = computed(() => {
  const map = new Map<number, ShortListItem>();
  if (initialRaw.value) map.set(initialRaw.value.id, initialRaw.value);

  for (const initial of props.initialItems) {
    const id = Number(initial.id);
    if (Number.isNaN(id)) continue;
    map.set(id, {
      code: initial.code ?? null,
      id,
      name: initial.name,
    });
  }

  for (const item of state.items) {
    const id = Number((item as any)[props.optionValue] ?? (item as any).id);
    if (Number.isNaN(id)) continue;
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
    emit("items-loaded", state.items);
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
  itemSize: props.disabledOptionReasonKey ? 54 : 44,
  lazy: true,
  onLazyLoad,
}));

const multiSelectPassThrough = computed(() => ({
  pcHeaderCheckbox: {
    root: {
      class: "hidden",
    },
  },
  pcOptionCheckbox: {
    root: {
      class: "hidden",
    },
  },
}));

const internalValue = computed<number | number[] | null>({
  get: () => props.modelValue,
  set: (value) => {
    (emit as (event: "update:modelValue", value: unknown) => void)(
      "update:modelValue",
      value
    );
  },
});

const loaderResetSignature = computed(() =>
  props.reloadKey !== "" ? props.reloadKey : props.loader
);

watch(loaderResetSignature, () => {
  state.items = [];
  state.page = 1;
  state.total = 0;
  state.totalPages = 1;
  state.search = "";
  state.initialized = false;
});

watch(
  () => props.modelValue,
  (value) => {
    if (effectiveMode.value === "multiple") {
      if (!Array.isArray(value) || !value.length) {
        emit("update:item", null);
        return;
      }

      emit("update:item", null);
      return;
    }

    if (value == null) {
      emit("update:item", null);
      return;
    }

    if (Array.isArray(value)) {
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
  if (props.modelValue == null) return false;
  if (Array.isArray(props.modelValue)) {
    return props.modelValue.includes(val);
  }
  return props.modelValue === val;
};

const toggleOption = (optVal: number) => {
  if (effectiveMode.value === "multiple") {
    const current = Array.isArray(internalValue.value)
      ? [...(internalValue.value as number[])]
      : [];
    const index = current.indexOf(optVal);
    if (index > -1) {
      current.splice(index, 1);
    } else {
      current.push(optVal);
    }
    internalValue.value = current;
  } else {
    internalValue.value = optVal;
    dialogVisible.value = false;
  }
};

const displayLabel = computed(() => {
  if (
    props.modelValue == null ||
    (Array.isArray(props.modelValue) && props.modelValue.length === 0)
  ) {
    return props.placeholder || "Select";
  }

  if (effectiveMode.value === "multiple" && Array.isArray(props.modelValue)) {
    const labels = props.modelValue
      .map((val) => options.value.find((opt) => opt.value === val)?.label)
      .filter(Boolean);
    if (labels.length <= props.maxSelectedLabels) {
      return labels.join(", ");
    }
    return `${labels.length} Selected`;
  }

  const found = options.value.find((opt) => opt.value === props.modelValue);
  return found ? found.label : String(props.modelValue);
});

const dialogBottomBarActions = computed<DialogMobileBottomBarAction[]>(() => {
  if (effectiveMode.value === "multiple") {
    return [
      {
        icon: "pi-filter-slash",
        key: "clear",
        label: "Clear All",
        onClick: () => {
          internalValue.value = [];
          dialogVisible.value = false;
        },
        type: "button",
      },
      {
        icon: "pi-check",
        key: "apply",
        label: "Apply",
        onClick: () => {
          dialogVisible.value = false;
        },
        type: "button",
      },
    ];
  } else {
    return [
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
    ];
  }
});
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
                'opacity-50 pointer-events-none': option.disabled,
              }"
              @click.stop="!option.disabled && toggleOption(option.value)"
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
                  v-if="option.disabledReason"
                  class="text-[12px] truncate mt-0.5 text-slate-400"
                >
                  {{
                    disabledOptionReasonPrefix
                      ? `${disabledOptionReasonPrefix} ${option.disabledReason}`
                      : option.disabledReason
                  }}
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
      <MultiSelect
        v-if="effectiveMode === 'multiple'"
        v-model="internalValue"
        :options="options"
        :virtual-scroller-options="virtualScrollerOptions"
        option-label="label"
        option-value="value"
        option-disabled="disabled"
        :placeholder="placeholder"
        filter
        display="chip"
        :max-selected-labels="maxSelectedLabels"
        :disabled="disabled || readonly"
        :invalid="!!error"
        :pt="multiSelectPassThrough"
        :show-toggle-all="showToggleAll"
        :loading="state.loading"
        class="w-full"
        @show="onShow"
        @filter="onFilter"
      >
        <template #option="{ option, selected }">
          <div
            class="simple-option"
            :class="{
              'simple-option--disabled': option.disabled,
              'simple-option--selected': selected,
            }"
          >
            <div class="simple-option__radio">
              <div class="simple-option__radio-dot" />
            </div>
            <div class="simple-option__content">
              <span class="simple-option__label">{{ option.label }}</span>
              <span
                v-if="option.disabledReason"
                class="simple-option__sub-label"
              >
                {{
                  disabledOptionReasonPrefix
                    ? `${disabledOptionReasonPrefix} ${option.disabledReason}`
                    : option.disabledReason
                }}
              </span>
            </div>
          </div>
        </template>
      </MultiSelect>

      <Select
        v-else
        v-model="internalValue"
        :options="options"
        :virtual-scroller-options="virtualScrollerOptions"
        option-label="label"
        option-value="value"
        option-disabled="disabled"
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
            class="simple-option"
            :class="{
              'simple-option--disabled': option.disabled,
              'simple-option--selected': selected,
            }"
          >
            <div class="simple-option__radio">
              <div class="simple-option__radio-dot" />
            </div>
            <div class="simple-option__content">
              <span class="simple-option__label">{{ option.label }}</span>
              <span
                v-if="option.disabledReason"
                class="simple-option__sub-label"
              >
                {{
                  disabledOptionReasonPrefix
                    ? `${disabledOptionReasonPrefix} ${option.disabledReason}`
                    : option.disabledReason
                }}
              </span>
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

.filter-select:has(.p-select.p-invalid),
.filter-select:has(.p-multiselect.p-invalid) {
  border-color: #fca5a5 !important;
}

.filter-select:has(.p-select.p-invalid):focus-within,
.filter-select:has(.p-multiselect.p-invalid):focus-within {
  box-shadow: 0 0 0 2px #fee2e2 !important;
}

.filter-select:not(:has(.p-select.p-invalid)):not(
    :has(.p-multiselect.p-invalid)
  ):not(:focus-within):hover {
  border-color: #cbd5e1;
}

:deep(.p-select),
:deep(.p-multiselect) {
  min-width: 0 !important;
  border: none !important;
  border-radius: 0 !important;
  background: transparent !important;
  box-shadow: none !important;
  height: 100% !important;
}

:deep(.p-select-label),
:deep(.p-multiselect-label) {
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

:deep(.p-select-label.p-placeholder),
:deep(.p-multiselect-label.p-placeholder) {
  color: #94a3b8 !important;
}

:deep(.p-select-dropdown),
:deep(.p-multiselect-dropdown) {
  color: #64748b !important;
  margin-right: 4px !important;
}

:deep(.p-select.p-disabled),
:deep(.p-multiselect.p-disabled) {
  background: transparent !important;
}

:deep(.p-select-overlay),
:deep(.p-multiselect-overlay) {
  border-radius: 14px !important;
  box-shadow:
    0 8px 24px rgba(15, 23, 42, 0.12),
    0 2px 8px rgba(15, 23, 42, 0.06) !important;
  border: 1px solid #e2e8f0 !important;
  margin-top: 4px !important;
}

:deep(.p-select-filter-container),
:deep(.p-multiselect-filter-container) {
  padding: 10px 12px 6px !important;
  background: #ffffff !important;
  border-bottom: 1px solid #f1f5f9 !important;
  position: relative !important;
}

:deep(.p-select-filter-container .p-select-filter-icon),
:deep(.p-multiselect-filter-container .p-multiselect-filter-icon) {
  position: absolute !important;
  left: 22px !important;
  top: 50% !important;
  transform: translateY(-50%) !important;
  color: #94a3b8 !important;
  font-size: 13px !important;
  margin-top: 2px !important;
  pointer-events: none !important;
}

:deep(.p-select-filter),
:deep(.p-multiselect-filter) {
  width: 100% !important;
  height: 36px !important;
  padding: 0 12px 0 36px !important;
  border: 1px solid #e2e8f0 !important;
  border-radius: 10px !important;
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

:deep(.p-select-filter:focus),
:deep(.p-multiselect-filter:focus) {
  border-color: #3b82f6 !important;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15) !important;
  background: #ffffff !important;
}

:deep(.p-select-filter::placeholder),
:deep(.p-multiselect-filter::placeholder) {
  color: #94a3b8 !important;
}

:deep(.p-select-items),
:deep(.p-multiselect-list) {
  padding: 6px !important;
}

:deep(.p-select-item),
:deep(.p-multiselect-option) {
  margin: 0 !important;
  padding: 0 !important;
  border-radius: 8px !important;
  background: transparent !important;
}

:deep(.p-select-item--selected),
:deep(.p-multiselect-option-selected) {
  background-color: #fff1f2 !important;
}

:deep(.p-select-item--selected):hover,
:deep(.p-multiselect-option-selected):hover {
  background-color: #fff1f2 !important;
}

:deep(.p-select-item):not(.p-select-item--selected):hover,
:deep(.p-multiselect-option):not(.p-multiselect-option-selected):hover {
  background-color: #fee2e2 !important;
}

:deep(.p-select-item--selected):hover .simple-option {
  background-color: #fff1f2 !important;
}

.simple-option {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  background: transparent !important;
}

.simple-option--selected {
  background-color: #fff1f2;
}

.simple-option:not(.simple-option--selected):hover {
  background-color: #fff1f2;
}

.simple-option--disabled {
  cursor: not-allowed;
}

.simple-option__radio {
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
  margin-top: 1px;
}

.simple-option--selected .simple-option__radio {
  border-color: #dc2626;
  background-color: #dc2626;
}

.simple-option__radio-dot {
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

.simple-option--selected .simple-option__radio-dot {
  background-color: #ffffff;
  transform: scale(1);
}

.simple-option__content {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 2px;
}

.simple-option__label {
  font-size: 13px;
  font-weight: 500;
  color: #0f172a;
  line-height: 1.3;
  white-space: nowrap;
}

.simple-option__sub-label {
  font-size: 10px;
  line-height: 1.25;
  color: #94a3b8;
  white-space: nowrap;
}

.simple-option--selected .simple-option__label {
  color: #dc2626;
  font-weight: 600;
}

.simple-option--disabled .simple-option__radio {
  border-color: #fecaca;
  background: #fff1f2;
}

.simple-option--disabled .simple-option__label,
.simple-option--disabled .simple-option__sub-label {
  color: #dc2626;
  text-decoration: line-through;
}

:deep(.p-select-option-check-icon) {
  display: none !important;
}

:deep(.p-select-option-blank-icon) {
  display: none !important;
}

:deep(.p-multiselect-option .p-checkbox),
:deep(.p-multiselect-option .p-checkbox-box),
:deep(.p-multiselect-option .p-multiselect-option-checkbox) {
  display: none !important;
}

:deep(.p-multiselect-option) {
  gap: 0 !important;
}

:deep(.p-select-empty-message) {
  text-align: center !important;
  padding: 20px !important;
  color: #94a3b8 !important;
  font-size: 13px !important;
}
</style>
