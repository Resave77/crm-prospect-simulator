import { computed, onMounted, onUnmounted, ref } from 'vue'

export function useIsMobile(breakpoint = 768) {
  const width = ref(typeof window === 'undefined' ? breakpoint : window.innerWidth)
  const update = () => { width.value = window.innerWidth }
  onMounted(() => window.addEventListener('resize', update))
  onUnmounted(() => window.removeEventListener('resize', update))
  return computed(() => width.value < breakpoint)
}
