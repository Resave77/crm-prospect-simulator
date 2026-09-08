<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();
const props = withDefaults(
  defineProps<{
    hideActionsWhenMobileBottomBar?: boolean;
  }>(),
  {
    hideActionsWhenMobileBottomBar: false,
  }
);
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
  updateViewport();
  window.addEventListener("resize", updateViewport);
});

onUnmounted(() => {
  window.removeEventListener("resize", updateViewport);
});
</script>

<template>
  <header
    class="mb-1 flex flex-col gap-3 px-2 md:flex-row md:items-start md:justify-between md:gap-10"
    :class="hasMobileBottomBar ? 'pt-2' : 'pt-2'"
  >
    <section class="min-w-0 flex-1">
      <slot name="section" />
    </section>

    <aside
      v-if="!hasMobileBottomBar"
      class="flex w-full shrink-0 flex-wrap items-center justify-start gap-2 md:w-auto md:justify-end"
    >
      <slot name="actions" />
    </aside>
  </header>
</template>
