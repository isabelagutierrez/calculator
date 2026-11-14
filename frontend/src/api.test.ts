import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, evaluateExpression, percentage, squareRoot } from './api'

const response = (body: unknown, ok = true, contentType = 'application/json') => ({
  ok, headers: new Headers({ 'Content-Type': contentType }), json: async () => body,
})

describe('API client', () => {
  afterEach(() => vi.restoreAllMocks())
  it('posts expressions and unary operations', async () => {
    const fetchMock = vi.fn().mockResolvedValue(response({ result: 4 }))
    vi.stubGlobal('fetch', fetchMock)
    await expect(evaluateExpression('2 + 2')).resolves.toBe(4)
    await expect(squareRoot(16)).resolves.toBe(4)
    await expect(percentage(400)).resolves.toBe(4)
    expect(fetchMock).toHaveBeenCalledTimes(3)
  })
  it('maps backend and invalid response errors', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response({ error: 'Bad input' }, false)))
    await expect(evaluateExpression('x')).rejects.toEqual(new ApiError('Bad input'))
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response({}, true)))
    await expect(evaluateExpression('x')).rejects.toThrow('invalid response')
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response('<html>', false, 'text/html')))
    await expect(evaluateExpression('x')).rejects.toThrow('unavailable')
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, headers: new Headers({ 'Content-Type': 'application/json' }), json: async () => { throw new Error() } }))
    await expect(evaluateExpression('x')).rejects.toThrow('unavailable')
  })
  it('maps network and timeout errors', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Failed to fetch')))
    await expect(evaluateExpression('x')).rejects.toThrow('Cannot reach')
    vi.useFakeTimers()
    vi.stubGlobal('fetch', vi.fn().mockImplementation((_input: URL, init: RequestInit) => new Promise((_resolve, reject) => {
      init.signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')))
    })))
    const promise = evaluateExpression('x')
    const assertion = expect(promise).rejects.toThrow('too long')
    await vi.advanceTimersByTimeAsync(8000)
    await assertion
    vi.useRealTimers()
  })
})
