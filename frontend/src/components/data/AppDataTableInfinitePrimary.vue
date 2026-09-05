<script setup lang="ts">
import Column from "primevue/column";
import DataTable, {
  type DataTableSelectAllChangeEvent,
} from "primevue/datatable";
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from "vue";

import type {
  LoadParams,
  LoadResult,
} from "@/components/data/AppDataTableV1.vue";

import { getObjectPathValue } from "@/lib/utils/objectPath";

const EXTRA_FILTERS_READY_META_KEY = "__appFiltersReady";

type SelectionMode = "multiple" | "single" | undefined;
type SelectionActionVariant = "primary" | "secondary";
type SelectionAction = {
  icon?: string;
  key: string;
  label: string;
  variant?: SelectionActionVariant;
};

const props = withDefaults(
  defineProps<{
    dataKey?: string;
    emptyLabel?: string;
    enabled?: boolean;
    endLabel?: string;
    error?: string;
    extraFilters?: Record<string, unknown> | null;
    initialPage?: number;
    loadFn: (params: LoadParams) => Promise<LoadResult<unknown>>;
    loadingLabel?: string;
    loadingMoreLabel?: string;
    pageSize?: number;
    rowClass?:
      | ((data: unknown) => string | Record<string, boolean>)
      | undefined;
    scrollThreshold?: number;
    selection?: unknown[] | null;
    selectionActions?: SelectionAction[];
    selectionColumnOnMobile?: boolean;
    selectionItemLabel?: string;
    selectionItemLabelPlural?: string;
    selectionMode?: SelectionMode;
    sortDirection?: "asc" | "desc";
    sortField?: string | null;
    tableClass?: string;
    valueKey?: string;
    withoutCheckBoxSelection?: boolean;
  }>(),
  {
    dataKey: "id",
    emptyLabel: "No data found.",
    enabled: true,
    endLabel: "All data loaded",
    error: "",
    extraFilters: () => ({}),
    initialPage: 1,
    loadingLabel: "Loading data...",
    loadingMoreLabel: "Loading more data...",
    pageSize: 10,
    rowClass: undefined,
    scrollThreshold: 160,
    selection: undefined,
    selectionActions: () => [],
    selectionColumnOnMobile: false,
    selectionItemLabel: "item",
    selectionItemLabelPlural: undefined,
    selectionMode: undefined,
    sortDirection: "desc",
    sortField: null,
    tableClass: "",
    valueKey: undefined,
    withoutCheckBoxSelection: false,
  }
);

const emit = defineEmits<{
  (e: "error", error: unknown): void;
  (e: "loaded", result: LoadResult<unknown>): void;
  (e: "selection-action", action: SelectionAction, selection: unknown[]): void;
  (e: "update:selection", selection: unknown[]): void;
}>();

const tableRef = ref<InstanceType<typeof DataTable> | null>(null);
const internalSelection = ref<unknown[]>([]);
const items = ref<unknown[]>([]);
const loading = ref(true);
const isLoadingRequest = ref(false);
const localError = ref("");
const page = ref(props.initialPage);
const totalPages = ref(1);
const scrollContainer = ref<HTMLElement | null>(null);

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
  () => `${selection.value.length} ${selectionItemText.value} selected`
);
const shouldRenderSelectionColumn = computed(
  () => props.selectionMode === "multiple" && !props.withoutCheckBoxSelection
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

const isCurrentPageFullySelected = computed(() => {
  if (props.selectionMode !== "multiple") return false;
  if (!items.value.length) return false;

  return items.value.every((item) => isSelected(item));
});

const clearSelection = () => {
  selection.value = [];
};

const emitSelectionAction = (action: SelectionAction) => {
  emit("selection-action", action, selection.value);
};

const selectionActionClass = (action: SelectionAction) =>
  action.variant === "secondary"
    ? "border border-[#dbeafe] bg-white text-[#1d4ed8] hover:bg-[#eff6ff]"
    : "border border-[#dc2626] bg-[#dc2626] text-white hover:bg-[#b91c1c]";

const forceSelectAllCurrentItems = () => {
  if (props.selectionMode !== "multiple") return;

  const fieldKey = selectionValueKey.value;
  if (!fieldKey) {
    selection.value = Array.from(new Set([...selection.value, ...items.value]));
    return;
  }

  const merged = new Map<string, unknown>();
  [...selection.value, ...items.value].forEach((item) => {
    const value = getObjectPathValue(item, fieldKey);
    if (value === undefined || value === null) return;
    merged.set(String(value), item);
  });

  selection.value = Array.from(merged.values());
};

const removeCurrentPageItemsFromSelection = () => {
  if (props.selectionMode !== "multiple") return;
  if (!selection.value.length || !items.value.length) return;

  const fieldKey = selectionValueKey.value;
  if (!fieldKey) {
    const currentItems = new Set(items.value);
    selection.value = selection.value.filter((item) => !currentItems.has(item));
    return;
  }

  const currentPageIds = new Set(
    items.value
      .map((item) => getObjectPathValue(item, fieldKey))
      .filter(
        (value): value is string | number =>
          typeof value === "string" || typeof value === "number"
      )
      .map(String)
  );

  selection.value = selection.value.filter((item) => {
    const value = getObjectPathValue(item, fieldKey);
    if (value === undefined || value === null) return true;
    return !currentPageIds.has(String(value));
  });
};

const onSelectAllChange = (event: DataTableSelectAllChangeEvent) => {
  if (event.checked) {
    forceSelectAllCurrentItems();
    return;
  }

  removeCurrentPageItemsFromSelection();
};

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
  if (shouldShowLoading) loading.value = true;
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
    if (shouldShowLoading) loading.value = false;
  }
};

const reload = (options?: { withoutLoading?: boolean }) =>
  loadData({ reset: true, withoutLoading: options?.withoutLoading });

const handleScroll = (event: Event) => {
  const target = event.target as HTMLElement | null;
  if (!target || !canLoadMore()) return;

  const distanceToBottom =
    target.scrollHeight - target.scrollTop - target.clientHeight;

  if (distanceToBottom <= props.scrollThreshold) void loadData();
};

const bindScrollContainer = async () => {
  await nextTick();

  scrollContainer.value?.removeEventListener("scroll", handleScroll);

  const root =
    tableRef.value && "$el" in tableRef.value ? tableRef.value.$el : null;
  const container = (root as HTMLElement | null)?.querySelector(
    ".p-datatable-table-container"
  ) as HTMLElement | null;

  scrollContainer.value = container;
  scrollContainer.value?.addEventListener("scroll", handleScroll, {
    passive: true,
  });
};

watch(
  () => [
    props.enabled,
    props.extraFilters,
    props.sortField,
    props.sortDirection,
    props.pageSize,
  ],
  ([enabled]) => {
    if (enabled && areExtraFiltersReady(props.extraFilters)) void reload();
  },
  { deep: true }
);

watch(items, () => {
  void bindScrollContainer();
});

onMounted(() => {
  if (props.enabled && areExtraFiltersReady(props.extraFilters)) void reload();
  void bindScrollContainer();
});

onBeforeUnmount(() => {
  scrollContainer.value?.removeEventListener("scroll", handleScroll);
});

defineExpose({
  reload,
});
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <slot name="error" :error="error || localError">
      <div
        v-if="error || localError"
        class="mb-3 border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
      >
        {{ error || localError }}
      </div>
    </slot>

    <div
      v-if="selectionMode && selection.length"
      class="mb-3 shrink-0 border border-[#bfdbfe] bg-gradient-to-b from-[#eff6ff] to-white px-[12px] py-[12px] shadow-[0_10px_24px_rgba(37,99,235,0.08)] sm:px-[16px]"
    >
      <slot
        name="selection"
        :selection="selection"
        :clear-selection="clearSelection"
        :selection-action="emitSelectionAction"
        :selection-actions="props.selectionActions"
      >
        <div class="flex items-center justify-between gap-[12px]">
          <span class="text-[13px] font-semibold text-[#1d4ed8]">
            {{ selectionSummaryText }}
          </span>
          <div class="flex items-center gap-[8px]">
            <button
              type="button"
              class="flex h-[30px] items-center justify-center border border-[#dbeafe] bg-white px-[14px] text-[12px] font-semibold text-[#1d4ed8] shadow-sm transition-colors hover:bg-[#eff6ff]"
              @click="clearSelection"
            >
              Clear
            </button>
            <button
              v-for="action in props.selectionActions"
              :key="action.key"
              type="button"
              class="flex h-[30px] items-center justify-center gap-[6px] px-[14px] text-[12px] font-semibold shadow-sm transition-colors"
              :class="selectionActionClass(action)"
              @click="emitSelectionAction(action)"
            >
              <span
                v-if="action.icon"
                :class="['pi text-[11px]', action.icon]"
              />
              <span>{{ action.label }}</span>
            </button>
          </div>
        </div>
      </slot>
    </div>

    <div
      class="min-h-0 flex-1 overflow-hidden border border-x-0 border-[#d8e2ef]"
    >
      <DataTable
        ref="tableRef"
        v-model:selection="selection"
        :class="['app-data-table-infinite-primary', tableClass]"
        :row-class="rowClass"
        :select-all="isCurrentPageFullySelected"
        :selection-mode="selectionMode"
        :value="items"
        :data-key="dataKey"
        scrollable
        scroll-height="flex"
        show-gridlines
        striped-rows
        @select-all-change="onSelectAllChange"
      >
        <template #empty>
          <slot name="empty" :loading="loading">
            <div class="px-4 py-10 text-center text-sm text-slate-500">
              <template v-if="loading">
                <span class="pi pi-spinner pi-spin mr-2 text-[13px]" />
                {{ loadingLabel }}
              </template>
              <template v-else>{{ emptyLabel }}</template>
            </div>
          </slot>
        </template>

        <Column
          v-if="shouldRenderSelectionColumn"
          :selection-mode="'multiple'"
          :class="
            props.selectionColumnOnMobile
              ? 'data-table-infinite__selection-column'
              : 'data-table-infinite__selection-column hidden md:table-cell'
          "
          :style="{ width: '50px', minWidth: '50px', maxWidth: '50px' }"
        >
          <template #header>
            <span />
          </template>
        </Column>

        <slot />
      </DataTable>
    </div>

    <div
      v-if="
        (loading && items.length) ||
        (items.length && !canLoadMore() && !loading)
      "
      class="flex justify-center py-4"
    >
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
</template>

<style scoped>
:deep(.app-data-table-infinite-primary.p-datatable) {
  height: 100%;
}

:deep(.app-data-table-infinite-primary .p-datatable-table-container) {
  height: 100%;
}

:deep(.app-data-table-infinite-primary .p-datatable-thead > tr > th) {
  border-color: #e2e8f0;
  background: #f8fafc;
  padding: 11px 12px;
  color: #64748b;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

:deep(.app-data-table-infinite-primary .p-datatable-tbody > tr > td) {
  border-color: #eef2f7;
  padding: 14px 12px;
  vertical-align: top;
}

:deep(
  .app-data-table-infinite-primary .p-datatable-tbody > tr:nth-child(even)
) {
  background: #fbfcfe;
}

:deep(.app-data-table-infinite-primary .p-column-header-content) {
  justify-content: flex-start;
}

:deep(
  .app-data-table-infinite-primary
    .p-datatable-thead
    > tr
    > th[data-p-selection-column="true"]
) {
  width: 55px !important;
  min-width: 55px !important;
  max-width: 55px !important;
  padding-left: 1.15rem !important;
  padding-right: 0 !important;
  text-align: left;
}

:deep(
  .app-data-table-infinite-primary
    .p-datatable-tbody
    > tr
    > td[data-p-selection-column="true"]
),
:deep(.app-data-table-infinite-primary .data-table-infinite__selection-column) {
  width: 55px !important;
  min-width: 55px !important;
  max-width: 55px !important;
  padding-left: 1.15rem !important;
  padding-right: 0 !important;
  text-align: left;
}

:deep(
  .app-data-table-infinite-primary
    .p-datatable-thead
    > tr
    > th[data-p-selection-column="true"]
    .p-column-header-content
) {
  align-items: center;
  justify-content: center;
  width: 100%;
}

:deep(
  .app-data-table-infinite-primary
    .p-datatable-thead
    > tr
    > th[data-p-selection-column="true"]
    .p-checkbox
) {
  display: none;
}

:deep(
  .app-data-table-infinite-primary
    .p-datatable-tbody
    > tr
    > td[data-p-selection-column="true"]
) {
  align-items: center;
  display: table-cell;
  text-align: center;
  vertical-align: middle;
}

:deep(
  .app-data-table-infinite-primary
    .p-datatable-tbody
    > tr
    > td[data-p-selection-column="true"]
    .p-checkbox
) {
  display: inline-flex;
  margin-left: 0;
  margin-right: 0;
  transform: translateX(-0.65rem);
}
</style>
