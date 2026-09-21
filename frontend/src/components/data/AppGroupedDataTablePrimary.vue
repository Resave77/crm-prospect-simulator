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
type RowClassValue = string | Record<string, boolean> | undefined;
type RowGroupMode = "subheader" | "rowspan" | null;

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
    /** Disable custom page size input and only allow preset page sizes */
    disableCustomPageSize?: boolean;
    /** Event bus names to listen for refresh events */
    eventBusNamesForRefresh?: string[];
    /** Additional filters to be passed to the load function for data fetching */
    extraFilters?: Record<string, unknown> | null;
    /** Field name(s) used by PrimeVue row grouping. Supports rowspan mode. */
    groupRowsBy?: string | string[] | null;
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
    /** Persist selection state to storage */
    persistSelection?: boolean;
    /** Row class resolver passed to the underlying PrimeVue DataTable */
    rowClass?: (data: unknown) => RowClassValue;
    /** PrimeVue row grouping mode. */
    rowGroupMode?: RowGroupMode;
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

    swapActionSlotAndPaginationAction?: boolean;

    /** Field name to use as the value key for selection (falls back to dataKey) */
    valueKey?: string;

    /** Whether to disable checkbox selection column */
    withoutCheckBoxSelection?: boolean;
  }>(),
  {
    columnControls: () => [],
    columnControlsKey: undefined,
    disableCustomPageSize: false,
    extraFilters: () => ({}),
    groupRowsBy: null,
    hidePaginationControls: false,
    initialPage: 1,
    initialPageSize: 25,
    maxHeight: undefined,
    pageSizePresets: () => [10, 15, 20, 25, 50, 100],
    permissions: undefined,
    persistSelection: true,
    rowClass: undefined,
    rowGroupMode: null,
    scrollable: true,
    selectionExternalInfo: "",
    selectionMode: null,
    sortDirection: "desc",
    sortField: null,
    swapActionSlotAndPaginationAction: false,
  }
);

const dataTablePreferences = useDataTablePreferencesStore();

const currentPageReportTemplate = computed(() => {
  return "Page {currentPage} of {totalPages}";
});

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
const loading = ref(true);
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
const pageSizeOptions = computed(() => {
  const presetOptions = resolvedPageSizePresets.value.map((value) => ({
    label: `${value}`,
    value,
  }));

  if (props.disableCustomPageSize) {
    return presetOptions;
  }

  return [...presetOptions, { label: "Custom", value: "custom" }];
});
const pageJumpOptions = computed(() =>
  Array.from({ length: totalPages.value }, (_, index) => {
    const value = index + 1;
    return {
      label: `${value}`,
      value,
    };
  })
);

watch(
  () => props.disableCustomPageSize,
  (disabled) => {
    if (!disabled || pageSizePreset.value !== "custom") return;

    const fallbackPageSize = resolvedPageSizePresets.value.includes(
      tableState.value.rows
    )
      ? tableState.value.rows
      : (resolvedPageSizePresets.value[0] ?? props.initialPageSize);

    pageSizePreset.value = fallbackPageSize;
    pageSizeInput.value = fallbackPageSize;
    tableState.value.rows = fallbackPageSize;
  },
  { immediate: true }
);

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
    const normalized = normalizeGroupedSelection(val ?? []);
    if (props.selection === undefined) {
      internalSelection.value = normalized;
    }
    emit("update:selection", normalized);
  },
});

const selectionStorageKey = computed(() =>
  props.persistSelection && props.columnControlsKey
    ? `${props.columnControlsKey}selection-field`
    : null
);
const selectionFieldKey = computed(() => props.valueKey || props.dataKey || "");
const selectionGroupFieldKey = computed(() => {
  if (props.rowGroupMode !== "rowspan") return "";
  if (Array.isArray(props.groupRowsBy)) return props.groupRowsBy[0] || "";

  return props.groupRowsBy || "";
});
const useCustomGroupedSelectionColumn = computed(
  () =>
    props.selectionMode === "multiple" &&
    !props.withoutCheckBoxSelection &&
    !!selectionGroupFieldKey.value &&
    !!selectionFieldKey.value
);
const isSelectionPrimitive = (value: unknown): value is string | number =>
  typeof value === "string" || typeof value === "number";

function getStringSelectionKey(item: unknown) {
  const fieldKey = selectionFieldKey.value;
  if (!fieldKey) return "";
  if (isSelectionPrimitive(item)) return String(item);

  const value = getObjectPathValue(item, fieldKey);

  return isSelectionPrimitive(value) ? String(value) : "";
}

function getStringGroupKey(item: unknown) {
  const fieldKey = selectionGroupFieldKey.value;
  if (!fieldKey) return "";

  const value = getObjectPathValue(item, fieldKey);

  return isSelectionPrimitive(value) ? String(value) : "";
}

function dedupeSelection(source: unknown[]) {
  const fieldKey = selectionFieldKey.value;
  if (!fieldKey) return Array.from(new Set(source));

  const withoutKey: unknown[] = [];
  const byKey = new Map<string, unknown>();

  for (const item of source) {
    const key = getStringSelectionKey(item);
    if (!key) {
      withoutKey.push(item);
      continue;
    }

    byKey.set(key, item);
  }

  return [...withoutKey, ...Array.from(byKey.values())];
}

function normalizeGroupedSelection(nextSelection: unknown[]) {
  if (useCustomGroupedSelectionColumn.value) {
    return dedupeSelection(nextSelection);
  }

  const groupFieldKey = selectionGroupFieldKey.value;
  const fieldKey = selectionFieldKey.value;
  if (props.selectionMode !== "multiple" || !groupFieldKey || !fieldKey) {
    return nextSelection;
  }

  const previousKeys = new Set(selection.value.map(getStringSelectionKey));
  const nextKeys = new Set(nextSelection.map(getStringSelectionKey));
  const addedGroups = new Set<string>();
  const removedGroups = new Set<string>();
  const visibleGroups = new Map<string, unknown[]>();

  for (const item of items.value) {
    const groupKey = getStringGroupKey(item);
    if (!groupKey) continue;

    const groupItems = visibleGroups.get(groupKey);
    if (groupItems) {
      groupItems.push(item);
      continue;
    }

    visibleGroups.set(groupKey, [item]);
  }

  for (const [groupKey, groupItems] of visibleGroups) {
    const representativeKey = getStringSelectionKey(groupItems[0]);
    if (!representativeKey) continue;

    const wasFullySelected = groupItems.every((item) => {
      const key = getStringSelectionKey(item);

      return key && previousKeys.has(key);
    });
    const isRepresentativeSelected = nextKeys.has(representativeKey);

    if (!wasFullySelected && isRepresentativeSelected)
      addedGroups.add(groupKey);
    if (wasFullySelected && !isRepresentativeSelected) {
      removedGroups.add(groupKey);
    }
  }

  if (!addedGroups.size && !removedGroups.size) return nextSelection;

  const normalized = nextSelection.filter((item) => {
    const groupKey = getStringGroupKey(item);

    return !groupKey || !removedGroups.has(groupKey);
  });

  for (const item of items.value) {
    const groupKey = getStringGroupKey(item);
    if (groupKey && addedGroups.has(groupKey)) {
      normalized.push(item);
    }
  }

  return dedupeSelection(normalized);
}

const tableSelection = computed<unknown[]>({
  get: () => {
    if (useCustomGroupedSelectionColumn.value) {
      return selection.value;
    }

    if (
      props.selectionMode !== "multiple" ||
      !selectionGroupFieldKey.value ||
      !selectionFieldKey.value
    ) {
      return selection.value;
    }

    const selectedKeys = new Set(selection.value.map(getStringSelectionKey));
    const visibleKeys = new Set(items.value.map(getStringSelectionKey));
    const visibleGroups = new Map<string, unknown[]>();
    const visualSelection = selection.value.filter((item) => {
      const key = getStringSelectionKey(item);

      return !key || !visibleKeys.has(key);
    });

    for (const item of items.value) {
      const groupKey = getStringGroupKey(item);
      if (!groupKey) continue;

      const groupItems = visibleGroups.get(groupKey);
      if (groupItems) {
        groupItems.push(item);
        continue;
      }

      visibleGroups.set(groupKey, [item]);
    }

    visualSelection.push(
      ...Array.from(visibleGroups.values())
        .filter((groupItems) =>
          groupItems.every((item) =>
            selectedKeys.has(getStringSelectionKey(item))
          )
        )
        .map((groupItems) => groupItems[0])
    );

    return visualSelection;
  },
  set: (value) => {
    selection.value = value;
  },
});

const groupedVisibleItems = computed(() => {
  const groups = new Map<string, unknown[]>();

  for (const item of items.value) {
    const groupKey = getStringGroupKey(item);
    if (!groupKey) continue;

    const groupItems = groups.get(groupKey);
    if (groupItems) {
      groupItems.push(item);
      continue;
    }

    groups.set(groupKey, [item]);
  }

  return groups;
});

const isGroupSelected = (item: unknown) => {
  const groupKey = getStringGroupKey(item);
  if (!groupKey) return false;

  const groupItems = groupedVisibleItems.value.get(groupKey) ?? [];
  return (
    groupItems.length > 0 &&
    groupItems.every((groupItem) => isItemSelected(groupItem))
  );
};

const toggleGroupSelection = (item: unknown) => {
  if (!useCustomGroupedSelectionColumn.value) return;

  const groupKey = getStringGroupKey(item);
  if (!groupKey) return;

  const groupItems = groupedVisibleItems.value.get(groupKey) ?? [];
  if (!groupItems.length) return;

  if (groupItems.every((groupItem) => isItemSelected(groupItem))) {
    const groupSelectionKeys = new Set(
      groupItems
        .map((groupItem) => getStringSelectionKey(groupItem))
        .filter(Boolean)
    );

    selection.value = selection.value.filter((selectedItem) => {
      const selectionKey = getStringSelectionKey(selectedItem);
      return !selectionKey || !groupSelectionKeys.has(selectionKey);
    });
    return;
  }

  selection.value = dedupeSelection([...selection.value, ...groupItems]);
};

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

const isItemSelected = (item: unknown) => {
  if (props.selectionMode !== "multiple" && props.selectionMode !== "single") {
    return false;
  }

  const fieldKey = selectionFieldKey.value;
  if (!fieldKey) {
    return selection.value.includes(item);
  }

  const itemValue = getObjectPathValue(item, fieldKey);
  return selection.value.some((selectedItem) => {
    const selectedValue = getObjectPathValue(selectedItem, fieldKey);
    return selectedValue === itemValue;
  });
};

const toggleItemSelection = (item: unknown) => {
  if (!props.selectionMode) return;

  if (props.selectionMode === "single") {
    selection.value = isItemSelected(item) ? [] : [item];
    return;
  }

  if (isItemSelected(item)) {
    const fieldKey = selectionFieldKey.value;
    if (!fieldKey) {
      selection.value = selection.value.filter(
        (selectedItem) => selectedItem !== item
      );
      return;
    }

    const itemValue = getObjectPathValue(item, fieldKey);
    selection.value = selection.value.filter((selectedItem) => {
      const selectedValue = getObjectPathValue(selectedItem, fieldKey);
      return selectedValue !== itemValue;
    });
    return;
  }

  selection.value = [...selection.value, item];
};

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
  withoutLoading?: boolean;
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

  return loadData({ withoutLoading: options?.withoutLoading });
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

const refreshTable = (options?: { withoutLoading?: boolean }) => {
  return loadData({ withoutLoading: options?.withoutLoading });
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
    if (
      requestId === latestLoadRequestId &&
      (shouldShowLoading || loading.value)
    ) {
      loading.value = false;
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
  groupRowsBy: props.groupRowsBy ?? undefined,
  lazy: true,
  loading: tableLoading.value,
  resizableColumns: true,
  responsiveLayout: "stack",
  rowClass: props.rowClass,
  rowGroupMode: props.rowGroupMode ?? undefined,
  rowHover: true,
  scrollable: props.scrollable,
  scrollHeight: scrollHeightValue.value,
  selectAll:
    props.selectionMode === "multiple" && !useCustomGroupedSelectionColumn.value
      ? isCurrentPageFullySelected.value
      : null,
  selection:
    props.selectionMode && !useCustomGroupedSelectionColumn.value
      ? tableSelection.value
      : undefined,
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
  refreshTable,
  resetAndLoad,
});
</script>

<template>
  <section
    class="app-data-table flex h-full min-h-0 min-w-0 flex-col overflow-hidden text-body"
  >
    <div v-if="hasPermission" class="data-table-card">
      <div
        class="data-table__toolbar flex flex-wrap items-center justify-between gap-3 text-table text-slate-700"
      >
        <div
          v-if="!props.hidePaginationControls"
          class="data-table__toolbar-left"
          :class="
            props.swapActionSlotAndPaginationAction
              ? 'order-2 justify-end'
              : 'order-1 justify-start'
          "
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

          <div v-if="!props.hidePaginationControls" class="data-table__meta">
            <div
              class="data-table__control-stack data-table__control-stack--page-size hidden md:flex"
            >
              <label
                for="page-size-select"
                class="data-table__stack-label text-[12px] font-semibold"
              >
                {{ pageSizeLabel }}
              </label>
              <div class="data-table__page-size-row">
                <div
                  class="data-table__stack-field data-table__stack-field--page-size-select"
                >
                  <Select
                    :model-value="pageSizePreset"
                    input-id="page-size-select"
                    :options="pageSizeOptions"
                    size="small"
                    option-label="label"
                    option-value="value"
                    class="page-jump-select text-select-item select-no-arrow text-[13px]"
                    @update:model-value="onSelectPageSizePreset"
                  />
                </div>

                <div
                  v-if="
                    !props.disableCustomPageSize && pageSizePreset === 'custom'
                  "
                  class="data-table__page-size-custom"
                >
                  <InputNumber
                    v-model="pageSizeInput"
                    input-class="!w-11 !h-[1.65rem] text-form !text-[12px] !px-1 text-center"
                    :min="minPageSize"
                    :max="pageSizeInputMax"
                    size="small"
                    :use-grouping="false"
                    @blur="applyCustomPageSize"
                  />
                  <button
                    type="button"
                    class="data-table__small-button whitespace-nowrap"
                    @click="applyCustomPageSize"
                  >
                    {{ setLabel }}
                  </button>
                </div>
              </div>
            </div>

            <div
              class="data-table__control-stack data-table__control-stack--jump"
            >
              <label
                for="page-jump-input"
                class="data-table__stack-label text-[12px] font-semibold"
              >
                {{ goToLabel }}
              </label>
              <div class="data-table__stack-field">
                <Select
                  :model-value="jumpPage"
                  input-id="page-jump-input"
                  :options="pageJumpOptions"
                  size="small"
                  option-label="label"
                  option-value="value"
                  class="page-jump-select text-select-item select-no-arrow text-[13px]"
                  @update:model-value="onJumpPage"
                />
              </div>
            </div>

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
            <button
              type="button"
              class="data-table__icon-button data-table__refresh-button"
              :aria-label="refreshLabel"
              :title="refreshLabel"
              :disabled="tableLoading"
              @click="() => refreshTable()"
            >
              <span
                :class="[
                  'pi pi-refresh text-sm',
                  tableLoading ? 'animate-spin' : '',
                ]"
              />
            </button>
          </div>
        </div>

        <div
          class="data-table__toolbar-right max-w-full"
          :class="
            props.swapActionSlotAndPaginationAction ? 'order-1' : 'order-2'
          "
        >
          <slot
            name="topbar-right"
            :loading="tableLoading"
            :refresh="refreshTable"
            :page="tableState.page"
            :page-size="tableState.rows"
            :page-size-options="pageSizeOptions"
            :page-size-preset="pageSizePreset"
            :set-page-size-preset="onSelectPageSizePreset"
            :set-page-size="applyCustomPageSize"
            :set-page="onJumpPage"
            :total-pages="totalPages"
          />
        </div>
      </div>

      <div class="data-table__bottom mb-2 w-full min-w-0">
        <slot
          name="bottom"
          :loading="tableLoading"
          :refresh="refreshTable"
          :page="tableState.page"
          :page-size="tableState.rows"
          :page-size-options="pageSizeOptions"
          :page-size-preset="pageSizePreset"
          :set-page-size-preset="onSelectPageSizePreset"
          :set-page-size="applyCustomPageSize"
          :set-page="onJumpPage"
          :total-pages="totalPages"
          :total-records="totalRecords"
        />
      </div>

      <div
        v-if="selectionMode && selection?.length"
        class="data-table__selection flex h-[44px] shrink-0 animate-in items-center justify-between gap-3 slide-in-from-top duration-300"
      >
        <slot
          name="selection"
          :selection="selection"
          :clear-selection="clearSelection"
        >
          <div class="flex min-w-0 items-center gap-2">
            <p
              class="truncate font-['Inter'] text-[13px] font-bold text-[#1d4ed8]"
            >
              {{ selectionCountLabel }}
            </p>
            <span
              v-if="selectionExternalInfo"
              class="truncate font-['Inter'] text-[12px] font-semibold text-[#2563eb]"
            >
              {{ selectionExternalInfo }}
            </span>
          </div>
          <button
            type="button"
            class="inline-flex h-[30px] items-center rounded-[6px] bg-white px-[16px] font-['Inter'] text-[12px] font-semibold text-[#1e293b] shadow-sm transition-colors hover:bg-[#f8fafc] focus:outline-none focus:ring-2 focus:ring-[#bfdbfe] focus:ring-offset-0"
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
            v-model:selection="tableSelection"
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
              :field="
                useCustomGroupedSelectionColumn
                  ? selectionGroupFieldKey
                  : undefined
              "
              :selection-mode="
                useCustomGroupedSelectionColumn ? undefined : selectionMode
              "
              class="data-table__selection-column hidden md:table-cell"
              :style="{ width: '50px', minWidth: '50px', maxWidth: '50px' }"
            >
              <template v-if="useCustomGroupedSelectionColumn" #header>
                <button
                  type="button"
                  class="mx-auto flex size-[20px] items-center justify-center rounded-[4px] border shadow-sm transition-colors hover:border-[#94a3b8]"
                  :class="
                    isCurrentPageFullySelected
                      ? 'border-[#ef4444] bg-[#ef4444] text-white'
                      : 'border-[#cbd5e1] bg-white text-transparent'
                  "
                  :aria-pressed="isCurrentPageFullySelected"
                  @click.stop="
                    isCurrentPageFullySelected
                      ? removeCurrentPageItemsFromSelection()
                      : forceSelectAllCurrentItems()
                  "
                >
                  <span class="pi pi-check text-[10px]" />
                </button>
              </template>

              <template v-if="useCustomGroupedSelectionColumn" #body="{ data }">
                <button
                  type="button"
                  class="mx-auto mt-[4px] flex size-[20px] items-center justify-center rounded-[4px] border shadow-sm transition-colors hover:border-[#94a3b8]"
                  :class="
                    isGroupSelected(data)
                      ? 'border-[#ef4444] bg-[#ef4444] text-white'
                      : 'border-[#cbd5e1] bg-white text-transparent'
                  "
                  :aria-pressed="isGroupSelected(data)"
                  @click.stop="toggleGroupSelection(data)"
                >
                  <span class="pi pi-check text-[10px]" />
                </button>
              </template>
            </Column>

            <slot
              :items="items"
              :loading="tableLoading"
              :selection="selection"
              :is-selected="isItemSelected"
              :is-group-selected="isGroupSelected"
              :toggle-group-selection="toggleGroupSelection"
              :toggle-selection="toggleItemSelection"
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
  min-width: 4.35rem !important;
}

.page-size-select {
  width: 100% !important;
  font-size: 13px;
  min-height: 1.5rem;
  height: 1.5rem;
  padding-top: 0.15rem;
  border: none !important;
  background: transparent !important;
  box-shadow: none !important;
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
  display: grid !important;
  grid-template-columns: minmax(0, max-content) minmax(0, 1fr);
  align-items: center;
  column-gap: 0.75rem;
  row-gap: 0.5rem;
  padding: 0.35rem 0.85rem;
  min-height: 2.65rem;
  background: #fff;
}

.data-table__toolbar-left,
.data-table__toolbar-right {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  color: #475569;
}

.data-table__toolbar-left {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 0.5rem;
  color: #475569;
  justify-self: start;
  min-width: 0;
  max-width: 100%;
}

.data-table__toolbar-right {
  justify-content: flex-end;
  justify-self: stretch;
  margin-left: 0;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
}

.data-table__toolbar-right :deep(> *) {
  min-width: 0;
  max-width: 100%;
}

@media (max-width: 1024px) {
  .data-table__toolbar {
    grid-template-columns: minmax(0, max-content) minmax(0, 1fr);
  }

  .data-table__toolbar-right {
    width: 100%;
  }
}

.data-table__bottom {
  margin-bottom: 1rem;
  margin-top: 0.5rem;
  padding: 0 0.85rem;
}

.data-table__bottom:empty {
  display: none;
}

.data-table__meta {
  /* display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 0.5rem; */
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 0.75rem;
  flex-shrink: 0; /* Cegah kontainer meta disusutkan oleh flexbox induk */
}

.data-table__control-stack,
.data-table__icon-group {
  align-items: center;
  gap: 0.4rem;
  min-height: 2.1rem;
}

.data-table__control-stack {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  min-width: 0;
  padding-top: 0.45rem;
}

.data-table__control-stack--page-size {
  width: fit-content;
  min-width: 3.8rem;
  flex-shrink: 0;
  transition: width 0.2s ease;
}

.data-table__control-stack--jump {
  width: 3.8rem;
}

.data-table__icon-group {
  display: flex;
  gap: 0.25rem;
  padding: 0.2rem;
}

.data-table__stack-label {
  position: absolute;
  top: 0;
  left: 0.5rem;
  z-index: 1;
  background: #fff;
  padding: 0 0.25rem;
  white-space: nowrap;
  line-height: 1;
  color: #64748b;
}

.data-table__stack-field {
  display: flex;
  min-width: 0;
  width: 100%;
  border: 1px solid #e2e8f0;
  border-radius: 0.4rem;
  background: #fff;
  box-shadow:
    0 1px 2px rgba(15, 23, 42, 0.04),
    inset 0 1px 0 rgba(255, 255, 255, 0.85);
  padding: 0.15rem 0.25rem 0.15rem 0.25rem;
}

.data-table__page-size-row {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  width: fit-content;
  flex-wrap: nowrap;
}

.data-table__stack-field--page-size-select {
  width: 4.75rem;
  flex-shrink: 0;
}

.data-table__page-size-custom {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  flex-shrink: 0;
}

.data-table__control-field {
  display: flex;
  min-width: 0;
}

.data-table__stack-field :deep(.p-select),
.data-table__stack-field :deep(.p-inputnumber) {
  width: 100%;
}

.data-table__stack-field :deep(.p-select-label) {
  padding: 0.08rem 0;
  font-size: 13px;
  scale: 0.9;
  line-height: 1.1;
}

.page-jump-select {
  width: 100% !important;
  font-size: 12px;
  min-height: 1.5rem;
  height: 1.5rem;
  padding-top: 0.15rem;
  border: none !important;
  background: transparent !important;
  box-shadow: none !important;
}

.page-jump-select :deep(.p-select-label) {
  padding-top: 0.15rem;
  padding-bottom: 0.15rem;
  font-size: calc(var(--font-select) * 0.84);
}

.data-table__stack-field :deep(.p-inputnumber-input) {
  width: 100%;
  border: 0;
  box-shadow: none;
  background: transparent;
  padding: 0.08rem 0;
  font-size: calc(var(--font-form) * 0.96);
  line-height: 1.1;
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

.data-table__refresh-button {
  margin-left: -0.45rem;
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
  background: #fff1f2;
  border-color: #fff1f2;
  border-radius: 9999px;
  color: #991b1b;
  font-weight: 700;
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
  padding: 0 24px;
  border-bottom: 1px solid #bfdbfe;
  border-top: 1px solid #bfdbfe;
  background: #eff6ff;
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
  font-size: 12px;
  padding-top: 0.3rem;
  padding-bottom: 0.3rem;
}

.data-table__body :deep(.p-datatable-thead > tr > th .p-column-title) {
  font-size: inherit;
  line-height: inherit;
}

.data-table__body
  :deep(.p-datatable-thead > tr > th[data-p-selection-column="true"]),
.data-table__body
  :deep(.p-datatable-tbody > tr > td[data-p-selection-column="true"]),
.data-table__body :deep(.data-table__selection-column) {
  width: 55px !important;
  min-width: 55px !important;
  max-width: 55px !important;
  padding-left: 0 !important;
  padding-right: 0 !important;
  text-align: center;
}

.data-table__body
  :deep(
    .p-datatable-thead
      > tr
      > th[data-p-selection-column="true"]
      .p-column-header-content
  ) {
  align-items: center;
  justify-content: center;
  width: 100%;
}

.data-table__body
  :deep(.p-datatable-tbody > tr > td[data-p-selection-column="true"]) {
  align-items: center;
  display: table-cell;
  text-align: center;
  vertical-align: middle;
}

.data-table__body
  :deep(
    .p-datatable-thead > tr > th[data-p-selection-column="true"] .p-checkbox
  ) {
  display: inline-flex;
  margin-left: 0;
  margin-right: 0;
  transform: none;
}

.data-table__body
  :deep(
    .p-datatable-tbody > tr > td[data-p-selection-column="true"] .p-checkbox
  ) {
  display: inline-flex;
  margin-left: 0;
  margin-right: 0;
  transform: none;
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
