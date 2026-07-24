export type ProspectStatus = 'NEW_LEAD' | 'CONTACTED' | 'INTERESTED' | 'QUALIFIED' | 'PROPOSAL_SENT' | 'NEGOTIATION' | 'WON' | 'LOST' | 'CONVERTED'
export type ParentMethod = 'MANUAL_ENTRY' | 'MATCH_CUSTOMER_NAME' | 'EXISTING_COMPANY'

export interface Prospect {
  id: string
  googlePlaceId: string
  placeName: string
  formattedAddress: string
  latitude: number | null
  longitude: number | null
  placeCategory: string
  industryGroup: string
  placeTypes: string[]
  phoneNumber: string
  websiteUrl: string
  googleMapsUrl: string
  assignedSalesExecutiveId: string
  assignedSalesExecutive: string
  visitNotes: string
  followUpNotes: string
  status: ProspectStatus
  convertedAt?: string
  createdAt: string
  updatedAt: string
}

export interface PlaceResult {
  googlePlaceId: string
  name: string
  category: string
  address: string
  distance: number
  rating: number
  userRatingCount: number
  businessStatus: string
  latitude: number | null
  longitude: number | null
  phone: string
  website: string
  googleMapsUrl: string
  markerCategory: string
  markerColor: string
  markerIcon: string
  placeTypes: string[]
}

export interface SalesExecutiveOption { id: string; fullName: string; activeProspectCount: number }

export interface ProspectHistory {
  id: string
  fromStatus: ProspectStatus | null
  toStatus: ProspectStatus
  changedByUserId: string
  changedByName: string
  notes: string
  createdAt: string
}

export interface ProspectReview {
  prospect: Prospect
  history: ProspectHistory[]
  visits: ProspectVisit[]
}

export interface ProspectVisit {
  id: string
  prospectId: string
  salesExecutiveId: string
  salesExecutiveName: string
  checkInAt: string
  checkOutAt?: string
  checkInLatitude: number
  checkInLongitude: number
  checkOutLatitude?: number
  checkOutLongitude?: number
  selfieReference: string
  visitNotes: string
  followUpNotes: string
}

export interface Address {
  mode: string
  province: string
  district: string
  subDistrict: string
  village: string
  latitude: number | null
  longitude: number | null
  previewAddress: string
}

export interface Contact {
  name: string
  position: string
  phone: string
  email: string
}

export interface PeriodAssignment {
  ownerId: string
  ownerName: string
  startMonth: number
  startYear: number
  end: string
}

export interface ParentCompany {
  id: string
  parentCode: string
  name: string
  address: Address
  contacts: Contact[]
  npwpName: string
  npwpAddress: string
  npwpNumber: string
  termOfPayment: string
  kamAssignments: PeriodAssignment[]
}

export interface CustomerSite {
  id: string
  customerCode: string
  parentCompanyId: string
  parentCode: string
  parentCompanyName: string
  sourceProspectId: string
  sourceGooglePlaceId: string
  name: string
  segment: string
  category: string
  region: string
  address: Address
  contacts: Contact[]
  salesExecutiveId: string
  salesExecutiveName: string
  convertedAt: string
  updatedAt: string
}

export interface CustomerDetail {
  customer: CustomerSite
  parentCompany: ParentCompany
  sourceProspectName: string
}

export interface UserOption {
  id: string
  fullName: string
}

export interface MasterOptions {
  segments: string[]
  categories: string[]
  shipmentCosts: string[]
  invoiceTypes: string[]
  termsOfPayment: string[]
  kams: string[]
  addressSuggestions: Address[]
}

export interface ConversionFormData {
  prospect: ProspectReview
  parentCompanies: ParentCompany[]
  salesExecutives: UserOption[]
  parentCodePreview: string
  customerCodePreview: string
  sellerIdentity: string
  options: MasterOptions
}

export interface ConversionInput {
  customerName: string
  customerSegment: string
  customerCategory: string
  parentMethod: ParentMethod | ''
  existingParentCompanyId: string | null
  parentCompanyName: string
  sameAsSiteAddress: boolean
  siteAddress: Address
  companyAddress: Address
  siteContacts: Contact[]
  companyContacts: Contact[]
  ppn: string
  idTkuNumber: string
  nik: string
  companyNpwpName: string
  companyNpwpAddress: string
  companyNpwpNumber: string
  shipmentCost: string
  invoiceType: string
  bankAccount: string
  termOfPayment: string
  billToSource: string
  shipToSource: string
  billingAddressPreview: string
  shippingAddressPreview: string
  salesExecutiveId: string
  salesAssignments: PeriodAssignment[]
  kamAssignments: PeriodAssignment[]
}

export interface CustomerListParams {
  page: number
  limit: number
  keyword: string
  segment: string
  category: string
  sales: string
  region: string
  sort: string
}

export interface CustomerListResult {
  items: CustomerSite[]
  total: number
  page: number
  limit: number
  pages: number
}

export interface ListFilterOptions {
  segments: string[]
  categories: string[]
  regions: string[]
  salesExecutives: SalesExecutiveOption[]
}
