<script setup lang="ts">
withDefaults(defineProps<{
  title: string
  subtitle?: string
  backLabel?: string
}>(), { subtitle: '', backLabel: '' })

const emit = defineEmits<{ back: [] }>()
</script>

<template>
  <header class="admin-page-header">
    <div class="admin-page-header__identity">
      <button v-if="backLabel" type="button" class="admin-page-header__back" @click="emit('back')">
        <i class="pi pi-arrow-left" aria-hidden="true" />
        <span>{{ backLabel }}</span>
      </button>
      <div>
        <h1>{{ title }}</h1>
        <p v-if="subtitle" class="admin-page-header__subtitle">{{ subtitle }}</p>
      </div>
    </div>
    <div v-if="$slots.actions" class="admin-page-header__actions"><slot name="actions" /></div>
  </header>
</template>

<style scoped>
.admin-page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1.5rem;
  margin: 0 0 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border-light);
}
.admin-page-header__identity { display: flex; align-items: flex-start; gap: 1rem; min-width: 0; }
.admin-page-header__back {
  display: inline-flex; align-items: center; gap: .4rem; min-height: 2rem; padding: 0;
  border: 0; background: transparent; color: var(--text-secondary); cursor: pointer; font-size: .8rem; white-space: nowrap;
}
.admin-page-header__back:hover { color: var(--brand-red); }
h1 { margin: 0; font-size: clamp(1.45rem, 2.3vw, 2rem); line-height: 1.15; letter-spacing: -.03em; }
.admin-page-header__subtitle { margin: .3rem 0 0; color: var(--text-muted); font-size: .78rem; }
.admin-page-header__actions { display: flex; align-items: center; gap: .5rem; flex-wrap: wrap; }
@media (max-width: 640px) {
  .admin-page-header { flex-direction: column; gap: .75rem; }
  .admin-page-header__actions { width: 100%; }
  .admin-page-header__actions :deep(button) { flex: 1; }
}
</style>
