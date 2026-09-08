<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

import type { DialogMobileBottomBarAction } from "@/composables/useDialogMobileBottomBar";

import AppDialog from "@/components/base/AppDialog.vue";

import type { FilterBarChipGroupV1 } from "./AppFilterBarChipsV1.vue";

import AppFilterBarChipsV1 from "./AppFilterBarChipsV1.vue";

interface Props {
  chipGroups?: FilterBarChipGroupV1[];
  fullScreen?: boolean;
  modelValue: boolean;
  resultTotal?: number | null;
  showChips?: boolean;
  title?: string;
}

const props = withDefaults(defineProps<Props>(), {
  chipGroups: () => [],
  fullScreen: false,
  resultTotal: undefined,
  showChips: true,
  title: "",
});

const resolvedTitle = computed(() => props.title || "Filter");
const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "apply"): void;
  (e: "reset"): void;
  (e: "chip-remove", payload: { key: string; value: number | string }): void;
  (e: "chip-clear-all"): void;
}>();

const hasActiveFilters = computed(() =>
  props.chipGroups.some((group) => group.items.length > 0)
);

const activeFilterCount = computed(() =>
  props.chipGroups.reduce((count, group) => count + group.items.length, 0)
);
const showResultInfo = computed(() => typeof props.resultTotal === "number");
const resultSummaryText = computed(() =>
  !showResultInfo.value
    ? ""
    : `Showing ${props.resultTotal!.toLocaleString("en-US")} results`
);

const close = () => {
  emit("update:modelValue", false);
};

const onApply = () => {
  emit("apply");
  close();
};

const onReset = () => {
  emit("reset");
};

const isMobile = ref(false);

const checkMobile = () => {
  isMobile.value = typeof window !== "undefined" && window.innerWidth < 768;
};

const drawerClasses = computed(() =>
  props.fullScreen || isMobile.value
    ? "filter-drawer-panel filter-drawer-panel--fullscreen fixed inset-0 z-[1000] flex min-h-0 flex-col overflow-hidden border border-slate-200 bg-white shadow-xl"
    : "filter-drawer-panel fixed right-0 top-0 bottom-0 z-[1000] flex min-h-0 w-96 max-w-[85vw] flex-col overflow-hidden rounded-l-2xl border-l border-slate-200 bg-white shadow-xl"
);

const handleEsc = (e: KeyboardEvent) => {
  if (e.key === "Escape") {
    close();
  }
};

onMounted(() => {
  checkMobile();
  window.addEventListener("resize", checkMobile);
  window.addEventListener("keydown", handleEsc);
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", checkMobile);
  window.removeEventListener("keydown", handleEsc);
});

const dialogBottomBarActions = computed<DialogMobileBottomBarAction[]>(() => [
  {
    icon: "pi-filter-slash",
    key: "reset",
    label: "Reset",
    onClick: onReset,
    type: "button",
  },
  {
    icon: "pi-check",
    key: "apply",
    label: "Apply",
    onClick: onApply,
    type: "button",
  },
]);
</script>

<template>
  <template v-if="!isMobile">
    <Teleport to="body">
      <!-- Backdrop -->
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[999] bg-black/30"
        @click="close"
      />

      <!-- Filter Drawer -->
      <div v-if="modelValue" :class="drawerClasses">
        <!-- Header -->
        <div
          class="sticky top-0 z-10 flex items-center justify-between border-b border-slate-100 bg-white px-4 py-3 sm:px-6"
        >
          <div class="flex items-center gap-3">
            <span class="pi pi-filter text-primary-600" />
            <div>
              <h2 class="text-heading text-slate-900">{{ resolvedTitle }}</h2>
              <p v-if="hasActiveFilters" class="text-xs text-slate-500">
                {{ activeFilterCount }} active filters
              </p>
            </div>
          </div>
          <button
            type="button"
            class="flex h-8 w-8 items-center justify-center rounded-full text-slate-400 transition hover:bg-slate-100 hover:text-slate-600"
            @click="close"
          >
            <span class="pi pi-times" />
          </button>
        </div>

        <!-- Filter Content Slot -->
        <div
          class="filter-drawer min-h-0 flex-1 overflow-y-auto p-3 pb-[calc(env(safe-area-inset-bottom)+5rem)]"
        >
          <slot />
        </div>

        <!-- Active Filters Chips -->
        <div
          v-if="showChips && hasActiveFilters && chipGroups.length > 0"
          class="filter-drawer-chips shrink-0 border-t border-slate-100 bg-slate-50/50 px-4 py-3 sm:px-6"
        >
          <AppFilterBarChipsV1
            :groups="chipGroups"
            @remove="(payload) => emit('chip-remove', payload)"
            @clear-all="emit('chip-clear-all')"
          />
          <p v-if="showResultInfo" class="mt-2 text-xs text-slate-500">
            {{ resultSummaryText }}
          </p>
        </div>

        <!-- Footer Actions -->
        <div
          class="filter-drawer-footer relative z-10 shrink-0 border-t border-slate-200 bg-white px-4 py-3 shadow-[0_-8px_18px_rgba(15,23,42,0.05)] sm:px-6"
        >
          <div class="flex items-center justify-end gap-3">
            <button
              type="button"
              class="text-button rounded-lg border border-slate-200 px-4 py-2 font-medium text-slate-600 transition hover:bg-slate-50"
              @click="onReset"
            >
              Reset
            </button>
            <button
              type="button"
              class="text-button rounded-lg bg-primary-600 px-4 py-2 font-medium text-white shadow-sm transition hover:bg-primary-700"
              @click="onApply"
            >
              Apply Filters
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </template>

  <!-- Mobile Dialog Layout -->
  <AppDialog
    v-else
    class="filter-drawer-dialog"
    :model-value="modelValue"
    :title="resolvedTitle"
    :mobile-actions="dialogBottomBarActions"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div
      class="filter-drawer filter-drawer--mobile flex flex-col gap-4 pb-[calc(env(safe-area-inset-bottom)+12px)]"
    >
      <!-- Active Filters Chips -->
      <div
        v-if="showChips && hasActiveFilters && chipGroups.length > 0"
        class="shrink-0 rounded-xl bg-slate-50/50 p-3 border border-slate-100"
      >
        <AppFilterBarChipsV1
          :groups="chipGroups"
          @remove="(payload) => emit('chip-remove', payload)"
          @clear-all="emit('chip-clear-all')"
        />
        <p v-if="showResultInfo" class="mt-2 text-xs text-slate-500">
          {{ resultSummaryText }}
        </p>
      </div>

      <slot />
    </div>
  </AppDialog>
</template>

<style scoped>
:deep(.filter-drawer-dialog.p-dialog) {
  z-index: 1000 !important;
}

:global(.p-dialog-mask:has(.filter-drawer-dialog)) {
  z-index: 999 !important;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-right-enter-active,
.slide-right-leave-active {
  transition:
    transform 0.3s ease,
    opacity 0.3s ease;
}

.slide-right-enter-from,
.slide-right-leave-to {
  transform: translateX(100%);
  opacity: 0;
}
/* --- PERBAIKAN PADA SELURUH BLOK CSS INI --- */

.filter-drawer {
  /* Default Desktop (Sidebar w-96): Susun vertikal ke bawah */
  display: flex;
  flex-direction: column;
  gap: 1.25rem; /* Jarak antar filter yang konsisten */
}

.filter-drawer :deep(> *),
.filter-drawer :deep(.filter-drawer-field),
.filter-drawer :deep(.filter-drawer-field > *) {
  width: 100% !important;
  max-width: 100% !important;
}

.filter-drawer :deep(.filter-input),
.filter-drawer :deep(.filter-select),
.filter-drawer :deep(.p-select),
.filter-drawer :deep(.p-multiselect),
.filter-drawer :deep(input),
.filter-drawer :deep(.p-inputtext),
.filter-drawer :deep(input[type="number"]),
.filter-drawer :deep(input[type="range"]) {
  width: 100% !important;
  max-width: 100% !important;
}

.filter-drawer--mobile :deep(.w-\[180px\]),
.filter-drawer--mobile :deep(.md\:w-\[180px\]),
.filter-drawer--mobile :deep(.md\:max-w-full) {
  width: 100% !important;
  max-width: 100% !important;
}

/* JIKA dalam mode Fullscreen di Desktop, baru gunakan layout Grid ke samping */
@media (min-width: 768px) {
  .filter-drawer-panel--fullscreen .filter-drawer {
    display: grid;
    /* Gunakan auto-fill agar filter tidak melar jika jumlahnya sedikit */
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    /* Gunakan stretch/start agar tinggi box filter seragam, tidak hancur karena flex-end */
    align-items: stretch;
    gap: 1rem;
  }
}

@media (min-width: 1024px) {
  .filter-drawer-panel--fullscreen .filter-drawer {
    grid-template-columns: repeat(4, 1fr);
  }
}

/* Mobile adjustments (Tetap pertahankan logika mobile Anda) */
@media (max-width: 640px) {
  .filter-drawer {
    /* Kembalikan ke flex-column/grid sesuai kebutuhan mobile Anda */
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .filter-drawer-panel:not(.filter-drawer-panel--fullscreen) {
    top: auto;
    right: 0.75rem;
    bottom: calc(env(safe-area-inset-bottom) + 4.75rem);
    left: 0.75rem;
    width: auto;
    max-width: none;
    height: min(78dvh, calc(100dvh - 5.75rem));
    max-height: min(78dvh, calc(100dvh - 5.75rem));
    border: 1px solid #e2e8f0;
    border-radius: 1rem;
  }

  .filter-drawer-panel--fullscreen {
    padding-bottom: env(safe-area-inset-bottom);
  }

  .filter-drawer-footer {
    padding-bottom: calc(env(safe-area-inset-bottom) + 0.75rem);
  }

  .filter-drawer-chips {
    max-height: 18dvh;
    overflow-y: auto;
  }

  .filter-drawer-footer > div {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  }

  .filter-drawer-footer button {
    width: 100%;
  }

  .slide-right-enter-from,
  .slide-right-leave-to {
    transform: translateY(0.75rem);
    opacity: 1;
  }

  /* Mengandalkan selector bawaan panel untuk merestrukturisasi grid anak di mobile */
  .filter-drawer-panel :deep(.grid) {
    grid-template-columns: 1fr !important;
    gap: 0.75rem !important;
  }
}
</style>
