<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";

const props = withDefaults(
  defineProps<{
    disabled?: boolean;
    onRefresh?: () => Promise<unknown> | void;
    pullThreshold?: number;
  }>(),
  {
    disabled: false,
    onRefresh: undefined,
    pullThreshold: 70,
  }
);

const emit = defineEmits<{
  (e: "refresh"): void;
}>();

const containerRef = ref<HTMLElement | null>(null);
const pullDistance = ref(0);
const isRefreshing = ref(false);
const isPulling = ref(false);

let startY = 0;
let currentY = 0;
let activeScrollableElement: any = null;
let isTopReached = false;

const pullRatio = computed(() => {
  const ratio = pullDistance.value / props.pullThreshold;
  return Math.min(Math.max(ratio, 0), 1);
});

const indicatorStyle = computed(() => {
  const translateY = isRefreshing.value
    ? props.pullThreshold
    : Math.min(pullDistance.value, props.pullThreshold * 1.3);

  return {
    opacity: isRefreshing.value ? 1 : pullRatio.value,
    transform: `translate3d(0, ${translateY}px, 0)`,
    transition: isPulling.value
      ? "none"
      : "all 0.3s cubic-bezier(0.25, 1, 0.5, 1)",
  };
});

const contentStyle = computed(() => {
  const translateY = isRefreshing.value
    ? props.pullThreshold * 0.7
    : Math.min(pullDistance.value * 0.4, props.pullThreshold * 0.7);

  return {
    transform: `translate3d(0, ${translateY}px, 0)`,
    transition: isPulling.value
      ? "none"
      : "transform 0.3s cubic-bezier(0.25, 1, 0.5, 1)",
  };
});

const getScrollTop = (target: any) => {
  if (!target) return 0;
  if (target === window) {
    return window.scrollY || document.documentElement.scrollTop || 0;
  }

  return Number(target?.scrollTop || 0);
};

const isElementScrollable = (element: HTMLElement) => {
  const style = window.getComputedStyle(element);
  const overflowY = style.overflowY;
  const canScrollY = ["auto", "scroll", "overlay"].includes(overflowY);
  return canScrollY && element.scrollHeight - element.clientHeight > 2;
};
// eslint-disable-next-line no-undef
type ScrollTarget = HTMLElement | Window;

const findScrollableElement = (target: HTMLElement | null): ScrollTarget => {
  let current = target;

  while (current) {
    if (isElementScrollable(current)) {
      return current;
    }

    current = current.parentElement;
  }

  return window;
};
const isAtScrollTop = (target: any) => getScrollTop(target) <= 2;

// eslint-disable-next-line no-undef
const handleTouchStart = (e: TouchEvent) => {
  if (props.disabled || isRefreshing.value || !e.touches[0]) {
    return;
  }

  const target = e.target as HTMLElement | null;

  activeScrollableElement = findScrollableElement(target);

  if (!isAtScrollTop(activeScrollableElement)) {
    resetPullState();
    return;
  }

  isTopReached = true;
  startY = e.touches[0].clientY;
  currentY = startY;
};

// eslint-disable-next-line no-undef
const handleTouchMove = (e: TouchEvent) => {
  if (props.disabled || isRefreshing.value || !isTopReached || !e.touches[0])
    return;

  if (!isAtScrollTop(activeScrollableElement)) {
    isTopReached = false;
    isPulling.value = false;
    pullDistance.value = 0;
    activeScrollableElement = null;
    return;
  }

  currentY = e.touches[0].clientY;
  const dy = currentY - startY;

  if (dy > 0) {
    if (e.cancelable) e.preventDefault();
    isPulling.value = true;
    pullDistance.value = Math.pow(dy, 0.85);
  } else {
    isPulling.value = false;
    pullDistance.value = 0;
  }
};

const triggerRefresh = async () => {
  isRefreshing.value = true;
  isPulling.value = false;
  emit("refresh");

  try {
    if (props.onRefresh) {
      await props.onRefresh();
    } else {
      await new Promise((resolve) => setTimeout(resolve, 800));
    }
  } catch (err) {
    console.warn("Pull to refresh failed", err);
  } finally {
    isRefreshing.value = false;
    pullDistance.value = 0;
  }
};

const resetPullState = () => {
  isTopReached = false;
  activeScrollableElement = null;
  isPulling.value = false;
  pullDistance.value = 0;
};

const handleTouchEnd = () => {
  if (props.disabled || isRefreshing.value || !isPulling.value) {
    resetPullState();
    return;
  }

  if (pullDistance.value >= props.pullThreshold) {
    activeScrollableElement = null;
    void triggerRefresh();
  } else {
    resetPullState();
  }
};

onMounted(() => {
  const el = containerRef.value;
  if (!el) return;

  el.addEventListener("touchstart", handleTouchStart, { passive: true });
  el.addEventListener("touchmove", handleTouchMove, { passive: false });
  el.addEventListener("touchend", handleTouchEnd, { passive: true });
});

onUnmounted(() => {
  const el = containerRef.value;
  if (!el) return;

  el.removeEventListener("touchstart", handleTouchStart);
  el.removeEventListener("touchmove", handleTouchMove);
  el.removeEventListener("touchend", handleTouchEnd);
});

defineExpose({
  triggerRefresh,
});
</script>

<template>
  <div
    ref="containerRef"
    class="relative flex h-full min-h-0 min-w-0 flex-1 flex-col"
  >
    <!-- Refresh Indicator Header -->
    <div
      class="pointer-events-none absolute left-0 right-0 -top-12 z-[90] flex items-center justify-center py-2"
      :style="indicatorStyle"
    >
      <div
        class="flex h-10 w-10 items-center justify-center rounded-full bg-white text-primary-600 shadow-md ring-1 ring-slate-200/80 transition-all"
      >
        <span
          v-if="isRefreshing"
          class="pi pi-spinner pi-spin text-[16px] text-primary-600"
        />
        <span
          v-else
          class="pi pi-arrow-down text-[14px] text-slate-600 transition-transform duration-200"
          :style="{
            transform: `rotate(${pullRatio * 180}deg)`,
            color: pullRatio >= 1 ? '#dc2626' : '#64748b',
          }"
        />
      </div>
    </div>

    <!-- Wrapped Content -->
    <div
      :style="contentStyle"
      class="flex h-full min-h-0 min-w-0 flex-1 flex-col"
    >
      <slot />
    </div>
  </div>
</template>
