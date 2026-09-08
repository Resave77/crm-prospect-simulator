<script setup lang="ts">
import { computed } from "vue";

interface Props {
  activeCount?: number;
  icon?: string;
  label?: string;
}

const props = withDefaults(defineProps<Props>(), {
  activeCount: 0,
  icon: "pi-filter",
  label: "",
});

const emit = defineEmits<{
  (e: "click"): void;
}>();

const resolvedLabel = computed(() => props.label || "Filter");
const hasFilters = computed(() => props.activeCount > 0);
</script>

<template>
  <div class="filter-button-container">
    <button
      type="button"
      class="filter-button"
      :class="{ 'has-filters': hasFilters }"
      @click="emit('click')"
    >
      <span class="pi" :class="icon" />
      <span class="filter-button-text">{{ resolvedLabel }}</span>
      <transition name="badge-pop">
        <span v-if="hasFilters" class="filter-button-badge">
          {{ activeCount }}
        </span>
      </transition>
    </button>
  </div>
</template>

<style scoped>
/* Filter Button - Normal button at top */
.filter-button-container {
  display: inline-block;
}

.filter-button {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: 1px solid #e2e8f0;
  background: white;
  color: #475569;
  font-size: 0.875rem;
  font-weight: 500;
  border-radius: 0.5rem;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  cursor: pointer;
  transition: all 0.2s ease;
}

.filter-button:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.filter-button.has-filters {
  background: #dc2626;
  color: white;
  border-color: #dc2626;
}

.filter-button.has-filters:hover {
  background: #b91c1c;
  border-color: #b91c1c;
}

.filter-button .pi {
  font-size: 0.875rem;
}

.filter-button-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 1.125rem;
  height: 1.125rem;
  padding: 0 0.3rem;
  border-radius: 9999px;
  background: white;
  color: #dc2626;
  font-size: 0.65rem;
  font-weight: 700;
}

.filter-button.has-filters .filter-button-badge {
  background: rgba(255, 255, 255, 0.9);
  color: #dc2626;
}

/* Badge pop animation */
.badge-pop-enter-active {
  animation: badge-pop-in 0.3s ease;
}

.badge-pop-leave-active {
  animation: badge-pop-in 0.2s ease reverse;
}

@keyframes badge-pop-in {
  0% {
    transform: scale(0);
    opacity: 0;
  }
  50% {
    transform: scale(1.2);
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

/* Responsive adjustments - Mobile */
@media (max-width: 640px) {
  .filter-button {
    padding: 0.5rem 0.875rem;
    font-size: 0.8125rem;
  }

  .filter-button-text {
    display: inline;
  }

  .filter-button-badge {
    min-width: 1rem;
    height: 1rem;
    font-size: 0.6rem;
  }

  .filter-button .pi {
    font-size: 0.8rem;
  }
}
</style>
