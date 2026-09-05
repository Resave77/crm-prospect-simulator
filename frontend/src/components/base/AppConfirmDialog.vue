<script setup lang="ts">
import Dialog from "primevue/dialog";
import { computed, nextTick } from "vue";

import AppButton from "./AppButton.vue";

export interface ConfirmDialogItem {
  id: number | string;
  label: string;
  sublabel?: string;
}

interface Props {
  cancelLabel?: string;
  confirmIcon?: string;
  confirmLabel?: string;
  error?: string;
  items?: ConfirmDialogItem[];
  labelIcon?: string;
  loading?: boolean;
  subtitle?: string;
  tag?: string;
  title: string;
  variant?: "danger" | "primary";
  visible: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  error: "",
  labelIcon: "pi-user",
  loading: false,
  variant: "danger",
});

const emit = defineEmits<{
  cancel: [];
  confirm: [];
  "update:visible": [value: boolean];
}>();

const localVisible = computed({
  get: () => props.visible,
  set: (value) => emit("update:visible", value),
});

const finalConfirmLabel = computed(() => props.confirmLabel || "Confirm");
const finalCancelLabel = computed(() => props.cancelLabel || "Cancel");

const handleConfirm = () => {
  emit("confirm");
};

const handleCancel = () => {
  emit("cancel");
  emit("update:visible", false);
};

const blurActiveElement = () => {
  if (document.activeElement instanceof HTMLElement) {
    document.activeElement.blur();
  }
};

const clearInitialFocus = async () => {
  await nextTick();
  blurActiveElement();
  window.requestAnimationFrame(blurActiveElement);
  window.setTimeout(blurActiveElement, 0);
};
</script>

<template>
  <Dialog
    v-model:visible="localVisible"
    :modal="true"
    :closable="true"
    :close-on-escape="true"
    :dismissable-mask="true"
    :style="{ width: 'min(32rem, calc(100vw - 2.5rem))' }"
    :pt="{
      root: { class: 'app-confirm-dialog rounded-2xl' },
      closeButton: {
        class: 'outline-none focus:outline-none focus:ring-0 focus:shadow-none',
        tabindex: -1,
      },
      pcCloseButton: {
        root: {
          class:
            'outline-none focus:outline-none focus:ring-0 focus:shadow-none',
          tabindex: -1,
        },
      },
    }"
    :header="title"
    @show="clearInitialFocus"
  >
    <template #header>
      <div>
        <p
          v-if="tag"
          class="text-[10px] font-semibold uppercase tracking-wide text-slate-400 sm:text-xs"
        >
          {{ tag }}
        </p>
        <h3 class="app-confirm-dialog__title">
          {{ title }}
        </h3>
        <p v-if="subtitle" class="app-confirm-dialog__subtitle">
          {{ subtitle }}
        </p>
      </div>
    </template>

    <div
      v-if="items && items.length"
      class="mt-4 max-h-40 space-y-2 overflow-y-auto rounded-xl border border-slate-100 bg-slate-50 px-3 py-2 text-xs text-slate-700 sm:text-sm"
    >
      <div v-for="item in items" :key="item.id" class="flex items-center gap-2">
        <span :class="`pi ${labelIcon} text-slate-400`" />
        <span class="font-semibold text-slate-900">{{ item.label }}</span>
        <template v-if="item.sublabel">
          <span class="text-slate-400">•</span>
          <span class="text-slate-500">{{ item.sublabel }}</span>
        </template>
      </div>
    </div>

    <slot />

    <div class="mt-5 flex items-center justify-end gap-2">
      <AppButton variant="secondary" :disabled="loading" @click="handleCancel">
        {{ finalCancelLabel }}
      </AppButton>
      <AppButton
        :variant="variant"
        :icon="confirmIcon"
        :loading="loading"
        @click="handleConfirm"
      >
        {{ finalConfirmLabel }}
      </AppButton>
    </div>

    <p v-if="error" class="mt-2 text-xs font-semibold text-red-700">
      {{ error }}
    </p>
  </Dialog>
</template>

<style scoped>
.app-confirm-dialog__title {
  margin: 0;
  font-size: 15px !important;
  font-weight: 600;
  line-height: 20px;
  color: #0f172a;
}

.app-confirm-dialog__subtitle {
  margin-top: 0.25rem;
  font-size: 12px !important;
  line-height: 17px;
  color: #64748b;
}

@media (min-width: 640px) {
  .app-confirm-dialog__title {
    font-size: var(--font-heading) !important;
    line-height: 1.25;
  }

  .app-confirm-dialog__subtitle {
    font-size: 14px !important;
    line-height: 20px;
  }
}

:deep(.app-confirm-dialog .p-dialog-close-button),
:deep(.app-confirm-dialog .p-dialog-close-button:focus),
:deep(.app-confirm-dialog .p-dialog-close-button:focus-visible),
:deep(.app-confirm-dialog .p-dialog-close-button:active),
:deep(.app-confirm-dialog .p-dialog-close-button.p-focus),
:deep(.app-confirm-dialog .p-dialog-close-button.p-button:focus),
:deep(.app-confirm-dialog .p-dialog-close-button.p-button:focus-visible) {
  outline: none !important;
  box-shadow: none !important;
  border-color: transparent !important;
}
</style>
