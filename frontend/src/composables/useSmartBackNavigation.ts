import { useRouter } from 'vue-router'

export function useSmartBackNavigation(fallback = '/') {
  const router = useRouter()
  const goBack = () => {
    if (window.history.length > 1) router.back()
    else router.push(fallback)
  }
  return { goBack, smartBack: goBack }
}
