<script setup lang="ts">
import { useSmartBackNavigation } from "@/composables/useSmartBackNavigation";

const { goBack } = useSmartBackNavigation();
</script>

<template>
  <div class="flex items-center justify-center">
    <div class="rounded-lg border bg-white p-10 text-center shadow-sm">
      <svg
        class="mx-auto h-16 w-16 text-gray-400"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6l4 2" />
      </svg>

      <h2 class="mt-4 text-xl font-semibold text-gray-800">
        Under Construction
      </h2>

      <p class="mt-1 text-gray-500">We're working on this page.</p>

      <button
        type="button"
        class="mx-auto mt-5 inline-flex items-center justify-center gap-2 rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 shadow-sm transition hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 active:scale-[0.98]"
        @click="goBack"
      >
        <span class="pi pi-arrow-left text-xs" />
        <span>Back</span>
      </button>
    </div>
  </div>
</template>
