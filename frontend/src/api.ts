type CalculateResponse = { result: number }
type ErrorResponse = { error?: string }

export async function calculateExpression(expression: string): Promise<number> {
  const response = await fetch('/api/calculate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ expression }),
  })

  const payload = await response.json() as CalculateResponse & ErrorResponse
  if (!response.ok) {
    throw new Error(payload.error ?? 'The calculation service is unavailable.')
  }
  if (typeof payload.result !== 'number' || !Number.isFinite(payload.result)) {
    throw new Error('The calculation service returned an invalid result.')
  }
  return payload.result
}

export async function calculateUnary(path: 'square-root' | 'percentage', value: number): Promise<number> {
  const response = await fetch(`/api/${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ value }),
  })

  const payload = await response.json() as CalculateResponse & ErrorResponse
  if (!response.ok) {
    throw new Error(payload.error ?? 'The calculation service is unavailable.')
  }
  if (typeof payload.result !== 'number' || !Number.isFinite(payload.result)) {
    throw new Error('The calculation service returned an invalid result.')
  }
  return payload.result
}
