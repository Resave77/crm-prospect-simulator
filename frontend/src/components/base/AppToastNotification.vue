<script setup lang="ts">
import { useToast } from "primevue";
import Toast from "primevue/toast";

import AppButton from "@/components/base/AppButton.vue";

/**
 * Action configuration for toast notifications
 * @example
 * const actions = [
 *   {
 *     label: 'Retry',
 *     variant: 'primary',
 *     onClick: () => retryAction()
 *   },
 *   {
 *     label: 'Cancel',
 *     variant: 'secondary',
 *     onClick: () => cancelAction()
 *   }
 * ]
 */
interface ToastAction {
  /** Label text for the action button */
  label: string;
  /** Click handler for the action */
  onClick: (event: MouseEvent) => void;
  /** Button variant style */
  variant?: "primary" | "danger" | "secondary" | "ghost";
}

const props = withDefaults(
  defineProps<{
    /** Single action label (legacy support) */
    actionLabel?: string | null;
    /** Array of action configurations for multiple buttons */
    actions?: ToastAction[];
    /** Single action variant (legacy support) */
    actionVariant?: "primary" | "danger" | "secondary" | "ghost";
    /** Toast notification group identifier */
    group: string;
    /** Icon class name to display */
    icon?: string | null;
    /** Position where toast should appear */
    position?: ToastPosition;
  }>(),
  {
    actionLabel: null,
    actions: () => [],
    actionVariant: "primary",
    icon: "pi pi-info-circle",
    position: "top-right",
  }
);
export type ToastPosition =
  | "top-left"
  | "top-right"
  | "top-center"
  | "center"
  | "bottom-left"
  | "bottom-right"
  | "bottom-center";

// Props are used directly in template via defineProps()

const emit = defineEmits<{
  (e: "action", event: MouseEvent): void;
}>();

const severityClass = (severity?: string) => {
  switch (severity) {
    case "error":
      return "text-red-500";
    case "success":
      return "text-emerald-500";
    case "warn":
      return "text-amber-500";
    default:
      return "text-blue-500";
  }
};

const onActionClick = (event: MouseEvent) => {
  emit("action", event);
};

const toast = useToast();
</script>

<template>
  <Toast
    :group="props.group"
    :position="props.position"
    :pt="{
      root: { class: 'app-toast-root' },
      message: { class: 'app-toast-message' },
    }"
  >
    <template #message="slotProps">
      <div class="app-toast">
        <div class="app-toast__content">
          <i
            v-if="icon"
            class="pi app-toast__icon"
            :class="[icon, severityClass(slotProps.message.severity)]"
          />
          <div class="app-toast__text">
            <p class="app-toast__summary">
              {{ slotProps.message.summary }}
            </p>
            <p v-if="slotProps.message.detail" class="app-toast__detail">
              {{ slotProps.message.detail }}
            </p>
          </div>
        </div>

        <div
          v-if="
            $slots.actions ||
            props.actionLabel ||
            (props.actions && props.actions.length > 0)
          "
          class="app-toast__actions"
        >
          <slot name="actions" :message="slotProps.message">
            <!-- Multiple actions from props.actions -->
            <template v-if="props.actions && props.actions.length > 0">
              <AppButton
                v-for="(action, index) in props.actions"
                :key="index"
                size="small"
                :variant="action.variant || 'primary'"
                @click="
                  (event) => {
                    action.onClick(event);
                    toast.remove(slotProps.message);
                  }
                "
              >
                {{ action.label }}
              </AppButton>
            </template>

            <!-- Legacy single action support -->
            <AppButton
              v-else-if="props.actionLabel"
              size="small"
              :variant="props.actionVariant"
              @click="onActionClick"
            >
              {{ props.actionLabel }}
            </AppButton>
          </slot>
        </div>
      </div>
    </template>
  </Toast>
</template>

<style scoped>
/* ======== DESKTOP / DEFAULT ======== */

:deep(.app-toast-root) {
  width: min(420px, calc(100vw - 2rem));
  max-width: 100%;
  padding: 0.25rem;
}

:deep(.app-toast-root.p-toast-top-right),
:deep(.app-toast-root.p-toast-bottom-right) {
  right: clamp(0.75rem, env(safe-area-inset-right) + 0.5rem, 2rem);
}

:deep(.app-toast-root.p-toast-top-left),
:deep(.app-toast-root.p-toast-bottom-left) {
  left: clamp(0.75rem, env(safe-area-inset-left) + 0.5rem, 2rem);
}

:deep(.app-toast-message) {
  width: 100%;
  padding: 1rem;
  border-radius: 0.75rem;
  background-color: #ffffff;
  box-shadow:
    0 8px 20px rgba(15, 23, 42, 0.1),
    0 4px 8px rgba(15, 23, 42, 0.08);
}

.app-toast {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  width: 100%;
}

.app-toast__content {
  display: flex;
  gap: 0.75rem;
}

.app-toast__icon {
  font-size: 1.35rem;
}

.app-toast__text {
  flex: 1;
  min-width: 0;
}

.app-toast__summary {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: #0f172a;
}

.app-toast__detail {
  margin: 0.25rem 0 0;
  font-size: 0.85rem;
  color: #475569;
  white-space: pre-line;
}

.app-toast__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
}

/* ======== MOBILE FRIENDLY STYLES ======== */
@media (max-width: 767px) {
  :deep(.app-toast-root) {
    width: calc(100vw - 1.5rem) !important;
    left: 0.75rem !important;
    right: 0.75rem !important;
    max-width: 100% !important;
    padding: 0 !important;
  }

  :deep(.app-toast-root.p-toast-top-right),
  :deep(.app-toast-root.p-toast-top-left),
  :deep(.app-toast-root.p-toast-top-center) {
    top: calc(env(safe-area-inset-top) + 0.75rem) !important;
  }

  :deep(.app-toast-root.p-toast-bottom-right),
  :deep(.app-toast-root.p-toast-bottom-left),
  :deep(.app-toast-root.p-toast-bottom-center) {
    bottom: calc(env(safe-area-inset-bottom) + 0.75rem) !important;
  }

  :deep(.app-toast-message) {
    padding: 0.85rem !important;
    border-radius: 0.5rem !important;
    margin-bottom: 0.5rem !important;
    box-shadow: 0 4px 12px rgba(15, 23, 42, 0.12) !important;
  }

  .app-toast__icon {
    font-size: 1.15rem !important;
  }

  .app-toast__summary {
    font-size: 0.85rem !important;
  }

  .app-toast__detail {
    font-size: 0.75rem !important;
  }
}
</style>
