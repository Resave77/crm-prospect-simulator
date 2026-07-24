export function openGoogleMapsNavigation(options: {
  latitude?: number | null
  longitude?: number | null
  address?: string
  googleMapsUrl?: string
}) {
  const { latitude, longitude, address, googleMapsUrl } = options
  if (latitude != null && longitude != null) {
    window.open(`https://www.google.com/maps/dir/?api=1&destination=${latitude},${longitude}`, '_blank', 'noopener,noreferrer')
  } else if (googleMapsUrl) {
    window.open(googleMapsUrl, '_blank', 'noopener,noreferrer')
  } else if (address) {
    window.open(`https://www.google.com/maps/dir/?api=1&destination=${encodeURIComponent(address)}`, '_blank', 'noopener,noreferrer')
  }
}

export function haversineKm(lat1: number, lng1: number, lat2: number, lng2: number): number {
  const R = 6371
  const dLat = ((lat2 - lat1) * Math.PI) / 180
  const dLng = ((lng2 - lng1) * Math.PI) / 180
  const a = Math.sin(dLat / 2) ** 2 + Math.cos((lat1 * Math.PI) / 180) * Math.cos((lat2 * Math.PI) / 180) * Math.sin(dLng / 2) ** 2
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}

export function formatDistance(km: number): string {
  if (km < 1) return `${Math.round(km * 1000)} m`
  return `${km.toFixed(1)} km`
}

export function getDistanceTo(
  targetLat: number | null,
  targetLng: number | null,
  userLat: number | null,
  userLng: number | null,
): number | null {
  if (targetLat == null || targetLng == null || userLat == null || userLng == null) return null
  return haversineKm(userLat, userLng, targetLat, targetLng)
}
