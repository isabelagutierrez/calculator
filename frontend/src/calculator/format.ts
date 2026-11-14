export function formatNumber(value: number): string {
  if (!Number.isFinite(value)) return 'Error'
  if (Object.is(value, -0)) return '0'
  return Number(value.toPrecision(12)).toString()
}

export function formatInput(value: string): string {
  if (value === '' || value === '-' || value === '0.') return value
  const digits = value.replace(/[-.]/g, '').length
  if (digits <= 15) return value
  const numeric = Number(value)
  return Number.isFinite(numeric) ? formatNumber(numeric) : 'Error'
}
