<!-- @ts-nocheck: legacy table API supports multiple pagination contracts. -->
<script setup lang="ts">
// @ts-nocheck
import Column from "primevue/column";
import DataTable, {
  type DataTableRowClickEvent,
  type DataTableSelectAllChangeEvent,
  type DataTableSortEvent,
} from "primevue/datatable";
import InputNumber from "primevue/inputnumber";
import Paginator, { type PageState } from "primevue/paginator";
import Popover from "primevue/popover";
import ProgressSpinner from "primevue/progressspinner";
import Select from "primevue/select";
import {
  type ComponentPublicInstance,
  computed,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from "vue";

import { useCurrentPageReportTemplate } from "@/composables/useCurrentPageReportTemplate";
import {
  normalizePageSizePresets,
  resolvePageSizeBounds,
} from "@/composables/useTablePagination";
import { useTableSelection } from "@/composables/useTableSelection";
import { getGlobalEventBus } from "@/lib/utils/broadcasterNotifier";
import {
  createObjectPathValue,
  getObjectPathValue,
} from "@/lib/utils/objectPath";
import { useAuthStore } from "@/stores/authStore";
import { useDataTablePreferencesStore } from "@/stores/dataTablePreferences";

// Constants
const DEBOUNCE_DELAY = 1000;
const EXTRA_FILTERS_READY_META_KEY = "__appFiltersReady";
/** Sort direction type for table columns */
type SortDirection = "asc" | "desc";

/** Selection mode for table rows */
type SelectionMode = "single" | "multiple" | null;

/** Configuration for column controls (show/hide, freeze/unfreeze) */
type ColumnControl = {
  /** Whether the column is frozen by default */
  frozen?: boolean;
  /** Whether the column is hidden by default */
  hidden?: boolean;
  /** Unique identifier for the column */
  key: string;
  /** Display label for the column in controls menu */
  label: string;
};

/** Parameters passed to the load function for data fetching */
export interface LoadParams {
  /** Additional filters applied to the data query */
  extraFilters?: Record<string, unknown> | null;
  /** Current page number (1-based) */
  page: number;
  /** Number of items per page */
  pageSize: number;
  /** Sort direction if sorting is applied */
  sortDirection?: SortDirection;
  /** Field name to sort by if sorting is applied */
  sortField?: string | null;
}

/** Pagination information returned from the server */
export interface PaginationInfo {
  /** Current page number */
  page: number;
  /** Total number of items across all pages */
  total: number;
  /** Total number of pages */
  total_pages: number;
}

/** Result structure expected from the load function */
export interface LoadResult<T> {
  /** Array of data items for the current page */
  items: T[];
  /** Pagination metadata */
  pagination: PaginationInfo;
}

const props = withDefaults(
  defineProps<{
    /** Array of column controls for show/hide and freeze/unfreeze functionality */
    columnControls?: ColumnControl[];
    /** Unique key for storing column preferences in localStorage */
    columnControlsKey?: string;
    /** Field name to use as unique identifier for each row (required for selection) */
    dataKey?: string;
    /** Event bus names to listen for refresh events */
    eventBusNamesForRefresh?: string[];
    /** Additional filters to be passed to the load function for data fetching */
    extraFilters?: Record<string, unknown> | null;
    /** Hide page size and go-to pagination controls */
    hidePaginationControls?: boolean;
    /** Initial page number to load (1-based) */
    initialPage?: number;
    /** Initial number of items per page */
    initialPageSize?: number;
    /** Function that loads data based on provided parameters (page, filters, etc.) */
    loadFn: (params: LoadParams) => Promise<LoadResult<unknown>>;
    /** Maximum height of the table in pixels (if not set, table will be flexible) */
    maxHeight?: number;
    /** Array of available page size options for the dropdown */
    pageSizePresets?: number[];
    /** Array of permission strings required to view this table */
    permissions?: string[];
    /** Enable/disable scrollable mode for the table */
    scrollable?: boolean;
    /** Current selection array (controlled mode) */
    selection?: unknown[] | null;
    /** Additional info text to show in selection bar */
    selectionExternalInfo?: string;
    /** Selection mode: 'single', 'multiple', or null for no selection */
    selectionMode?: SelectionMode;
    /** Initial sort direction: 'asc' or 'desc' */
    sortDirection?: SortDirection;
    /** Initial field name to sort by */
    sortField?: string | null;

    /** Field name to use as the value key for selection (falls back to dataKey) */
    valueKey?: string;

    /** Whether to disable checkbox selection column */
    withoutCheckBoxSelection?: boolean;
  }>(),
  {
    columnControls: () => [],
    columnControlsKey: undefined,
    extraFilters: () => ({}),
    hidePaginationControls: false,
    initialPage: 1,
    initialPageSize: 25,
    maxHeight: undefined,
    pageSizePresets: () => [10, 15, 20, 25, 50, 100],
    permissions: undefined,
    scrollable: true,
    selectionExternalInfo: "",
    selectionMode: null,
    sortDirection: "desc",
    sortField: null,
  }
);

const { currentPageReportTemplate } = useCurrentPageReportTemplate();
const dataTablePreferences = useDataTablePreferencesStore();

dataTablePreferences.hydrate();

const columnStateKey = computed(() => {
  if (props.columnControlsKey) return props.columnControlsKey;
  return "app-data-table-v1";
});

const authStore = useAuthStore();
const hasPermission = computed(() => {
  if (!props.permissions || props.permissions.length === 0) return true;
  return props.permissions.some((perm) => authStore.permissions.includes(perm));
});

type RowClickPayload = {
  data: unknown;
  originalEvent: Event | null;
};

const emit = defineEmits<{
  (e: "update:selection", value: unknown[]): void;
  (e: "row-click", value: RowClickPayload): void;
  (
    e: "loaded",
    value: { items: unknown[]; pagination: PaginationInfo | null }
  ): void;
  (e: "error", error: unknown): void;
  (
    e: "toggle-column-visibility",
    value: { hidden: boolean; key: string }
  ): void;
  (e: "toggle-column-freeze", value: { frozen: boolean; key: string }): void;
}>();

const items = ref<unknown[]>([]);
const pagination = ref<PaginationInfo | null>(null);
const loading = ref(false);
const minTableHeight = 320;
const tableRef = ref<ComponentPublicInstance | null>(null);
const resolvedPageSizePresets = computed(() =>
  normalizePageSizePresets(props.pageSizePresets)
);
const pageSizeBounds = computed(() =>
  resolvePageSizeBounds(resolvedPageSizePresets.value)
);
const tableState = ref({
  page: props.initialPage,
  rows: props.initialPageSize,
  sortDirection: props.sortDirection,
  sortField: props.sortField,
});
const jumpPage = ref(tableState.value.page);
const pageSizeInput = ref(tableState.value.rows);
const pageSizePreset = ref<number | "custom">(
  resolvedPageSizePresets.value.includes(tableState.value.rows)
    ? tableState.value.rows
    : "custom"
);
const minPageSize = computed(() => pageSizeBounds.value.min);
const maxPageSize = computed(() => pageSizeBounds.value.max);

const getTableRoot = () => {
  const root =
    tableRef.value && "$el" in tableRef.value ? tableRef.value.$el : null;
  return (root as HTMLElement | null) ?? null;
};

const getScrollContainer = () => {
  const root = getTableRoot();
  if (!root) return null;
  return root.querySelector(
    ".p-datatable-scrollable-body"
  ) as HTMLElement | null;
};

const resetScrollPosition = () => {
  const scroller = getScrollContainer() ?? getTableRoot();
  if (scroller) {
    scroller.scrollTop = 0;
  }
};

const {
  clearSelection: clearSelectionFromComposable,
  internalSelection,
  syncSelectionWithData: syncSelectionFromComposable,
} = useTableSelection({
  dataKey: props.dataKey,
  items,
  selectionMode: props.selectionMode,
  valueKey: props.valueKey,
});

const totalRecords = computed(() => pagination.value?.total ?? 0);
const totalPages = computed(() => pagination.value?.total_pages ?? 1);
const pageSizeInputMax = computed(() => {
  const currentMaxPage = totalPages.value || 1;
  return Math.max(
    minPageSize.value,
    Math.min(maxPageSize.value, currentMaxPage)
  );
});
const pageSizeOptions = computed(() => [
  ...resolvedPageSizePresets.value.map((value) => ({
    label: `${value}`,
    value,
  })),
  { label: "Custom", value: "custom" as const },
]);

const tableLoading = computed(() => loading.value);
const paginatorFirst = computed(
  () => (tableState.value.page - 1) * tableState.value.rows
);
const scrollHeightValue = computed(() =>
  !props.scrollable
    ? undefined
    : typeof props.maxHeight === "number" && props.maxHeight > 0
      ? `${props.maxHeight}px`
      : "flex"
);
const tableStyle = computed(() => ({
  maxHeight:
    typeof props.maxHeight === "number" && props.maxHeight > 0
      ? `${props.maxHeight}px`
      : undefined,
  minHeight: `${minTableHeight}px`,
}));

const isCurrentPageFullySelected = computed(() => {
  if (props.selectionMode !== "multiple") return false;
  if (!items.value.length) return false;

  const fieldKey = selectionFieldKey.value;
  if (fieldKey) {
    const selectedIds = new Set(
      (selection.value || [])
        .map((item) => getObjectPathValue(item, fieldKey))
        .filter(
          (value): value is string | number =>
            typeof value === "string" || typeof value === "number"
        )
        .map(String)
    );

    return items.value.every((item) => {
      const id = getObjectPathValue(item, fieldKey);
      if (id === undefined || id === null) return false;
      return selectedIds.has(String(id));
    });
  }

  return items.value.every((item) => selection.value.includes(item));
});

const selection = computed<unknown[]>({
  get: () =>
    props.selection !== undefined && props.selection !== null
      ? (props.selection ?? [])
      : internalSelection.value,
  set: (val) => {
    const normalized = val ?? [];
    if (props.selection === undefined) {
      internalSelection.value = normalized;
    }
    emit("update:selection", normalized);
  },
});

const selectionCountLabel = computed(
  () => `${selection.value.length} selected`
);
const clearLabel = "Clear";
const refreshLabel = "Refresh";
const columnMenuTitle = "Columns";
const columnMenuSubtitle = "Show, hide, or freeze columns";
const columnMenuCount = computed(
  () => hiddenColumns.value.size + frozenColumns.value.size
);
const pageSizeLabel = "Page size";
const goToLabel = "Go to";
const setLabel = "Set";
const emptyMessage = "No data found";
const loadingMessage = "Loading data...";
const noAccessMessage = "You do not have access to view this table";

const selectionStorageKey = computed(() =>
  props.columnControlsKey ? `${props.columnControlsKey}selection-field` : null
);
const selectionFieldKey = computed(() => props.valueKey || props.dataKey || "");
const isSelectionPrimitive = (value: unknown): value is string | number =>
  typeof value === "string" || typeof value === "number";

const persistSelectionToStorage = (nextSelection: unknown[]) => {
  if (typeof window === "undefined") return;
  const storageKey = selectionStorageKey.value;
  const fieldKey = selectionFieldKey.value;
  if (!storageKey || !fieldKey) return;

  const compactSelectionValues = Array.from(
    new Map(
      (Array.isArray(nextSelection) ? nextSelection : [])
        .map((item) => {
          if (isSelectionPrimitive(item)) return item;
          return getObjectPathValue(item, fieldKey);
        })
        .filter(isSelectionPrimitive)
        .map((value) => [String(value), value])
    ).values()
  );

  if (compactSelectionValues.length === 0) {
    window.sessionStorage.removeItem(storageKey);
    return;
  }

  window.sessionStorage.setItem(
    storageKey,
    JSON.stringify(compactSelectionValues)
  );
};

const hydrateSelectionFromStorage = () => {
  if (typeof window === "undefined") return;
  const storageKey = selectionStorageKey.value;
  const fieldKey = selectionFieldKey.value;
  if (!storageKey || !fieldKey) return;

  const raw = window.sessionStorage.getItem(storageKey);
  if (!raw) return;

  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed) || parsed.length === 0) return;

    const compactSelectionValues = Array.from(
      new Map(
        parsed
          .map((value) => {
            if (isSelectionPrimitive(value)) {
              return value;
            }
            if (
              value !== null &&
              typeof value === "object" &&
              !Array.isArray(value)
            ) {
              const selectionValue = getObjectPathValue(value, fieldKey);
              if (isSelectionPrimitive(selectionValue)) {
                return selectionValue;
              }
            }
            return null;
          })
          .filter((value): value is string | number => value !== null)
          .map((value) => [String(value), value])
      ).values()
    );

    if (compactSelectionValues.length) {
      const placeholders = compactSelectionValues.map((value) =>
        createObjectPathValue(fieldKey, value)
      );
      internalSelection.value = placeholders;
      selection.value = placeholders;
    }
  } catch {
    window.sessionStorage.removeItem(storageKey);
  }
};

watch(
  selection,
  (nextSelection) => {
    persistSelectionToStorage(nextSelection ?? []);
  },
  { deep: true }
);

const resetAndLoad = (options?: {
  clearSelection?: boolean;
  keepPage?: boolean;
}) => {
  if (!options?.keepPage) {
    tableState.value.page = 1;
  }
  jumpPage.value = tableState.value.page;
  // items.value = [];
  // pagination.value = null;
  resetScrollPosition();

  if (options?.clearSelection !== false) {
    clearSelection();
  }

  loadData();
};

const columnMenuPanel = ref<InstanceType<typeof Popover> | null>(null);
const hiddenColumns = ref<Set<string>>(new Set());
const frozenColumns = ref<Set<string>>(new Set());

const persistColumnState = (
  hiddenSet: Set<string> = hiddenColumns.value,
  frozenSet: Set<string> = frozenColumns.value
) => {
  const key = columnStateKey.value;
  if (!key) return;
  dataTablePreferences.setColumnState(key, {
    frozen: Array.from(frozenSet),
    hidden: Array.from(hiddenSet),
  });
};

const applyColumnPreferences = (cols: ColumnControl[] = []) => {
  const availableKeys = new Set(cols.map((col) => col.key));
  const storedState = dataTablePreferences.getColumnState(columnStateKey.value);

  const hiddenSet = new Set(
    storedState.hidden.filter((key) => availableKeys.has(key))
  );
  const frozenSet = new Set(
    storedState.frozen.filter((key) => availableKeys.has(key))
  );

  cols.forEach((col) => {
    if (col.hidden && !hiddenSet.has(col.key)) {
      hiddenSet.add(col.key);
    }
    if (col.frozen && !frozenSet.has(col.key)) {
      frozenSet.add(col.key);
    }
  });

  hiddenColumns.value = hiddenSet;
  frozenColumns.value = frozenSet;
  persistColumnState(hiddenSet, frozenSet);
};

watch(
  () => props.columnControls,
  (cols) => {
    applyColumnPreferences(cols || []);
  },
  { deep: true, immediate: true }
);

watch(columnStateKey, () => {
  applyColumnPreferences(props.columnControls || []);
});

const columnMenuItems = computed(() => props.columnControls || []);

const toggleColumnVisibility = (key: string) => {
  const next = new Set(hiddenColumns.value);
  if (next.has(key)) next.delete(key);
  else next.add(key);
  hiddenColumns.value = next;
  emit("toggle-column-visibility", { hidden: next.has(key), key });
  persistColumnState(next, frozenColumns.value);
};

const toggleColumnFreeze = (key: string, force?: boolean) => {
  const next = new Set(frozenColumns.value);
  const willFreeze = typeof force === "boolean" ? force : !next.has(key);
  if (willFreeze) next.add(key);
  else next.delete(key);
  frozenColumns.value = next;
  emit("toggle-column-freeze", { frozen: next.has(key), key });
  persistColumnState(hiddenColumns.value, next);
};

const toggleColumnMenu = (event: MouseEvent) => {
  columnMenuPanel.value?.toggle(event);
};

const resetColumnStates = () => {
  hiddenColumns.value = new Set();
  frozenColumns.value = new Set();
  emit("toggle-column-visibility", { hidden: false, key: "*" });
  emit("toggle-column-freeze", { frozen: false, key: "*" });
  persistColumnState();
  columnMenuPanel.value?.hide();
};

const refreshTable = () => {
  loadData();
};

const closeColumnMenu = () => {
  columnMenuPanel.value?.hide();
};

const clearSelection = () => {
  clearSelectionFromComposable();
  selection.value = [];
};

const syncSelectionWithData = () => {
  internalSelection.value = Array.isArray(selection.value)
    ? [...selection.value]
    : [];
  syncSelectionFromComposable();
  selection.value = internalSelection.value;
};

const buildParams = (page: number): LoadParams => ({
  extraFilters: props.extraFilters ?? null,
  page,
  pageSize: tableState.value.rows,
  sortDirection: tableState.value.sortDirection,
  sortField: tableState.value.sortField,
});

let latestLoadRequestId = 0;
const hasTriggeredInitialLoad = ref(false);
const createExtraFiltersSignature = (
  value: Record<string, unknown> | null | undefined
) => {
  try {
    return JSON.stringify(value ?? null);
  } catch {
    return String(value ?? null);
  }
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
const triggerInitialLoad = () => {
  if (hasTriggeredInitialLoad.value) return;
  if (!areExtraFiltersReady(props.extraFilters)) return;

  hasTriggeredInitialLoad.value = true;
  lastExtraFiltersSignature = createExtraFiltersSignature(props.extraFilters);
  loadData();
};

const initEventBusSubscriptions = () => {
  if (
    !props.eventBusNamesForRefresh ||
    props.eventBusNamesForRefresh.length === 0
  )
    return;
  const bus = getGlobalEventBus();

  for (const eventName of props.eventBusNamesForRefresh) {
    bus.subscribe(eventName, () => {
      loadData({
        withoutLoading: true,
      });
    });
  }
};

const loadData = async (options?: {
  page?: number;
  withoutLoading?: boolean;
}) => {
  const targetPage = options?.page ?? tableState.value.page;
  const requestId = ++latestLoadRequestId;
  const shouldShowLoading = !options?.withoutLoading;

  if (shouldShowLoading) {
    loading.value = true;
  }

  try {
    const { items: nextItems, pagination: pageInfo } = await props.loadFn(
      buildParams(targetPage)
    );
    if (requestId !== latestLoadRequestId) {
      return;
    }
    pagination.value = pageInfo ?? null;
    tableState.value.page = pageInfo?.page ?? targetPage;
    items.value = nextItems;

    // Sync selection with current page items
    syncSelectionWithData();

    emit("loaded", { items: items.value, pagination: pagination.value });
  } catch (err) {
    if (requestId === latestLoadRequestId) {
      emit("error", err);
    }
  } finally {
    if (requestId === latestLoadRequestId) {
      if (shouldShowLoading) {
        loading.value = false;
      }
    }
  }
};

let extraFiltersTimer: number | undefined;
let lastExtraFiltersSignature = createExtraFiltersSignature(props.extraFilters);
watch(
  () => props.extraFilters,
  (nextFilters) => {
    if (!areExtraFiltersReady(nextFilters)) return;
    if (!hasTriggeredInitialLoad.value) {
      triggerInitialLoad();
      return;
    }

    const nextSignature = createExtraFiltersSignature(nextFilters);
    if (nextSignature === lastExtraFiltersSignature) return;
    lastExtraFiltersSignature = nextSignature;

    if (extraFiltersTimer) window.clearTimeout(extraFiltersTimer);
    extraFiltersTimer = window.setTimeout(() => {
      jumpPage.value = 1;
      resetAndLoad({ clearSelection: true });
    }, DEBOUNCE_DELAY);
  },
  { deep: true }
);

watch(
  () => areExtraFiltersReady(props.extraFilters),
  (ready) => {
    if (ready) {
      triggerInitialLoad();
    }
  }
);

watch(
  () => [props.sortField, props.sortDirection],
  ([field, direction]) => {
    const normalizedField = field ?? null;
    const normalizedDirection: SortDirection =
      direction === "asc" ? "asc" : "desc";
    const changed =
      normalizedField !== tableState.value.sortField ||
      normalizedDirection !== tableState.value.sortDirection;
    tableState.value.sortField = normalizedField;
    tableState.value.sortDirection = normalizedDirection;
    if (changed) {
      resetAndLoad({ clearSelection: false });
    }
  }
);

const onPageChange = (event: PageState) => {
  if (extraFiltersTimer) {
    window.clearTimeout(extraFiltersTimer);
    extraFiltersTimer = undefined;
  }
  const nextPage = (event.page ?? 0) + 1;
  const rows = event.rows ?? tableState.value.rows;
  tableState.value.page = nextPage;
  tableState.value.rows = rows;
  pageSizeInput.value = rows;
  loadData({ page: nextPage });
};

watch(
  () => pagination.value?.page,
  (val) => {
    if (typeof val === "number") {
      jumpPage.value = val;
    }
  }
);

const onJumpPage = (value?: number) => {
  const requested = typeof value === "number" ? value : jumpPage.value;
  const target = Math.max(
    1,
    Math.min(requested || 1, pagination.value?.total_pages ?? 1)
  );
  if (target === tableState.value.page) return;
  jumpPage.value = target;
  tableState.value.page = target;
  items.value = [];
  resetScrollPosition();
  loadData({ page: tableState.value.page });
};

const applyCustomPageSize = () => {
  pageSizeInput.value = Math.max(
    minPageSize.value,
    Math.min(
      pageSizeInput.value || tableState.value.rows,
      pageSizeInputMax.value
    )
  );
  pageSizePreset.value = "custom";
  if (pageSizeInput.value === tableState.value.rows) return;
  tableState.value.rows = pageSizeInput.value;
  tableState.value.page = 1;
  jumpPage.value = 1;
  items.value = [];
  pagination.value = null;
  resetScrollPosition();
  loadData();
};

const onSelectPageSizePreset = (value: number | "custom") => {
  pageSizePreset.value = value;
  if (value === "custom") {
    pageSizeInput.value = tableState.value.rows;
    return;
  }
  if (value === tableState.value.rows) return;
  pageSizeInput.value = value;
  tableState.value.rows = value;
  tableState.value.page = 1;
  jumpPage.value = 1;
  items.value = [];
  pagination.value = null;
  resetScrollPosition();
  loadData();
};

watch(
  () => tableState.value.rows,
  (val) => {
    pageSizeInput.value = val;
    pageSizePreset.value = resolvedPageSizePresets.value.includes(val)
      ? val
      : "custom";
  }
);

const onSort = (event: DataTableSortEvent) => {
  const sortField =
    typeof event.sortField === "string" ? event.sortField : null;
  const direction = event.sortOrder === 1 ? "asc" : "desc";
  tableState.value.sortField = sortField;
  tableState.value.sortDirection = direction;
  resetAndLoad({ clearSelection: false });
};

const onRowClick = (event: DataTableRowClickEvent) => {
  emit("row-click", { data: event.data, originalEvent: event.originalEvent });
};

const forceSelectAllCurrentItems = () => {
  if (props.selectionMode !== "multiple") return;
  const mergedSelection = [...selection.value, ...items.value];
  const fieldKey = selectionFieldKey.value;

  if (!fieldKey) {
    selection.value = Array.from(new Set(mergedSelection));
    return;
  }

  const mergedById = new Map<string, unknown>();
  const itemsWithoutId: unknown[] = [];
  const seenItemsWithoutId = new Set<unknown>();

  for (const item of mergedSelection) {
    const id = getObjectPathValue(item, fieldKey);
    if (id === undefined || id === null) {
      if (!seenItemsWithoutId.has(item)) {
        seenItemsWithoutId.add(item);
        itemsWithoutId.push(item);
      }
      continue;
    }
    mergedById.set(String(id), item);
  }

  selection.value = [...itemsWithoutId, ...Array.from(mergedById.values())];
};

const removeCurrentPageItemsFromSelection = () => {
  if (props.selectionMode !== "multiple") return;
  if (!selection.value.length || !items.value.length) return;

  const fieldKey = selectionFieldKey.value;

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

  if (!currentPageIds.size) return;

  selection.value = selection.value.filter((selected) => {
    const id = getObjectPathValue(selected, fieldKey);
    if (id === undefined || id === null) return true;
    return !currentPageIds.has(String(id));
  });
};

const onSelectAllChange = (event: DataTableSelectAllChangeEvent) => {
  if (!event.checked) {
    removeCurrentPageItemsFromSelection();
    return;
  }
  forceSelectAllCurrentItems();
};

const tableProps = computed(() => ({
  dataKey: props.dataKey,
  lazy: true,
  loading: tableLoading.value,
  resizableColumns: true,
  responsiveLayout: "stack",
  rowHover: true,
  scrollable: props.scrollable,
  scrollHeight: scrollHeightValue.value,
  selectAll:
    props.selectionMode === "multiple"
      ? isCurrentPageFullySelected.value
      : null,
  selection: props.selectionMode ? selection.value : undefined,
  showGridlines: true,
  size: "small",
  sortField: tableState.value.sortField ?? undefined,
  sortOrder: tableState.value.sortDirection === "asc" ? 1 : -1,
  value: items.value,
}));

onMounted(() => {
  initEventBusSubscriptions();
  hydrateSelectionFromStorage();
  triggerInitialLoad();
});

onBeforeUnmount(() => {
  if (extraFiltersTimer) {
    window.clearTimeout(extraFiltersTimer);
  }
});

// Expose methods for parent component
defineExpose({
  clearSelection,
  loadData,
  resetAndLoad,
});
</script>

<template>
  <section
    class="app-data-table flex h-full min-h-0 min-w-0 flex-col overflow-hidden text-body"
  >
    <div v-if="hasPermission" class="data-table-card">
      <div
        class="data-table__toolbar flex flex-nowrap items-center justify-between gap-3 text-table text-slate-700"
      >
        <Paginator
          :key="currentPageReportTemplate"
          :rows="tableState.rows"
          :total-records="totalRecords"
          :first="paginatorFirst"
          :always-show="true"
          template="PageLinks CurrentPageReport"
          :current-page-report-template="currentPageReportTemplate"
          @page="onPageChange"
        />
        <div class="data-table__toolbar-actions">
          <div
            v-if="!props.hidePaginationControls"
            class="data-table__control-group hidden md:flex"
          >
            <label for="page-size-select" class="data-table__control-label">
              {{ pageSizeLabel }}
            </label>
            <div class="data-table__control-field">
              <Select
                :model-value="pageSizePreset"
                input-id="page-size-select"
                :options="pageSizeOptions"
                size="small"
                option-label="label"
                option-value="value"
                class="page-size-select text-select-item select-no-arrow"
                @update:model-value="onSelectPageSizePreset"
              />
            </div>
            <div
              v-if="pageSizePreset === 'custom'"
              class="flex items-center gap-1.5"
            >
              <InputNumber
                v-model="pageSizeInput"
                input-class="!w-14 !h-8 text-form"
                :min="minPageSize"
                :max="pageSizeInputMax"
                size="small"
                :use-grouping="false"
                @blur="applyCustomPageSize"
              />
              <button
                type="button"
                class="data-table__small-button"
                @click="applyCustomPageSize"
              >
                {{ setLabel }}
              </button>
            </div>
          </div>
          <form
            v-if="!props.hidePaginationControls"
            class="data-table__control-group hidden md:flex"
            @submit.prevent="onJumpPage()"
          >
            <label for="page-jump-input" class="data-table__control-label">
              {{ goToLabel }}
            </label>
            <InputNumber
              v-model="jumpPage"
              input-id="page-jump-input"
              input-class="!w-14 !h-8 text-form"
              size="small"
              :min="1"
              :max="totalPages || 1"
              :use-grouping="false"
              @blur="onJumpPage()"
            />
            <button type="submit" class="data-table__small-button">
              {{ setLabel }}
            </button>
          </form>
          <div class="data-table__icon-group relative">
            <button
              v-if="columnControls.length > 0"
              type="button"
              class="data-table__icon-button relative"
              :title="columnMenuTitle"
              :disabled="!columnMenuItems.length"
              @click.stop="toggleColumnMenu"
            >
              <span class="pi pi-sliders-h text-sm" />
              <span
                v-if="columnMenuCount"
                class="absolute -right-1 -top-1 grid h-4 min-w-4 place-items-center rounded-full bg-primary-600 px-1 text-[10px] font-bold text-white shadow-sm"
              >
                {{ columnMenuCount }}
              </span>
            </button>
            <button
              type="button"
              class="data-table__icon-button"
              :aria-label="refreshLabel"
              :title="refreshLabel"
              :disabled="tableLoading"
              @click="refreshTable"
            >
              <span
                :class="[
                  'pi pi-refresh text-sm',
                  tableLoading ? 'animate-spin' : '',
                ]"
              />
            </button>

            <Popover
              ref="columnMenuPanel"
              :dismissable="true"
              :breakpoints="{ '640px': '92vw', '480px': '94vw' }"
              append-to="body"
              style="width: 18rem; max-width: 92vw"
            >
              <div
                class="flex items-start justify-between gap-2 border-b border-slate-100 px-3 py-2"
              >
                <div>
                  <p class="text-sm font-semibold text-slate-800">
                    {{ columnMenuTitle }}
                  </p>
                  <p class="text-xs text-slate-500">
                    {{ columnMenuSubtitle }}
                  </p>
                </div>
                <button
                  type="button"
                  class="grid h-8 w-8 place-items-center rounded-full text-slate-500 transition hover:bg-slate-100 hover:text-primary-700"
                  aria-label="Close"
                  @click.stop="closeColumnMenu"
                >
                  <span class="pi pi-times text-sm" />
                </button>
              </div>
              <div
                class="flex items-center justify-end gap-2 border-b border-slate-100 px-3 py-2"
              >
                <button
                  type="button"
                  class="rounded-md border border-slate-200 px-3 py-1 text-xs font-semibold text-slate-600 transition hover:border-primary-200 hover:text-primary-700"
                  @click.stop="resetColumnStates"
                >
                  Reset
                </button>
              </div>
              <div class="max-h-64 overflow-auto px-2 py-2">
                <div
                  v-for="col in columnMenuItems"
                  :key="col.key"
                  class="data-table__column-menu-row flex items-center justify-between rounded-lg px-2 py-2 text-sm text-slate-800"
                >
                  <span class="truncate">{{ col.label }}</span>
                  <div class="flex items-center gap-1">
                    <button
                      type="button"
                      class="data-table__column-menu-button"
                      :title="hiddenColumns.has(col.key) ? 'Show' : 'Hide'"
                      @click.stop="toggleColumnVisibility(col.key)"
                    >
                      <span
                        :class="
                          hiddenColumns.has(col.key)
                            ? 'pi pi-eye-slash'
                            : 'pi pi-eye'
                        "
                      />
                    </button>
                    <button
                      type="button"
                      class="data-table__column-menu-button"
                      :title="
                        frozenColumns.has(col.key) ? 'Unfreeze' : 'Freeze'
                      "
                      @click.stop="toggleColumnFreeze(col.key)"
                    >
                      <span
                        :class="
                          frozenColumns.has(col.key)
                            ? 'pi pi-lock'
                            : 'pi pi-lock-open'
                        "
                      />
                    </button>
                  </div>
                </div>
              </div>
            </Popover>
          </div>
        </div>
      </div>

      <div
        v-if="selectionMode && selection?.length"
        class="data-table__selection flex flex-wrap items-center justify-between gap-3 text-table font-semibold text-slate-700"
      >
        <slot
          name="selection"
          :selection="selection"
          :clear-selection="clearSelection"
        >
          <div class="flex items-center gap-2">
            <span>{{ selectionCountLabel }}</span>
            <span v-if="selectionExternalInfo" class="text-slate-600">
              {{ selectionExternalInfo }}
            </span>
          </div>
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-button font-semibold text-slate-700 shadow-sm transition hover:border-slate-300 hover:text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-200 focus:ring-offset-0"
            @click="clearSelection"
          >
            <span>{{ clearLabel }}</span>
          </button>
        </slot>
      </div>

      <div class="data-table__body flex-1 min-h-0">
        <div
          :class="[
            props.scrollable ? 'data-table__inner' : 'data-table__outer-scroll',
            'h-full',
            'min-h-0',
          ]"
        >
          <DataTable
            ref="tableRef"
            v-bind="tableProps"
            v-model:selection="selection"
            column-resize-mode="fit"
            class="data-table__grid h-full"
            :style="tableStyle"
            @row-click="onRowClick"
            @select-all-change="onSelectAllChange"
            @sort="onSort"
          >
            <template #empty>
              <div class="py-6 text-center text-slate-500">
                {{ emptyMessage }}
              </div>
            </template>
            <template #loading>
              <div
                class="data-table__loading flex flex-col items-center justify-center gap-3 text-slate-600"
              >
                <ProgressSpinner
                  stroke-width="6"
                  style="width: 2.75rem; height: 2.75rem"
                />
                <span class="text-sm font-semibold">{{ loadingMessage }}</span>
              </div>
            </template>
            <template #footer>
              <slot name="footer" />
            </template>

            <Column
              v-if="selectionMode && !withoutCheckBoxSelection"
              :selection-mode="selectionMode"
              class="data-table__selection-column"
              :style="{ width: '3rem', minWidth: '3rem', maxWidth: '3rem' }"
            />

            <slot
              :items="items"
              :loading="tableLoading"
              :selection="selection"
              :table-props="tableProps"
              :on-row-click="onRowClick"
              :hidden-columns="hiddenColumns"
              :frozen-columns="frozenColumns"
              :toggle-column-freeze="toggleColumnFreeze"
              :toggle-column-visibility="toggleColumnVisibility"
            />
          </DataTable>
        </div>
      </div>
    </div>
    <div
      v-else
      class="flex items-center justify-center min-h-[200px] text-slate-400 text-lg font-semibold"
    >
      {{ noAccessMessage }}
    </div>
  </section>
</template>

<style scoped>
.select-no-arrow :deep(.p-select-dropdown) {
  display: none;
}

.page-size-container {
  width: auto !important;
  min-width: 5.25rem !important;
}

.page-size-select {
  width: auto !important;
  min-width: 5.25rem !important;
  min-height: 1.75rem;
  height: 1.75rem;
}

.page-size-select :deep(.p-select-label) {
  padding-top: 0.2rem;
  padding-bottom: 0.2rem;
  font-size: calc(var(--font-select) * 0.92);
}

.app-data-table {
  height: 100%;
}

.data-table-card {
  background: #fff;
  border-top: 1px solid #e2e8f0;
  margin-top: 0;
  overflow: visible;
  min-width: 0;
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.data-table__toolbar {
  padding: 0.35rem 0.85rem;
  min-height: 2.65rem;
  background: #fff;
}

.data-table__toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
  color: #475569;
}

.data-table__control-group,
.data-table__icon-group {
  align-items: center;
  gap: 0.4rem;
  min-height: 2.1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  box-shadow:
    0 1px 2px rgba(15, 23, 42, 0.04),
    inset 0 1px 0 rgba(255, 255, 255, 0.85);
  padding: 0.18rem 0.28rem 0.18rem 0.6rem;
}

.data-table__icon-group {
  display: flex;
  gap: 0.25rem;
  padding: 0.2rem;
}

.data-table__control-label {
  white-space: nowrap;
  font-size: calc(var(--font-table) * 0.88);
  font-weight: 600;
  line-height: 1;
  color: #64748b;
}

.data-table__control-field {
  display: flex;
  min-width: 0;
}

.data-table__small-button {
  display: inline-flex;
  min-height: 1.7rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.55rem;
  border: 1px solid #e2e8f0;
  background: #fff;
  padding: 0 0.6rem;
  font-size: calc(var(--font-button) * 0.88);
  font-weight: 800;
  line-height: 1;
  color: #475569;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);
  transition:
    transform 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease,
    background-color 0.15s ease,
    box-shadow 0.15s ease;
}

.data-table__small-button:hover {
  transform: translateY(-1px);
  border-color: #fecaca;
  background: #fff5f5;
  color: #b91c1c;
  box-shadow: 0 5px 12px rgba(220, 38, 38, 0.1);
}

.data-table__icon-button {
  position: relative;
  display: inline-flex;
  height: 1.9rem;
  width: 1.9rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 0.65rem;
  color: #64748b;
  background: transparent;
  transition:
    transform 0.15s ease,
    background-color 0.15s ease,
    color 0.15s ease,
    box-shadow 0.15s ease;
}

.data-table__icon-button:hover {
  transform: translateY(-1px);
  background: #fff;
  color: #b91c1c;
  box-shadow: 0 7px 18px rgba(15, 23, 42, 0.09);
}

.data-table__icon-button:active {
  transform: translateY(0) scale(0.97);
}

.data-table__icon-button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
  transform: none;
  box-shadow: none;
}

.data-table__toolbar :deep(.p-paginator) {
  min-height: 2rem;
  padding: 0;
  gap: 0.15rem;
}

.data-table__toolbar :deep(.p-paginator-page),
.data-table__toolbar :deep(.p-paginator-first),
.data-table__toolbar :deep(.p-paginator-prev),
.data-table__toolbar :deep(.p-paginator-next),
.data-table__toolbar :deep(.p-paginator-last) {
  height: 1.75rem;
  min-width: 1.75rem;
  padding: 0 0.35rem;
  border-radius: 0.375rem;
  font-size: calc(var(--font-table) * 0.92);
}

.data-table__toolbar :deep(.p-paginator-page.p-paginator-page-selected) {
  background: #dc2626;
  color: #fff;
  box-shadow: 0 7px 16px rgba(220, 38, 38, 0.18);
}

.data-table__toolbar :deep(.p-paginator-current) {
  margin-left: 0.35rem;
  border-radius: 999px;
  background: #f8fafc;
  padding: 0.2rem 0.55rem;
  font-size: calc(var(--font-table) * 0.88);
  font-weight: 700;
  color: #64748b;
}

.data-table__column-menu-row {
  border: 1px solid transparent;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease,
    box-shadow 0.15s ease;
}

.data-table__column-menu-row:hover {
  border-color: #e2e8f0;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.data-table__column-menu-button {
  display: grid;
  height: 2rem;
  width: 2rem;
  place-items: center;
  border-radius: 0.6rem;
  color: #64748b;
  transition:
    transform 0.15s ease,
    background-color 0.15s ease,
    color 0.15s ease;
}

.data-table__column-menu-button:hover {
  transform: translateY(-1px);
  background: #fff1f2;
  color: #b91c1c;
}

.data-table__column-menu-button:active {
  transform: translateY(0) scale(0.97);
}

.data-table__loading {
  padding: 1.5rem 0;
}

.data-table__selection {
  padding: 0.5rem 0.85rem;
  border-bottom: 1px solid #e2e8f0;
  background: #f87171;
}

.data-table__body {
  padding: 0;
  position: relative;
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
}

.data-table__grid {
  height: 100%;
}

.data-table__body :deep(.p-datatable) {
  border: none;
  height: 100%;
}

.data-table__body :deep(.p-datatable-wrapper) {
  border-radius: 0.6rem;
  width: 100%;
  max-width: 100%;
  overflow-x: auto;
  flex: 1;
  min-height: 0;
  height: 100%;
}

.data-table__body :deep(.p-datatable-mask),
.data-table__body :deep(.p-datatable-loading-overlay) {
  background-color: transparent !important;
  backdrop-filter: none !important;
}

.data-table__body :deep(.p-datatable-scrollable-body) {
  min-height: 100%;
}

.data-table__body :deep(.p-datatable-thead > tr > th) {
  background: #f4f4f4;
  text-transform: uppercase;
}

.data-table__body
  :deep(.p-datatable-thead > tr > th[data-p-selection-column="true"]),
.data-table__body
  :deep(.p-datatable-tbody > tr > td[data-p-selection-column="true"]),
.data-table__body :deep(.data-table__selection-column) {
  width: 3rem !important;
  min-width: 3rem !important;
  max-width: 3rem !important;
  padding-left: 0.9rem !important;
  padding-right: 0.5rem !important;
  text-align: center;
}

.data-table__body
  :deep(
    .p-datatable-thead
      > tr
      > th[data-p-selection-column="true"]
      .p-column-header-content
  ) {
  justify-content: center;
}

.data-table__body
  :deep(
    .p-datatable-thead > tr > th[data-p-selection-column="true"] .p-checkbox
  ) {
  margin-left: auto;
  margin-right: auto;
}

.data-table__body
  :deep(
    .p-datatable-tbody > tr > td[data-p-selection-column="true"] .p-checkbox
  ) {
  margin-left: auto;
  margin-right: auto;
  transform: translateX(-0.15rem);
}

/* Frozen column lock indicator */
:deep(.p-datatable-thead > tr > th.p-frozen-column) {
  background: #f4f4f4;
  border-right: 2px solid #cbd5e1 !important;
  position: relative;
}

:deep(.p-datatable-thead > tr > th.p-frozen-column .p-column-header-content) {
  padding-left: 2rem;
  position: relative;
}

:deep(
  .p-datatable-thead > tr > th.p-frozen-column .p-column-header-content
)::before {
  content: "\e923";
  font-family: "primeicons";
  position: absolute;
  left: 0.5rem;
  top: 50%;
  transform: translateY(-50%);
  font-size: 0.8rem;
  color: #64748b;
  font-weight: normal;
}

:deep(.p-datatable-tbody > tr > td.p-frozen-column) {
  background: linear-gradient(90deg, #ffffff 0%, #fafbfc 100%);
  border-right: 1px solid #e2e8f0 !important;
}

.data-table__outer-scroll {
  overflow: auto;
  max-height: 100%;
  min-height: 0;
  -webkit-overflow-scrolling: touch;
}

.data-table__inner {
  height: 100%;
}
</style>
