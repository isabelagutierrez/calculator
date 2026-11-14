import { afterEach, describe, expect, it, vi } from 'vitest'
import { calculateExpression, calculateUnary } from './api'

const jsonResponse = (payload: unknown, ok = true) => ({
  ok,
  headers: new Headers({ 'Content-Type': 'application/json' }),
  json: async () => payload,
})

describe('API client', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  it('sends expression and unary request bodies exactly', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(jsonResponse({ result: 16 }))
      .mockResolvedValueOnce(jsonResponse({ result: 9 }))
      .mockResolvedValueOnce(jsonResponse({ result: 0.25 }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(calculateExpression('8 + 8')).resolves.toBe(16)
    await expect(calculateUnary('square-root', 81)).resolves.toBe(9)
    await expect(calculateUnary('percentage', 25)).resolves.toBe(0.25)

    expect(fetchMock.mock.calls[0][1]).toMatchObject({
      method: 'POST',
      body: JSON.stringify({ expression: '8 + 8' }),
    })
    expect(fetchMock.mock.calls[1][1]).toMatchObject({
      method: 'POST',
      body: JSON.stringify({ value: 81 }),
    })
    expect(fetchMock.mock.calls[2][1]).toMatchObject({
      method: 'POST',
      body: JSON.stringify({ value: 25 }),
    })
  })

  it('maps JSON API errors and invalid responses', async () => {
    vi.stubGlobal('fetch', vi.fn()
      .mockResolvedValueOnce(jsonResponse({ error: 'Cannot divide by zero.' }, false))
      .mockResolvedValueOnce(jsonResponse({ result: Infinity }))
      .mockResolvedValueOnce({ ok: false, headers: new Headers({ 'Content-Type': 'text/html' }), json: async () => '<html>' })
      .mockResolvedValueOnce({ ok: false, headers: new Headers(), json: async () => { throw new Error('not json') } }))

    await expect(calculateExpression('8 / 0')).rejects.toThrow('Cannot divide by zero.')
    await expect(calculateExpression('bad')).rejects.toThrow('invalid response')
    await expect(calculateExpression('bad')).rejects.toThrow('unavailable')
    await expect(calculateExpression('bad')).rejects.toThrow('unavailable')
  })

  it('maps network failures and abort timeouts', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValueOnce(new Error('Failed to fetch')))
    await expect(calculateExpression('1')).rejects.toThrow('Cannot reach the calculation service')

    vi.useFakeTimers()
    vi.stubGlobal('fetch', vi.fn().mockImplementation((_input: string, init: RequestInit) => new Promise((_resolve, reject) => {
      init.signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')))
    })))
    const result = calculateExpression('1')
    const assertion = expect(result).rejects.toThrow('too long')
    await vi.advanceTimersByTimeAsync(8000)
    await assertion
  })
})
