<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    cols?: number;
    description?: string;
    groupLabel?: string;
    title?: string;
  }>(),
  {
    cols: 2,
    description: "",
    groupLabel: "",
    title: "",
  }
);

const displayTitle = computed(() => props.title || props.groupLabel || "");
</script>

<template>
  <div class="space-y-[12px]">
    <div v-if="displayTitle">
      <p class="font-['Inter'] text-[11px] font-bold uppercase text-[#94a3b8]">
        {{ displayTitle }}
      </p>
      <p
        v-if="description"
        class="mt-[3px] font-['Inter'] text-[10px] leading-[14px] text-[#94a3b8]"
      >
        {{ description }}
      </p>
    </div>
    <div
      class="grid gap-x-[8px] gap-y-[10px]"
      :class="cols === 1 ? 'grid-cols-1' : 'grid-cols-2'"
    >
      <slot />
    </div>
  </div>
</template>
