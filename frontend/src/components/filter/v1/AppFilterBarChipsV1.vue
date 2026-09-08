<script setup lang="ts">
import { computed } from "vue";

export interface FilterBarChipGroupV1 {
  items: { label: string; value: number | string }[];
  key: string;
  label: string;
}

type FilterChipItemV1 = {
  groupKey: string;
  groupLabel: string;
  itemLabel: string;
  itemValue: number | string;
};

interface Props {
  groups: FilterBarChipGroupV1[];
  showClearAll?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showClearAll: true,
});

const emit = defineEmits<{
  (e: "clear-all"): void;
  (e: "remove", payload: { key: string; value: number | string }): void;
}>();

const chips = computed<FilterChipItemV1[]>(() =>
  props.groups.flatMap((group) =>
    group.items.map((item) => ({
      groupKey: group.key,
      groupLabel: group.label,
      itemLabel: item.label,
      itemValue: item.value,
    }))
  )
);
const hasAnyFilter = computed(() => chips.value.length > 0);

const visibleChips = computed(() => chips.value);
</script>

<template>
  <div
    v-if="hasAnyFilter"
    class="filter-bar-chips flex max-w-full flex-nowrap items-center gap-1.5 overflow-x-auto overflow-y-hidden whitespace-nowrap"
  >
    <span
      v-for="chip in visibleChips"
      :key="`${chip.groupKey}-${chip.itemValue}`"
      class="inline-flex max-w-[240px] shrink-0 items-center gap-1.5 rounded-full border border-red-100 bg-red-50 px-2.5 py-1 text-[11px] font-medium text-red-700"
    >
      <span class="min-w-0 truncate">
        <span class="font-semibold">{{ chip.groupLabel }}:</span>
        {{ chip.itemLabel }}
      </span>
      <button
        type="button"
        class="inline-flex h-4 w-4 shrink-0 items-center justify-center rounded-full text-red-500 transition hover:bg-red-100 hover:text-red-700"
        @click="emit('remove', { key: chip.groupKey, value: chip.itemValue })"
      >
        <span class="pi pi-times text-[9px]" />
      </button>
    </span>

    <button
      v-if="showClearAll"
      type="button"
      class="inline-flex shrink-0 items-center gap-1 rounded-full px-2 py-1 text-[11px] font-semibold text-red-600 transition hover:bg-red-50"
      @click="emit('clear-all')"
    >
      <span class="pi pi-trash text-[10px]" />
      Clear
    </button>
  </div>
</template>

<style scoped>
.filter-bar-chips {
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.filter-bar-chips::-webkit-scrollbar {
  display: none;
}
</style>
