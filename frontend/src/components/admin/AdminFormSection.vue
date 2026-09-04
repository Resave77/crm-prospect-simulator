<script setup lang="ts">defineProps<{ title: string; description?: string }>()</script>
<template><section class="admin-form-section"><header><div><h2>{{ title }}</h2><p v-if="description">{{ description }}</p></div><slot name="icon" /></header><div class="admin-form-section__body"><slot /></div></section></template>
<style scoped>
.admin-form-section{border:1px solid var(--border-light);border-radius:var(--radius-lg);background:var(--surface-card);overflow:hidden}.admin-form-section>header{display:flex;align-items:flex-start;justify-content:space-between;gap:1rem;padding:1rem 1.25rem;border-bottom:1px solid var(--border-light)}h2{margin:0;font-size:1rem}.admin-form-section p{margin:.25rem 0 0;color:var(--text-muted);font-size:.78rem}.admin-form-section__body{padding:1.25rem}
</style>
