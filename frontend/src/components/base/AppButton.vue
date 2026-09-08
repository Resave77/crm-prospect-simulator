<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";

import { useAuthStore } from "@/stores/authStore";

type ButtonVariant =
  | "primary"
  | "danger"
  | "secondary"
  | "ghost"
  | "info"
  | "success"
  | "custom"
  | "dangerSoft";
type ButtonSize = "small" | "medium" | "large";

interface Props {
  disabled?: boolean;
  form?: string;
  icon?: string;
  iconPos?: "left" | "right";
  loading?: boolean;
  permissions?: string[];
  size?: ButtonSize | null;
  type?: "button" | "submit" | "reset";
  variant?: ButtonVariant;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  form: undefined,
  iconPos: "left",
  loading: false,
  permissions: undefined,
  size: null,
  type: "button",
  variant: "primary",
});

const emit = defineEmits<{
  (e: "click", event: MouseEvent): void;
}>();

const authStore = useAuthStore();
const isCompactViewport = ref(false);

// Detect if viewport should use compact button sizing
const checkCompactViewport = () => {
  isCompactViewport.value = window.innerWidth <= 1500;
};

onMounted(() => {
  checkCompactViewport();
  window.addEventListener("resize", checkCompactViewport);
});

onUnmounted(() => {
  window.removeEventListener("resize", checkCompactViewport);
});

// Compute actual size based on viewport and prop
const actualSize = computed<ButtonSize>(() => {
  if (isCompactViewport.value) {
    return "small";
  }

  // If size is explicitly provided, use it on larger viewports
  if (props.size !== null) {
    return props.size;
  }

  return "medium";
});

// Check if user has at least one of the required permissions
const hasPermission = computed(() => {
  // If no permissions specified, always render
  if (!props.permissions || props.permissions.length === 0) {
    return true;
  }
  // Check if user has at least one of the required permissions
  return props.permissions.some((perm) => authStore.permissions.includes(perm));
});

const baseClasses =
  "inline-flex items-center justify-center border font-semibold tracking-normal shadow-[0_1px_2px_rgba(15,23,42,0.08)] transition-all duration-150 ease-out [-webkit-tap-highlight-color:transparent] hover:-translate-y-px active:translate-y-0 active:scale-[0.99] active:shadow-[0_1px_1px_rgba(15,23,42,0.08)] focus:outline-none focus:ring-1 focus:ring-offset-0 disabled:translate-y-0 disabled:scale-100 disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none";

const variantClasses: Record<ButtonVariant, string> = {
  custom: "",
  danger:
    "border-red-600 bg-red-600 text-white shadow-[0_3px_8px_rgba(220,38,38,0.14)] hover:border-red-700 hover:bg-red-700 hover:shadow-[0_4px_10px_rgba(220,38,38,0.16)] focus:ring-red-100",
  dangerSoft:
    "border-[rgba(221,0,51,0.16)] bg-[rgba(221,0,51,0.08)] text-[#d03] shadow-[0_1px_3px_rgba(221,0,51,0.08)] hover:bg-[rgba(221,0,51,0.14)] hover:shadow-[0_2px_6px_rgba(221,0,51,0.1)] focus:ring-[rgba(221,0,51,0.16)]",
  ghost:
    "border-transparent bg-transparent text-slate-600 shadow-none hover:border-slate-200 hover:bg-slate-50 hover:text-slate-900 focus:ring-slate-100",
  info: "border-blue-400 bg-blue-400 text-white shadow-[0_3px_8px_rgba(96,165,250,0.14)] hover:border-blue-500 hover:bg-blue-500 hover:shadow-[0_4px_10px_rgba(96,165,250,0.16)] focus:ring-blue-100",
  primary:
    "border-primary-600 bg-primary-600 text-white shadow-[0_3px_8px_rgba(220,38,38,0.14)] hover:border-primary-700 hover:bg-primary-700 hover:shadow-[0_4px_10px_rgba(220,38,38,0.16)] focus:ring-primary-100",
  secondary:
    "border-slate-200 bg-white text-slate-700 shadow-[0_1px_3px_rgba(15,23,42,0.06)] hover:border-slate-300 hover:bg-slate-50 hover:text-slate-900 hover:shadow-[0_2px_6px_rgba(15,23,42,0.08)] focus:ring-slate-100",
  success:
    "border-green-500 bg-green-500 text-white shadow-[0_3px_8px_rgba(34,197,94,0.14)] hover:border-green-600 hover:bg-green-600 hover:shadow-[0_4px_10px_rgba(34,197,94,0.16)] focus:ring-green-100",
};

const sizeClasses: Record<ButtonSize, string> = {
  large: "gap-2.5 rounded-lg px-4 py-2.5 text-sm",
  medium: "gap-2 rounded-lg px-3.5 py-2 text-[13px]",
  small: "gap-1.5 rounded-md px-2.5 py-1.5 text-[11.5px]",
};

const iconSizeClasses: Record<ButtonSize, string> = {
  large: "text-base",
  medium: "text-sm",
  small: "text-xs",
};

const buttonClasses = computed(() => [
  baseClasses,
  variantClasses[props.variant],
  sizeClasses[actualSize.value],
]);

const onClick = (event: MouseEvent) => {
  if (!props.disabled && !props.loading) {
    emit("click", event);
  }
};
</script>

<template>
  <button
    v-if="hasPermission"
    :type="type"
    :form="form"
    :class="buttonClasses"
    :disabled="disabled || loading"
    @click="onClick"
  >
    <!-- Loading spinner -->
    <i
      v-if="loading"
      class="pi pi-spinner animate-spin"
      :class="iconSizeClasses[actualSize]"
    />

    <!-- Left icon -->
    <i
      v-else-if="icon && iconPos === 'left'"
      class="pi"
      :class="[icon, iconSizeClasses[actualSize], 'shrink-0']"
    />

    <!-- Slot content -->
    <slot />

    <!-- Right icon -->
    <i
      v-if="icon && iconPos === 'right' && !loading"
      class="pi"
      :class="[icon, iconSizeClasses[actualSize]]"
    />
  </button>
</template>
