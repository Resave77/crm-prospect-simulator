// Shared category emoji icons used by Master Data and Prospect Finder.
// Legacy keys are intentionally kept so no icon ever disappears when a
// category is renamed.
const CATEGORY_ICONS: Record<string, string> = {
  // B2B categories
  'resto & cafe': '🍽️',
  'resto & café': '🍽️',
  'qsr / fast food': '🍔',
  'bakery & dessert': '🎂',
  'hotel & accommodation': '🏨',
  'hotels & accommodation': '🏨',
  'catering & event': '🎪',
  'industry / manufacturer': '🏭',
  'potential institutional': '🏫',
  institutional: '🏫',
  'distributor / agent': '📦',
  // B2C categories
  'modern trade': '🛒',
  'convenience store': '🏪',
  'general trade': '🏬',
  'toko bahan kue / baking supply': '🥣',
  // Legacy consumer categories
  'individual consumer': '🧍',
  'household consumer': '🏠',
  'online consumer': '🛍️'
}

export function categoryIcon(name: string): string {
  const key = name.trim().toLowerCase()
  return CATEGORY_ICONS[key] ?? name.charAt(0).toUpperCase()
}
