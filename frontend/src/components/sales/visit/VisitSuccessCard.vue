<script setup lang="ts">
defineProps<{
  icon?: string
  title: string
  message: string
  primaryLabel: string
  primaryTo: string | Record<string, unknown>
  secondaryLabel?: string
  secondaryTo?: string | Record<string, unknown>
  tertiaryLabel?: string
  tertiaryTo?: string | Record<string, unknown>
}>()
</script>

<template>
  <div class="vs-card">
    <div class="vs-icon">
      <i :class="icon ?? 'pi pi-check'" />
    </div>
    <h1 class="vs-title">{{ title }}</h1>
    <p class="vs-message">{{ message }}</p>
    <div class="vs-actions">
      <RouterLink class="vs-btn vs-btn-primary" :to="primaryTo">{{ primaryLabel }}</RouterLink>
      <RouterLink v-if="secondaryLabel && secondaryTo" class="vs-btn vs-btn-secondary" :to="secondaryTo">{{ secondaryLabel }}</RouterLink>
      <RouterLink v-if="tertiaryLabel && tertiaryTo" class="vs-btn vs-btn-tertiary" :to="tertiaryTo">{{ tertiaryLabel }}</RouterLink>
    </div>
  </div>
</template>

<style scoped>
.vs-card {
  display: flex; flex-direction: column; align-items: center; text-align: center;
  padding: 2.5rem 1.5rem; background: var(--surface-card); border: 1px solid var(--border-light);
  border-radius: var(--radius-xl); box-shadow: var(--shadow-sm);
}
.vs-icon {
  width: 72px; height: 72px; border-radius: 50%; display: grid; place-items: center;
  background: #f0fdf4; color: #16a34a; font-size: 1.8rem; margin-bottom: 1.25rem;
  box-shadow: 0 4px 20px rgba(22, 163, 74, 0.15);
}
.vs-title { margin: 0 0 0.5rem; font-size: 1.2rem; font-weight: 800; color: var(--text-primary); }
.vs-message { margin: 0 0 1.75rem; font-size: 0.82rem; color: var(--text-muted); line-height: 1.5; max-width: 300px; }
.vs-actions { display: flex; flex-direction: column; gap: 0.6rem; width: 100%; max-width: 320px; }
.vs-btn {
  display: flex; align-items: center; justify-content: center; gap: 0.4rem;
  padding: 0.75rem 1rem; border-radius: 14px; font-size: 0.82rem; font-weight: 700;
  text-decoration: none; text-align: center; transition: all 0.15s ease; border: none; cursor: pointer;
}
.vs-btn-primary { background: var(--brand-blue); color: #fff; }
.vs-btn-primary:hover { background: #bb3342; }
.vs-btn-secondary { background: #f8fafc; color: var(--text-primary); border: 1px solid var(--border-light); }
.vs-btn-secondary:hover { background: #f1f5f9; }
.vs-btn-tertiary { background: transparent; color: var(--text-muted); font-weight: 600; }
.vs-btn-tertiary:hover { color: var(--text-primary); }
</style>
