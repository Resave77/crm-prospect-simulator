import { getMyProspect, getMyCustomer } from '../api/crm'
import type { Prospect, CustomerSite, ProspectReview, CustomerDetail } from '../types/crm'

export type VisitEntityType = 'prospect' | 'customer'

export interface VisitEntityContext {
  entityType: VisitEntityType
  entityId: string
  name: string
  subtitle: string
  address: string
  latitude: number | null
  longitude: number | null
  googleMapsUrl: string
  attendanceRadiusMeters: number
  phone: string
  status: string
}

export const FALLBACK_ATTENDANCE_RADIUS_METERS = 100

export function prospectToVisitEntity(prospect: Prospect): VisitEntityContext {
  return {
    entityType: 'prospect',
    entityId: prospect.id,
    name: prospect.placeName || 'Prospect',
    subtitle: prospect.placeCategory || 'Prospect',
    address: prospect.formattedAddress || 'No address',
    latitude: prospect.latitude,
    longitude: prospect.longitude,
    googleMapsUrl: prospect.googleMapsUrl || '',
    attendanceRadiusMeters: FALLBACK_ATTENDANCE_RADIUS_METERS,
    phone: prospect.phoneNumber || '',
    status: prospect.status,
  }
}

export function customerToVisitEntity(customer: CustomerSite): VisitEntityContext {
  return {
    entityType: 'customer',
    entityId: customer.id,
    name: customer.name || 'Customer',
    subtitle: customer.parentCompanyName || 'Customer Existing',
    address: customer.address?.previewAddress || 'No address',
    latitude: customer.address?.latitude ?? null,
    longitude: customer.address?.longitude ?? null,
    googleMapsUrl: '',
    attendanceRadiusMeters: FALLBACK_ATTENDANCE_RADIUS_METERS,
    phone: customer.contacts?.[0]?.phone || '',
    status: 'ACTIVE',
  }
}

export async function fetchProspectVisitData(entityId: string): Promise<{ entity: VisitEntityContext; review: ProspectReview }> {
  const review = await getMyProspect(entityId)
  return { entity: prospectToVisitEntity(review.prospect), review }
}

export async function fetchCustomerVisitData(entityId: string): Promise<{ entity: VisitEntityContext; detail: CustomerDetail }> {
  const detail = await getMyCustomer(entityId)
  return { entity: customerToVisitEntity(detail.customer), detail }
}

export function isValidEntityType(value: unknown): value is VisitEntityType {
  return value === 'prospect' || value === 'customer'
}

export function normalizeRouteId(value: unknown): string {
  if (Array.isArray(value)) return String(value[0] ?? '').trim()
  return String(value ?? '').trim()
}
