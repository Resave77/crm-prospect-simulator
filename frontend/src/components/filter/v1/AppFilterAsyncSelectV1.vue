<script setup lang="ts">
import { useInfiniteQuery } from "@tanstack/vue-query";
import MultiSelect from "primevue/multiselect";
import Select from "primevue/select";
import {
  computed,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
  watch,
} from "vue";

import type {
  ShortListItem,
  ShortListParams,
  ShortListResult,
} from "@/api/general";
import type { DialogMobileBottomBarAction } from "@/composables/useDialogMobileBottomBar";

import AppDialog from "@/components/base/AppDialog.vue";
import isMobile from "@/composables/useIsMobile";

import {
  filterSizeClasses,
  useFilterSize,
  type FilterSize,
} from "./composables/useFilterSize";

type SelectValue = number | string;
type LoaderFn = (
  params: ShortListParams
) => Promise<ShortListResult<SelectValue>>;
type SelectMode = "single" | "multiple";

interface Props {
  cacheKey: string;
  disabled?: boolean;
  disabledItemValues?: (string | number)[];
  idKey?: string;
  inputClass?: string;
  labelKey?: string;
  loader: LoaderFn;
  maxSelectedLabels?: number;
  mobileModal?: boolean;
  mode?: SelectMode;
  modelValue: SelectValue | SelectValue[] | null;
  placeholder?: string;
  selectedLabelResolver?: (value: SelectValue) => string | null | undefined;
  size?: FilterSize | null;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  disabledItemValues: () => [],
  idKey: "id",
  inputClass: "",
  labelKey: "name",
  mobileModal: true,
  mode: "single",
  placeholder: "",
  selectedLabelResolver: undefined,
  size: null,
});

const emit = defineEmits<{
  (e: "update:modelValue", value: SelectValue | SelectValue[] | null): void;
  (e: "items-loaded", items: ShortListItem<SelectValue>[]): void;
}>();

const { actualSize } = useFilterSize({ size: props.size });

const filterFieldRef = ref<HTMLElement | null>(null);
const filterFieldWidth = ref(0);
let filterFieldResizeObserver: ResizeObserver | null = null;

const PAGE_SIZE = 25;

const hasInitialSelection = computed(() =>
  Array.isArray(props.modelValue)
    ? props.modelValue.length > 0
    : props.modelValue != null
);

const state = reactive({
  initialized: hasInitialSelection.value,
  search: "",
});

const RESET_OPTION_VALUE = "__app_filter_select_reset__";

const placeholderValue = computed(() => props.placeholder || "Select");

const resetOption = computed(() => ({
  disabled: false,
  label: props.mode === "multiple" ? "Unselect All" : "Clear Selection",
  subLabel:
    props.mode === "multiple"
      ? "Remove all selected options"
      : "Remove the selected option",
  value: RESET_OPTION_VALUE,
}));

const isResetOption = (value: SelectValue | null) =>
  value === RESET_OPTION_VALUE;

const listPassThrough = computed(() => ({
  filterIcon: { class: "!h-3.5 !w-3.5 !text-slate-400" },
  header: {
    class:
      "!m-0 !flex !items-center !gap-0 !border-b !border-slate-200 !bg-white !px-2 !py-2",
  },
  list: { class: "!px-1 !pb-1 !pt-px" },
  option: { class: "!m-0 !p-0 !rounded-xl !bg-transparent" },
  overlay: {
    class:
      "!mt-2 !min-w-[18rem] !max-w-[calc(100vw-2rem)] !overflow-hidden !rounded-2xl !border !border-slate-200 !p-0 !shadow-[0_18px_40px_rgba(15,23,42,0.12),0_4px_12px_rgba(15,23,42,0.06)]",
  },
  panel: {
    class:
      "!mt-2 !min-w-[18rem] !max-w-[calc(100vw-2rem)] !overflow-hidden !rounded-2xl !border !border-slate-200 !p-0 !shadow-[0_18px_40px_rgba(15,23,42,0.12),0_4px_12px_rgba(15,23,42,0.06)]",
  },
  pcFilter: {
    root: {
      class:
        "!h-8 !w-full !rounded-lg !border !border-slate-200 !bg-slate-50 !py-0 !pl-3 !pr-9 !text-xs !font-normal !leading-8 !text-slate-900 !shadow-none placeholder:!text-slate-400 focus:!border-blue-300 focus:!bg-white focus:!shadow-[0_0_0_3px_rgba(59,130,246,0.10)]",
    },
  },
  pcFilterContainer: {
    root: { class: "!w-full !min-w-0" },
  },
  pcFilterIconContainer: {
    root: { class: "!right-2.5 !text-slate-400" },
  },
}));

const multiSelectPassThrough = computed(() => ({
  ...listPassThrough.value,
  headerCheckboxContainer: { class: "!hidden" },
  pcHeaderCheckbox: { root: { class: "!hidden" } },
  pcOptionCheckbox: { root: { class: "hidden" } },
  toggleAllIcon: { class: "!hidden" },
}));

const queryKey = computed(
  () => ["filter-shortlist", props.cacheKey, state.search] as const
);

const shortlistQuery = useInfiniteQuery<ShortListResult<SelectValue>, Error>({
  enabled: computed(() => state.initialized),
  getNextPageParam: (lastPage) => {
    const currentPage = lastPage.pagination?.page ?? 1;
    const totalPages = lastPage.pagination?.total_pages ?? 1;
    return currentPage < totalPages ? currentPage + 1 : undefined;
  },
  initialPageParam: 1,
  queryFn: async ({ pageParam }) => {
    return props.loader({
      page: pageParam as number,
      page_size: PAGE_SIZE,
      search: state.search || undefined,
    });
  },
  queryKey,
  staleTime: 1000 * 60 * 5,
});

const items = computed(
  () => shortlistQuery.data.value?.pages.flatMap((page) => page.items) ?? []
);

const loadedOptions = computed(() =>
  items.value.map((item: ShortListItem<SelectValue>) => {
    const raw = item as Record<string, unknown>;
    const label = raw[props.labelKey] ?? raw["name"] ?? "";
    const value = (raw[props.idKey] ?? raw["id"]) as SelectValue;

    const isExcluded = props.disabledItemValues?.includes(value) ?? false;
    const isCurrent = isSelected(value);

    return {
      disabled: isExcluded && !isCurrent,
      label:
        isExcluded && !isCurrent
          ? `${String(label)} (Sudah Terpilih)`
          : String(label),
      subLabel: "",
      value,
    };
  })
);

const options = computed(() => [resetOption.value, ...loadedOptions.value]);

const isLoading = computed(
  () =>
    shortlistQuery.isLoading.value ||
    shortlistQuery.isFetching.value ||
    shortlistQuery.isFetchingNextPage.value
);

const resolveModelValue = (val: SelectValue | SelectValue[] | null) => {
  if (Array.isArray(val)) return val.includes(RESET_OPTION_VALUE) ? [] : val;
  return val === RESET_OPTION_VALUE ? null : val;
};

const internalValue = computed<SelectValue | SelectValue[] | null>({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", resolveModelValue(val)),
});

const onShow = () => {
  if (!state.initialized) {
    state.initialized = true;
  }
};

let filterTimer: number | undefined;

const onFilter = (event: { value?: string }) => {
  const newSearch = event.value?.trim() ?? "";
  if (newSearch === state.search) return;

  if (filterTimer) {
    window.clearTimeout(filterTimer);
  }

  filterTimer = window.setTimeout(() => {
    state.search = newSearch;
  }, 350);
};

const onLazyLoad = async (event: { last: number }) => {
  if (
    shortlistQuery.isFetchingNextPage.value ||
    !shortlistQuery.hasNextPage.value
  ) {
    return;
  }

  const threshold = 5;
  if (event.last >= options.value.length - threshold) {
    await shortlistQuery.fetchNextPage();
  }
};

const virtualScrollerOptions = computed(() => ({
  autoSize: true,
  itemSize: 36,
  lazy: true,
  onLazyLoad,
}));

watch(
  items,
  (val) => {
    emit("items-loaded", val);
  },
  { immediate: true }
);

watch(
  () => [props.loader, props.cacheKey],
  () => {
    state.initialized = false;
    state.search = "";
  }
);

onMounted(() => {
  const updateFieldWidth = () => {
    filterFieldWidth.value = filterFieldRef.value?.clientWidth ?? 0;
  };

  updateFieldWidth();

  if (filterFieldRef.value && typeof ResizeObserver !== "undefined") {
    filterFieldResizeObserver = new ResizeObserver(updateFieldWidth);
    filterFieldResizeObserver.observe(filterFieldRef.value);
  }
});

onBeforeUnmount(() => {
  if (filterTimer) {
    window.clearTimeout(filterTimer);
  }

  filterFieldResizeObserver?.disconnect();
});

const dialogVisible = ref(false);
const localSearch = ref("");

const openMobileDialog = () => {
  if (!state.initialized) {
    state.initialized = true;
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
  }, 350);
});

const onScrollList = async (e: Event) => {
  const target = e.target as HTMLElement;
  const threshold = 50; // pixels from bottom
  if (
    target.scrollHeight - target.scrollTop - target.clientHeight <
    threshold
  ) {
    if (
      !shortlistQuery.isFetchingNextPage.value &&
      shortlistQuery.hasNextPage.value
    ) {
      await shortlistQuery.fetchNextPage();
    }
  }
};

const isSelected = (val: SelectValue | null) => {
  if (isResetOption(val)) return false;

  if (!props.modelValue || val == null) return false;
  if (Array.isArray(props.modelValue)) return props.modelValue.includes(val);
  return props.modelValue === val;
};

const toggleOption = (optVal: SelectValue) => {
  if (isResetOption(optVal)) {
    internalValue.value = props.mode === "multiple" ? [] : null;
    if (props.mode !== "multiple") dialogVisible.value = false;
    return;
  }

  if (props.mode === "multiple") {
    const current = Array.isArray(internalValue.value)
      ? [...internalValue.value]
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

const resolveSelectedLabel = (value: SelectValue) =>
  loadedOptions.value.find((option) => option.value === value)?.label ??
  props.selectedLabelResolver?.(value) ??
  null;

const selectedLabels = computed(() => {
  if (!props.modelValue) return [];
  const vals = Array.isArray(props.modelValue)
    ? props.modelValue
    : [props.modelValue];
  return vals
    .map((val) => resolveSelectedLabel(val))
    .filter((label): label is string => Boolean(label));
});

const selectedChipLabels = computed(() => {
  if (!Array.isArray(props.modelValue) || props.modelValue.length === 0) {
    return [];
  }

  return props.modelValue.map(
    (value) => resolveSelectedLabel(value) ?? String(value)
  );
});

const estimateChipWidth = (label: string) =>
  Math.min(Math.max(label.length * 7 + 30, 54), 168);

const shouldCompactSelectedChips = computed(() => {
  const labels = selectedChipLabels.value;
  if (labels.length <= 1 || filterFieldWidth.value === 0) return false;

  const availableWidth = Math.max(filterFieldWidth.value - 48, 80);
  const totalChipWidth =
    labels.reduce((total, label) => total + estimateChipWidth(label), 0) +
    (labels.length - 1) * 5;

  return totalChipWidth > availableWidth;
});

const selectedSummary = computed(() => {
  const labels = selectedChipLabels.value;
  if (labels.length === 0) return null;

  return {
    extraCount: Math.max(labels.length - 1, 0),
    label: labels[0],
  };
});

const displayLabel = computed(() => {
  if (
    !props.modelValue ||
    (Array.isArray(props.modelValue) && props.modelValue.length === 0)
  ) {
    return props.placeholder || "Select";
  }

  const labels = selectedLabels.value;
  if (labels.length === 0) {
    if (Array.isArray(props.modelValue)) {
      return `${props.modelValue.length} Selected`;
    }
    return String(props.modelValue);
  }

  if (props.mode === "multiple") {
    if (
      typeof props.maxSelectedLabels === "number" &&
      labels.length > props.maxSelectedLabels
    ) {
      return `${labels.length} Selected`;
    }
    return labels.join(", ");
  }

  return labels[0];
});

const hasSelection = computed(() => {
  if (!props.modelValue) return false;
  if (Array.isArray(props.modelValue) && props.modelValue.length === 0)
    return false;
  return true;
});

const dialogBottomBarActions = computed<DialogMobileBottomBarAction[]>(() => {
  if (props.mode === "multiple") {
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
  <template v-if="isMobile && mobileModal">
    <!-- Trigger Button that looks like a select box -->
    <div
      class="filter-select flex w-full items-center border border-[#d9e2ec] bg-white transition cursor-pointer px-3 justify-between"
      :class="[
        filterSizeClasses[actualSize].wrapper,
        { 'bg-[#f1f5f9] opacity-60 pointer-events-none': disabled },
        inputClass,
      ]"
      @click.stop="openMobileDialog"
    >
      <span
        v-if="hasSelection && mode === 'single'"
        class="filter-select-summary-chip"
      >
        <span class="min-w-0 truncate">{{ displayLabel }}</span>
      </span>
      <span v-else class="truncate text-[13px] text-[#94a3b8]">
        {{ displayLabel }}
      </span>
      <span
        class="pi pi-chevron-down text-slate-400 text-[12px] shrink-0 ml-2"
      />
    </div>

    <!-- Mobile Dialog Search & Select -->
    <AppDialog
      :model-value="dialogVisible"
      :title="placeholderValue"
      :mobile-actions="dialogBottomBarActions"
      @update:model-value="dialogVisible = $event"
    >
      <div class="flex flex-col h-full min-h-0 gap-3">
        <!-- Search bar fixed at top -->
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

        <!-- Scrollable Options List -->
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
            <!-- Checkbox/Radio selection indicator -->
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

            <div class="flex-1 min-w-0">
              <span
                class="text-[14px] font-medium truncate block"
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
                class="text-[12px] truncate mt-0.5 block"
                :class="
                  isSelected(option.value) ? 'text-red-500' : 'text-slate-400'
                "
              >
                {{ option.subLabel }}
              </span>
            </div>
          </div>

          <!-- Loading state -->
          <div
            v-if="isLoading"
            class="py-4 text-center text-slate-400 text-xs flex items-center justify-center gap-2"
          >
            <span class="pi pi-spinner pi-spin text-sm" />
            Loading items...
          </div>

          <!-- Empty state -->
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
    ref="filterFieldRef"
    class="filter-select flex items-center border border-[#d9e2ec] bg-white transition"
    :class="[
      filterSizeClasses[actualSize].wrapper,
      { 'bg-[#f1f5f9] opacity-60 pointer-events-none': disabled },
      inputClass,
    ]"
  >
    <MultiSelect
      v-if="mode === 'multiple'"
      v-model="internalValue"
      :options="options"
      :virtual-scroller-options="virtualScrollerOptions"
      option-label="label"
      option-value="value"
      option-disabled="disabled"
      :placeholder="placeholderValue"
      filter
      display="chip"
      :class="['w-full', filterSizeClasses[actualSize].input]"
      :max-selected-labels="maxSelectedLabels"
      :loading="isLoading"
      :pt="multiSelectPassThrough"
      :show-toggle-all="false"
      :disabled="disabled"
      @show="onShow"
      @filter="onFilter"
    >
      <template #value="{ placeholder: valuePlaceholder }">
        <span v-if="selectedChipLabels.length === 0" class="text-slate-400">
          {{ valuePlaceholder }}
        </span>
        <span
          v-else-if="shouldCompactSelectedChips && selectedSummary"
          class="filter-select-summary-chip"
        >
          <span class="min-w-0 truncate">{{ selectedSummary.label }}</span>
          <span v-if="selectedSummary.extraCount > 0" class="shrink-0">
            +{{ selectedSummary.extraCount }}
          </span>
        </span>
        <span v-else class="filter-select-chip-list">
          <span
            v-for="(label, index) in selectedChipLabels"
            :key="`${label}-${index}`"
            class="filter-select-summary-chip"
          >
            <span class="min-w-0 truncate">{{ label }}</span>
          </span>
        </span>
      </template>
      <template #option="{ option }">
        <div
          :class="[
            'relative flex w-full min-w-0 items-start gap-2 rounded-xl border px-2.5 py-1 text-left transition-colors',
            isResetOption(option.value) && isSelected(option.value)
              ? 'border-red-200 bg-red-50 text-slate-800'
              : isSelected(option.value)
                ? 'border-blue-200 bg-blue-50 text-slate-800'
                : 'border-transparent text-slate-700 hover:bg-slate-50',
            option.disabled ? 'cursor-not-allowed opacity-50' : '',
          ]"
        >
          <span
            :class="[
              'mt-0.5 flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded border-2 transition-colors',
              isResetOption(option.value) && isSelected(option.value)
                ? 'border-red-600 bg-red-600'
                : isSelected(option.value)
                  ? 'border-blue-600 bg-blue-600'
                  : 'border-slate-300 bg-white',
            ]"
          >
            <span
              v-if="isSelected(option.value)"
              class="pi pi-check text-[10px] leading-none text-white"
            />
          </span>
          <span class="min-w-0 flex-1">
            <span
              :class="[
                'block truncate text-[13px] leading-5',
                isSelected(option.value)
                  ? 'font-semibold text-slate-800'
                  : 'font-medium text-slate-700',
              ]"
            >
              {{ option.label }}
            </span>
            <span
              v-if="option.subLabel"
              class="block truncate text-[11px] leading-4 text-slate-400"
            >
              {{ option.subLabel }}
            </span>
          </span>
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
      :placeholder="placeholderValue"
      filter
      :class="['w-full', filterSizeClasses[actualSize].input]"
      :loading="isLoading"
      :pt="listPassThrough"
      :disabled="disabled"
      @show="onShow"
      @filter="onFilter"
    >
      <template #value="{ placeholder: valuePlaceholder }">
        <span v-if="hasSelection" class="filter-select-summary-chip">
          <span class="min-w-0 truncate">{{ displayLabel }}</span>
        </span>
        <span v-else class="text-slate-400">{{ valuePlaceholder }}</span>
      </template>
      <template #option="{ option }">
        <div
          :class="[
            'relative flex w-full min-w-0 items-start gap-2 rounded-xl border px-2.5 py-1 text-left transition-colors',
            isResetOption(option.value) && isSelected(option.value)
              ? 'border-red-200 bg-red-50 text-slate-800'
              : isSelected(option.value)
                ? 'border-blue-200 bg-blue-50 text-slate-800'
                : 'border-transparent text-slate-700 hover:bg-slate-50',
            option.disabled ? 'cursor-not-allowed opacity-50' : '',
          ]"
        >
          <span
            :class="[
              'mt-0.5 flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded border-2 transition-colors',
              isResetOption(option.value) && isSelected(option.value)
                ? 'border-red-600 bg-red-600'
                : isSelected(option.value)
                  ? 'border-blue-600 bg-blue-600'
                  : 'border-slate-300 bg-white',
            ]"
          >
            <span
              v-if="isSelected(option.value)"
              class="pi pi-check text-[10px] leading-none text-white"
            />
          </span>
          <span class="min-w-0 flex-1">
            <span
              :class="[
                'block truncate text-[13px] leading-5',
                isSelected(option.value)
                  ? 'font-semibold text-slate-800'
                  : 'font-medium text-slate-700',
              ]"
            >
              {{ option.label }}
            </span>
            <span
              v-if="option.subLabel"
              class="block truncate text-[11px] leading-4 text-slate-400"
            >
              {{ option.subLabel }}
            </span>
          </span>
        </div>
      </template>
    </Select>
  </div>
</template>

<style scoped>
.filter-select {
  border-radius: 12px;
  box-shadow: 0 1px 2px 0 rgba(15, 23, 42, 0.04);
  overflow: visible;
}

.filter-select:focus-within {
  border-color: #94a3b8;
  box-shadow: 0 0 0 2px #dbeafe;
}

.filter-select:hover:not(:focus-within) {
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
  font-weight: 400 !important;
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
  font-size: 13px;
}

:deep(.p-multiselect-label) {
  display: flex !important;
  max-width: 100% !important;
  align-items: center !important;
  gap: 5px !important;
  overflow: hidden !important;
  padding: 0 10px !important;
  flex-wrap: nowrap !important;
}

:deep(.p-multiselect-chip) {
  display: inline-flex !important;
  max-width: 10rem !important;
  min-width: max-content !important;
  height: 23px !important;
  flex: 0 0 auto !important;
  align-items: center !important;
  gap: 5px !important;
  border: 1px solid #cbd5e1 !important;
  border-radius: 10px !important;
  background: linear-gradient(180deg, #ffffff 0%, #f1f5f9 100%) !important;
  padding: 0 7px !important;
  color: #475569 !important;
  font-size: 11px !important;
  font-weight: 700 !important;
  line-height: 1.15 !important;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    0 1px 2px rgba(15, 23, 42, 0.08) !important;
}

:deep(.p-multiselect-chip .p-chip-label),
:deep(.p-multiselect-chip-label) {
  min-width: 0 !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
}

:deep(.p-multiselect-chip .p-chip-remove-icon),
:deep(.p-multiselect-chip-remove-icon) {
  width: 12px !important;
  height: 12px !important;
  min-width: 12px !important;
  border-radius: 999px !important;
  color: #64748b !important;
  opacity: 0.8 !important;
}

.filter-select-chip-list {
  display: inline-flex;
  max-width: 100%;
  min-width: 0;
  align-items: center;
  gap: 5px;
  overflow: hidden;
}

.filter-select-summary-chip {
  display: inline-flex;
  max-width: 100%;
  min-width: 0;
  height: 23px;
  align-items: center;
  gap: 5px;
  overflow: hidden;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  background: linear-gradient(180deg, #ffffff 0%, #f1f5f9 100%);
  padding: 0 7px;
  color: #475569;
  font-size: 11px;
  font-weight: 700;
  line-height: 1.15;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    0 1px 2px rgba(15, 23, 42, 0.08);
}

:deep(.p-select-dropdown),
:deep(.p-multiselect-dropdown) {
  color: #64748b !important;
  margin-right: 4px !important;
}

:deep(.p-select-overlay),
:deep(.p-select-panel),
:deep(.p-multiselect-overlay),
:deep(.p-multiselect-panel) {
  border-radius: 14px !important;
  box-shadow:
    0 10px 30px rgba(15, 23, 42, 0.1),
    0 2px 8px rgba(15, 23, 42, 0.06) !important;
  border: 1px solid #e2e8f0 !important;
  margin-top: 4px !important;
  padding: 0 !important;
}

:deep(.p-select-header),
:deep(.p-multiselect-header) {
  margin: 0 !important;
  padding: 8px !important;
  background: #ffffff !important;
  border-bottom: 1px solid #e2e8f0 !important;
  gap: 0 !important;
}

:deep(.p-select-header .p-iconfield),
:deep(.p-multiselect-header .p-iconfield) {
  width: 100% !important;
  min-width: 0 !important;
}

:deep(.p-select-header .p-inputtext),
:deep(.p-multiselect-header .p-inputtext) {
  width: 100% !important;
  height: 32px !important;
  padding: 0 36px 0 12px !important;
  border: 1px solid #e2e8f0 !important;
  border-radius: 8px !important;
  background: #f8fafc !important;
  color: #0f172a !important;
  font-size: 12px !important;
  font-weight: 400 !important;
  line-height: 32px !important;
  outline: none !important;
  box-shadow: none !important;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease !important;
}

:deep(.p-select-header .p-inputtext:focus),
:deep(.p-multiselect-header .p-inputtext:focus) {
  border-color: #93c5fd !important;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1) !important;
  background: #ffffff !important;
}

:deep(.p-select-header .p-inputtext::placeholder),
:deep(.p-multiselect-header .p-inputtext::placeholder) {
  color: #94a3b8 !important;
}

:deep(.p-select-header .p-inputicon),
:deep(.p-multiselect-header .p-inputicon) {
  right: 12px !important;
  color: #94a3b8 !important;
}

:deep(.p-select-header .p-icon),
:deep(.p-multiselect-header .p-icon) {
  display: block !important;
  width: 14px !important;
  height: 14px !important;
  color: #94a3b8 !important;
}

:deep(.p-select-items),
:deep(.p-select-list),
:deep(.p-multiselect-items),
:deep(.p-multiselect-list) {
  padding: 1px 4px 4px !important;
}

:deep(.p-select-item),
:deep(.p-select-option),
:deep(.p-multiselect-item),
:deep(.p-multiselect-option) {
  margin-top: 0 !important;
  padding: 0 !important;
  border-radius: 12px !important;
  background: transparent !important;
  font-size: 13px !important;
  color: #334155 !important;
}

:deep(.p-multiselect-item .p-checkbox),
:deep(.p-multiselect-option .p-checkbox),
:deep(.p-multiselect-item [data-pc-name="checkbox"]),
:deep(.p-multiselect-option [data-pc-name="checkbox"]) {
  display: none !important;
}

:deep(.p-select-item:not(.p-highlight):not(.p-select-item--selected):hover),
:deep(
  .p-multiselect-item:not(.p-highlight):not(.p-multiselect-item--selected):hover
) {
  background-color: #f1f5f9 !important;
  color: #0f172a !important;
}

:deep(.p-select-item.p-highlight),
:deep(.p-select-option.p-highlight),
:deep(.p-select-option-selected),
:deep(.p-multiselect-item.p-highlight),
:deep(.p-multiselect-option.p-highlight),
:deep(.p-multiselect-option-selected) {
  background-color: transparent !important;
  color: inherit !important;
  font-weight: inherit !important;
}

:deep(.p-select-empty-message),
:deep(.p-multiselect-empty-message) {
  text-align: center !important;
  padding: 20px !important;
  color: #94a3b8 !important;
  font-size: 13px !important;
}
</style>
