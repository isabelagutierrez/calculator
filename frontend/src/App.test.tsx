import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'

const response = (result: number, ok = true) => ({
  ok,
  headers: new Headers({ 'Content-Type': 'application/json' }),
  json: async () => ok ? { result } : { error: 'Cannot divide by zero.' },
})

describe('Calculator interface', () => {
  afterEach(() => cleanup())
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response(16)))
  })

  it.each([
    ['addition', ['8', '+', '8'], '8 + 8'],
    ['subtraction', ['8', '−', '3'], '8 − 3'],
    ['multiplication', ['2', '×', '3'], '2 × 3'],
    ['division', ['8', '÷', '2'], '8 ÷ 2'],
    ['power', ['2', 'xʸ', '3'], '2 xʸ 3'],
    ['parentheses', ['(', '2', '+', '3', ')'], '(2 + 3)'],
  ])('sends the exact expression for %s', async (_name, keys, expression) => {
    const user = userEvent.setup()
    render(<App />)
    for (const key of keys) await user.click(screen.getByRole('button', { name: key }))
    await user.click(screen.getByRole('button', { name: '=' }))
    await waitFor(() => expect(screen.getAllByText('16').length).toBeGreaterThan(0))
    expect(fetch).toHaveBeenCalledWith(expect.any(URL), expect.objectContaining({
      method: 'POST',
      body: JSON.stringify({ expression }),
    }))
  })

  it('sends square-root and percentage values and records history', async () => {
    const user = userEvent.setup()
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(response(9))
      .mockResolvedValueOnce(response(0.09))
    vi.stubGlobal('fetch', fetchMock)
    render(<App />)

    await user.click(screen.getByRole('button', { name: '9' }))
    await user.click(screen.getByRole('button', { name: '√' }))
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(expect.any(URL), expect.objectContaining({
      body: JSON.stringify({ value: 9 }),
    })))
    await user.click(screen.getByRole('button', { name: '%' }))
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(expect.any(URL), expect.objectContaining({
      body: JSON.stringify({ value: 9 }),
    })))

    await user.click(screen.getByRole('button', { name: 'History' }))
    expect(screen.getByText('√(9) =')).toBeInTheDocument()
    expect(screen.getByText('%(9) =')).toBeInTheDocument()
  })

  it('handles API errors, clearing, backspace, sign and history', async () => {
    const user = userEvent.setup()
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response(0, false)))
    render(<App />)
    await user.click(screen.getByRole('button', { name: '8' }))
    await user.click(screen.getByRole('button', { name: '÷' }))
    await user.click(screen.getByRole('button', { name: '0' }))
    await user.click(screen.getByRole('button', { name: '=' }))
    await waitFor(() => expect(screen.getByText('Cannot divide by zero.')).toBeInTheDocument())

    await user.click(screen.getByRole('button', { name: 'C' }))
    await user.click(screen.getByRole('button', { name: '1' }))
    await user.click(screen.getByRole('button', { name: '2' }))
    await user.click(screen.getByRole('button', { name: '⌫' }))
    expect(screen.getByText('1', { selector: 'strong' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: '±' }))
    expect(document.querySelector('.display strong')).toHaveTextContent('-1')
    await user.click(screen.getByRole('button', { name: 'C' }))
    expect(document.querySelector('.display strong')).toHaveTextContent('')
  })

  it('continues a calculation from the previous result', async () => {
    const user = userEvent.setup()
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(response(10))
      .mockResolvedValueOnce(response(15))
    vi.stubGlobal('fetch', fetchMock)
    render(<App />)

    await user.click(screen.getByRole('button', { name: '5' }))
    await user.click(screen.getByRole('button', { name: '+' }))
    await user.click(screen.getByRole('button', { name: '5' }))
    await user.click(screen.getByRole('button', { name: '=' }))
    await waitFor(() => expect(document.querySelector('.display strong')).toHaveTextContent('10'))

    await user.click(screen.getByRole('button', { name: '+' }))
    await user.click(screen.getByRole('button', { name: '5' }))
    await user.click(screen.getByRole('button', { name: '=' }))
    await waitFor(() => expect(document.querySelector('.display strong')).toHaveTextContent('15'))
    expect(fetchMock).toHaveBeenLastCalledWith(expect.any(URL), expect.objectContaining({
      body: JSON.stringify({ expression: '10 + 5' }),
    }))
  })

  it('supports keyboard input and history drawer controls', async () => {
    const user = userEvent.setup()
    render(<App />)
    await user.keyboard('8+8=')
    await waitFor(() => expect(screen.getAllByText('16').length).toBeGreaterThan(0))
    await user.click(screen.getByRole('button', { name: 'History' }))
    expect(screen.getByRole('heading', { name: 'History' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Clear' }))
    await user.click(screen.getAllByRole('button', { name: 'Close history' })[2])
    expect(screen.getByText(/Your operations/)).toBeInTheDocument()
  })

  it('shows a friendly network error', async () => {
    const user = userEvent.setup()
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Failed to fetch')))
    render(<App />)
    await user.click(screen.getByRole('button', { name: '8' }))
    await user.click(screen.getByRole('button', { name: '=' }))
    await waitFor(() => expect(screen.getByText('Cannot reach the calculation service. Please check that the backend is running.')).toBeInTheDocument())
  })

  it('handles decimals, unary operators and invalid input without crashing', async () => {
    const user = userEvent.setup()
    render(<App />)

    await user.click(screen.getByRole('button', { name: '−' }))
    await user.click(screen.getByRole('button', { name: '2' }))
    await user.click(screen.getByRole('button', { name: '−' }))
    await user.click(screen.getByRole('button', { name: '−' }))
    await user.click(screen.getByRole('button', { name: '3' }))
    await user.click(screen.getByRole('button', { name: '=' }))
    await waitFor(() => expect(fetch).toHaveBeenCalledWith(expect.any(URL), expect.objectContaining({
      body: JSON.stringify({ expression: '-2 − − 3' }),
    })))

    await user.click(screen.getByRole('button', { name: 'C' }))
    await user.click(screen.getByRole('button', { name: '1' }))
    await user.click(screen.getByRole('button', { name: '.' }))
    await user.click(screen.getByRole('button', { name: '.' }))
    await user.click(screen.getByRole('button', { name: '2' }))
    expect(document.querySelector('.display strong')).toHaveTextContent('1.2')

    await user.click(screen.getByRole('button', { name: 'C' }))
    const nine = screen.getByRole('button', { name: '9' })
    for (let index = 0; index < 309; index += 1) {
      fireEvent.click(nine)
    }
    expect(screen.getByText('Number is too large.')).toBeInTheDocument()
  })
})
