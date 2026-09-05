<script setup lang="ts">
import { computed } from "vue";

export interface FilterChipGroup {
  items: { label: string; value: number | string }[];
  key: string;
  label: string;
}

interface Props {
  groups: FilterChipGroup[];
  showClearAll?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showClearAll: true,
});

const emit = defineEmits<{
  (e: "remove", payload: { key: string; value: number | string }): void;
  (e: "clear-all"): void;
}>();

const clearLabel = computed(() => "Clear");

const hasAnyFilter = computed(() =>
  props.groups.some((group) => group.items.length > 0)
);

const onRemove = (key: string, value: number | string) => {
  emit("remove", { key, value });
};

const onClearAll = () => {
  emit("clear-all");
};
</script>

<template>
  <div
    v-if="hasAnyFilter"
    class="app-filter-chips pl-2 animate-in fade-in slide-in-from-left-2 flex flex-wrap items-center gap-3 duration-300"
  >
    <div
      v-for="group in groups"
      v-show="group.items.length > 0"
      :key="group.key"
      class="filter-group flex items-center gap-2 rounded-[8px] border border-dashed border-slate-200 bg-slate-50/50 px-3 py-1.5"
    >
      <span class="text-[12px] font-medium text-slate-500"
        >{{ group.label }}:</span
      >
      <div class="flex flex-wrap items-center gap-1.5">
        <div
          v-for="item in group.items"
          :key="item.value"
          class="flex items-center gap-1.5 rounded-full border border-slate-200 bg-white px-2.5 py-0.5 text-slate-900 shadow-sm"
        >
          <span class="text-[11px] leading-4">{{ item.label }}</span>
          <button
            type="button"
            class="text-slate-400 transition-colors hover:text-red-600"
            @click="onRemove(group.key, item.value)"
          >
            <i class="pi pi-times text-[10px]" />
          </button>
        </div>
      </div>
    </div>

    <button
      v-if="showClearAll && hasAnyFilter"
      type="button"
      class="clear-all-btn group ml-1 flex items-center gap-1.5 rounded-[6px] px-2 py-1 text-red-500 transition-all hover:bg-red-50"
      @click="onClearAll"
    >
      <i
        class="pi pi-trash text-[13px] transition-transform group-hover:scale-110"
      />
      <span class="text-[12px] font-semibold">{{ clearLabel }}</span>
    </button>
  </div>
</template>
