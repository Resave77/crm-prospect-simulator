<!-- @ts-nocheck: supports both inline and drawer display modes. -->
<script setup lang="ts">
// @ts-nocheck
import { computed, useSlots } from "vue";

import AppFilterChips, {
  type FilterChipGroup,
} from "@/components/filter/AppFilterChips.vue";
import AppFilterBarDrawer from "@/components/filter/v1/AppFilterBarDrawer.vue";
import { useUserPreferenceStore } from "@/stores/userPreference";

type FilterDisplayModeOverride = "drawer" | "inline" | "preference";
type InlineFilterBehavior = "always" | "header-toggle";

interface Props {
  chipGroups?: FilterChipGroup[];
  displayMode?: FilterDisplayModeOverride;
  drawerOpen?: boolean;
  filterOpen?: boolean;
  inlineBehavior?: InlineFilterBehavior;
  resultTotal?: number | null;
  title?: string;
}

const props = withDefaults(defineProps<Props>(), {
  chipGroups: () => [],
  displayMode: "preference",
  drawerOpen: false,
  filterOpen: false,
  inlineBehavior: "always",
  resultTotal: undefined,
  title: "",
});

const emit = defineEmits<{
  (e: "chip-remove", payload: { key: string; value: number | string }): void;
  (e: "clear-all"): void;
  (e: "update:drawerOpen", value: boolean): void;
}>();

const userPreferences = useUserPreferenceStore();
const slots = useSlots();
const filterDisplayMode = computed(() => userPreferences.filterDisplayMode);
const filterChipPosition = computed(() => userPreferences.filterChipPosition);
const isMobile = computed(() => userPreferences.isMobile);

const resolvedTitle = computed(() => props.title || "Filter");
const hasActiveFilters = computed(() =>
  props.chipGroups.some((group) => group.items.length > 0)
);

const resolvedDisplayMode = computed(() => {
  if (isMobile.value) return "drawer";
  if (props.displayMode !== "preference") return props.displayMode;
  return filterDisplayMode.value;
});

const isDrawerMode = computed(() => resolvedDisplayMode.value === "drawer");
const hasInlineSlot = computed(() => Boolean(slots.inline));

const showInlineControls = computed(() => {
  if (isDrawerMode.value) return false;
  if (!hasInlineSlot.value) return false;
  if (props.inlineBehavior === "header-toggle") {
    return props.filterOpen;
  }
  return true;
});

const showInlineChips = computed(() => {
  if (isMobile.value) return false;
  if (resolvedDisplayMode.value === "inline") {
    if (props.inlineBehavior === "header-toggle") {
      return props.filterOpen && hasActiveFilters.value;
    }
    return hasActiveFilters.value;
  }
  return filterChipPosition.value === "inline" && hasActiveFilters.value;
});

const shouldRenderInlineContainer = computed(
  () =>
    !isDrawerMode.value && (showInlineControls.value || showInlineChips.value)
);

// const showResultInfo = computed(() => typeof props.resultTotal === "number");

const drawerChipGroups = computed(() =>
  filterChipPosition.value === "drawer" ? props.chipGroups : []
);

const onChipRemove = (payload: { key: string; value: number | string }) => {
  emit("chip-remove", payload);
};

const onClearAll = () => {
  emit("clear-all");
};
</script>

<template>
  <!-- Inline Mode -->
  <template v-if="!isDrawerMode">
    <div v-if="shouldRenderInlineContainer" class="p-2">
      <!-- Inline Filter Controls -->
      <div v-if="showInlineControls" class="filter-controls-inline">
        <slot name="inline" />
      </div>

      <!-- Active Filter Chips -->
      <div
        v-if="showInlineChips"
        :class="['chips-summary-wrapper', { 'mt-3': showInlineControls }]"
      >
        <AppFilterChips
          :groups="chipGroups"
          @remove="onChipRemove"
          @clear-all="onClearAll"
        />
        <!-- <p
          v-if="showResultInfo"
          class="result-summary pl-2 text-xs text-slate-500"
        >
          {{ resultSummaryText }}
        </p> -->
      </div>
    </div>
  </template>

  <!-- Drawer Mode -->
  <template v-else>
    <!-- Active Filter Chips (if position is inline and not mobile) -->
    <div v-if="showInlineChips" class="chips-summary-wrapper">
      <AppFilterChips
        :groups="chipGroups"
        @remove="onChipRemove"
        @clear-all="onClearAll"
      />
      <!-- <p
          v-if="showResultInfo"
          class="result-summary pl-2 text-xs text-slate-500"
        >
          {{ resultSummaryText }}
        </p> -->
    </div>

    <!-- Filter Drawer -->
    <AppFilterBarDrawer
      :model-value="drawerOpen"
      :title="resolvedTitle"
      :chip-groups="drawerChipGroups"
      :full-screen="isMobile"
      :result-total="props.resultTotal"
      @update:model-value="(val) => emit('update:drawerOpen', val)"
      @chip-remove="onChipRemove"
      @chip-clear-all="onClearAll"
      @reset="onClearAll"
    >
      <slot name="drawer" />
    </AppFilterBarDrawer>
  </template>
</template>

<style scoped>
.filter-controls-inline {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  align-items: flex-end;
  gap: 0.75rem;
}

@media (min-width: 640px) {
  .filter-controls-inline {
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  }
}

@media (min-width: 1024px) {
  .filter-controls-inline {
    grid-template-columns: repeat(4, 1fr);
  }
}

.chips-summary-wrapper {
  display: flex;
  flex-direction: column;
}

.result-summary {
  margin-top: 0.35rem;
}
</style>
