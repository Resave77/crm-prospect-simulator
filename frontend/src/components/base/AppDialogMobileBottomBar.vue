<script setup lang="ts">
import { computed, onMounted, onUnmounted, unref, ref } from "vue";

import type {
  DialogMobileBottomBarAction,
  DialogMobileBottomBarSubmenuItem,
} from "@/composables/useDialogMobileBottomBar";

import AppButton from "@/components/base/AppButton.vue";

const props = defineProps<{
  actions: DialogMobileBottomBarAction[];
}>();

type DialogMobileBottomBarSubmenuAction = Extract<
  DialogMobileBottomBarAction,
  { type: "submenu" }
>;

const submenuOpenKey = ref<string | null>(null);
const bottomBarRef = ref<HTMLElement | null>(null);
const submenuPanelRef = ref<HTMLElement | null>(null);

const activeSubmenuAction = computed<DialogMobileBottomBarSubmenuAction | null>(
  () => {
    if (!submenuOpenKey.value) return null;
    return (
      (props.actions.find(
        (action) =>
          action.type === "submenu" && action.key === submenuOpenKey.value
      ) as DialogMobileBottomBarSubmenuAction | undefined) ?? null
    );
  }
);

const visibleSubmenuItems = computed(() => {
  const action = activeSubmenuAction.value;
  if (!action) return [];
  return action.items.filter((item) => unref(item.visible) !== false);
});

const visibleActions = computed(() =>
  props.actions.filter((action) => unref(action.visible) !== false)
);

const handleActionClick = (action: DialogMobileBottomBarAction) => {
  if (unref(action.disabled) || unref(action.loading)) return;

  if (action.type === "submenu") {
    submenuOpenKey.value =
      submenuOpenKey.value === action.key ? null : action.key;
    return;
  }

  submenuOpenKey.value = null;
  action.onClick();
};

const handleSubmenuItemClick = (item: DialogMobileBottomBarSubmenuItem) => {
  if (unref(item.disabled) || unref(item.loading)) return;

  submenuOpenKey.value = null;
  item.onClick();
};

const isActionActive = (action: DialogMobileBottomBarAction) => {
  if (unref(action.active)) return true;
  if (action.type === "submenu") return submenuOpenKey.value === action.key;
  return false;
};

// eslint-disable-next-line no-undef
const handleDocumentPointerDown = (event: PointerEvent) => {
  if (!submenuOpenKey.value) return;

  const target = event.target;
  // eslint-disable-next-line no-undef
  if (!(target instanceof Node)) return;

  if (submenuPanelRef.value?.contains(target)) return;
  if (bottomBarRef.value?.contains(target)) return;
  if ((target as HTMLElement).closest(".dialog-mobile-submenu-trigger")) return;

  submenuOpenKey.value = null;
};

onMounted(() => {
  document.addEventListener("pointerdown", handleDocumentPointerDown);
});

onUnmounted(() => {
  document.removeEventListener("pointerdown", handleDocumentPointerDown);
});
</script>

<template>
  <Transition
    enter-active-class="transition duration-200 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition duration-150 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <button
      v-if="submenuOpenKey"
      type="button"
      class="fixed inset-0 -z-10 bg-transparent"
      aria-label="Close dialog action menu"
      @click="submenuOpenKey = null"
    />
  </Transition>

  <Transition
    enter-active-class="transition duration-150 ease-out"
    enter-from-class="translate-y-2 opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition duration-100 ease-in"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="translate-y-2 opacity-0"
  >
    <div
      v-if="activeSubmenuAction"
      ref="submenuPanelRef"
      class="fixed right-3 bottom-[calc(env(safe-area-inset-bottom)+4rem)] left-3 z-[70] max-h-[min(20rem,calc(100vh-8rem))] overflow-hidden rounded-xl border border-slate-200 bg-white p-1.5 shadow-lg"
    >
      <div
        class="flex items-center justify-between border-b border-slate-100 px-2.5 py-2 text-xs font-semibold text-slate-500"
      >
        <span class="min-w-0 truncate">
          {{ activeSubmenuAction.panelTitle || activeSubmenuAction.label }}
        </span>
        <button
          type="button"
          class="flex size-6 shrink-0 items-center justify-center rounded-md text-slate-400 hover:bg-slate-100 hover:text-slate-700"
          aria-label="Close dialog action menu"
          @click="submenuOpenKey = null"
        >
          <span class="pi pi-times text-[10px]" />
        </button>
      </div>
      <div
        class="max-h-[calc(min(20rem,calc(100vh-8rem))-3.5rem)] overflow-y-auto overscroll-contain py-1"
      >
        <AppButton
          v-for="item in visibleSubmenuItems"
          :key="item.key"
          variant="custom"
          :permissions="item.permissions"
          :disabled="unref(item.disabled) || unref(item.loading)"
          class="!flex !min-h-10 !w-full !items-center !gap-2.5 !rounded-lg !border-0 !px-2.5 !py-2 !text-left !text-xs !font-semibold !text-slate-700 !shadow-none hover:!bg-slate-50"
          @click="handleSubmenuItemClick(item)"
        >
          <span
            class="pi w-4 shrink-0 text-center text-[13px] text-slate-400"
            :class="unref(item.loading) ? 'pi-spinner pi-spin' : item.icon"
          />
          <span class="min-w-0 flex-1 truncate">{{ item.label }}</span>
          <span class="pi pi-chevron-right text-[10px] text-slate-300" />
        </AppButton>
      </div>
    </div>
  </Transition>

  <div
    ref="bottomBarRef"
    class="mx-auto grid w-full max-w-md grid-flow-col auto-cols-fr gap-2 items-center"
  >
    <AppButton
      v-for="action in visibleActions"
      :key="action.key"
      variant="custom"
      size="small"
      :permissions="action.permissions"
      :disabled="unref(action.disabled) || unref(action.loading)"
      class="group !flex !min-h-[48px] !min-w-0 !items-center !justify-center !rounded-2xl !border-0 !bg-transparent !p-0 !text-[10px] !font-semibold !shadow-none !outline-none !ring-0 hover:!translate-y-0 focus:!outline-none focus:!ring-0 focus-visible:!outline-none focus-visible:!ring-0 [-webkit-tap-highlight-color:transparent]"
      :class="action.type === 'submenu' ? 'dialog-mobile-submenu-trigger' : ''"
      @click.stop="handleActionClick(action)"
    >
      <span
        class="flex min-w-[52px] flex-col items-center justify-center gap-0.5 rounded-2xl px-2 py-1 transition-all duration-200"
        :class="isActionActive(action) ? 'bg-red-50 shadow-sm' : ''"
      >
        <span
          class="flex h-7 w-7 items-center justify-center rounded-xl transition-all duration-200"
          :class="
            isActionActive(action)
              ? 'bg-[#dc2626] text-white shadow-[0_8px_18px_rgba(220,38,38,0.24)]'
              : 'bg-slate-100 text-slate-500 group-active:bg-slate-200'
          "
        >
          <span
            class="pi text-[14px] leading-none"
            :class="unref(action.loading) ? 'pi-spinner pi-spin' : action.icon"
          />
        </span>
        <span
          class="block w-full max-w-[64px] truncate !whitespace-nowrap text-center leading-none transition-colors duration-200"
          :class="isActionActive(action) ? 'text-[#dc2626]' : 'text-slate-600'"
        >
          {{ action.label }}
        </span>
      </span>
    </AppButton>
  </div>
</template>
