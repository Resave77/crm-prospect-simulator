import { defineStore } from 'pinia'
export const useDataTablePreferencesStore = defineStore('dataTablePreferences', () => { const state = new Map<string, Record<string, unknown>>(); return { hydrate: () => undefined, getColumnState: (key: string) => state.get(key) || {}, setColumnState: (key: string, value: Record<string, unknown>) => state.set(key, value) } })
