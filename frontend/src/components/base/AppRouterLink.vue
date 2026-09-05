<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router";

import { computed, useAttrs } from "vue";

import isMobile from "@/composables/useIsMobile";
import { useAuthStore } from "@/stores/authStore";

type LinkSize = "small" | "medium" | "large";

defineOptions({
  inheritAttrs: false,
});

const props = withDefaults(
  defineProps<{
    custom?: boolean;
    disabled?: boolean;
    permissions?: string[];
    renderSlot?: boolean;
    replace?: boolean;
    size?: LinkSize | null;
    to: RouteLocationRaw;
    visibility?: boolean;
  }>(),
  {
    custom: false,
    disabled: false,
    permissions: undefined,
    renderSlot: false,
    replace: false,
    size: null,
    visibility: false,
  }
);

const authStore = useAuthStore();
const attrs = useAttrs();
const isDisabled = computed(() => props.disabled === true);

const actualSize = computed<LinkSize>(() => {
  if (props.size !== null && !isMobile.value) {
    return props.size;
  }

  return isMobile.value ? "small" : "medium";
});

const hasPermission = computed(() => {
  if (!props.permissions || props.permissions.length === 0) return true;
  return props.permissions.some((perm) => authStore.permissions.includes(perm));
});

const sizeClasses: Record<LinkSize, string> = {
  large: "text-base",
  medium: "text-sm",
  small: "text-xs",
};

const linkClasses = computed(() => [
  sizeClasses[actualSize.value],
  attrs.class,
]);

const rootAttrs = computed(() => {
  const { class: _class, ...rest } = attrs;
  return rest;
});
</script>
<template>
  <!-- NORMAL LINK -->
  <RouterLink
    v-if="hasPermission && !isDisabled"
    v-bind="rootAttrs"
    :to="to"
    :replace="replace"
    :custom="custom"
    :class="linkClasses"
  >
    <slot :size="actualSize" :actual-size="actualSize" :is-mobile="isMobile" />
  </RouterLink>

  <!-- DISABLED / NO PERMISSION -->
  <span
    v-else-if="renderSlot || visibility || isDisabled"
    v-bind="rootAttrs"
    :class="[linkClasses, 'router-link--disabled']"
    aria-disabled="true"
  >
    <slot :size="actualSize" :actual-size="actualSize" :is-mobile="isMobile" />
  </span>
</template>

<style scoped>
/* .router-link--disabled {
  pointer-events: none;
  cursor: not-allowed;
  opacity: 0.5;
} */
</style>
