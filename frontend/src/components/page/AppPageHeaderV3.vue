<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useRoute } from "vue-router";

interface Props {
  backLabel?: string;
  breadcrumb?: string;
  containerClass?: string;
  containerMaxWidth?: string;
  fullWidth?: boolean;
  hideActionsWhenMobileBottomBar?: boolean;
  title: string;
}

const props = withDefaults(defineProps<Props>(), {
  containerClass: "",
  containerMaxWidth: "max-w-[1200px]",
  fullWidth: false,
  hideActionsWhenMobileBottomBar: false,
});

const emit = defineEmits<{
  (event: "back"): void;
}>();

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

const containerClasses = computed(() => [
  props.fullWidth ? "max-w-full" : props.containerMaxWidth,
  props.containerClass,
]);

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
    class="sticky top-0 z-40 border-b border-[#e2e8f0] bg-white/95 px-[16px] py-[8px] shadow-[0px_2px_8px_0px_rgba(15,23,42,0.04)] backdrop-blur sm:px-[24px] lg:px-[24px]"
  >
    <div
      class="mx-auto flex flex-wrap items-center justify-between gap-[12px]"
      :class="containerClasses"
    >
      <div class="min-w-0 flex items-center gap-[12px]">
        <button
          v-if="backLabel"
          type="button"
          class="flex shrink-0 items-center gap-[6px] font-semibold text-[#64748b] transition-colors hover:text-[#dc2626]"
          @click="emit('back')"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="size-[14px]"
          >
            <path d="m12 19-7-7 7-7" />
            <path d="M19 12H5" />
          </svg>
          {{ backLabel }}
        </button>

        <div
          v-if="backLabel"
          class="hidden h-[28px] w-px bg-[#e2e8f0] sm:block"
        />

        <div class="min-w-0 flex items-center gap-[8px]">
          <h5
            class="truncate font-['Inter'] text-[15px] font-bold leading-[20px] text-[#1e293b]"
          >
            <slot name="title">
              {{ title }}
            </slot>
          </h5>
          <span
            v-if="breadcrumb"
            class="hidden truncate text-[11px] leading-[15px] text-[#94a3b8] sm:block"
          >
            {{ breadcrumb }}
          </span>
        </div>
      </div>

      <section v-if="!hasMobileBottomBar" class="flex items-center gap-[8px]">
        <slot name="actions" />
      </section>
    </div>
  </header>
</template>
