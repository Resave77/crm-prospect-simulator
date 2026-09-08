<script setup lang="ts">
import { useSmartBackNavigation } from "@/composables/useSmartBackNavigation";

const { goBack } = useSmartBackNavigation();
</script>

<template>
  <section class="flex min-h-full flex-1 items-center justify-center p-4">
    <div
      class="w-full max-w-md rounded-2xl border border-slate-200 bg-white px-6 py-8 text-center shadow-sm sm:px-8 sm:py-10"
    >
      <div
        class="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-primary-50 text-primary-600"
      >
        <span class="text-2xl font-bold">404</span>
      </div>

      <p
        class="mt-5 text-xs font-bold uppercase tracking-[0.12em] text-slate-400"
      >
        Page Not Found
      </p>
      <h1 class="mt-2 text-xl font-bold text-slate-900">
        Halaman tidak ditemukan
      </h1>
      <p class="mt-2 text-sm leading-6 text-slate-500">
        Link yang kamu buka mungkin sudah berubah, dihapus, atau tidak tersedia
        untuk aksesmu saat ini.
      </p>

      <button
        type="button"
        class="mx-auto mt-6 inline-flex items-center justify-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-primary-700 active:scale-[0.98]"
        @click="goBack"
      >
        <span class="pi pi-arrow-left text-xs" />
        <span>Back</span>
      </button>
    </div>
  </section>
</template>
