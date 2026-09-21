export type ShortListItem<T = number> = { id: T; name: string; [key: string]: unknown }
export type ShortListParams = { page: number; page_size: number; search?: string; [key: string]: unknown }
export type ShortListResult<T = number> = { items: ShortListItem<T>[]; pagination?: { page: number; total_pages: number; total?: number } }
