export class ApiError extends Error {}

type ResponsePayload = { result?: number; error?: string }

const baseUrl = () => import.meta.env.VITE_API_BASE_URL || window.location.origin

async function postJson(path: string, body: unknown): Promise<number> {
  const controller = new AbortController()
  const timeout = window.setTimeout(() => controller.abort(), 8000)
  try {
    let response: Response
    try {
      response = await fetch(new URL(path, baseUrl()), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
        signal: controller.signal,
      })
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') {
        throw new ApiError('The calculation service took too long to respond.')
      }
      throw new ApiError('Cannot reach the calculation service. Please check that the backend is running.')
    }
    if (!response.headers.get('Content-Type')?.toLowerCase().startsWith('application/json')) {
      throw new ApiError('The calculation service is unavailable. Please try again.')
    }
    let payload: ResponsePayload
    try {
      payload = await response.json() as ResponsePayload
    } catch {
      throw new ApiError('The calculation service is unavailable. Please try again.')
    }
    if (!response.ok) throw new ApiError(payload.error ?? 'The calculation service is unavailable. Please try again.')
    if (typeof payload.result !== 'number' || !Number.isFinite(payload.result)) {
      throw new ApiError('The calculation service returned an invalid response.')
    }
    return payload.result
  } finally {
    window.clearTimeout(timeout)
  }
}

export const evaluateExpression = (expression: string) => postJson('/api/calculate', { expression })
export const squareRoot = (value: number) => postJson('/api/square-root', { value })
export const percentage = (value: number) => postJson('/api/percentage', { value })
