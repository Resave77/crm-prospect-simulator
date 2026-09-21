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
import { type ComponentPublicInstance, computed, ref, watch } from "vue";

import { useCurrentPageReportTemplate } from "@/composables/useCurrentPageReportTemplate";
import isMobile from "@/composables/useIsMobile";
import {
  normalizePageSizePresets,
  resolvePageSizeBounds,
} from "@/composables/useTablePagination";
import { useTableSelection } from "@/composables/useTableSelection";
import { getObjectPathValue } from "@/lib/utils/objectPath";
import { useAuthStore } from "@/stores/authStore";
import { useDataTablePreferencesStore } from "@/stores/dataTablePreferences";

type SortDirection = "asc" | "desc";
type SelectionMode = "single" | "multiple" | null;
type RowClassValue = string | Record<string, boolean> | undefined;

type ColumnControl = {
  frozen?: boolean;
  hidden?: boolean;
  key: string;
  label: string;
};

export interface AppDataTableV4PaginationInfo {
  page: number;
  total: number;
  total_pages: number;
}

const props = withDefaults(
  defineProps<{
    columnControls?: readonly ColumnControl[];
    columnControlsKey?: string;
    dataKey?: string;
    initialPage?: number;
    initialPageSize?: number;
    items: unknown[];
    loading?: boolean;
    maxHeight?: number;
    pageSizePresets?: number[];
    permissions?: string[];
    rowClass?: (data: unknown) => RowClassValue;
    scrollable?: boolean;
    selection?: unknown[] | null;
    selectionExternalInfo?: string;
    selectionMode?: SelectionMode;
    sortDirection?: SortDirection;
    sortField?: string | null;
    valueKey?: string;
    withoutCheckBoxSelection?: boolean;
  }>(),
  {
    columnControls: () => [],
    columnControlsKey: undefined,
    initialPage: 1,
    initialPageSize: 25,
    loading: false,
    maxHeight: undefined,
    pageSizePresets: () => [10, 15, 20, 25, 50, 100],
    permissions: undefined,
    rowClass: undefined,
    scrollable: true,
    selectionExternalInfo: "",
    selectionMode: null,
    sortDirection: "desc",
    sortField: null,
    withoutCheckBoxSelection: false,
  }
);

const emit = defineEmits<{
  (e: "update:selection", value: unknown[]): void;
  (e: "row-click", value: { data: unknown; originalEvent: Event | null }): void;
  (
    e: "loaded",
    value: { items: unknown[]; pagination: AppDataTableV4PaginationInfo | null }
  ): void;
  (e: "error", error: unknown): void;
  (
    e: "toggle-column-visibility",
    value: { hidden: boolean; key: string }
  ): void;
  (e: "toggle-column-freeze", value: { frozen: boolean; key: string }): void;
}>();

const { currentPageReportTemplate } = useCurrentPageReportTemplate();
const dataTablePreferences = useDataTablePreferencesStore();

dataTablePreferences.hydrate();

const authStore = useAuthStore();
const hasPermission = computed(() => {
  if (!props.permissions?.length) return true;
  return props.permissions.some((perm) => authStore.permissions.includes(perm));
});

const columnStateKey = computed(
  () => props.columnControlsKey || "app-data-table-v4"
);
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
  if (scroller) scroller.scrollTop = 0;
};

const localItems = computed(() => props.items ?? []);

const {
  clearSelection: clearSelectionFromComposable,
  internalSelection,
  syncSelectionWithData: syncSelectionFromComposable,
} = useTableSelection({
  dataKey: props.dataKey,
  items: computed(() => paginatedItems.value),
  selectionMode: props.selectionMode,
  valueKey: props.valueKey,
});

const totalRecords = computed(() => localItems.value.length);
const totalPages = computed(() =>
  Math.max(1, Math.ceil(totalRecords.value / tableState.value.rows))
);
const pageSizeInputMax = computed(() =>
  Math.max(minPageSize.value, maxPageSize.value)
);
const pageSizeOptions = computed(() => [
  ...resolvedPageSizePresets.value.map((value) => ({
    label: `${value}`,
    value,
  })),
  { label: "Custom", value: "custom" as const },
]);

const paginatorFirst = computed(
  () => (tableState.value.page - 1) * tableState.value.rows
);
const paginatorTemplate = computed(() =>
  isMobile.value
    ? "PrevPageLink CurrentPageReport NextPageLink"
    : "PageLinks CurrentPageReport"
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
  minHeight: `0px`,
}));

const selection = computed<unknown[]>({
  get: () =>
    props.selection !== undefined && props.selection !== null
      ? (props.selection ?? [])
      : internalSelection.value,
  set: (value) => {
    const normalized = value ?? [];
    if (props.selection === undefined) {
      internalSelection.value = normalized;
    }
    emit("update:selection", normalized);
  },
});

const selectionFieldKey = computed(() => props.valueKey || props.dataKey || "");
const isSelectionPrimitive = (value: unknown): value is string | number =>
  typeof value === "string" || typeof value === "number";
const selectionStorageKey = computed(() =>
  props.columnControlsKey ? `${props.columnControlsKey}selection-field` : null
);

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

  if (!compactSelectionValues.length) {
    window.sessionStorage.removeItem(storageKey);
    return;
  }

  window.sessionStorage.setItem(
    storageKey,
    JSON.stringify(compactSelectionValues)
  );
};

const syncSelectionWithData = () => {
  internalSelection.value = Array.isArray(selection.value)
    ? [...selection.value]
    : [];
  syncSelectionFromComposable();
  selection.value = internalSelection.value;
};

watch(
  selection,
  (nextSelection) => {
    persistSelectionToStorage(nextSelection ?? []);
  },
  { deep: true }
);

const sortedItems = computed(() => {
  const source = [...localItems.value];
  const sortField = tableState.value.sortField;

  if (!sortField) return source;

  const direction = tableState.value.sortDirection === "asc" ? 1 : -1;

  return source.sort((left, right) => {
    const leftValue = getObjectPathValue(left, sortField);
    const rightValue = getObjectPathValue(right, sortField);

    if (leftValue === rightValue) return 0;
    if (leftValue === null || leftValue === undefined || leftValue === "") {
      return 1;
    }
    if (rightValue === null || rightValue === undefined || rightValue === "") {
      return -1;
    }

    if (typeof leftValue === "number" && typeof rightValue === "number") {
      return (leftValue - rightValue) * direction;
    }

    return String(leftValue).localeCompare(String(rightValue)) * direction;
  });
});

const pagination = computed<AppDataTableV4PaginationInfo>(() => ({
  page: tableState.value.page,
  total: totalRecords.value,
  total_pages: totalPages.value,
}));

const paginatedItems = computed(() => {
  const start = (tableState.value.page - 1) * tableState.value.rows;
  return sortedItems.value.slice(start, start + tableState.value.rows);
});

const tableLoading = computed(() => props.loading);
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

const isCurrentPageFullySelected = computed(() => {
  if (props.selectionMode !== "multiple") return false;
  if (!paginatedItems.value.length) return false;

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

    return paginatedItems.value.every((item) => {
      const id = getObjectPathValue(item, fieldKey);
      if (id === undefined || id === null) return false;
      return selectedIds.has(String(id));
    });
  }

  return paginatedItems.value.every((item) => selection.value.includes(item));
});

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

const applyColumnPreferences = (cols: readonly ColumnControl[] = []) => {
  const availableKeys = new Set(cols.map((col) => col.key));
  const storedState = dataTablePreferences.getColumnState(columnStateKey.value);

  const hiddenSet = new Set(
    storedState.hidden.filter((key) => availableKeys.has(key))
  );
  const frozenSet = new Set(
    storedState.frozen.filter((key) => availableKeys.has(key))
  );

  cols.forEach((col) => {
    if (col.hidden && !hiddenSet.has(col.key)) hiddenSet.add(col.key);
    if (col.frozen && !frozenSet.has(col.key)) frozenSet.add(col.key);
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

const closeColumnMenu = () => {
  columnMenuPanel.value?.hide();
};

const clearSelection = () => {
  clearSelectionFromComposable();
  selection.value = [];
};

const emitLoadedState = () => {
  emit("loaded", {
    items: paginatedItems.value,
    pagination: pagination.value,
  });
};

const clampPage = () => {
  const nextPage = Math.max(
    1,
    Math.min(tableState.value.page, totalPages.value)
  );
  if (nextPage !== tableState.value.page) {
    tableState.value.page = nextPage;
  }
  jumpPage.value = tableState.value.page;
};

const loadData = () => {
  clampPage();
  syncSelectionWithData();
  emitLoadedState();
};

const resetAndLoad = (options?: {
  clearSelection?: boolean;
  keepPage?: boolean;
}) => {
  if (!options?.keepPage) {
    tableState.value.page = 1;
  }
  jumpPage.value = tableState.value.page;
  resetScrollPosition();

  if (options?.clearSelection !== false) {
    clearSelection();
  }

  loadData();
};

const refreshTable = () => {
  loadData();
};

const onPageChange = (event: PageState) => {
  const nextPage = (event.page ?? 0) + 1;
  const rows = event.rows ?? tableState.value.rows;
  tableState.value.page = nextPage;
  tableState.value.rows = rows;
  pageSizeInput.value = rows;
  resetScrollPosition();
  loadData();
};

const onJumpPage = (value?: number) => {
  const requested = typeof value === "number" ? value : jumpPage.value;
  const target = Math.max(1, Math.min(requested || 1, totalPages.value));
  if (target === tableState.value.page) return;
  tableState.value.page = target;
  jumpPage.value = target;
  resetScrollPosition();
  loadData();
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
  tableState.value.rows = value;
  tableState.value.page = 1;
  pageSizeInput.value = value;
  jumpPage.value = 1;
  resetScrollPosition();
  loadData();
};

watch(
  () => tableState.value.rows,
  (value) => {
    pageSizeInput.value = value;
    pageSizePreset.value = resolvedPageSizePresets.value.includes(value)
      ? value
      : "custom";
  }
);

const onSort = (event: DataTableSortEvent) => {
  tableState.value.sortField =
    typeof event.sortField === "string" ? event.sortField : null;
  tableState.value.sortDirection = event.sortOrder === 1 ? "asc" : "desc";
  tableState.value.page = 1;
  jumpPage.value = 1;
  resetScrollPosition();
  loadData();
};

const onRowClick = (event: DataTableRowClickEvent) => {
  const target = event.originalEvent?.target;

  if (
    target instanceof globalThis.Element &&
    target.closest(
      [
        "button",
        "a",
        "input",
        "select",
        "textarea",
        "[role='button']",
        "[role='combobox']",
        "[data-stop-row-click='true']",
        ".p-select",
        ".p-inputtext",
        ".p-checkbox",
        ".p-radiobutton",
      ].join(",")
    )
  ) {
    return;
  }

  emit("row-click", { data: event.data, originalEvent: event.originalEvent });
};

const forceSelectAllCurrentItems = () => {
  if (props.selectionMode !== "multiple") return;
  const mergedSelection = [...selection.value, ...paginatedItems.value];
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
  if (!selection.value.length || !paginatedItems.value.length) return;

  const fieldKey = selectionFieldKey.value;

  if (!fieldKey) {
    const currentItems = new Set(paginatedItems.value);
    selection.value = selection.value.filter((item) => !currentItems.has(item));
    return;
  }

  const currentPageIds = new Set(
    paginatedItems.value
      .map((item) => getObjectPathValue(item, fieldKey))
      .filter(
        (value): value is string | number =>
          typeof value === "string" || typeof value === "number"
      )
      .map(String)
  );

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
  loading: tableLoading.value,
  resizableColumns: true,
  responsiveLayout: "stack",
  rowClass: props.rowClass,
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
  value: paginatedItems.value,
}));

watch(
  () => [
    props.items,
    props.loading,
    tableState.value.page,
    tableState.value.rows,
  ],
  () => {
    clampPage();
    syncSelectionWithData();
    emitLoadedState();
  },
  { deep: true, immediate: true }
);

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
        class="data-table__toolbar flex items-start justify-between gap-3 text-table text-slate-700"
      >
        <Paginator
          :key="currentPageReportTemplate"
          :rows="tableState.rows"
          :total-records="totalRecords"
          :first="paginatorFirst"
          :always-show="true"
          :template="paginatorTemplate"
          :current-page-report-template="currentPageReportTemplate"
          class="data-table__paginator"
          @page="onPageChange"
        />
        <div class="data-table__toolbar-actions">
          <div class="data-table__control-group hidden xl:flex">
            <label for="page-size-select-v4" class="data-table__control-label">
              {{ pageSizeLabel }}
            </label>
            <div class="data-table__control-field">
              <Select
                :model-value="pageSizePreset"
                input-id="page-size-select-v4"
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
            class="data-table__control-group hidden xl:flex"
            @submit.prevent="onJumpPage()"
          >
            <label for="page-jump-input-v4" class="data-table__control-label">
              {{ goToLabel }}
            </label>
            <InputNumber
              v-model="jumpPage"
              input-id="page-jump-input-v4"
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
                  class="flex items-center justify-between rounded-md px-2 py-2 text-sm text-slate-800 hover:bg-slate-50"
                >
                  <span class="truncate">{{ col.label }}</span>
                  <div class="flex items-center gap-1">
                    <button
                      type="button"
                      class="grid h-8 w-8 place-items-center rounded-md text-slate-500 transition hover:bg-slate-100 hover:text-primary-700"
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
                      class="grid h-8 w-8 place-items-center rounded-md text-slate-500 transition hover:bg-slate-100 hover:text-primary-700"
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
            />

            <slot
              :items="paginatedItems"
              :loading="tableLoading"
              :selection="selection"
              :table-props="tableProps"
              :on-row-click="onRowClick"
              :hidden-columns="hiddenColumns"
              :frozen-columns="frozenColumns"
              :toggle-column-freeze="toggleColumnFreeze"
              :toggle-column-visibility="toggleColumnVisibility"
              :pagination="pagination"
              :total-items="sortedItems"
            />
          </DataTable>
        </div>
      </div>
    </div>
    <div
      v-else
      class="flex min-h-[200px] items-center justify-center text-lg font-semibold text-slate-400"
    >
      {{ noAccessMessage }}
    </div>
  </section>
</template>

<style scoped>
.select-no-arrow :deep(.p-select-dropdown) {
  display: none;
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
  /* border-top: 1px solid #e2e8f0; */
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
  flex-wrap: wrap;
}

.data-table__paginator {
  min-width: 0;
}

.data-table__toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 0.45rem;
  color: #475569;
  margin-left: auto;
}

.data-table__control-group,
.data-table__icon-group {
  align-items: center;
  gap: 0.4rem;
  min-height: 2rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  background: #f8fafc;
  padding: 0.15rem 0.25rem 0.15rem 0.5rem;
}

.data-table__control-label {
  font-size: calc(var(--font-table) * 0.92);
  font-weight: 600;
  color: #475569;
  white-space: nowrap;
}

.data-table__control-field {
  display: flex;
  align-items: center;
}

.data-table__icon-button,
.data-table__small-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.45rem;
  border: 1px solid transparent;
  background: #fff;
  color: #475569;
  transition: all 0.2s ease;
}

.data-table__icon-button {
  width: 1.95rem;
  height: 1.95rem;
}

.data-table__small-button {
  min-height: 1.95rem;
  padding: 0 0.75rem;
  font-size: calc(var(--font-button) * 0.92);
  font-weight: 600;
}

.data-table__icon-button:hover,
.data-table__small-button:hover {
  border-color: #cbd5e1;
  color: #0f172a;
}

.data-table__selection {
  padding: 0.75rem 0.85rem;
  border-top: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
}

.data-table__body {
  min-height: 0;
}

.data-table__inner,
.data-table__outer-scroll {
  min-height: 0;
}

.data-table__loading {
  min-height: 12rem;
}

.data-table__grid :deep(.p-datatable-table-container) {
  min-height: inherit;
}

.data-table__grid :deep(.p-datatable-thead > tr > th) {
  white-space: nowrap;
}

@media (max-width: 767px) {
  .data-table__toolbar {
    align-items: center;
  }

  .data-table__paginator {
    flex: 1 1 auto;
  }

  .data-table__paginator :deep(.p-paginator) {
    justify-content: flex-start;
    flex-wrap: wrap;
    gap: 0.35rem;
    padding: 0;
    border: 0;
    background: transparent;
  }

  .data-table__paginator :deep(.p-paginator-pages) {
    display: none;
  }

  .data-table__paginator :deep(.p-paginator-current) {
    margin-inline-start: 0;
    font-size: calc(var(--font-table) * 0.92);
    line-height: 1.25rem;
    white-space: normal;
  }

  .data-table__paginator :deep(.p-paginator-prev),
  .data-table__paginator :deep(.p-paginator-next) {
    width: 1.95rem;
    height: 1.95rem;
    min-width: 1.95rem;
    border: 1px solid #e2e8f0;
    border-radius: 0.5rem;
    background: #fff;
    color: #475569;
  }

  .data-table__toolbar-actions {
    flex: 0 0 auto;
    align-self: flex-start;
  }
}
</style>
