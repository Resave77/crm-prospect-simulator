export interface VisitRecord {
  id: string
  sales: string
  customer: string
  location: string
  checkIn: string
  checkOut: string | null
  duration: string
  distance: string
  latitude: number
  longitude: number
  radiusStatus: string
  result: string
  selfie?: string
}
