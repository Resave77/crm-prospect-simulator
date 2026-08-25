import { api } from './client'
import type { ApiEnvelope } from '../types/auth'

export interface MasterDataSegment {
  id: string
  name: string
  description: string
  status: 'ACTIVE' | 'INACTIVE'
  categoryCount: number
  createdAt: string
  updatedAt: string
  deletedAt?: string | null
}

export interface MasterDataCategory {
  id: string
  segmentId: string
  segmentName: string
  name: string
  description: string
  placeApi: string
  status: 'ACTIVE' | 'INACTIVE'
  createdAt: string
  updatedAt: string
  deletedAt?: string | null
}

export interface SegmentInput {
  name: string
  description: string
  status: 'ACTIVE' | 'INACTIVE'
}

export interface CategoryInput {
  segmentId: string
  name: string
  description: string
  placeApi?: string
  status: 'ACTIVE' | 'INACTIVE'
}

export async function listSegments(params: { search?: string; status?: string } = {}) {
  return (await api.get<ApiEnvelope<MasterDataSegment[]>>('/master-data/segments', { params })).data.data
}

export async function createSegment(input: SegmentInput) {
  return (await api.post<ApiEnvelope<MasterDataSegment>>('/master-data/segments', input)).data.data
}

export async function updateSegment(id: string, input: SegmentInput) {
  return (await api.put<ApiEnvelope<MasterDataSegment>>(`/master-data/segments/${id}`, input)).data.data
}

export async function deleteSegment(id: string) {
  await api.delete(`/master-data/segments/${id}`)
}

export async function listCategories(params: { search?: string; segmentId?: string; status?: string } = {}) {
  return (await api.get<ApiEnvelope<MasterDataCategory[]>>('/master-data/categories', { params })).data.data
}

export async function listCategoriesBySegment(segmentId: string) {
  return (await api.get<ApiEnvelope<MasterDataCategory[]>>(`/master-data/segments/${segmentId}/categories`)).data.data
}

export async function createCategory(input: CategoryInput) {
  return (await api.post<ApiEnvelope<MasterDataCategory>>('/master-data/categories', input)).data.data
}

export async function updateCategory(id: string, input: CategoryInput) {
  return (await api.put<ApiEnvelope<MasterDataCategory>>(`/master-data/categories/${id}`, input)).data.data
}

export async function deleteCategory(id: string) {
  await api.delete(`/master-data/categories/${id}`)
}

export async function listTrashedSegments() {
  return (await api.get<ApiEnvelope<MasterDataSegment[]>>('/master-data/trash/segments')).data.data
}

export async function restoreSegment(id: string) {
  await api.post(`/master-data/trash/segments/${id}/restore`)
}

export async function listTrashedCategories() {
  return (await api.get<ApiEnvelope<MasterDataCategory[]>>('/master-data/trash/categories')).data.data
}

export async function restoreCategory(id: string) {
  await api.post(`/master-data/trash/categories/${id}/restore`)
}
