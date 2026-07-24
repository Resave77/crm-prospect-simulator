import { api } from './client'
import type { ApiEnvelope } from '../types/auth'
import type { ConversionFormData, ConversionInput, CustomerDetail, CustomerListParams, CustomerListResult, CustomerSite, ListFilterOptions, ParentCompany, PlaceResult, Prospect, ProspectReview, ProspectStatus, ProspectVisit, SalesExecutiveOption } from '../types/crm'

export async function getMyProspects() {
  return (await api.get<ApiEnvelope<Prospect[]>>('/sales/prospects')).data.data
}

export async function transitionProspect(id: string, status: ProspectStatus, notes: string) {
  return (await api.patch<ApiEnvelope<Prospect>>(`/sales/prospects/${id}/transition`, { status, notes })).data.data
}

export async function getMyProspect(id: string) { return (await api.get<ApiEnvelope<ProspectReview>>(`/sales/prospects/${id}`)).data.data }
export async function checkInProspect(id: string, input: { latitude: number; longitude: number; visitNotes: string; selfiePlaceholder: boolean }) {
  return (await api.post<ApiEnvelope<ProspectVisit>>(`/sales/prospects/${id}/visits/check-in`, input)).data.data
}
export async function checkOutProspect(id: string, visitId: string, input: { latitude: number; longitude: number; followUpNotes: string }) {
  return (await api.patch<ApiEnvelope<ProspectVisit>>(`/sales/prospects/${id}/visits/${visitId}/check-out`, input)).data.data
}
export async function getPipeline() { return (await api.get<ApiEnvelope<Prospect[]>>('/admin/prospects/pipeline')).data.data }
export async function getSalesExecutives() { return (await api.get<ApiEnvelope<SalesExecutiveOption[]>>('/admin/sales-executives')).data.data }
export async function searchPlaces(params: { keyword: string; categories: string; radius: number; latitude: number; longitude: number }) { return (await api.get<ApiEnvelope<PlaceResult[]>>('/admin/prospect-finder/search', { params })).data.data }
export async function saveProspect(place: PlaceResult, industryGroup: string, assignedSalesExecutiveId: string) { return (await api.post<ApiEnvelope<Prospect>>('/admin/prospects', { place, industryGroup, assignedSalesExecutiveId })).data.data }

export async function getWonProspects() {
  return (await api.get<ApiEnvelope<Prospect[]>>('/admin/prospects/won')).data.data
}

export async function getProspectReview(id: string) {
  return (await api.get<ApiEnvelope<ProspectReview>>(`/admin/prospects/${id}`)).data.data
}

export async function getConversionForm(id: string) {
  return (await api.get<ApiEnvelope<ConversionFormData>>(`/admin/prospects/${id}/conversion-form`)).data.data
}

export async function searchParentCompanies(search = '') {
  return (await api.get<ApiEnvelope<ParentCompany[]>>('/admin/parent-companies', { params: { search } })).data.data
}

export async function convertProspect(id: string, input: ConversionInput) {
  return (await api.post<ApiEnvelope<CustomerSite>>(`/admin/prospects/${id}/convert`, input)).data.data
}

export async function getAdminCustomers() {
  return (await api.get<ApiEnvelope<CustomerSite[]>>('/admin/customers')).data.data
}

export async function getMyCustomers() {
  return (await api.get<ApiEnvelope<CustomerSite[]>>('/sales/customers')).data.data
}

export async function getMyCustomer(id: string) {
  return (await api.get<ApiEnvelope<CustomerDetail>>(`/sales/customers/${id}`)).data.data
}

export async function getAdminCustomersList(params: CustomerListParams) {
  return (await api.get<ApiEnvelope<CustomerListResult>>('/admin/customers/list', { params })).data.data
}

export async function getCustomerFilterOptions() {
  return (await api.get<ApiEnvelope<ListFilterOptions>>('/admin/customers/filter-options')).data.data
}

export async function getAdminCustomer(id: string) {
  return (await api.get<ApiEnvelope<CustomerDetail>>(`/admin/customers/${id}`)).data.data
}

export async function deleteCustomer(id: string) {
  await api.delete(`/admin/customers/${id}`)
}
