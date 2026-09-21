import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
export const useUserPreferenceStore = defineStore('userPreference', () => {
  const filterDisplayMode = ref<'inline' | 'drawer'>('inline')
  const filterChipPosition = ref<'inline' | 'top' | 'bottom'>('top')
  const isMobile = computed(() => typeof window !== 'undefined' && window.innerWidth < 768)
  return { filterDisplayMode, filterChipPosition, isMobile }
})
