import { useCallback, useEffect, useReducer, useRef, useState } from 'react'
import './index.css'
import { evaluateExpression, percentage, squareRoot } from './api'
import { reducer, initialState, expressionTokens, serializeTokens, type Operator } from './calculator/state'
import { Display } from './components/Display'
import { HistoryPanel } from './components/HistoryPanel'
import { Keypad } from './components/Keypad'

function App() {
  const [state, dispatch] = useReducer(reducer, initialState)
  const [showHistory, setShowHistory] = useState(false)

  const handleKey = useCallback(async (key: string) => {
    if (state.pending) return
    if (/^\d$/.test(key)) return dispatch({ type: 'digit', value: key })
    if (key === '.') return dispatch({ type: 'decimal' })
    if (key === 'C') return dispatch({ type: 'clear' })
    if (key === '⌫') return dispatch({ type: 'backspace' })
    if (key === '±') return dispatch({ type: 'sign' })
    if (key === '(') return dispatch({ type: 'openParen' })
    if (key === ')') return dispatch({ type: 'closeParen' })
    if (['+', '−', '×', '÷', 'xʸ'].includes(key)) {
      const operator: Operator = ({ '+': '+', '−': '-', '×': '*', '÷': '/', xʸ: '^' } as Record<string, Operator>)[key]
      return dispatch({ type: 'operator', value: operator })
    }
    if (key === '=') {
      const tokens = expressionTokens(state, true)
      const last = tokens[tokens.length - 1]
      if (!tokens.length) return
      if (last?.kind === 'operator' || last?.value === '(') return dispatch({ type: 'hint', value: 'Complete the expression first.' })
      const expression = serializeTokens(tokens, 'api')
      dispatch({ type: 'pending', value: true })
      try {
        const result = await evaluateExpression(expression)
        dispatch({ type: 'result', expression: serializeTokens(tokens, 'display'), value: result })
      } catch (error) {
        dispatch({ type: 'error', value: error instanceof Error ? error.message : 'The calculation service is unavailable. Please try again.' })
      } finally {
        dispatch({ type: 'pending', value: false })
      }
      return
    }
    if (key === '√' || key === '%') {
      if (!state.current) return
      const value = Number(state.current)
      if (!Number.isFinite(value)) return
      dispatch({ type: 'pending', value: true })
      try {
        const result = key === '√' ? await squareRoot(value) : await percentage(value)
        dispatch({ type: 'unaryResult', expression: `${key}(${state.current})`, value: result })
      } catch (error) {
        dispatch({ type: 'error', value: error instanceof Error ? error.message : 'The calculation service is unavailable. Please try again.' })
      } finally {
        dispatch({ type: 'pending', value: false })
      }
    }
  }, [state])

  const keyRef = useRef(handleKey)
  useEffect(() => { keyRef.current = handleKey }, [handleKey])
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.ctrlKey || event.metaKey || event.altKey) return
      const mapped: Record<string, string> = { '*': '×', '/': '÷', '^': 'xʸ', Enter: '=', Escape: 'C', Delete: 'C', Backspace: '⌫', c: 'C', C: 'C', '-': '−' }
      const key = mapped[event.key] ?? event.key
      if (/^\d$/.test(key) || ['.', '+', '−', '×', '÷', 'xʸ', '(', ')', '%', '√', '±', '=', 'C', '⌫'].includes(key)) {
        event.preventDefault()
        void keyRef.current(key)
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [])

  return <main className="app-shell"><section className="calculator"><div className="calculator-main">
    <button className="history-toggle" onClick={() => setShowHistory((open) => !open)} aria-expanded={showHistory} aria-controls="history-panel">
      {showHistory ? 'Close history' : 'History'}
    </button>
    <Display state={state} />
    <Keypad onKey={(key) => void handleKey(key)} disabled={state.pending} />
  </div>
    <HistoryPanel history={state.history} open={showHistory} onClear={() => dispatch({ type: 'clearHistory' })} onClose={() => setShowHistory(false)} />
  </section></main>
}

export default App
