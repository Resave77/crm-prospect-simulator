import type { Prospect, CustomerSite } from '../types/crm'

export interface EntityPlaceInfo {
  placeId: string
  name: string
  formattedAddress: string
  latitude: number | null
  longitude: number | null
  phone: string
  website: string
  googleMapsUrl: string
  categories: string[]
  placeTypes: string[]
  industryGroup: string
}

export function normalizeProspectPlace(p: Prospect): EntityPlaceInfo {
  return {
    placeId: p.googlePlaceId ?? '',
    name: p.placeName ?? '',
    formattedAddress: p.formattedAddress ?? '',
    latitude: safeNum(p.latitude),
    longitude: safeNum(p.longitude),
    phone: p.phoneNumber ?? '',
    website: p.websiteUrl ?? '',
    googleMapsUrl: p.googleMapsUrl ?? '',
    categories: [p.placeCategory, p.industryGroup].filter(Boolean),
    placeTypes: Array.isArray(p.placeTypes) ? p.placeTypes : [],
    industryGroup: p.industryGroup ?? '',
  }
}

export function normalizeCustomerPlace(customer: CustomerSite, sourceProspectName: string): EntityPlaceInfo {
  return {
    placeId: customer.sourceGooglePlaceId ?? '',
    name: customer.name ?? '',
    formattedAddress: customer.address?.previewAddress ?? '',
    latitude: safeNum(customer.address?.latitude),
    longitude: safeNum(customer.address?.longitude),
    phone: customer.contacts?.[0]?.phone ?? '',
    website: '',
    googleMapsUrl: '',
    categories: [customer.category, customer.segment].filter(Boolean),
    placeTypes: [],
    industryGroup: '',
  }
}

function safeNum(v: number | null | undefined): number | null {
  if (v == null) return null
  const n = Number(v)
  return Number.isFinite(n) ? n : null
}

export function formatPlaceType(t: string): string {
  return t.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}

export function businessStatusLabel(s: string): string {
  if (s === 'OPERATIONAL') return 'Operational'
  if (s === 'CLOSED_TEMPORARILY') return 'Temporarily closed'
  if (s === 'CLOSED_PERMANENTLY') return 'Permanently closed'
  return s
}

export function isValidWebsite(url: string): boolean {
  if (!url) return false
  try {
    const u = new URL(url.startsWith('http') ? url : `https://${url}`)
    return u.protocol === 'http:' || u.protocol === 'https:'
  } catch { return false }
}

export function websiteDisplayUrl(url: string): string {
  return url.replace(/^https?:\/\//, '').replace(/\/+$/, '')
}

export function isValidPhone(phone: string): boolean {
  return !!phone && /^[\d\s+\-().]{7,}$/.test(phone)
}

export function copyToClipboard(text: string): void {
  navigator.clipboard?.writeText(text).catch(() => {})
}
