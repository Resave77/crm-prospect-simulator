export type ProspectStatus = 'NEW_LEAD' | 'CONTACTED' | 'INTERESTED' | 'QUALIFIED' | 'PROPOSAL_SENT' | 'NEGOTIATION' | 'WON' | 'LOST' | 'CONVERTED'
export type ParentMethod = 'MANUAL_ENTRY' | 'MATCH_CUSTOMER_NAME' | 'EXISTING_COMPANY'

export interface AIStatus {
  enabled: boolean
  configured: boolean
  modelConfigured: boolean
}

export interface ProspectInitialAnalysis {
  prospectId: string
  summary?: Record<string, unknown> | null
  menu?: Record<string, unknown> | null
  status: 'PENDING' | 'SUCCESS' | 'FAILED'
  errorCode?: string
}
export interface ProspectMenuProfileItem {
  menuName: string
  profile: string
  yoghurtFit: 'HIGH' | 'MEDIUM' | 'LOW' | 'UNKNOWN'
  opportunity: string
  reason: string
  recommendedSalesAction: string
  confidence: 'HIGH' | 'MEDIUM' | 'LOW'
}
export interface ProspectMenuProfile {
  menus?: ProspectMenuProfileItem[]
  menuOpportunity?: 'HIGH' | 'MEDIUM' | 'LOW' | 'UNKNOWN'
  yoghurtFit?: 'HIGH' | 'MEDIUM' | 'LOW' | 'UNKNOWN'
  topOpportunity: string
  why?: string
  recommendedAction: string
  confidence?: 'HIGH' | 'MEDIUM' | 'LOW'
  sourcePhotoNames?: string[]
  state?: string
  message?: string
}
export type MenuBranchMatch = 'exact_branch' | 'likely_same_branch' | 'brand_only' | 'uncertain'
export interface ProspectMenuFindingSource { name: string; url: string; branchMatch: MenuBranchMatch }
export interface ProspectMenuFindingPrice { source: string; sourceUrl: string; price: number; currency: 'IDR' | string }
export interface ProspectMenuFindingItem { name: string; description: string | null; imageUrl: string | null; prices: ProspectMenuFindingPrice[]; availability: string; branchMatch: MenuBranchMatch; confidence: number }
export interface ProspectMenuFindingCategory { name: string; items: ProspectMenuFindingItem[] }
export interface ProspectMenuFinding { status: 'FOUND' | 'NOT_FOUND' | 'MENU_SOURCE_NOT_AVAILABLE'; message?: string; place: { name: string; branch: string; address: string; googlePlaceId: string; googleMapsUrl: string }; sources: ProspectMenuFindingSource[]; categories: ProspectMenuFindingCategory[]; summary: { sourceCount: number; totalCategories: number; totalItems: number } }
export interface ProspectMenuDocument { discovery?: ProspectMenuFinding | null; profiling?: ProspectMenuProfile | null; finding?: ProspectMenuFinding | null; profile?: ProspectMenuProfile | null }

export interface ProspectChatResponse {
  answer: string
  skill: string
  insight: string
  why: string
  recommendedAction: string
}
export interface ProspectAIChatHistory extends ProspectChatResponse {
  id: string
  message: string
  userId?: string
  authorName?: string
  authorRole?: string
  createdAt: string
}

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
  deletionRequested: boolean
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
  hasMenuPhotos?: boolean
  isCustomer?: boolean
  customerId?: string
}

export interface CustomerMarker {
  customerId: string
  customerCode: string
  googlePlaceId: string
  name: string
  address: string
  latitude: number | null
  longitude: number | null
}

export interface SalesExecutiveOption { id: string; fullName: string; activeProspectCount: number }

export interface TeamDashboardMember {
  userId: string
  fullName: string
  roleName: string
  roleLevel: number
  parentUserId: string
  activeProspects: number
  customers: number
  visitsToday: number
  completedVisits: number
  pendingVisits: number
}

export interface TeamDashboard {
  lead: {
    id: string
    fullName: string
    roleName: string
    roleLevel: number
    effectiveFrom: string
  }
  hasTeam: boolean
  directMemberCount: number
  totalDescendantCount: number
  activeProspects: number
  customers: number
  visitsToday: number
  completedVisits: number
  pendingVisits: number
  pipelineCounts: Partial<Record<ProspectStatus, number>>
  members: TeamDashboardMember[]
}

export interface AssignedSalesInfo {
  userId: string
  fullName: string
  roleName: string
  roleLevel: number
}

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
  visitResult: string
  visitOutcome: string
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
  assignedSales?: AssignedSalesInfo | null
  convertedAt: string
  updatedAt: string
}

export interface TeamCustomers {
  hasTeam: boolean
  directMemberCount: number
  totalDescendantCount: number
  customers: CustomerSite[]
}

export interface CustomerDetail {
  customer: CustomerSite
  parentCompany: ParentCompany
  sourceProspectName: string
}

export interface PlacePhoto {
  name: string
  photoUrl: string
  widthPx: number
  heightPx: number
  attribution: string
  isMenu: boolean
}

export interface MenuImage {
  title: string
  imageUrl: string
  sourceUrl: string
  sourceSite: string
}

export interface PlaceOpeningHours {
  openNow: boolean
  weekdays: string[]
}

export interface PlaceReview {
  authorName: string
  authorPhoto: string
  rating: number
  text: string
  time: string
  languageCode: string
}

export interface PlaceDetails {
  googlePlaceId: string
  placeName: string
  formattedAddress: string
  latitude: number
  longitude: number
  placeCategory: string
  placeTypes: string[]
  phoneNumber: string
  internationalPhone: string
  websiteUrl: string
  googleMapsUrl: string
  rating: number
  userRatingCount: number
  businessStatus: string
  priceLevel: string
  editorialSummary: string
  utcOffsetMinutes: number
  photos: PlacePhoto[]
  openingHours: PlaceOpeningHours | null
  reviews: PlaceReview[]
  delivery: boolean
  dineIn: boolean
  takeout: boolean
  curbsidePickup: boolean
  parkingOptions: PlaceParking | null
  paymentOptions: PlacePayments | null
  accessibilityOptions: PlaceAccessibility | null
}

export interface PlaceParking {
  paidStreetParking: boolean
  paidParkingLot: boolean
  freeStreetParking: boolean
  freeParkingLot: boolean
  valetParking: boolean
  garageParking: boolean
}

export interface PlacePayments {
  cashOnly: boolean
  creditCardOnly: boolean
  debitCardOnly: boolean
  nfcOnly: boolean
}

export interface PlaceAccessibility {
  wheelchairAccessibleEntrance: boolean
  wheelchairAccessibleParking: boolean
  wheelchairAccessibleRestroom: boolean
  wheelchairAccessibleSeating: boolean
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
  placeDetails?: PlaceDetails
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

export interface VisitMonitoringItem {
  id: string
  prospectId: string
  customerId?: string
  entityType: 'prospect' | 'customer'
  customerName: string
  customerCategory: string
  industryGroup: string
  formattedAddress: string
  phoneNumber: string
  prospectLatitude: number | null
  prospectLongitude: number | null
  salesExecutiveId: string
  salesExecutiveName: string
  checkInAt: string
  checkOutAt?: string
  checkInLatitude: number
  checkInLongitude: number
  checkOutLatitude?: number
  checkOutLongitude?: number
  distanceMeters: number
  durationSeconds?: number
  radiusStatus: 'INSIDE' | 'OUTSIDE' | 'UNKNOWN'
  prospectStatus: ProspectStatus
  selfieReference: string
  visitNotes: string
  followUpNotes: string
  visitResult: string
  visitOutcome: string
  visitCount: number
}

export interface VisitMonitoringFilters {
  dateFrom: string
  dateTo: string
  salesExecutiveId: string
  customerName: string
  radiusStatus: string
}

export interface AdminReport {
  kpi: { totalVisits: number; withinRadius: number; outsideRadius: number; wonProspects: number }
  trends: { label: string; withinRadius: number; outsideRadius: number }[]
  stages: { label: string; count: number }[]
  performance: { salesExecutiveId: string; salesExecutiveName: string; territory: string; visits: number; withinRadius: number; prospectsWon: number; conversion: number; performance: number }[]
  territories: string[]
}

export interface UpdateParentCompanyInput {
  name: string
  termOfPayment: string
  npwpName: string
  npwpAddress: string
  npwpNumber: string
  companyAddress?: Address
  companyContacts?: Contact[]
}

export interface ProspectComment {
  id: string
  prospectId: string
  userId: string
  userName: string
  content: string
  attachments: ProspectCommentAttachment[]
  createdAt: string
  updatedAt: string
}

export interface ProspectCommentAttachment {
  id: string
  name: string
  contentType: string
  size: number
}

export type PhotoCategory = 'MENU' | 'PLACE'

export interface ProspectPhotoTag {
  id: string
  prospectId: string
  photoName: string | null
  photoIndex: number | null
  category: PhotoCategory
  updatedBy: string
  createdAt: string
  updatedAt: string
}
