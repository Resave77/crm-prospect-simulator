<script setup lang="ts">
import {
  Comment,
  computed,
  Fragment,
  h,
  inject,
  onBeforeUnmount,
  onMounted,
  ref,
  Text,
  toRef,
  useSlots,
  type VNode,
} from "vue";

import AppButton from "@/components/base/AppButton.vue";
import { appPageHeaderV4FilterContextKey } from "@/components/page/AppPageHeaderV4FilterContext";

import type { FilterBarChipGroupV1 } from "./AppFilterBarChipsV1.vue";

import AppFilterBarChipsV1 from "./AppFilterBarChipsV1.vue";
import AppFilterBarDrawer from "./AppFilterBarDrawer.vue";

interface Props {
  chipGroups?: FilterBarChipGroupV1[];
  drawerOpen?: boolean;
  drawerTitle?: string;
  fullScreenDrawer?: boolean;
  inlineLimit?: number;
  resultTotal?: number | null;
  showChips?: boolean;
  wrapInline?: boolean;
}

type FilterNodeItem = {
  id: number | string | symbol;
  label: string;
  node: VNode;
  showDrawerLabel: boolean;
};

const props = withDefaults(defineProps<Props>(), {
  chipGroups: () => [],
  drawerOpen: undefined,
  drawerTitle: "",
  fullScreenDrawer: false,
  inlineLimit: 3,
  resultTotal: undefined,
  showChips: true,
  wrapInline: false,
});

const emit = defineEmits<{
  (e: "apply"): void;
  (e: "chip-clear-all"): void;
  (e: "chip-remove", payload: { key: string; value: number | string }): void;
  (e: "clear-all"): void;
  (e: "update:drawerOpen", value: boolean): void;
}>();

const slots = useSlots();
const internalDrawerOpen = ref(false);
const isMobile = ref(false);
const pageHeaderContext = inject(appPageHeaderV4FilterContextKey, null);
let unregisterFromPageHeader: (() => void) | null = null;

const checkMobile = () => {
  isMobile.value = window.innerWidth < 768;
};

const drawerVisible = computed({
  get: () => props.drawerOpen ?? internalDrawerOpen.value,
  set: (value: boolean) => {
    internalDrawerOpen.value = value;
    emit("update:drawerOpen", value);
  },
});

const resolvedDrawerTitle = computed(() => props.drawerTitle || "Filter");

const flattenNodes = (nodes: VNode[]): VNode[] =>
  nodes.flatMap((node) => {
    if (node.type === Fragment && Array.isArray(node.children)) {
      return flattenNodes(node.children as VNode[]);
    }

    if (node.type === Comment) {
      return [];
    }

    if (node.type === Text && String(node.children ?? "").trim() === "") {
      return [];
    }

    return [node];
  });

const getFilterNodeLabel = (node: VNode) => {
  const props = node.props ?? {};
  const label = props.label;
  const placeholder = props.placeholder;

  if (typeof label === "string" && label.trim()) return label.trim();
  if (typeof placeholder === "string" && placeholder.trim()) {
    return placeholder.trim();
  }

  return "Filter";
};

const shouldShowDrawerLabel = (node: VNode) => {
  const label = node.props?.label;
  return !(typeof label === "string" && label.trim());
};

const filterNodes = computed(() => flattenNodes(slots.default?.() ?? []));
const filterNodeItems = computed<FilterNodeItem[]>(() =>
  filterNodes.value.map((node, index) => ({
    id: node.key ?? index,
    label: getFilterNodeLabel(node),
    node,
    showDrawerLabel: shouldShowDrawerLabel(node),
  }))
);
const effectiveInlineLimit = computed(() =>
  isMobile.value ? 1 : Math.max(0, props.inlineLimit)
);
const visibleFilterNodeItems = computed(() =>
  filterNodeItems.value.slice(0, effectiveInlineLimit.value)
);
const overflowFilterNodeItems = computed(() =>
  filterNodeItems.value.slice(effectiveInlineLimit.value)
);
const drawerFilterNodeItems = computed(() => filterNodeItems.value);
const hasDrawerFilters = computed(
  () => overflowFilterNodeItems.value.length > 0
);
const hasActiveFilters = computed(() =>
  props.chipGroups.some((group) => group.items.length > 0)
);
const activeFilterCount = computed(() =>
  props.chipGroups.reduce((count, group) => count + group.items.length, 0)
);
const useCompactFilterButton = computed(() => isMobile.value);

const RenderNode = (nodeProps: {
  key?: number | string | symbol;
  node: VNode;
}) => h(nodeProps.node);

const openDrawer = () => {
  drawerVisible.value = true;
};

const onDrawerReset = () => {
  emit("clear-all");
};

onMounted(() => {
  checkMobile();
  window.addEventListener("resize", checkMobile);

  if (pageHeaderContext) {
    unregisterFromPageHeader = pageHeaderContext.register({
      chipGroups: toRef(props, "chipGroups"),
      onClearAll: () => emit("chip-clear-all"),
      onRemove: (payload) => emit("chip-remove", payload),
    });
  }
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", checkMobile);
  unregisterFromPageHeader?.();
});
</script>

<template>
  <div class="flex min-w-0 flex-col gap-2">
    <div
      class="flex min-w-0 items-end gap-2 md:gap-3"
      :class="
        props.wrapInline
          ? 'flex-wrap justify-start md:justify-end'
          : 'flex-nowrap'
      "
    >
      <template v-for="item in visibleFilterNodeItems" :key="item.id">
        <div
          class="min-w-0"
          :class="
            props.wrapInline ? 'flex-1 basis-0 md:flex-none' : 'flex-1 basis-0'
          "
        >
          <RenderNode :node="item.node" />
        </div>
      </template>

      <AppButton
        v-if="hasDrawerFilters"
        type="button"
        size="small"
        :variant="
          useCompactFilterButton
            ? 'secondary'
            : hasActiveFilters
              ? 'danger'
              : 'secondary'
        "
        :icon="useCompactFilterButton ? undefined : 'pi-filter'"
        :class="[
          'shrink-0',
          useCompactFilterButton
            ? '!relative !h-[34px] !w-[34px] !rounded-lg !p-0 !text-slate-700'
            : '',
        ]"
        @click="openDrawer"
      >
        <template v-if="useCompactFilterButton">
          <span class="pi pi-sliders-h text-[15px] leading-none" />
          <span
            v-if="hasActiveFilters"
            class="absolute -right-1 -top-1 inline-flex h-[17px] min-w-[17px] items-center justify-center rounded-full bg-red-600 px-[4px] text-[10px] font-bold leading-none text-white ring-2 ring-white"
          >
            {{ activeFilterCount }}
          </span>
        </template>
        <template v-else>
          More Filter
          <span
            v-if="hasActiveFilters"
            class="inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-white px-1 text-[10px] font-bold text-red-600"
          >
            {{ activeFilterCount }}
          </span>
        </template>
      </AppButton>

      <slot name="actions" />
    </div>

    <AppFilterBarChipsV1
      v-if="
        props.showChips && !pageHeaderContext && !isMobile && hasActiveFilters
      "
      :groups="chipGroups"
      @remove="(payload) => emit('chip-remove', payload)"
      @clear-all="emit('chip-clear-all')"
    />

    <AppFilterBarDrawer
      v-if="hasDrawerFilters || slots.drawer"
      v-model="drawerVisible"
      :title="resolvedDrawerTitle"
      :chip-groups="chipGroups"
      :show-chips="isMobile"
      :full-screen="fullScreenDrawer"
      :result-total="resultTotal"
      @apply="emit('apply')"
      @reset="onDrawerReset"
      @chip-remove="(payload) => emit('chip-remove', payload)"
      @chip-clear-all="emit('chip-clear-all')"
    >
      <slot name="drawer">
        <template v-for="item in drawerFilterNodeItems" :key="item.id">
          <div class="filter-drawer-field w-full max-w-full space-y-[6px]">
            <span
              v-if="item.showDrawerLabel"
              class="filter-drawer-field-label block text-[12px] font-semibold leading-[16px] text-[#475569]"
            >
              {{ item.label }}
            </span>
            <RenderNode :node="item.node" />
          </div>
        </template>
      </slot>
    </AppFilterBarDrawer>
  </div>
</template>
