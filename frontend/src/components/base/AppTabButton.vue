<script setup lang="ts">
import { computed } from "vue";

import isMobile from "@/composables/useIsMobile";

type ButtonSize = "small" | "medium" | "large";

interface Props {
  active?: boolean;
  badge?: number | string;
  disabled?: boolean;
  size?: ButtonSize | null;
  type?: "button" | "submit" | "reset";
}

const props = withDefaults(defineProps<Props>(), {
  active: false,
  badge: undefined,
  disabled: false,
  size: null,
  type: "button",
});

const emit = defineEmits<{
  (e: "click", event: MouseEvent): void;
}>();

const actualSize = computed<ButtonSize>(() => {
  if (props.size !== null && !isMobile.value) {
    return props.size;
  }

  return isMobile.value ? "small" : "medium";
});

const sizeClasses: Record<ButtonSize, string> = {
  large: "h-[72px] rounded-[18px] px-5",
  medium: "h-[64px] rounded-[16px] px-4",
  small: "h-[52px] rounded-[14px] px-3",
};

const buttonClasses = computed(() => [
  "flex items-center border text-left shadow-sm transition-all focus:outline-none focus:ring-2 focus:ring-offset-0 disabled:cursor-not-allowed disabled:opacity-50",
  sizeClasses[actualSize.value],
  props.active
    ? "border-[#dc2626] bg-[#dc2626] text-white shadow-red-100 focus:ring-red-200"
    : "border-slate-200 bg-[#f5f8fa] text-slate-600 hover:border-slate-300 hover:bg-slate-100 focus:ring-slate-200",
]);

const badgeSizeClasses: Record<ButtonSize, string> = {
  large: "mr-3.5 size-9",
  medium: "mr-3 size-8",
  small: "mr-2.5 size-7",
};

const badgeClasses = computed(() => [
  "flex shrink-0 items-center justify-center rounded-full transition-colors",
  badgeSizeClasses[actualSize.value],
  props.active ? "bg-white/20 text-white" : "bg-white text-slate-500",
]);

const badgeTextClasses: Record<ButtonSize, string> = {
  large: "text-[13px]",
  medium: "text-[12px]",
  small: "text-[11px]",
};

const labelClasses: Record<ButtonSize, string> = {
  large: "text-[15px]",
  medium: "text-[14px]",
  small: "text-[13px]",
};

const onClick = (event: MouseEvent) => {
  if (!props.disabled) {
    emit("click", event);
  }
};
</script>

<template>
  <button
    :type="type"
    :disabled="disabled"
    :class="buttonClasses"
    @click="onClick"
  >
    <div v-if="$slots.badge || badge !== undefined" :class="badgeClasses">
      <slot name="badge">
        <span :class="[badgeTextClasses[actualSize], 'font-bold']">{{
          badge
        }}</span>
      </slot>
    </div>

    <span :class="[labelClasses[actualSize], 'font-semibold leading-tight']">
      <slot />
    </span>
  </button>
</template>
