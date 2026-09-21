export function useDialogBackStack() {
  return { push: () => undefined, pop: () => undefined, clear: () => undefined }
}

export function registerDialogBackHandler(handler: () => void) {
  return () => { void handler }
}
