<script setup lang="ts">
import Column from "primevue/column";
import DataTable from "primevue/datatable";
import { computed, ref, watch } from "vue";

import { getObjectPathValue } from "@/lib/utils/objectPath";

type SortDirection = "asc" | "desc";

export interface AppDataTableV3LoadParams {
  extraFilters?: Record<string, unknown> | null;
  page: number;
  pageSize: number;
  sortDirection?: SortDirection;
  sortField?: string | null;
}

export interface AppDataTableV3PaginationInfo {
  page: number;
  total: number;
  total_pages: number;
}

export interface AppDataTableV3LoadResult<T> {
  items: T[];
  pagination: AppDataTableV3PaginationInfo;
}

export type AppDataTableV3Column = {
  align?: "left" | "right";
  cellClass?: string;
  field?: string;
  header: string;
  key: string;
  width?: string;
};

const EXTRA_FILTERS_READY_META_KEY = "__appFiltersReady";

interface Props {
  columns: AppDataTableV3Column[];
  emptyText?: string;
  errorMessage?: string;
  extraFilters?: Record<string, unknown> | null;
  initialPageSize?: number;
  items?: Record<string, unknown>[];
  loadFn?: (
    params: AppDataTableV3LoadParams
  ) => Promise<AppDataTableV3LoadResult<unknown>>;
  loading?: boolean;
  loadingText?: string;
  minColumnWidth?: number;
  minTableWidth?: string;
  rowClickable?: boolean;
  rowKey?: string;
  showLessLabel?: string;
  showMoreLabel?: string;
  showMoreThreshold?: number;
}

const props = withDefaults(defineProps<Props>(), {
  emptyText: "No data found.",
  errorMessage: "",
  extraFilters: () => ({}),
  initialPageSize: 50,
  items: () => [],
  loadFn: undefined,
  loading: false,
  loadingText: "Loading data...",
  minColumnWidth: 160,
  minTableWidth: "",
  rowClickable: false,
  rowKey: "id",
  showLessLabel: "Show Less",
  showMoreLabel: "Show All",
  showMoreThreshold: 3,
});

const emit = defineEmits<{
  (e: "error", error: unknown): void;
  (
    e: "loaded",
    value: {
      items: unknown[];
      pagination: AppDataTableV3PaginationInfo | null;
    }
  ): void;
  (e: "row-click", value: { data: unknown; originalEvent: Event | null }): void;
}>();

const localItems = ref<unknown[]>([]);
const localLoading = ref(false);
const localError = ref("");
const isExpanded = ref(false);

const resolvedItems = computed(() =>
  props.loadFn
    ? (localItems.value as Record<string, unknown>[])
    : (props.items as Record<string, unknown>[])
);

const resolvedLoading = computed(() =>
  props.loadFn ? localLoading.value : props.loading
);

const resolvedErrorMessage = computed(() =>
  props.loadFn ? localError.value : props.errorMessage
);

const shouldShowMore = computed(
  () =>
    props.showMoreThreshold > 0 &&
    resolvedItems.value.length > props.showMoreThreshold
);

const displayedItems = computed(() => {
  if (!shouldShowMore.value || isExpanded.value) {
    return resolvedItems.value;
  }

  return resolvedItems.value.slice(0, props.showMoreThreshold);
});

const getColumnMinWidth = (column: AppDataTableV3Column) => {
  const width = column.width?.trim();
  if (!width) return props.minColumnWidth;
  if (width === "auto") return 96;

  const pixelMatch = width.match(/^(\d*\.?\d+)px$/);
  if (pixelMatch?.[1]) {
    return Number(pixelMatch[1]);
  }

  const fractionMatch = width.match(/^(\d*\.?\d+)fr$/);
  if (fractionMatch?.[1]) {
    return Number(fractionMatch[1]) * props.minColumnWidth;
  }

  return props.minColumnWidth;
};

const tableMinWidth = computed(() => {
  if (props.minTableWidth) return props.minTableWidth;

  const width = props.columns.reduce(
    (total, column) => total + getColumnMinWidth(column),
    0
  );

  return `${Math.max(width, props.minColumnWidth)}px`;
});

const tableStyle = computed(() => ({
  minWidth: tableMinWidth.value,
}));

const loadingSkeletonRows = computed(() =>
  Array.from({ length: Math.max(3, Math.min(props.showMoreThreshold, 5)) })
);

const getColumnStyle = (column: AppDataTableV3Column) => {
  const width = column.width?.trim();
  const minWidth = `${getColumnMinWidth(column)}px`;

  if (!width || width.endsWith("fr")) {
    return { minWidth };
  }

  if (width === "auto") {
    return { minWidth, width: "1%" };
  }

  if (width.endsWith("px")) {
    return { maxWidth: width, minWidth: width, width };
  }

  return { minWidth, width };
};

const getCellValue = (
  row: Record<string, unknown>,
  column: AppDataTableV3Column
) => {
  if (!column.field) return "-";
  const value = getObjectPathValue(row, column.field);
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  return value;
};

const loadData = async () => {
  if (!props.loadFn) return;

  localLoading.value = true;
  localError.value = "";

  try {
    const result = await props.loadFn({
      extraFilters: props.extraFilters ?? null,
      page: 1,
      pageSize: props.initialPageSize,
      sortDirection: undefined,
      sortField: null,
    });

    localItems.value = result.items ?? [];
    emit("loaded", {
      items: localItems.value,
      pagination: result.pagination ?? null,
    });
  } catch (err: unknown) {
    localError.value =
      (err as { message?: string; response?: { data?: { message?: string } } })
        ?.response?.data?.message ||
      (err as { message?: string })?.message ||
      "Failed to load table data.";
    emit("error", err);
  } finally {
    localLoading.value = false;
  }
};

const resetAndLoad = async () => {
  isExpanded.value = false;
  await loadData();
};

watch(
  () => props.extraFilters,
  (nextFilters) => {
    if (!props.loadFn) return;

    const filtersRecord = (nextFilters as Record<string, unknown> | null) ?? {};
    const isReady =
      !Object.prototype.hasOwnProperty.call(
        filtersRecord,
        EXTRA_FILTERS_READY_META_KEY
      ) || filtersRecord[EXTRA_FILTERS_READY_META_KEY] !== false;

    if (!isReady) return;
    loadData();
  },
  { deep: true, immediate: true }
);

watch(
  () => resolvedItems.value.length,
  () => {
    isExpanded.value = false;
  }
);

const onRowClick = (row: Record<string, unknown>, event: Event | null) => {
  emit("row-click", {
    data: row,
    originalEvent: event,
  });
};

const onPrimeRowClick = (event: {
  data: Record<string, unknown>;
  originalEvent: Event;
}) => {
  onRowClick(event.data, event.originalEvent);
};

const showMore = () => {
  isExpanded.value = true;
};

const showLess = () => {
  isExpanded.value = false;
};

defineExpose({
  loadData,
  resetAndLoad,
});
</script>

<template>
  <div
    class="app-data-table-v3 relative overflow-hidden rounded-xl border border-slate-200 bg-white"
  >
    <DataTable
      :value="resolvedErrorMessage ? [] : displayedItems"
      :loading="resolvedLoading"
      :data-key="rowKey"
      :row-hover="rowClickable"
      :table-style="tableStyle"
      size="small"
      @row-click="onPrimeRowClick"
    >
      <template #empty>
        <div
          class="flex min-h-[160px] items-center justify-center px-4 text-center text-sm"
          :class="resolvedErrorMessage ? 'text-rose-600' : 'text-slate-500'"
        >
          {{ resolvedErrorMessage || emptyText }}
        </div>
      </template>

      <template #loading>
        <div class="app-data-table-v3__loading min-w-full p-4">
          <div
            class="mb-3 flex items-center gap-2 text-xs font-semibold text-slate-500"
          >
            <span class="h-2 w-2 animate-pulse rounded-full bg-primary-500" />
            <span>{{ loadingText }}</span>
          </div>
          <div class="space-y-2">
            <div
              v-for="(_, index) in loadingSkeletonRows"
              :key="index"
              class="grid gap-3"
              :style="{
                gridTemplateColumns: columns
                  .map(() => 'minmax(80px, 1fr)')
                  .join(' '),
              }"
            >
              <span
                v-for="column in columns"
                :key="column.key"
                class="h-4 rounded-full bg-slate-100"
              />
            </div>
          </div>
        </div>
      </template>

      <Column
        v-for="column in columns"
        :key="column.key"
        :field="column.field"
        :header="column.header"
        :style="getColumnStyle(column)"
        :header-class="[
          column.align === 'right' ? '!text-right' : '!text-left',
        ]"
        :body-class="[
          column.align === 'right' ? '!text-right' : '!text-left',
          column.cellClass,
        ]"
      >
        <template #body="{ data }">
          <span class="text-sm text-slate-600">
            <slot
              :name="`cell-${column.key}`"
              :row="data"
              :value="getCellValue(data, column)"
            >
              {{ getCellValue(data, column) }}
            </slot>
          </span>
        </template>
      </Column>
    </DataTable>

    <div
      v-if="shouldShowMore && !isExpanded"
      class="flex justify-center border-t border-slate-100 px-3 py-2"
    >
      <button
        type="button"
        class="rounded-full border border-slate-200 bg-white px-4 py-1.5 text-xs font-semibold text-slate-600 transition hover:bg-slate-50"
        @click="showMore"
      >
        {{ showMoreLabel }}
        <i class="pi pi-chevron-down ml-1 text-[10px]" />
      </button>
    </div>

    <div
      v-if="shouldShowMore && isExpanded"
      class="flex justify-center border-t border-slate-100 px-3 py-2"
    >
      <button
        type="button"
        class="rounded-full border border-slate-200 bg-white px-4 py-1.5 text-xs font-semibold text-slate-600 transition hover:bg-slate-50"
        @click="showLess"
      >
        {{ showLessLabel }}
        <i class="pi pi-chevron-up ml-1 text-[10px]" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.app-data-table-v3 :deep(.p-datatable),
.app-data-table-v3 :deep(.p-datatable-wrapper) {
  max-height: none !important;
  overflow-y: hidden !important;
}

.app-data-table-v3 :deep(.p-datatable-table-container) {
  max-height: none !important;
  overflow-x: auto !important;
  overflow-y: hidden !important;
}

.app-data-table-v3 :deep(.p-datatable-table) {
  border-collapse: separate;
  border-spacing: 0;
}

.app-data-table-v3 :deep(.p-datatable-thead > tr > th) {
  border-color: #e2e8f0;
  background: #f8fafc;
  color: #64748b;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.08em;
  padding: 8px 16px;
  text-transform: uppercase;
}

.app-data-table-v3 :deep(.p-datatable-tbody > tr > td) {
  border-color: #f1f5f9;
  color: #475569;
  padding: 12px 16px;
}

.app-data-table-v3 :deep(.p-datatable-tbody > tr:last-child > td) {
  border-bottom: 0;
}

.app-data-table-v3 :deep(.p-datatable-tbody > tr) {
  transition: background-color 150ms ease;
}

.app-data-table-v3 :deep(.p-datatable-tbody > tr:hover) {
  background: #f8fafc;
}

.app-data-table-v3 :deep(.p-datatable-loading-overlay) {
  background: rgba(255, 255, 255, 0.88) !important;
  backdrop-filter: blur(2px);
  color: #64748b;
}

.app-data-table-v3 :deep(.p-datatable-mask) {
  background: rgba(255, 255, 255, 0.88) !important;
  backdrop-filter: blur(2px);
  color: #64748b;
}

.app-data-table-v3__loading {
  background:
    linear-gradient(
      90deg,
      rgba(248, 250, 252, 0.92),
      rgba(255, 255, 255, 0.96)
    ),
    #ffffff;
}
</style>
