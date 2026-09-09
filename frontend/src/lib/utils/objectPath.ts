export function getObjectPathValue(value: unknown, path: string) {
  return path.split('.').reduce((current: any, key) => current?.[key], value as any)
}
export const createObjectPathValue = getObjectPathValue
