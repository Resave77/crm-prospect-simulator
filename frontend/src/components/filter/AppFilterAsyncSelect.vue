<!-- TODO: masih ada bug, dia hit ke api padahal scroll belum sampai habis -->
<script setup lang="ts">
import { useInfiniteQuery } from "@tanstack/vue-query";
import FloatLabel from "primevue/floatlabel";
import MultiSelect from "primevue/multiselect";
import Select from "primevue/select";
import { computed, onBeforeUnmount, reactive, watch } from "vue";

import type {
  ShortListItem,
  ShortListParams,
  ShortListResult,
} from "@/api/general";

type LoaderFn = (params: ShortListParams) => Promise<ShortListResult>;
type SelectMode = "single" | "multiple";
type LabelPosition = "none" | "top" | "float";

interface Props {
  cacheKey: string;
  disabled?: boolean;
  inputClass?: string;
  label?: string;
  labelPosition?: LabelPosition;
  loader: LoaderFn;
  maxSelectedLabels?: number;
  mode?: SelectMode;
  modelValue: number | number[] | null;
  placeholder?: string;
  showToggleAll?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  inputClass: "",
  label: "",
  labelPosition: "none",
  maxSelectedLabels: 2,
  mode: "single",
  placeholder: "",
  showToggleAll: false,
});

const emit = defineEmits<{
  (e: "update:modelValue", value: number | number[] | null): void;
  (e: "items-loaded", items: ShortListItem[]): void;
}>();

const PAGE_SIZE = 25;

const state = reactive({
  initialized: false,
  search: "",
});

const placeholderValue = computed(() => props.placeholder || "Select");

// stable id, jangan computed random
const inputId = `select-${Math.random().toString(36).slice(2, 9)}`;

const queryKey = computed(
  () => ["shortlist", props.cacheKey, state.search] as const
);

const shortlistQuery = useInfiniteQuery<ShortListResult, Error>({
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

const options = computed(() =>
  items.value.map((item: ShortListItem) => ({
    label: item.name,
    value: item.id,
  }))
);

const isLoading = computed(
  () =>
    shortlistQuery.isLoading.value ||
    shortlistQuery.isFetching.value ||
    shortlistQuery.isFetchingNextPage.value
);

const internalValue = computed<number | number[] | null>({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", val),
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

watch(items, (val) => {
  emit("items-loaded", val);
});

watch(
  () => [props.loader, props.cacheKey],
  () => {
    state.initialized = false;
    state.search = "";
  }
);

onBeforeUnmount(() => {
  if (filterTimer) {
    window.clearTimeout(filterTimer);
  }
});
</script>

<template>
  <div v-if="labelPosition === 'top'" class="app-filter-select-wrapper">
    <label
      v-if="label"
      :for="inputId"
      class="mb-0.5 block text-xs font-medium text-slate-500"
    >
      {{ label }}
    </label>

    <MultiSelect
      v-if="mode === 'multiple'"
      v-model="internalValue"
      :input-id="inputId"
      :options="options"
      :virtual-scroller-options="virtualScrollerOptions"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      filter
      display="chip"
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :max-selected-labels="maxSelectedLabels"
      :loading="isLoading"
      :show-toggle-all="showToggleAll"
      :disabled="disabled"
      @show="onShow"
      @filter="onFilter"
    />

    <Select
      v-else
      v-model="internalValue"
      :input-id="inputId"
      :options="options"
      :virtual-scroller-options="virtualScrollerOptions"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      filter
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :loading="isLoading"
      :disabled="disabled"
      @show="onShow"
      @filter="onFilter"
    />
  </div>

  <FloatLabel v-else-if="labelPosition === 'float'" variant="on">
    <MultiSelect
      v-if="mode === 'multiple'"
      v-model="internalValue"
      :input-id="inputId"
      :options="options"
      :virtual-scroller-options="virtualScrollerOptions"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      filter
      display="chip"
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :max-selected-labels="maxSelectedLabels"
      :loading="isLoading"
      :show-toggle-all="showToggleAll"
      :disabled="disabled"
      @show="onShow"
      @filter="onFilter"
    />

    <Select
      v-else
      v-model="internalValue"
      :input-id="inputId"
      :options="options"
      :virtual-scroller-options="virtualScrollerOptions"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      filter
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :loading="isLoading"
      :disabled="disabled"
      @show="onShow"
      @filter="onFilter"
    />

    <label :for="inputId" class="text-xs font-medium text-slate-500">
      {{ label }}
    </label>
  </FloatLabel>

  <template v-else>
    <MultiSelect
      v-if="mode === 'multiple'"
      v-model="internalValue"
      :input-id="inputId"
      :options="options"
      :virtual-scroller-options="virtualScrollerOptions"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      filter
      display="chip"
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :max-selected-labels="maxSelectedLabels"
      :loading="isLoading"
      :show-toggle-all="showToggleAll"
      :disabled="disabled"
      @show="onShow"
      @filter="onFilter"
    />

    <Select
      v-else
      v-model="internalValue"
      :input-id="inputId"
      :options="options"
      :virtual-scroller-options="virtualScrollerOptions"
      option-label="label"
      option-value="value"
      :placeholder="placeholderValue"
      filter
      :class="[
        '!bg-slate-100 !border-none !shadow-none hover:!bg-slate-200 focus:!bg-slate-200 focus:!outline focus:!outline-2 focus:!outline-blue-500',
        inputClass,
      ]"
      :loading="isLoading"
      :disabled="disabled"
      @show="onShow"
      @filter="onFilter"
    />
  </template>
</template>

<style scoped>
:deep(.p-multiselect),
:deep(.p-select) {
  font-size: var(--font-select);
  min-height: 2rem;
  height: 2rem;
}

:deep(.p-multiselect-label),
:deep(.p-select-label) {
  padding-top: 0.25rem;
  padding-bottom: 0.25rem;
}

:deep(.p-multiselect-header) {
  padding: 0.35rem 0.5rem;
}

:deep(.p-multiselect-filter),
:deep(.p-select-filter) {
  height: 2rem;
  font-size: var(--font-select);
}

:deep(.p-multiselect-items .p-multiselect-item),
:deep(.p-select-items .p-select-item) {
  font-size: calc(var(--font-select) * 0.92);
  padding-top: 0.35rem;
  padding-bottom: 0.35rem;
}

:deep(.p-multiselect-label.p-placeholder),
:deep(.p-select-label.p-placeholder) {
  color: #64748b !important;
}
</style>
