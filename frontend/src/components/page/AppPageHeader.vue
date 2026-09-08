<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useRoute } from "vue-router";

interface Props {
  hideActionsWhenMobileBottomBar?: boolean;
  section?: string;
  subtitle?: string;
  title: string;
}

const props = withDefaults(defineProps<Props>(), {
  hideActionsWhenMobileBottomBar: false,
});

const route = useRoute();
const isCompactViewport = ref(
  typeof window !== "undefined" && window.innerWidth < 1024
);

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

onMounted(() => {
  if (typeof window !== "undefined") {
    updateViewport();
    window.addEventListener("resize", updateViewport);
  }
});

onUnmounted(() => {
  if (typeof window !== "undefined") {
    window.removeEventListener("resize", updateViewport);
  }
});
</script>

<template>
  <header
    class="flex-col gap-4 rounded-md pt-2 ps-2 pe-2 md:flex md:flex-row md:items-center md:justify-between"
  >
    <div class="space-y-1">
      <p
        v-if="section"
        class="hidden text-navbar font-semibold uppercase tracking-wide text-slate-400 md:block"
        style="font-size: calc(var(--font-body) * 0.9)"
      >
        {{ section }}
      </p>
      <div class="hidden flex-wrap items-center gap-2 md:flex">
        <h2
          class="text-heading font-bold text-slate-900"
          style="font-size: calc(var(--font-heading) * 1.15)"
        >
          {{ title }}
        </h2>
        <slot name="title-suffix" />
      </div>
      <p v-if="subtitle" class="hidden text-sm text-primary-700 md:block">
        {{ subtitle }}
      </p>
    </div>

    <div v-if="!hasMobileBottomBar" class="flex flex-wrap items-center gap-2">
      <slot name="filter-button" />
      <slot name="actions" />
    </div>
  </header>

  <!-- Mobile: floating action buttons at bottom only -->
  <!-- <div
    class="fixed bottom-0 left-0 right-0 z-50 flex items-center justify-end gap-2 bg-white/95 p-4 shadow-lg backdrop-blur-sm md:hidden"
  >
    <slot name="filter-button" />
    <slot name="actions" />
  </div> -->
</template>
