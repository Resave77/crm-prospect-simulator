/** Safe defaults for the currently deployed development backend. */
export const DEV_CAPABILITIES = {
  prospectTrash: import.meta.env.VITE_ENABLE_PROSPECT_TRASH === 'true' || import.meta.env.VITE_ENABLE_PROSPECT_TRASH === undefined,
  customerTrash: import.meta.env.VITE_ENABLE_CUSTOMER_TRASH === 'true' || import.meta.env.VITE_ENABLE_CUSTOMER_TRASH === undefined,
} as const
