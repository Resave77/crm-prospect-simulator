<script setup lang="ts">
import InputNumber from "primevue/inputnumber";
import Paginator, { type PageState } from "primevue/paginator";
import ProgressSpinner from "primevue/progressspinner";
import Select from "primevue/select";
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";

import { useCurrentPageReportTemplate } from "@/composables/useCurrentPageReportTemplate";
import {
  normalizePageSizePresets,
  resolvePageSizeBounds,
} from "@/composables/useTablePagination";
import { getObjectPathValue } from "@/lib/utils/objectPath";

type SelectionMode = "single" | "multiple" | null;
type SortDirection = "asc" | "desc";

export interface AppDataTableSelectedLoadParams {
  extraFilters?: Record<string, unknown> | null;
  page: number;
  pageSize: number;
  sortDirection?: SortDirection;
  sortField?: string | null;
}

export interface AppDataTableSelectedPaginationInfo {
  page: number;
  total: number;
  total_pages: number;
}

export interface AppDataTableSelectedLoadResult<T> {
  items: T[];
  pagination: AppDataTableSelectedPaginationInfo;
}

export type AppDataTableSelectedColumn = {
  cellClass?: string;
  field?: string;
  headerClass?: string;
  key: string;
  label: string;
  width?: string;
};

const EXTRA_FILTERS_READY_META_KEY = "__appFiltersReady";
const FILTER_DEBOUNCE_DELAY = 400;

const props = withDefaults(
  defineProps<{
    columns: AppDataTableSelectedColumn[];
    dataKey?: string;
    emptyMessage?: string;
    extraFilters?: Record<string, unknown> | null;
    initialPage?: number;
    initialPageSize?: number;
    loadFn: (
      params: AppDataTableSelectedLoadParams
    ) => Promise<AppDataTableSelectedLoadResult<unknown>>;
    loadingMessage?: string;
    maxHeight?: number | string;
    minHeight?: number | string;
    pageSizePresets?: number[];
    selection?: unknown[] | null;
    selectionMode?: SelectionMode;
    valueKey?: string;
  }>(),
  {
    dataKey: undefined,
    emptyMessage: "No data found",
    extraFilters: () => ({}),
    initialPage: 1,
    initialPageSize: 10,
    loadingMessage: "Loading data...",
    maxHeight: 480,
    minHeight: 420,
    pageSizePresets: () => [5, 10, 15, 20, 25, 50],
    selection: undefined,
    selectionMode: null,
    valueKey: undefined,
  }
);

const emit = defineEmits<{
  (e: "update:selection", value: unknown[]): void;
  (e: "row-click", value: { data: unknown; originalEvent: Event | null }): void;
  (
    e: "loaded",
    value: {
      items: unknown[];
      pagination: AppDataTableSelectedPaginationInfo | null;
    }
  ): void;
  (e: "error", error: unknown): void;
}>();

const { currentPageReportTemplate } = useCurrentPageReportTemplate();

const items = ref<unknown[]>([]);
const pagination = ref<AppDataTableSelectedPaginationInfo | null>(null);
const loading = ref(false);
const internalSelection = ref<unknown[]>([]);
const tableState = ref({
  page: props.initialPage,
  rows: props.initialPageSize,
});
const jumpPage = ref(tableState.value.page);
const pageSizeInput = ref(tableState.value.rows);
const resolvedPageSizePresets = computed(() =>
  normalizePageSizePresets(props.pageSizePresets)
);
const pageSizeBounds = computed(() =>
  resolvePageSizeBounds(resolvedPageSizePresets.value)
);
const pageSizePreset = ref<number | "custom">(
  resolvedPageSizePresets.value.includes(tableState.value.rows)
    ? tableState.value.rows
    : "custom"
);

const minPageSize = computed(() => pageSizeBounds.value.min);
const maxPageSize = computed(() => pageSizeBounds.value.max);
const totalRecords = computed(() => pagination.value?.total ?? 0);
const totalPages = computed(() => pagination.value?.total_pages ?? 1);
const paginatorFirst = computed(
  () => (tableState.value.page - 1) * tableState.value.rows
);
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
const gridTemplateColumns = computed(() =>
  props.columns.map((column) => column.width ?? "minmax(0,1fr)").join(" ")
);
const tableStyle = computed(() => ({
  maxHeight:
    typeof props.maxHeight === "number"
      ? `${props.maxHeight}px`
      : props.maxHeight,
  minHeight:
    typeof props.minHeight === "number"
      ? `${props.minHeight}px`
      : props.minHeight,
}));

const selection = computed<unknown[]>({
  get: () =>
    props.selection !== undefined && props.selection !== null
      ? props.selection
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

const getSelectionValue = (item: unknown) => {
  const fieldKey = selectionFieldKey.value;
  if (!fieldKey) return item;
  return getObjectPathValue(item, fieldKey);
};

const isSameSelectionValue = (left: unknown, right: unknown) => {
  const leftValue = getSelectionValue(left);
  const rightValue = getSelectionValue(right);
  if (leftValue === undefined || leftValue === null) return left === right;
  if (rightValue === undefined || rightValue === null) return left === right;
  return String(leftValue) === String(rightValue);
};

const isSelected = (item: unknown) =>
  selection.value.some((selected) => isSameSelectionValue(selected, item));

const clearSelection = () => {
  selection.value = [];
};

const toggleSelection = (item: unknown) => {
  if (!props.selectionMode) return;

  if (props.selectionMode === "single") {
    selection.value = isSelected(item) ? [] : [item];
    return;
  }

  selection.value = isSelected(item)
    ? selection.value.filter(
        (selected) => !isSameSelectionValue(selected, item)
      )
    : [...selection.value, item];
};

const buildParams = (page: number): AppDataTableSelectedLoadParams => ({
  extraFilters: props.extraFilters ?? null,
  page,
  pageSize: tableState.value.rows,
});

let latestLoadRequestId = 0;
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
    if (requestId !== latestLoadRequestId) return;

    pagination.value = pageInfo ?? null;
    tableState.value.page = pageInfo?.page ?? targetPage;
    items.value = nextItems;
    emit("loaded", { items: items.value, pagination: pagination.value });
  } catch (err) {
    if (requestId === latestLoadRequestId) {
      emit("error", err);
    }
  } finally {
    if (requestId === latestLoadRequestId && shouldShowLoading) {
      loading.value = false;
    }
  }
};

const resetAndLoad = (options?: {
  clearSelection?: boolean;
  keepPage?: boolean;
}) => {
  if (!options?.keepPage) {
    tableState.value.page = 1;
  }
  jumpPage.value = tableState.value.page;

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
  loadData({ page: nextPage });
};

const onJumpPage = () => {
  const target = Math.max(
    1,
    Math.min(jumpPage.value || 1, pagination.value?.total_pages ?? 1)
  );
  if (target === tableState.value.page) return;
  jumpPage.value = target;
  tableState.value.page = target;
  items.value = [];
  loadData({ page: target });
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
  loadData();
};

const getCellValue = (item: unknown, column: AppDataTableSelectedColumn) => {
  const value = getObjectPathValue(item, column.field ?? column.key);
  return value === undefined || value === null || value === "" ? "-" : value;
};

const onRowClick = (item: unknown, event: MouseEvent) => {
  toggleSelection(item);
  emit("row-click", { data: item, originalEvent: event });
};

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

let filterTimer: number | undefined;
let lastExtraFiltersSignature = createExtraFiltersSignature(props.extraFilters);

watch(
  () => props.extraFilters,
  (nextFilters) => {
    if (!areExtraFiltersReady(nextFilters)) return;

    const nextSignature = createExtraFiltersSignature(nextFilters);
    if (nextSignature === lastExtraFiltersSignature) return;
    lastExtraFiltersSignature = nextSignature;

    if (filterTimer) window.clearTimeout(filterTimer);
    filterTimer = window.setTimeout(() => {
      resetAndLoad({ clearSelection: true });
    }, FILTER_DEBOUNCE_DELAY);
  },
  { deep: true }
);

watch(
  () => pagination.value?.page,
  (value) => {
    if (typeof value === "number") {
      jumpPage.value = value;
    }
  }
);

watch(
  () => tableState.value.rows,
  (value) => {
    pageSizeInput.value = value;
    pageSizePreset.value = resolvedPageSizePresets.value.includes(value)
      ? value
      : "custom";
  }
);

onMounted(() => {
  if (areExtraFiltersReady(props.extraFilters)) {
    loadData();
  }
});

onBeforeUnmount(() => {
  if (filterTimer) {
    window.clearTimeout(filterTimer);
  }
});

defineExpose({
  clearSelection,
  loadData,
  resetAndLoad,
});
</script>

<template>
  <section
    class="app-data-table-selected flex h-full min-h-0 min-w-0 flex-col overflow-hidden"
  >
    <div class="selected-table__toolbar">
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

      <div class="selected-table__toolbar-actions">
        <div class="selected-table__control-group">
          <label
            for="selected-page-size-select"
            class="selected-table__control-label"
          >
            Page size
          </label>
          <Select
            :model-value="pageSizePreset"
            input-id="selected-page-size-select"
            :options="pageSizeOptions"
            size="small"
            option-label="label"
            option-value="value"
            class="selected-table__page-size-select selected-table__select-no-arrow"
            @update:model-value="onSelectPageSizePreset"
          />
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
              class="selected-table__small-button"
              @click="applyCustomPageSize"
            >
              Set
            </button>
          </div>
        </div>

        <form
          class="selected-table__control-group"
          @submit.prevent="onJumpPage"
        >
          <label
            for="selected-page-jump-input"
            class="selected-table__control-label"
          >
            Go to
          </label>
          <InputNumber
            v-model="jumpPage"
            input-id="selected-page-jump-input"
            input-class="!w-14 !h-8 text-form"
            size="small"
            :min="1"
            :max="totalPages || 1"
            :use-grouping="false"
            @blur="onJumpPage"
          />
          <button type="submit" class="selected-table__small-button">
            Set
          </button>
        </form>

        <button
          type="button"
          class="selected-table__icon-button"
          aria-label="Refresh"
          title="Refresh"
          :disabled="loading"
          @click="refreshTable"
        >
          <span
            :class="['pi pi-refresh text-sm', loading ? 'animate-spin' : '']"
          />
        </button>
      </div>
    </div>

    <div class="selected-table__frame" :style="tableStyle">
      <div class="selected-table__header" :style="{ gridTemplateColumns }">
        <span
          v-for="column in columns"
          :key="column.key"
          :class="column.headerClass"
        >
          {{ column.label }}
        </span>
      </div>

      <div class="selected-table__body">
        <div v-if="loading" class="selected-table__loading">
          <ProgressSpinner
            stroke-width="6"
            style="width: 2.75rem; height: 2.75rem"
          />
          <span>{{ loadingMessage }}</span>
        </div>

        <div v-else-if="!items.length" class="selected-table__empty">
          {{ emptyMessage }}
        </div>

        <button
          v-for="item in items"
          v-else
          :key="String(getSelectionValue(item) ?? items.indexOf(item))"
          type="button"
          class="selected-table__row"
          :class="{ 'is-selected': isSelected(item) }"
          :style="{ gridTemplateColumns }"
          @click="onRowClick(item, $event)"
        >
          <span
            v-for="column in columns"
            :key="column.key"
            class="selected-table__cell"
            :class="[
              column.cellClass,
              { 'selected-table__cell--select': column.key === 'select' },
            ]"
          >
            <slot
              :name="`cell-${column.key}`"
              :item="item"
              :data="item"
              :selected="isSelected(item)"
              :select="() => toggleSelection(item)"
            >
              {{ getCellValue(item, column) }}
            </slot>
          </span>
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.app-data-table-selected {
  color: #0f172a;
}

.selected-table__toolbar {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding-bottom: 0.75rem;
}

.selected-table__toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 0.45rem;
}

.selected-table__control-group {
  display: flex;
  min-height: 2rem;
  align-items: center;
  gap: 0.4rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  background: #f8fafc;
  padding: 0.15rem 0.25rem 0.15rem 0.5rem;
}

.selected-table__control-label {
  white-space: nowrap;
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
  color: #64748b;
}

.selected-table__page-size-select {
  width: auto !important;
  min-width: 5.25rem !important;
  min-height: 1.75rem;
  height: 1.75rem;
}

.selected-table__page-size-select :deep(.p-select-label) {
  padding-top: 0.2rem;
  padding-bottom: 0.2rem;
  font-size: 12px;
}

.selected-table__select-no-arrow :deep(.p-select-dropdown) {
  display: none;
}

.selected-table__small-button,
.selected-table__icon-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #cbd5e1;
  background: #fff;
  font-size: 11px;
  font-weight: 700;
  color: #475569;
  transition:
    border-color 0.15s ease,
    color 0.15s ease,
    background-color 0.15s ease;
}

.selected-table__small-button {
  min-height: 1.65rem;
  border-radius: 0.375rem;
  padding: 0 0.5rem;
}

.selected-table__icon-button {
  height: 2rem;
  width: 2rem;
  flex-shrink: 0;
  border-radius: 0.5rem;
}

.selected-table__small-button:hover,
.selected-table__icon-button:hover {
  border-color: #fecaca;
  color: #b91c1c;
}

.selected-table__icon-button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.selected-table__toolbar :deep(.p-paginator) {
  min-height: 2rem;
  padding: 0;
  gap: 0.15rem;
}

.selected-table__toolbar :deep(.p-paginator-page),
.selected-table__toolbar :deep(.p-paginator-first),
.selected-table__toolbar :deep(.p-paginator-prev),
.selected-table__toolbar :deep(.p-paginator-next),
.selected-table__toolbar :deep(.p-paginator-last) {
  height: 1.75rem;
  min-width: 1.75rem;
  padding: 0 0.35rem;
  border-radius: 0.375rem;
  font-size: 12px;
}

.selected-table__toolbar :deep(.p-paginator-page.p-paginator-page-selected) {
  background: #dc2626;
  color: #fff;
}

.selected-table__toolbar :deep(.p-paginator-current) {
  margin-left: 0.35rem;
  font-size: 11px;
  color: #64748b;
}

.selected-table__frame {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  border-radius: 18px;
  background: #fff;
}

.selected-table__header,
.selected-table__row {
  display: grid;
  gap: 12px;
}

.selected-table__header {
  flex-shrink: 0;
  border-bottom: 1px solid #eef2f7;
  background: #f8fafc;
  padding: 12px 16px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.2;
  text-transform: uppercase;
  color: #64748b;
}

.selected-table__body {
  min-height: 0;
  flex: 1;
  overflow-x: hidden;
  overflow-y: auto;
}

.selected-table__row {
  width: 100%;
  align-items: center;
  border-bottom: 1px solid #eef2f7;
  background: #fff;
  padding: 16px;
  text-align: left;
  transition:
    background-color 0.15s ease,
    box-shadow 0.15s ease;
}

.selected-table__row:hover {
  background: #f8fafc;
}

.selected-table__row.is-selected {
  background: #fff7f7;
}

.selected-table__row:last-child {
  border-bottom: 0;
}

.selected-table__cell {
  min-width: 0;
  padding-right: 8px;
}

.selected-table__loading,
.selected-table__empty {
  display: flex;
  min-height: 12rem;
  align-items: center;
  justify-content: center;
  color: #64748b;
}

.selected-table__loading {
  flex-direction: column;
  gap: 0.75rem;
  font-size: 13px;
  font-weight: 700;
}

.selected-table__empty {
  font-size: 13px;
  font-weight: 600;
}

@media (max-width: 767px) {
  .selected-table__toolbar {
    flex-wrap: nowrap;
    align-items: center;
    gap: 0.5rem;
    padding-bottom: 0.625rem;
  }

  .selected-table__toolbar :deep(.p-paginator) {
    min-width: 0;
    flex: 1 1 auto;
    justify-content: flex-start;
    overflow-x: auto;
    overflow-y: hidden;
  }

  .selected-table__toolbar-actions {
    width: auto;
    flex: 0 0 auto;
    justify-content: flex-end;
    gap: 0.35rem;
  }

  .selected-table__control-group {
    display: none;
  }

  .selected-table__icon-button {
    flex: 0 0 2rem;
  }

  .selected-table__frame {
    border-radius: 14px;
    overflow-x: hidden;
    overflow-y: hidden;
    min-height: 0 !important;
  }

  .selected-table__body {
    min-width: 0;
  }

  .selected-table__header {
    display: none;
  }

  .selected-table__row {
    grid-template-columns: minmax(0, 1fr) !important;
    gap: 6px;
    align-items: stretch;
    position: relative;
    padding: 14px 52px 14px 16px;
  }

  .selected-table__cell {
    width: 100%;
    padding-right: 0;
  }

  .selected-table__cell--select {
    position: absolute;
    top: 14px;
    right: 16px;
    width: auto;
    padding-right: 0;
  }

  .selected-table__body {
    overflow-x: visible;
  }
}
</style>
