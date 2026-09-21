import { useIsMobile } from './useIsMobile'

import type { MaybeRef } from 'vue'

export type DialogMobileBottomBarSubmenuItem = {
  key: string
  label: string
  icon?: string
  visible?: MaybeRef<boolean>
  disabled?: MaybeRef<boolean>
  loading?: MaybeRef<boolean>
  onClick: () => void
  permissions?: string[]
}

export type DialogMobileBottomBarAction = {
  key: string
  label: string
  icon?: string
  type?: 'button'
  visible?: MaybeRef<boolean>
  disabled?: MaybeRef<boolean>
  loading?: MaybeRef<boolean>
  active?: MaybeRef<boolean>
  onClick: () => void
  permissions?: string[]
} | {
  key: string
  label: string
  icon?: string
  type: 'submenu'
  items: DialogMobileBottomBarSubmenuItem[]
  visible?: MaybeRef<boolean>
  disabled?: MaybeRef<boolean>
  loading?: MaybeRef<boolean>
  active?: MaybeRef<boolean>
  panelTitle?: string
  permissions?: string[]
}

export type DialogMobileStepperConfig = {
  activeStep: number
  totalSteps: number
  canGoNext: boolean
  onBack: () => void
  onNext: () => void
  onSubmit: () => void
  submitLoading?: boolean
}

export function useDialogMobileBottomBar() {
  const isMobile = useIsMobile()
  return { isMobile, shouldUseMobileBottomBar: isMobile }
}
