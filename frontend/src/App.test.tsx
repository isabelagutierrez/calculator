import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'

const mockFetch = (result: number, ok = true) => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
    ok,
    json: async () => ok ? { result } : { error: 'Cannot divide by zero.' },
  }))
}

describe('Calculator', () => {
  afterEach(() => cleanup())

  beforeEach(() => {
    vi.restoreAllMocks()
    mockFetch(16)
  })

  it('sends an accumulated expression to the backend', async () => {
    render(<App />)

    fireEvent.click(screen.getByRole('button', { name: '8' }))
    fireEvent.click(screen.getByRole('button', { name: '+' }))
    fireEvent.click(screen.getByRole('button', { name: '8' }))
    fireEvent.click(screen.getByRole('button', { name: '=' }))

    await waitFor(() => expect(screen.getAllByText('16')).not.toHaveLength(0))
    expect(fetch).toHaveBeenCalledWith('/api/calculate', expect.objectContaining({
      method: 'POST',
      body: JSON.stringify({ expression: '8 + 8' }),
    }))
  })

  it('supports exponent and parentheses buttons', () => {
    render(<App />)

    expect(screen.getByRole('button', { name: 'xʸ' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '(' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: ')' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '×' })).toBeInTheDocument()
  })

  it('shows backend errors', async () => {
    mockFetch(0, false)
    render(<App />)

    fireEvent.click(screen.getByRole('button', { name: '8' }))
    fireEvent.click(screen.getByRole('button', { name: '÷' }))
    fireEvent.click(screen.getByRole('button', { name: '0' }))
    fireEvent.click(screen.getByRole('button', { name: '=' }))

    await waitFor(() => expect(screen.getByText('Cannot divide by zero.')).toBeInTheDocument())
  })

  it('opens and closes the mobile history control', () => {
    const { container } = render(<App />)

    const historyButton = screen.getByRole('button', { name: 'History' })
    fireEvent.click(historyButton)
    expect(container.querySelector('.history-panel.is-open')).toBeInTheDocument()
    fireEvent.click(container.querySelector('.history-close') as HTMLButtonElement)
    expect(screen.getByRole('button', { name: 'History' })).toBeInTheDocument()
  })
})
