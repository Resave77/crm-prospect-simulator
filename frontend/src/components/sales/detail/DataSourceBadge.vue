<template>
  <span class="ds-badge" :class="`ds-${source}`" :title="tooltip" role="img" :aria-label="tooltip">
    <i :class="iconClass" aria-hidden="true" />
    {{ displayLabel }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type DataSourceType = 'google' | 'manual' | 'system' | 'prospect'

const props = withDefaults(defineProps<{
  source: DataSourceType
  label?: string
}>(), {})

const labels: Record<DataSourceType, string> = {
  google: 'Google Places',
  manual: 'Manual Admin',
  system: 'CRM System',
  prospect: 'From Prospect',
}

const icons: Record<DataSourceType, string> = {
  google: 'pi pi-map-marker',
  manual: 'pi pi-pencil',
  system: 'pi pi-cog',
  prospect: 'pi pi-arrow-right',
}

const displayLabel = computed(() => props.label ?? labels[props.source])
const tooltip = computed(() => `Data source: ${displayLabel.value}`)
const iconClass = computed(() => icons[props.source])
</script>

<style scoped>
.ds-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  font-size: 0.55rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  white-space: nowrap;
  line-height: 1.4;
  vertical-align: middle;
}
.ds-badge i { font-size: 0.5rem; }
.ds-google { background: #fff0f1; color: #e63946; }
.ds-manual { background: #fffbeb; color: #b45309; }
.ds-system { background: #f5f3ff; color: #c54b59; }
.ds-prospect { background: #ecfdf5; color: #059669; }
</style>
