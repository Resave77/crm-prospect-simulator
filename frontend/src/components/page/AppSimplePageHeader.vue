<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    backLabel?: string;
    subtitle?: string | null;
    title: string;
  }>(),
  {
    backLabel: "Back",
    subtitle: null,
  }
);

const emit = defineEmits<{
  (event: "back"): void;
}>();
</script>

<template>
  <header
    class="sticky top-0 z-30 border-b border-[#e2e8f0] bg-white/95 px-3 pb-2 pt-2 shadow-[0_2px_10px_rgba(15,23,42,0.05)] backdrop-blur-xl lg:hidden"
  >
    <div class="flex min-h-11 items-center justify-between gap-2">
      <div class="flex min-w-0 flex-1 items-center gap-2">
        <button
          type="button"
          class="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl border border-[#e2e8f0] bg-white text-[#334155] shadow-sm transition active:scale-[0.98]"
          :aria-label="props.backLabel"
          @click="emit('back')"
        >
          <span class="pi pi-arrow-left text-[12px] leading-none" />
        </button>

        <div class="min-w-0 flex-1">
          <h1
            class="app-mobile-page-header-title truncate font-bold text-[#0f172a]"
          >
            {{ props.title }}
          </h1>
          <slot name="subtitle">
            <p
              v-if="props.subtitle"
              class="app-mobile-page-header-subtitle mt-0.5 truncate font-semibold text-[#64748b]"
            >
              {{ props.subtitle }}
            </p>
          </slot>
        </div>
      </div>

      <div class="flex shrink-0 items-center gap-1.5">
        <slot name="actions" />
      </div>
    </div>
  </header>
</template>

<style scoped>
.app-mobile-page-header-title {
  font-size: 0.7875rem !important;
  line-height: 0.875rem !important;
}

.app-mobile-page-header-subtitle {
  font-size: 0.6625rem !important;
  line-height: 0.75rem !important;
}
</style>
