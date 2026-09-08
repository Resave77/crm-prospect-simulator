<script setup lang="ts">
import DataView from "primevue/dataview";
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";

import type {
  LoadParams,
  LoadResult,
} from "@/components/data/AppDataTableV1.vue";

import { getGlobalEventBus } from "@/lib/utils/broadcasterNotifier";

const DEBOUNCE_DELAY = 1000;
const EXTRA_FILTERS_READY_META_KEY = "__appFiltersReady";

type CheckboxPlacement =
  | "bottom-left"
  | "bottom-right"
  | "top-left"
  | "top-right";
type SelectionMode = "multiple" | "single" | null;
type SelectionActionsPlacement = "bottom" | "top";
type SelectionActionVariant = "primary" | "secondary";
type SelectionAction = {
  icon?: string;
  key: string;
  label: string;
  variant?: SelectionActionVariant;
};

const getObjectPathValue = (item: unknown, path?: string) => {
  if (!path || item == null || typeof item !== "object") return item;

  return path.split(".").reduce<unknown>((current, key) => {
    if (current == null || typeof current !== "object") return undefined;
    return (current as Record<string, unknown>)[key];
  }, item);
};

const props = withDefaults(
  defineProps<{
    checkboxPlacement?: CheckboxPlacement;
    dataKey?: string;
    emptyLabel?: string;
    enabled?: boolean;
    endLabel?: string;
    error?: string;
    eventBusNamesForRefresh?: string[];
    extraFilters?: Record<string, unknown> | null;
    initialPage?: number;
    listClass?: string;
    loadFn: (params: LoadParams) => Promise<LoadResult<unknown>>;
    loadingLabel?: string;
    loadingMoreLabel?: string;
    pageSize?: number;
    scrollThreshold?: number;
    selection?: unknown[] | null;
    selectionActions?: SelectionAction[];
    selectionActionsAlwaysVisible?: boolean;
    selectionActionsClass?: string;
    selectionActionsPlacement?: SelectionActionsPlacement;
    selectionItemLabel?: string;
    selectionItemLabelPlural?: string;
    selectionMode?: SelectionMode;
    sortDirection?: "asc" | "desc";
    sortField?: string | null;
    valueKey?: string;
    withCheckbox?: boolean;
  }>(),
  {
    checkboxPlacement: "top-right",
    dataKey: "id",
    emptyLabel: "No data found.",
    enabled: true,
    endLabel: "All data loaded",
    error: "",
    eventBusNamesForRefresh: () => [],
    extraFilters: () => ({}),
    initialPage: 1,
    loadingLabel: "Loading data...",
    loadingMoreLabel: "Loading more data...",
    pageSize: 10,
    scrollThreshold: 160,
    selection: undefined,
    selectionActions: () => [
      {
        icon: "pi-pencil",
        key: "bulk-edit",
        label: "Bulk Edit",
        variant: "primary",
      },
      {
        icon: "pi-file-excel",
        key: "download-excel",
        label: "Download Excel",
        variant: "secondary",
      },
    ],
    selectionActionsAlwaysVisible: false,
    selectionActionsClass: "",
    selectionActionsPlacement: "top",
    selectionItemLabel: "item",
    selectionItemLabelPlural: undefined,
    selectionMode: null,
    sortDirection: "desc",
    sortField: null,
    valueKey: undefined,
    withCheckbox: false,
  }
);

const emit = defineEmits<{
  (e: "error", error: unknown): void;
  (e: "loaded", result: LoadResult<unknown>): void;
  (e: "selection-action", action: SelectionAction, selection: unknown[]): void;
  (e: "selection-download", selection: unknown[]): void;
  (e: "update:selection", selection: unknown[]): void;
}>();

const internalSelection = ref<unknown[]>([]);
const items = ref<unknown[]>([]);
const loading = ref(true);
const isLoadingRequest = ref(false);
const localError = ref("");
const page = ref(props.initialPage);
const totalPages = ref(1);

const selection = computed<unknown[]>({
  get: () =>
    props.selection !== undefined && props.selection !== null
      ? props.selection
      : internalSelection.value,
  set: (value) => {
    const normalized = value ?? [];
    if (props.selection === undefined) internalSelection.value = normalized;
    emit("update:selection", normalized);
  },
});

const selectionValueKey = computed(() => props.valueKey ?? props.dataKey);
const selectionItemText = computed(() => {
  const count = selection.value.length;
  return count === 1
    ? props.selectionItemLabel
    : props.selectionItemLabelPlural || `${props.selectionItemLabel}s`;
});
const selectionSummaryText = computed(
  () => `${selectionItemText.value} selected`
);

const canLoadMore = () =>
  page.value < totalPages.value && !isLoadingRequest.value;

const getSelectionValue = (item: unknown) =>
  getObjectPathValue(item, selectionValueKey.value);

const isSelected = (item: unknown) => {
  const fieldKey = selectionValueKey.value;
  if (!fieldKey) return selection.value.includes(item);

  const itemValue = getSelectionValue(item);
  return selection.value.some(
    (selectedItem) => getSelectionValue(selectedItem) === itemValue
  );
};

const clearSelection = () => {
  selection.value = [];
};

const emitSelectionAction = (action: SelectionAction) => {
  emit("selection-action", action, selection.value);

  if (action.key === "download-excel") {
    emit("selection-download", selection.value);
  }
};

const selectionActionClass = (action: SelectionAction) =>
  action.variant === "secondary"
    ? "border border-[#bfdbfe] bg-white text-[#1d4ed8] shadow-[0_4px_12px_rgba(29,78,216,0.08)] hover:border-[#93c5fd] hover:bg-[#dbeafe] active:scale-[0.99] focus:ring-[#bfdbfe]"
    : "border border-[#dc2626] bg-[#dc2626] text-white shadow-[0_10px_20px_rgba(220,38,38,0.22)] hover:bg-[#b91c1c] active:scale-[0.99] focus:ring-[#fecaca]";

const toggleSelection = (item: unknown) => {
  if (!props.selectionMode) return;

  if (props.selectionMode === "single") {
    selection.value = isSelected(item) ? [] : [item];
    return;
  }

  if (isSelected(item)) {
    const itemValue = getSelectionValue(item);
    selection.value = selection.value.filter((selectedItem) =>
      selectionValueKey.value
        ? getSelectionValue(selectedItem) !== itemValue
        : selectedItem !== item
    );
    return;
  }

  selection.value = [...selection.value, item];
};

const checkboxPlacementClass = computed(() => {
  const placementClass: Record<CheckboxPlacement, string> = {
    "bottom-left": "bottom-4 left-4",
    "bottom-right": "bottom-4 right-4",
    "top-left": "left-4 top-4",
    "top-right": "right-4 top-4",
  };

  return placementClass[props.checkboxPlacement];
});

const areExtraFiltersReady = (
  value: Record<string, unknown> | null | undefined
) => {
  if (!value || typeof value !== "object") return true;
  return (
    (
      value as Record<string, unknown> & {
        [EXTRA_FILTERS_READY_META_KEY]?: boolean;
      }
    )[EXTRA_FILTERS_READY_META_KEY] !== false
  );
};

const loadData = async (
  options: { reset?: boolean; withoutLoading?: boolean } = {}
) => {
  const reset = options.reset ?? false;
  const shouldShowLoading = !options.withoutLoading;
  if (
    !props.enabled ||
    isLoadingRequest.value ||
    !areExtraFiltersReady(props.extraFilters)
  ) {
    return;
  }

  isLoadingRequest.value = true;
  if (shouldShowLoading) {
    loading.value = true;
  }
  localError.value = "";

  try {
    const nextPage = reset ? props.initialPage : page.value + 1;
    const result = await props.loadFn({
      extraFilters: props.extraFilters,
      page: nextPage,
      pageSize: props.pageSize,
      sortDirection: props.sortDirection,
      sortField: props.sortField,
    });

    items.value = reset ? result.items : [...items.value, ...result.items];
    page.value = result.pagination?.page ?? nextPage;
    totalPages.value = result.pagination?.total_pages ?? 1;

    emit("loaded", result);
  } catch (error: unknown) {
    localError.value =
      error instanceof Error ? error.message : "Failed to load data.";
    emit("error", error);
  } finally {
    isLoadingRequest.value = false;
    if (shouldShowLoading) {
      loading.value = false;
    }
  }
};

const reload = (options?: { withoutLoading?: boolean }) =>
  loadData({ reset: true, withoutLoading: options?.withoutLoading });

const onScroll = (event: Event) => {
  const target = event.target as HTMLElement | null;
  if (!target || !canLoadMore()) return;

  const distanceToBottom =
    target.scrollHeight - target.scrollTop - target.clientHeight;

  if (distanceToBottom <= props.scrollThreshold) void loadData();
};

let extraFiltersTimer: number | undefined;
watch(
  () => props.extraFilters,
  (nextFilters) => {
    if (!props.enabled || !areExtraFiltersReady(nextFilters)) return;
    if (extraFiltersTimer) window.clearTimeout(extraFiltersTimer);
    extraFiltersTimer = window.setTimeout(() => {
      extraFiltersTimer = undefined;
      void reload();
    }, DEBOUNCE_DELAY);
  },
  { deep: true }
);

watch(
  () => [props.enabled, props.sortField, props.sortDirection],
  ([enabled]) => {
    if (enabled && areExtraFiltersReady(props.extraFilters)) void reload();
  }
);

onMounted(() => {
  if (props.enabled && areExtraFiltersReady(props.extraFilters)) void reload();

  if (props.eventBusNamesForRefresh?.length) {
    const bus = getGlobalEventBus();
    for (const eventName of props.eventBusNamesForRefresh) {
      bus.subscribe(eventName, onEventBusRefresh);
    }
  }
});

onBeforeUnmount(() => {
  if (extraFiltersTimer) {
    window.clearTimeout(extraFiltersTimer);
  }

  if (props.eventBusNamesForRefresh?.length) {
    const bus = getGlobalEventBus();
    for (const eventName of props.eventBusNamesForRefresh) {
      bus.unsubscribe(eventName, onEventBusRefresh);
    }
  }
});

const onEventBusRefresh = () => {
  void reload({ withoutLoading: true });
};

defineExpose({
  reload,
});
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <slot name="error" :error="error || localError">
      <div
        v-if="error || localError"
        class="mb-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
      >
        {{ error || localError }}
      </div>
    </slot>

    <div
      v-if="
        selectionMode &&
        (selection.length > 0 || props.selectionActionsAlwaysVisible)
      "
      :class="[
        'shrink-0 rounded-[16px] border border-[#bfdbfe] bg-gradient-to-b from-[#eff6ff] to-white px-[12px] py-[12px] shadow-[0_10px_24px_rgba(37,99,235,0.08)] sm:px-[16px]',
        props.selectionActionsPlacement === 'bottom' ? 'order-3 mt-3' : 'mb-3',
        props.selectionActionsClass,
      ]"
    >
      <slot
        name="selection"
        :selection="selection"
        :clear-selection="clearSelection"
        :selection-action="emitSelectionAction"
        :selection-actions="props.selectionActions"
      >
        <div class="mb-[10px] flex items-center justify-between gap-[10px]">
          <span
            class="inline-flex min-w-0 items-center gap-2 rounded-full bg-white px-[10px] py-[6px] font-['Inter'] text-[12px] font-bold text-[#1d4ed8] shadow-sm"
          >
            <span
              class="flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded-full bg-[#dbeafe] text-[10px] text-[#1d4ed8]"
            >
              {{ selection.length }}
            </span>
            <span class="truncate">{{ selectionSummaryText }}</span>
          </span>
          <button
            type="button"
            class="shrink-0 rounded-full bg-white px-[11px] py-[6px] font-['Inter'] text-[11px] font-bold text-[#1d4ed8] shadow-sm transition hover:bg-[#dbeafe] active:scale-[0.98] focus:outline-none focus:ring-2 focus:ring-[#bfdbfe] focus:ring-offset-0"
            :disabled="!selection.length"
            @click="clearSelection"
          >
            Clear
          </button>
        </div>
        <div
          class="grid gap-2"
          :class="
            props.selectionActions.length > 1 ? 'grid-cols-2' : 'grid-cols-1'
          "
        >
          <button
            v-for="action in props.selectionActions"
            :key="action.key"
            type="button"
            class="inline-flex h-[40px] min-w-0 items-center justify-center gap-1.5 rounded-[12px] px-[10px] font-['Inter'] text-[11.5px] font-bold transition focus:outline-none focus:ring-2 focus:ring-offset-0 sm:px-[12px] sm:text-[12px]"
            :class="selectionActionClass(action)"
            :disabled="!selection.length"
            @click="emitSelectionAction(action)"
          >
            <span
              v-if="action.icon"
              :class="['pi shrink-0 text-[12px] leading-none', action.icon]"
            />
            <span class="min-w-0 truncate">{{ action.label }}</span>
          </button>
        </div>
      </slot>
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto" @scroll="onScroll">
      <DataView :value="items" :data-key="dataKey">
        <template #empty>
          <slot name="empty" :loading="loading">
            <div
              class="rounded-2xl border border-dashed border-slate-200 bg-white px-4 py-10 text-center text-sm text-slate-500"
            >
              <template v-if="loading">
                <span class="pi pi-spinner pi-spin mr-2 text-[13px]" />
                {{ loadingLabel }}
              </template>
              <template v-else>{{ emptyLabel }}</template>
            </div>
          </slot>
        </template>

        <template #list="slotProps">
          <div :class="props.listClass || 'space-y-3'">
            <div
              v-for="(item, index) in slotProps.items"
              :key="item?.[dataKey]"
              class="relative"
            >
              <button
                v-if="withCheckbox && selectionMode"
                type="button"
                class="absolute z-10 inline-flex h-6 w-6 items-center justify-center rounded-lg border shadow-sm transition"
                :class="checkboxPlacementClass"
                :aria-pressed="isSelected(item)"
                :aria-label="isSelected(item) ? 'Unselect item' : 'Select item'"
                @click.stop="toggleSelection(item)"
              >
                <span
                  class="flex h-4 w-4 items-center justify-center rounded-[4px] text-[10px] font-bold transition"
                  :class="
                    isSelected(item)
                      ? 'border-[#dc2626] bg-[#dc2626] text-white'
                      : 'border border-slate-300 bg-white text-transparent'
                  "
                >
                  <span class="pi pi-check" />
                </span>
              </button>
              <slot
                name="item"
                :item="item"
                :index="index"
                :selected="isSelected(item)"
                :toggle-selection="() => toggleSelection(item)"
              />
            </div>
          </div>
        </template>
      </DataView>

      <div class="flex justify-center py-4">
        <slot
          name="loading-more"
          :loading="loading"
          :has-items="Boolean(items.length)"
        >
          <p
            v-if="loading && items.length"
            class="text-center text-[11px] text-slate-400"
          >
            <span class="pi pi-spinner pi-spin mr-1 text-[10px]" />
            {{ loadingMoreLabel }}
          </p>
        </slot>

        <slot
          name="end"
          :has-items="Boolean(items.length)"
          :can-load-more="canLoadMore()"
          :loading="loading"
        >
          <p
            v-if="items.length && !canLoadMore() && !loading"
            class="text-center text-[11px] text-slate-400"
          >
            {{ endLabel }}
          </p>
        </slot>
      </div>
    </div>
  </div>
</template>
