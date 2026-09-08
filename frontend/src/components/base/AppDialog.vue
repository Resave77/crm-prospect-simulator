<!--
  AppDialog Mobile Bottom Bar Usages:

  A. Slots:
     1. #bottom-bar:
        - src/modules/billing/invoice-mapping/components/ActionDialog/CustomDialog.vue (delegates to AppDialog)
        - src/modules/billing/invoice-mapping/pages/Index.vue (via CustomDialog wrapper)

  B. Props (:mobile-actions / mobileActions):
     - src/components/filter/v1/AppFilterAsyncSelectNestedV1.vue
     - src/components/filter/v1/AppFilterAsyncSelectV1.vue
     - src/components/filter/v1/AppFilterBarDrawer.vue
     - src/modules/billing/do-delivered/components/DialogCreateInvoice.vue
     - src/modules/billing/invoice-list/components/DialogCreateReceipt.vue
     - src/modules/billing/receipt/components/DialogCreateBLR.vue
     - src/modules/billing/shared/components/blr/DialogDetailBlr.vue
     - src/modules/billing/shared/components/blr/DialogEditBlr.vue
     - src/modules/billing/shared/components/invoice/DialogDetailInvoice.vue
     - src/modules/billing/shared/components/invoice/DialogEditInvoice.vue
     - src/modules/billing/shared/components/receipt/DialogDetailReceipt.vue
     - src/modules/billing/shared/components/receipt/DialogEditReceipt.vue
     - src/modules/order/sales-order/components/SalesOrderAddPackSupportDialog.vue
     - src/modules/order/sales-order/components/SalesOrderAddProductDialog.vue
     - src/modules/product-pricing/product-to-customer/components/ProductPricingCustomerPricingDialog.vue

  C. Props (:mobile-stepper / mobileStepper):
     - src/modules/product-pricing/customer-to-product/components/ProductPricingCustToProductCreateDialog.vue
-->
<script setup lang="ts">
import Dialog from "primevue/dialog";
import { computed, onBeforeUnmount, useSlots, watch } from "vue";

import type {
  DialogMobileBottomBarAction,
  DialogMobileStepperConfig,
} from "@/composables/useDialogMobileBottomBar";

import { registerDialogBackHandler } from "@/composables/useDialogBackStack";
import defaultMobileLayout from "@/composables/useIsMobile";

import AppDialogMobileBottomBar from "./AppDialogMobileBottomBar.vue";

const props = withDefaults(
  defineProps<{
    breakpoints?: Record<string, string>;
    closable?: boolean;
    closeOnBack?: boolean;
    dismissableMask?: boolean;
    footerCompact?: boolean;
    height?: string;
    maxBodyHeight?: string;
    maxHeight?: string;
    maxWidth?: string;
    mobileActions?: DialogMobileBottomBarAction[];
    mobileLayout?: boolean;
    mobileStepper?: DialogMobileStepperConfig;
    modelValue: boolean;
    position?: "center" | "top" | "bottom" | "left" | "right";
    subtitle?: string;
    tagsLeft?: string[];
    tagsRight?: string[];
    title?: string;
    width?: string;
  }>(),
  {
    breakpoints: () => ({}),
    closable: true,
    closeOnBack: true,
    dismissableMask: true,
    footerCompact: true,
    height: undefined,
    maxBodyHeight: undefined,
    maxHeight: undefined,
    maxWidth: undefined,
    mobileActions: () => [],
    mobileLayout: undefined,
    mobileStepper: undefined,
    position: "center",
    subtitle: "",
    tagsLeft: () => [],
    tagsRight: () => [],
    title: "",
    width: undefined,
  }
);

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "show"): void;
  (e: "hide"): void;
}>();

const slots = useSlots();
const hasHeaderContent = computed(
  () => !!props.title || !!props.subtitle || !!slots.header
);

const isMobileLayout = computed(
  () => props.mobileLayout ?? defaultMobileLayout.value
);

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", val),
});

let unregisterBackHandler: (() => void) | null = null;

const unregisterDialogBackHandler = () => {
  unregisterBackHandler?.();
  unregisterBackHandler = null;
};

const syncDialogBackHandler = () => {
  if (!props.closeOnBack || !visible.value) {
    unregisterDialogBackHandler();
    return;
  }

  if (unregisterBackHandler) return;

  unregisterBackHandler = registerDialogBackHandler(() => {
    emit("update:modelValue", false);
  });
};

watch(
  () => [props.closeOnBack, props.modelValue] as const,
  syncDialogBackHandler,
  { immediate: true }
);

onBeforeUnmount(unregisterDialogBackHandler);

const handleHide = () => {
  unregisterDialogBackHandler();
  emit("update:modelValue", false);
  emit("hide");
};

const handleShow = () => {
  emit("show");
};
</script>

<template>
  <Dialog
    v-model:visible="visible"
    append-to="body"
    modal
    :dismissable-mask="dismissableMask"
    :draggable="false"
    :position="position"
    :breakpoints="breakpoints === undefined ? undefined : breakpoints"
    :closable="false"
    :pt="{
      content: {
        class: 'p-0 outline-none focus:outline-none focus-visible:outline-none',
      },
      header: { class: 'hidden' },
      footer: { class: 'p-0' },
      mask: {
        class:
          !isMobileLayout && position === 'top' ? 'app-dialog-mask-top' : '',
      },
      root: {
        class: [
          isMobileLayout
            ? 'rounded-none overflow-hidden'
            : 'rounded-2xl overflow-hidden',
          'outline-none focus:outline-none focus-visible:outline-none focus:ring-0 focus-visible:ring-0 border-0',
        ],
      },
    }"
    :class="[
      'app-dialog-fullscreen w-full h-full min-w-0 min-h-0',
      isMobileLayout ? 'app-dialog-mobile-layout rounded-none' : 'rounded-2xl',
      position === 'top' ? 'app-dialog-position-top' : '',
    ]"
    :style="{
      '--app-dialog-width': width ?? undefined,
      '--app-dialog-max-width': maxWidth ?? undefined,
      '--app-dialog-height': height ?? undefined,
      '--app-dialog-max-height': maxHeight ?? undefined,
    }"
    @hide="handleHide"
    @show="handleShow"
  >
    <div
      class="flex h-full min-h-0 w-full min-w-0 flex-1 flex-col overflow-hidden text-body"
      :style="{ maxHeight: maxBodyHeight ?? undefined }"
    >
      <div
        v-if="hasHeaderContent"
        class="dialog-header flex shrink-0 items-start justify-between gap-3 border-b border-slate-100 px-5 py-4"
      >
        <slot name="header">
          <div class="space-y-1">
            <div class="flex flex-wrap items-center gap-2">
              <p
                v-if="title"
                class="text-[14px] font-semibold leading-[19px] text-slate-900 lg:text-lg lg:leading-normal"
              >
                {{ title }}
              </p>
              <span
                v-if="title && tagsLeft.length"
                class="text-sm font-medium text-slate-300"
                aria-hidden="true"
              >
                |
              </span>
              <div
                v-if="tagsLeft.length"
                class="flex flex-wrap items-center gap-2"
              >
                <span
                  v-for="tag in tagsLeft"
                  :key="`left-${tag}`"
                  class="rounded-full border border-[#dbe3ef] bg-[#f8fafc] px-[8px] py-[4px] text-[10px] font-semibold text-[#334155]"
                >
                  {{ tag }}
                </span>
              </div>
            </div>
            <p
              v-if="subtitle"
              class="text-[11px] leading-[16px] text-slate-600 lg:text-sm lg:leading-normal"
            >
              {{ subtitle }}
            </p>
          </div>
        </slot>
        <div class="flex items-center gap-2">
          <div
            v-if="tagsRight.length"
            class="flex flex-wrap items-center justify-end gap-2"
          >
            <span
              v-for="tag in tagsRight"
              :key="`right-${tag}`"
              class="rounded-full border border-[#dbe3ef] bg-[#f8fafc] px-[8px] py-[4px] text-[10px] font-semibold text-[#334155]"
            >
              {{ tag }}
            </span>
          </div>
          <slot name="action" />
          <slot name="header-extra" />
          <button
            v-if="closable"
            type="button"
            class="grid h-9 w-9 place-items-center rounded-full text-slate-500 transition hover:bg-slate-100 hover:text-slate-800"
            aria-label="Close"
            @click.stop="handleHide"
          >
            <span class="pi pi-times" />
          </button>
        </div>
      </div>

      <div v-if="$slots.top" class="app-dialog-top mt-2 shrink-0 ps-5 pe-5">
        <slot name="top" />
      </div>

      <div
        class="app-dialog-body min-h-0 min-w-0 flex-1 overflow-auto px-5 py-4"
        :class="{
          'app-dialog-body-no-header': isMobileLayout && !hasHeaderContent,
          'pb-[calc(env(safe-area-inset-bottom)+3.5rem)]':
            isMobileLayout &&
            ((mobileActions && mobileActions.length > 0) ||
              mobileStepper ||
              $slots['bottom-bar'] ||
              $slots.footer),
        }"
      >
        <slot />
      </div>

      <!-- Desktop Footer -->
      <div
        v-if="$slots.footer && !isMobileLayout"
        class="shrink-0 border-t border-slate-100"
        :class="footerCompact ? 'px-3 py-2' : 'px-5 py-4'"
      >
        <slot name="footer" />
      </div>
    </div>

    <nav
      v-if="
        isMobileLayout &&
        ((mobileActions && mobileActions.length > 0) ||
          mobileStepper ||
          $slots['bottom-bar'] ||
          $slots.footer)
      "
      class="app-dialog-mobile-nav fixed inset-x-0 bottom-0 z-50 border-t border-slate-200 bg-white/95 px-4 pb-[calc(env(safe-area-inset-bottom)+0.25rem)] pt-1 shadow-[0_-8px_24px_rgba(15,23,42,0.08)] backdrop-blur-xl"
    >
      <AppDialogMobileBottomBar
        v-if="mobileActions && mobileActions.length > 0"
        :actions="mobileActions"
      />
      <div
        v-else-if="mobileStepper"
        class="grid grid-cols-[64px_1fr_64px] items-center w-full px-2"
      >
        <!-- Back Button -->
        <div class="flex justify-start">
          <button
            v-if="mobileStepper.activeStep > 1"
            type="button"
            class="group !flex !min-h-[48px] !min-w-0 !items-center !justify-center !rounded-2xl !border-0 !bg-transparent !p-0 !text-[10px] !font-semibold !shadow-none hover:!translate-y-0 active:scale-[0.97]"
            @click="mobileStepper.onBack"
          >
            <span
              class="flex min-w-[52px] flex-col items-center justify-center gap-0.5 rounded-2xl px-2 py-1 transition-all duration-200"
            >
              <span
                class="flex h-7 w-7 items-center justify-center rounded-xl bg-slate-100 text-slate-500 transition-all duration-200 group-active:bg-slate-200"
              >
                <span class="pi pi-arrow-left text-[14px] leading-none" />
              </span>
              <span
                class="block w-full max-w-[64px] truncate !whitespace-nowrap text-center leading-none text-slate-600 transition-colors duration-200"
              >
                Back
              </span>
            </span>
          </button>
        </div>

        <!-- Middle Step Info -->
        <div class="text-center">
          <p class="text-sm font-semibold text-slate-600">
            Step {{ mobileStepper.activeStep }} of
            {{ mobileStepper.totalSteps }}
          </p>
        </div>

        <!-- Next / Submit Button -->
        <div class="flex justify-end">
          <!-- Next Button -->
          <button
            v-if="mobileStepper.activeStep < mobileStepper.totalSteps"
            type="button"
            :disabled="!mobileStepper.canGoNext"
            class="group !flex !min-h-[48px] !min-w-0 !items-center !justify-center !rounded-2xl !border-0 !bg-transparent !p-0 !text-[10px] !font-semibold !shadow-none hover:!translate-y-0 active:scale-[0.97]"
            @click="mobileStepper.onNext"
          >
            <span
              class="flex min-w-[52px] flex-col items-center justify-center gap-0.5 rounded-2xl px-2 py-1 transition-all duration-200"
            >
              <span
                :class="[
                  'flex h-7 w-7 items-center justify-center rounded-xl transition-all duration-200 text-white bg-[#dc2626]',
                  mobileStepper.canGoNext
                    ? 'shadow-[0_8px_18px_rgba(220,38,38,0.24)]'
                    : 'opacity-50',
                ]"
              >
                <span class="pi pi-arrow-right text-[14px] leading-none" />
              </span>
              <span
                class="block w-full max-w-[64px] truncate !whitespace-nowrap text-center leading-none text-[#dc2626] transition-colors duration-200"
              >
                Next
              </span>
            </span>
          </button>

          <!-- Submit Button -->
          <button
            v-else
            type="button"
            class="group !flex !min-h-[48px] !min-w-0 !items-center !justify-center !rounded-2xl !border-0 !bg-transparent !p-0 !text-[10px] !font-semibold !shadow-none hover:!translate-y-0 active:scale-[0.97]"
            @click="mobileStepper.onSubmit"
          >
            <span
              class="flex min-w-[52px] flex-col items-center justify-center gap-0.5 rounded-2xl px-2 py-1 transition-all duration-200"
            >
              <span
                class="flex h-7 w-7 items-center justify-center rounded-xl bg-[#dc2626] text-white shadow-[0_8px_18px_rgba(220,38,38,0.24)] transition-all duration-200 animate-fade-in"
              >
                <span
                  v-if="mobileStepper.submitLoading"
                  class="pi pi-spinner pi-spin text-[14px] leading-none"
                />
                <span v-else class="pi pi-check text-[14px] leading-none" />
              </span>
              <span
                class="block w-full max-w-[64px] truncate !whitespace-nowrap text-center leading-none text-[#dc2626] transition-colors duration-200"
              >
                Submit
              </span>
            </span>
          </button>
        </div>
      </div>
      <div
        v-else
        class="mx-auto grid max-w-md grid-flow-col auto-cols-fr gap-2 items-center"
      >
        <slot name="bottom-bar">
          <slot name="footer" />
        </slot>
      </div>
    </nav>
  </Dialog>
</template>
<style>
/* ============================
   FULLSCREEN MODE – MOBILE LAYOUT
   Applies both by viewport and by explicit mobileLayout prop.
   ============================ */
.app-dialog-mobile-layout.p-dialog {
  width: 100vw !important;
  height: 100dvh !important;
  max-width: 100vw !important;
  max-height: 100dvh !important;
  top: 0 !important;
  left: 0 !important;
  transform: none !important;
  display: flex !important;
  flex-direction: column !important;
  border-radius: 0 !important;
  margin: 0 !important;
  overflow: hidden !important;
  padding: 0 !important;
}

.app-dialog-mobile-layout .p-dialog-content {
  width: 100vw !important;
  height: 100% !important;
  max-height: 100% !important;
  padding: 0 !important;
  display: flex !important;
  flex-direction: column !important;
  overflow: hidden !important;
  -webkit-overflow-scrolling: touch;
}

.app-dialog-mobile-layout .dialog-header {
  padding-top: calc(env(safe-area-inset-top) + 1rem) !important;
  padding-left: calc(env(safe-area-inset-left) + 1.25rem) !important;
  padding-right: calc(env(safe-area-inset-right) + 1.25rem) !important;
}

.app-dialog-mobile-layout .app-dialog-top {
  padding-left: calc(env(safe-area-inset-left) + 1.25rem) !important;
  padding-right: calc(env(safe-area-inset-right) + 1.25rem) !important;
}

.app-dialog-mobile-layout .app-dialog-body {
  padding-left: calc(env(safe-area-inset-left) + 1.25rem) !important;
  padding-right: calc(env(safe-area-inset-right) + 1.25rem) !important;
}

.app-dialog-mobile-layout .app-dialog-body-no-header {
  padding-top: calc(env(safe-area-inset-top) + 1rem) !important;
}

.app-dialog-mobile-nav {
  padding-left: calc(env(safe-area-inset-left) + 1rem) !important;
  padding-right: calc(env(safe-area-inset-right) + 1rem) !important;
  padding-bottom: calc(env(safe-area-inset-bottom) + 0.25rem) !important;
}

@media (max-width: 768px) {
  .app-dialog-fullscreen.p-dialog {
    width: 100vw !important;
    /* Prefer dynamic viewport units for iOS Safari, with fallbacks */
    height: 100dvh !important;
    max-width: 100vw !important;
    max-height: 100dvh !important;
    top: 0 !important;
    left: 0 !important;
    transform: none !important;
    display: flex !important;
    flex-direction: column !important;
    border-radius: 0 !important;
    margin: 0 !important;
    overflow: hidden !important;
    padding: 0 !important;
  }

  .app-dialog-fullscreen .p-dialog-content {
    /* Ensure content fills dynamic viewport and keeps internal scrolling working on iOS */
    width: 100vw !important;
    height: 100% !important;
    max-height: 100% !important;
    padding: 0 !important;
    display: flex !important;
    flex-direction: column !important;
    overflow: hidden !important;
    -webkit-overflow-scrolling: touch;
  }

  .app-dialog-fullscreen .dialog-header {
    padding-top: calc(env(safe-area-inset-top) + 1rem) !important;
    padding-left: calc(env(safe-area-inset-left) + 1.25rem) !important;
    padding-right: calc(env(safe-area-inset-right) + 1.25rem) !important;
  }

  .app-dialog-fullscreen .app-dialog-top {
    padding-left: calc(env(safe-area-inset-left) + 1.25rem) !important;
    padding-right: calc(env(safe-area-inset-right) + 1.25rem) !important;
  }

  .app-dialog-fullscreen .app-dialog-body {
    padding-left: calc(env(safe-area-inset-left) + 1.25rem) !important;
    padding-right: calc(env(safe-area-inset-right) + 1.25rem) !important;
  }

  .app-dialog-fullscreen .app-dialog-body-no-header {
    padding-top: calc(env(safe-area-inset-top) + 1rem) !important;
  }
}

/* ============================
   DESKTOP MODE – NORMAL DIALOG
   ============================ */
@media (min-width: 769px) {
  .app-dialog-fullscreen.p-dialog:not(.app-dialog-mobile-layout) {
    width: var(--app-dialog-width, 70vw) !important;
    /* Keep the dialog inside the visible viewport when browser zoom reduces height */
    height: var(
      --app-dialog-height,
      min(90dvh, calc(100dvh - 2rem))
    ) !important;
    max-width: min(
      var(--app-dialog-max-width, 70vw),
      calc(100vw - 2rem)
    ) !important;
    max-height: min(
      var(--app-dialog-max-height, 90dvh),
      calc(100dvh - 2rem)
    ) !important;
    top: 0 !important;
    left: 0 !important;
    transform: none !important;
    display: flex !important;
    flex-direction: column !important;
    border-radius: 1rem !important;
    overflow: hidden !important;
    margin: 0 !important;
    /* Respect iPhone safe area so buttons don't get hidden behind home indicator */
    padding-bottom: env(safe-area-inset-bottom);
  }

  .app-dialog-mask-top {
    align-items: flex-start !important;
    padding-top: clamp(0.5rem, 2dvh, 1.5rem) !important;
    padding-bottom: clamp(0.5rem, 2dvh, 1.5rem) !important;
  }

  .app-dialog-fullscreen.p-dialog:not(.app-dialog-mobile-layout)
    .p-dialog-content {
    /* Ensure content fills the constrained dialog and keeps internal scrolling working */
    width: 100% !important;
    min-height: 0 !important;
    height: 100% !important;
    max-height: 100% !important;
    padding: 0 !important;
    display: flex !important;
    flex: 1 1 auto !important;
    flex-direction: column !important;
    overflow: hidden !important;
    -webkit-overflow-scrolling: touch;
  }
}

.app-dialog-fullscreen.p-dialog,
.app-dialog-fullscreen.p-dialog:focus,
.app-dialog-fullscreen.p-dialog:focus-visible,
.app-dialog-fullscreen.p-dialog:focus-within,
.app-dialog-fullscreen .p-dialog-content,
.app-dialog-fullscreen .p-dialog-content:focus,
.app-dialog-fullscreen .p-dialog-content:focus-visible {
  outline: none !important;
  border: none !important;
  box-shadow:
    0 20px 25px -5px rgb(0 0 0 / 0.1),
    0 8px 10px -6px rgb(0 0 0 / 0.1) !important;
}
</style>
