import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'

describe('Calculator', () => {
  afterEach(() => cleanup())
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => ({ result: 16 }),
    }))
  })

  it('sends the API expression in ASCII format', async () => {
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: '8' }))
    fireEvent.click(screen.getByRole('button', { name: 'Add' }))
    fireEvent.click(screen.getByRole('button', { name: '8' }))
    fireEvent.click(screen.getByRole('button', { name: 'Equals' }))
    await waitFor(() => expect(document.querySelector('.display strong')).toHaveTextContent('16'))
    expect(fetch).toHaveBeenCalledWith(expect.any(URL), expect.objectContaining({ body: JSON.stringify({ expression: '8 + 8' }) }))
  })

  it('renders the supported symbolic controls', () => {
    render(<App />)
    for (const name of ['Power', 'Open parenthesis', 'Close parenthesis', 'Multiply', 'Toggle sign']) {
      expect(screen.getByRole('button', { name })).toBeInTheDocument()
    }
  })

  it('shows backend errors and can recover with a digit', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => ({ error: 'Cannot divide by zero' }),
    }))
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: '8' }))
    fireEvent.click(screen.getByRole('button', { name: 'Divide' }))
    fireEvent.click(screen.getByRole('button', { name: '0' }))
    fireEvent.click(screen.getByRole('button', { name: 'Equals' }))
    await waitFor(() => expect(screen.getByText('Error')).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: '3' }))
    expect(document.querySelector('.display strong')).toHaveTextContent('3')
  })

  it('opens and closes history', () => {
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: 'History' }))
    expect(screen.getByRole('complementary')).toHaveClass('is-open')
    fireEvent.click(screen.getAllByRole('button', { name: 'Close history' })[1])
    expect(screen.getByRole('complementary')).not.toHaveClass('is-open')
  })

  it('calls unary operations and ignores modified keyboard shortcuts', async () => {
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: '9' }))
    fireEvent.click(screen.getByRole('button', { name: 'Square root' }))
    await waitFor(() => expect(document.querySelector('.display strong')).toHaveTextContent('16'))
    const preventDefault = vi.fn()
    fireEvent.keyDown(window, { key: '+', ctrlKey: true, preventDefault })
    expect(preventDefault).not.toHaveBeenCalled()
  })

  it('shows unary service errors', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('offline')))
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: '9' }))
    fireEvent.click(screen.getByRole('button', { name: 'Percent' }))
    await waitFor(() => expect(screen.getByText('Cannot reach the calculation service. Please check that the backend is running.')).toBeInTheDocument())
  })

  it('handles physical keyboard input', () => {
    render(<App />)
    fireEvent.keyDown(window, { key: '5' })
    fireEvent.keyDown(window, { key: 'Delete' })
    fireEvent.keyDown(window, { key: 'c' })
    expect(document.querySelector('.display strong')).toHaveTextContent('0')
  })
})
