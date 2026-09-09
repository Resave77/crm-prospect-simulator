import { useIsMobile } from './useIsMobile'

export function useDialogMobileBottomBar() {
  const isMobile = useIsMobile()
  return { isMobile, shouldUseMobileBottomBar: isMobile }
}
