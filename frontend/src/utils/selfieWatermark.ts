export interface WatermarkMeta {
  salesName: string
  entityType: 'prospect' | 'customer'
  entityName: string
  latitude: number
  longitude: number
  accuracyMeters: number
  insideRadius: boolean
}

export function applySelfieWatermark(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  meta: WatermarkMeta,
) {
  const fontSize = Math.max(10, Math.round(width / 40))
  const padding = Math.round(width / 25)
  const lineHeight = fontSize * 1.4

  const lines = [
    meta.salesName,
    new Date().toLocaleString(),
    `${meta.latitude.toFixed(6)}, ${meta.longitude.toFixed(6)} • Accuracy ±${Math.round(meta.accuracyMeters)}m`,
    `${meta.entityName} • ${meta.entityType === 'prospect' ? 'Prospect' : 'Customer'}`,
    meta.insideRadius ? 'Inside radius' : 'Outside radius',
  ]

  const textHeight = lines.length * lineHeight + padding * 2
  const boxY = height - textHeight

  ctx.fillStyle = 'rgba(0, 0, 0, 0.55)'
  ctx.fillRect(0, boxY, width, textHeight)

  ctx.fillStyle = '#ffffff'
  ctx.font = `bold ${fontSize}px system-ui, -apple-system, sans-serif`
  ctx.textBaseline = 'top'

  lines.forEach((line, i) => {
    ctx.fillText(line, padding, boxY + padding + i * lineHeight)
  })
}
