<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router";

import {
  computed,
  onMounted,
  onUnmounted,
  provide,
  ref,
  shallowRef,
} from "vue";
import { useRoute } from "vue-router";

import AppRouterLink from "@/components/base/AppRouterLink.vue";
import AppFilterBarChipsV1 from "@/components/filter/v1/AppFilterBarChipsV1.vue";

import {
  appPageHeaderV4FilterContextKey,
  type AppPageHeaderV4FilterRegistration,
} from "./AppPageHeaderV4FilterContext";

export type AppPageHeaderV4Tab = {
  key: string;
  label: string;
  permissions?: string[];
  to: RouteLocationRaw;
};

const props = withDefaults(
  defineProps<{
    activeTab?: string;
    hideActionsWhenMobileBottomBar?: boolean;
    hideOnMobile?: boolean;
    hideTabsWhenMobileBottomBar?: boolean;
    tabs?: AppPageHeaderV4Tab[];
  }>(),
  {
    activeTab: "",
    hideActionsWhenMobileBottomBar: false,
    hideOnMobile: false,
    hideTabsWhenMobileBottomBar: false,
    tabs: () => [],
  }
);

const emit = defineEmits<{
  (event: "tab-change", value: string): void;
}>();

const route = useRoute();
const isCompactViewport = ref(
  typeof window !== "undefined" && window.innerWidth < 1024
);
const isMobileTabsOpen = ref(false);
const filterRegistrations = shallowRef<AppPageHeaderV4FilterRegistration[]>([]);

const updateViewport = () => {
  isCompactViewport.value =
    typeof window !== "undefined" && window.innerWidth < 1024;
};

const hasMobileBottomBar = computed(
  () =>
    props.hideActionsWhenMobileBottomBar &&
    Boolean(route.meta.mobileBottomBar) &&
    isCompactViewport.value
);

const hideTabsForMobileBottomBar = computed(
  () =>
    props.hideTabsWhenMobileBottomBar &&
    Boolean(route.meta.mobileBottomBar) &&
    isCompactViewport.value
);

const isHidden = computed(() => props.hideOnMobile && isCompactViewport.value);

const activeTabLabel = computed(
  () =>
    props.tabs.find((tab) => tab.key === props.activeTab)?.label ?? "Sub Menu"
);
const activeFilterRegistration = computed(
  () => filterRegistrations.value[filterRegistrations.value.length - 1] ?? null
);
const activeFilterChipGroups = computed(
  () => activeFilterRegistration.value?.chipGroups.value ?? []
);
const hasActiveFilterChips = computed(() =>
  activeFilterChipGroups.value.some((group) => group.items.length > 0)
);

provide(appPageHeaderV4FilterContextKey, {
  register: (registration) => {
    filterRegistrations.value = [...filterRegistrations.value, registration];

    return () => {
      filterRegistrations.value = filterRegistrations.value.filter(
        (item) => item !== registration
      );
    };
  },
});

const onTabClick = (key: string) => {
  emit("tab-change", key);
  isMobileTabsOpen.value = false;
};

onMounted(() => {
  updateViewport();
  window.addEventListener("resize", updateViewport);
});

onUnmounted(() => {
  window.removeEventListener("resize", updateViewport);
});
</script>

<template>
  <header v-if="!isHidden" class="mb-1 px-2 pt-1 md:mb-2 md:pt-2 text-body">
    <div
      class="flex w-full min-w-0 flex-col gap-1 rounded-none border-0 bg-transparent px-1.5 py-1 shadow-none md:rounded-2xl md:border md:border-slate-200 md:bg-white md:px-3 md:py-2 md:shadow-[0_8px_24px_rgba(15,23,42,0.04)] md:flex-row md:flex-wrap md:items-center md:gap-3"
    >
      <div
        class="flex min-w-0 flex-1 flex-col gap-1 border-t border-slate-100 pt-1 md:flex-row md:items-center md:justify-start md:gap-3 md:border-t-0 md:pt-0"
        :class="{
          'border-t-0 pt-0':
            (!$slots.tabs && tabs.length === 0) || hideTabsForMobileBottomBar,
        }"
      >
        <section
          v-if="!hideTabsForMobileBottomBar && ($slots.tabs || tabs.length)"
          class="min-w-0 shrink-0 md:flex-[0_1_auto] md:shrink md:grow-0"
        >
          <slot name="tabs">
            <div
              class="hidden min-w-0 overflow-x-auto overflow-y-hidden md:block"
            >
              <div class="inline-flex min-w-max flex-nowrap gap-1.5 pb-0.5">
                <AppRouterLink
                  v-for="tab in tabs"
                  :key="tab.key"
                  :to="tab.to"
                  :permissions="tab.permissions"
                  class="shrink-0"
                >
                  <button
                    type="button"
                    class="rounded-lg px-3.5 py-2 text-xs font-semibold whitespace-nowrap transition"
                    :class="
                      activeTab === tab.key
                        ? 'bg-[#dc2626] text-white shadow-sm'
                        : 'bg-slate-100 text-slate-600 hover:bg-slate-200 hover:text-slate-900'
                    "
                    @click="onTabClick(tab.key)"
                  >
                    {{ tab.label }}
                  </button>
                </AppRouterLink>
              </div>
            </div>

            <div class="relative md:hidden">
              <button
                type="button"
                class="inline-flex max-w-full items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 shadow-sm transition hover:bg-slate-50"
                @click="isMobileTabsOpen = !isMobileTabsOpen"
              >
                <span class="pi pi-bars text-sm" />
                <span class="truncate">{{ activeTabLabel }}</span>
                <span
                  class="pi pi-chevron-down text-[10px] transition-transform"
                  :class="{ 'rotate-180': isMobileTabsOpen }"
                />
              </button>

              <Transition
                enter-active-class="transition duration-200 ease-out"
                enter-from-class="-translate-y-2 opacity-0"
                enter-to-class="translate-y-0 opacity-100"
                leave-active-class="transition duration-150 ease-in"
                leave-from-class="translate-y-0 opacity-100"
                leave-to-class="-translate-y-2 opacity-0"
              >
                <div
                  v-if="isMobileTabsOpen"
                  class="absolute left-0 top-[calc(100%+8px)] z-50 w-[min(20rem,calc(100vw-2rem))] rounded-xl border border-slate-200 bg-white p-1.5 shadow-xl"
                >
                  <AppRouterLink
                    v-for="tab in tabs"
                    :key="tab.key"
                    :to="tab.to"
                    :permissions="tab.permissions"
                    class="block"
                  >
                    <button
                      type="button"
                      class="w-full rounded-lg px-3 py-2.5 text-left text-xs font-semibold transition"
                      :class="
                        activeTab === tab.key
                          ? 'bg-red-50 text-[#dc2626]'
                          : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
                      "
                      @click="onTabClick(tab.key)"
                    >
                      {{ tab.label }}
                    </button>
                  </AppRouterLink>
                </div>
              </Transition>

              <button
                v-if="isMobileTabsOpen"
                type="button"
                class="fixed inset-0 z-40 cursor-default bg-transparent"
                aria-label="Close submenu"
                @click="isMobileTabsOpen = false"
              />
            </div>
          </slot>
        </section>

        <section class="min-w-0 flex-1">
          <slot name="filter" />
        </section>
      </div>

      <aside
        v-if="!hasMobileBottomBar"
        class="flex shrink-0 flex-wrap items-center justify-start gap-2 md:ml-auto md:justify-end"
      >
        <slot name="actions" />
      </aside>

      <AppFilterBarChipsV1
        v-if="hasActiveFilterChips && !isCompactViewport"
        class="w-full min-w-0 md:basis-full"
        :groups="activeFilterChipGroups"
        @remove="(payload) => activeFilterRegistration?.onRemove(payload)"
        @clear-all="activeFilterRegistration?.onClearAll()"
      />
    </div>
  </header>
</template>
