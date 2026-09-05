<script setup lang="ts">
import DataView from "primevue/dataview";
import { computed, onMounted, ref, watch } from "vue";

import type {
  LoadParams,
  LoadResult,
} from "@/components/data/AppDataTableV1.vue";

type CheckboxPlacement =
  | "bottom-left"
  | "bottom-right"
  | "top-left"
  | "top-right";
type SelectionMode = "multiple" | "single" | null;

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
    extraFilters?: Record<string, unknown> | null;
    initialPage?: number;
    listClass?: string;
    loadFn: (params: LoadParams) => Promise<LoadResult<unknown>>;
    loadingLabel?: string;
    loadingMoreLabel?: string;
    pageSize?: number;
    scrollThreshold?: number;
    selection?: unknown[] | null;
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
    extraFilters: () => ({}),
    initialPage: 1,
    loadingLabel: "Loading data...",
    loadingMoreLabel: "Loading more data...",
    pageSize: 10,
    scrollThreshold: 160,
    selection: undefined,
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
  (e: "update:selection", selection: unknown[]): void;
}>();

const internalSelection = ref<unknown[]>([]);
const items = ref<unknown[]>([]);
const loading = ref(false);
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
const selectionCountLabel = computed(
  () => `${selection.value.length} selected`
);

const canLoadMore = () => page.value < totalPages.value && !loading.value;

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

const loadData = async (options: { reset?: boolean } = {}) => {
  const reset = options.reset ?? false;
  if (!props.enabled || loading.value) return;

  loading.value = true;
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
    loading.value = false;
  }
};

const reload = () => loadData({ reset: true });

const onScroll = (event: Event) => {
  const target = event.target as HTMLElement | null;
  if (!target || !canLoadMore()) return;

  const distanceToBottom =
    target.scrollHeight - target.scrollTop - target.clientHeight;

  if (distanceToBottom <= props.scrollThreshold) void loadData();
};

watch(
  () => [
    props.enabled,
    props.extraFilters,
    props.sortField,
    props.sortDirection,
  ],
  ([enabled]) => {
    if (enabled) void reload();
  },
  { deep: true }
);

onMounted(() => {
  if (props.enabled) void reload();
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
        class="mb-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
      >
        {{ error || localError }}
      </div>
    </slot>

    <div
      v-if="selectionMode && selection.length"
      class="mb-3 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-slate-200 bg-white px-4 py-3 text-table font-semibold text-slate-700 shadow-sm"
    >
      <slot
        name="selection"
        :selection="selection"
        :clear-selection="clearSelection"
      >
        <span>{{ selectionCountLabel }}</span>
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-button font-semibold text-slate-700 shadow-sm transition hover:border-slate-300 hover:text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-200 focus:ring-offset-0"
          @click="clearSelection"
        >
          Clear
        </button>
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
              v-for="item in slotProps.items"
              :key="item?.[dataKey]"
              class="relative"
            >
              <label
                v-if="withCheckbox && selectionMode"
                class="absolute z-10 inline-flex h-6 w-6 items-center justify-center rounded-lg border border-slate-200 bg-white shadow-sm"
                :class="checkboxPlacementClass"
                @click.stop
              >
                <input
                  type="checkbox"
                  class="h-4 w-4 accent-[#dc2626]"
                  :checked="isSelected(item)"
                  @change="toggleSelection(item)"
                />
              </label>
              <slot
                name="item"
                :item="item"
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
