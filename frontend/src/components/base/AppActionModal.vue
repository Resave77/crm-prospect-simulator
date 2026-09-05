<script setup lang="ts">
import Dialog from "primevue/dialog";
import { computed } from "vue";

import { useAuthStore } from "@/stores/authStore";

import AppButton from "./AppButton.vue";

export interface ActionModalButton {
  description?: string;
  hide?: () => boolean;
  icon: string;
  label: string;
  onClick?: () => void;
  permissions?: string[];
  to?: { name: string; params?: Record<string, unknown> };
  variant?: "primary" | "secondary";
}

export interface ActionModalDangerZone {
  buttonLabel: string;
  description: string;
  error?: string;
  hide?: () => boolean;
  onAction: () => void;
  permissions?: string[];
  title: string;
}

interface Props {
  actions?: ActionModalButton[];
  dangerZone?: ActionModalDangerZone;
  subtitle?: string;
  tag?: string;
  title: string;
  visible?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  actions: () => [],
  visible: false,
});

const authStore = useAuthStore();

function hasPermission(permissions?: string[]): boolean {
  if (!permissions || permissions.length === 0) return true;
  return permissions.some((perm) => authStore.permissions.includes(perm));
}

const filteredActions = computed(() => {
  return props.actions.filter(
    (action) => hasPermission(action.permissions) && !action.hide?.()
  );
});

const emit = defineEmits<{
  close: [];
  "update:visible": [value: boolean];
}>();

const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit("update:visible", value),
});

const closeModal = () => {
  emit("close");
  emit("update:visible", false);
};
</script>

<template>
  <Dialog
    v-model:visible="dialogVisible"
    :modal="true"
    :closable="true"
    :close-on-escape="true"
    :dismissable-mask="true"
    :style="{ width: '28rem' }"
    :pt="{ root: { class: 'rounded-2xl' } }"
  >
    <template #header>
      <div>
        <p
          v-if="tag"
          class="text-xs font-semibold uppercase tracking-wide text-slate-400"
        >
          {{ tag }}
        </p>
        <h3 class="text-heading font-semibold text-slate-900">
          {{ title }}
        </h3>
        <p v-if="subtitle" class="text-sm text-slate-500">
          {{ subtitle }}
        </p>
      </div>
    </template>

    <!-- Action Buttons Grid -->
    <div v-if="filteredActions.length" class="mt-4 grid gap-3 sm:grid-cols-2">
      <component
        :is="action.to ? 'RouterLink' : 'button'"
        v-for="(action, index) in filteredActions"
        :key="index"
        :to="action.to"
        type="button"
        class="block"
        @click="
          () => {
            action.onClick?.();
            if (action.to) closeModal();
          }
        "
      >
        <div
          class="flex items-center gap-3 rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 transition hover:border-primary-200 hover:bg-primary-50"
        >
          <span :class="[action.icon, 'text-primary-600']" />
          <div class="text-left">
            <p class="font-semibold text-slate-900">
              {{ action.label }}
            </p>
            <p v-if="action.description" class="text-xs text-slate-500">
              {{ action.description }}
            </p>
          </div>
        </div>
      </component>
    </div>

    <!-- Danger Zone -->
    <div
      v-if="
        dangerZone &&
        !dangerZone.hide?.() &&
        (!dangerZone.permissions || hasPermission(dangerZone.permissions))
      "
      class="mt-5 rounded-xl border border-red-100 bg-red-50 px-4 py-3 text-red-800"
    >
      <div class="flex items-start gap-2">
        <span class="pi pi-exclamation-triangle mt-0.5" />
        <div class="space-y-2 flex-1">
          <div class="flex items-center justify-between">
            <p class="font-semibold">
              {{ dangerZone.title }}
            </p>
            <AppButton
              variant="danger"
              size="small"
              icon="pi-trash"
              @click="dangerZone.onAction"
            >
              {{ dangerZone.buttonLabel }}
            </AppButton>
          </div>
          <p class="text-xs text-red-700">
            {{ dangerZone.description }}
          </p>
          <p v-if="dangerZone.error" class="text-xs font-semibold text-red-700">
            {{ dangerZone.error }}
          </p>
        </div>
      </div>
    </div>
  </Dialog>
</template>
