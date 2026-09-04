<script setup lang="ts">defineProps<{ label: string; tone?: string }>()</script>
<template><span class="admin-status-badge" :class="tone ? `admin-status-badge--${tone}` : ''">{{ label }}</span></template>
<style scoped>
.admin-status-badge{display:inline-flex;align-items:center;min-height:1.5rem;padding:.2rem .55rem;border:1px solid var(--border-default);border-radius:999px;background:var(--surface-subtle);color:var(--text-secondary);font-size:.7rem;font-weight:700;line-height:1}.admin-status-badge--success{background:#ecfdf3;color:#16804b;border-color:#a7ebc5}.admin-status-badge--danger{background:var(--brand-red-50);color:var(--brand-red-700);border-color:var(--brand-red-100)}.admin-status-badge--warning{background:#fff8e6;color:#9a6700;border-color:#f5d98b}
</style>
