<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

type Destination = 'pipeline' | 'my-prospects'

const props = withDefaults(defineProps<{
  destinations?: Destination[]
}>(), {
  destinations: () => ['pipeline', 'my-prospects'],
})

interface SwitcherItem {
  id: Destination
  label: string
  route: string
  icon: string
}

const ITEMS: SwitcherItem[] = [
  { id: 'pipeline', label: 'Prospect Pipeline', route: '/sales/pipeline', icon: 'pi pi-chart-line' },
  { id: 'my-prospects', label: 'My Prospects', route: '/sales/my-prospects', icon: 'pi pi-briefcase' },
]

const route = useRoute()

const visibleItems = computed(() => ITEMS.filter((item) => props.destinations.includes(item.id)))

function isActive(item: SwitcherItem): boolean {
  if (item.route === '/sales/pipeline') return route.path === '/sales/pipeline'
  return route.path === '/sales/my-prospects' || route.path.startsWith('/sales/my-prospects/')
}
</script>

<template>
  <nav
    class="psw"
    aria-label="Prospect workspace"
    :style="{ gridTemplateColumns: `repeat(${visibleItems.length}, minmax(0, 1fr))` }"
  >
    <RouterLink
      v-for="item in visibleItems"
      :key="item.route"
      :to="item.route"
      class="psw-item"
      :class="{ 'psw-item-active': isActive(item) }"
      :aria-current="isActive(item) ? 'page' : undefined"
      :title="item.label"
    >
      <i :class="item.icon" aria-hidden="true" />
      <span>{{ item.label }}</span>
    </RouterLink>
  </nav>
</template>

<style scoped>
.psw {
  display: grid;
  gap: 0.25rem;
  padding: 0.25rem;
  min-width: 0;
  width: 100%;
  box-sizing: border-box;
  background: var(--surface-subtle);
  border: 1px solid var(--border-light);
  border-radius: 12px;
}

.psw-item {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  min-width: 0;
  min-height: 44px;
  padding: 0.5rem 0.6rem;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.74rem;
  font-weight: 600;
  line-height: 1.2;
  text-align: center;
  text-decoration: none;
  white-space: nowrap;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  transition: background var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast), box-shadow var(--transition-fast);
}

.psw-item i {
  font-size: 0.9rem;
  flex-shrink: 0;
  color: inherit;
}

.psw-item:hover {
  color: var(--brand-blue);
  background: #ffffff;
  border-color: var(--border-default);
}

.psw-item:active {
  transform: scale(0.98);
}

.psw-item:focus-visible {
  outline: 2px solid var(--brand-blue);
  outline-offset: 1px;
}

.psw-item-active,
.psw-item-active:hover {
  color: var(--brand-blue);
  background: var(--brand-blue-50);
  border-color: var(--brand-blue-100);
  font-weight: 700;
  box-shadow: 0 1px 2px rgba(73, 34, 41, 0.06);
}

@media (max-width: 360px) {
  .psw-item {
    padding: 0.5rem 0.35rem;
    gap: 0.3rem;
    font-size: 0.7rem;
  }
  .psw-item i { font-size: 0.85rem; }
}
</style>
