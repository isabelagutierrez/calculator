import { useEffect, useRef, useState } from 'react'
import './index.css'
import { calculateExpression as calculateWithApi, calculateUnary } from './api'

type Operator = '+' | '−' | '×' | '÷' | 'xʸ'

const keys = [
  ['C', '⌫', '%', '÷'],
  ['√', 'xʸ', '(', ')'],
  ['7', '8', '9', '×'],
  ['4', '5', '6', '−'],
  ['1', '2', '3', '+'],
  ['±', '0', '.', '='],
]

type HistoryItem = { expression: string; result: string }

const formatNumber = (value: number) => {
  const formatted = value.toPrecision(8)
  return formatted.includes('e')
    ? formatted.replace(/\.?0+e/, 'e')
    : formatted.replace(/\.?0+$/, '')
}

const formatInput = (value: string) => {
  if (value === '' || value === '0.') return value
  const digitCount = value.replace(/[-.]/g, '').length
  if (digitCount <= 8) return value
  const numericValue = Number(value)
  return Number.isFinite(numericValue) ? formatNumber(numericValue) : 'Error'
}

function App() {
  const [display, setDisplay] = useState('')
  const [rawInput, setRawInput] = useState('')
  const [expression, setExpression] = useState('')
  const [waitingForOperand, setWaitingForOperand] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [history, setHistory] = useState<HistoryItem[]>([])
  const [showHistory, setShowHistory] = useState(false)
  const [isCalculating, setIsCalculating] = useState(false)

  const showError = (message: string) => {
    setError(message)
    setDisplay('Error')
    setWaitingForOperand(true)
  }

  const clear = () => {
    setDisplay('')
    setRawInput('')
    setExpression('')
    setWaitingForOperand(false)
    setError(null)
  }

  const clearHistory = () => setHistory([])

  const addHistoryItem = (expression: string, result: number) => {
    setHistory((current) => [
      { expression: `${expression} =`, result: formatNumber(result) },
      ...current,
    ].slice(0, 20))
  }

  const getValue = () => {
    if (display.trim() === '') throw new Error('Enter a number.')
    const value = Number(display)
    if (!Number.isFinite(value)) throw new Error('Invalid number.')
    return value
  }

  const enterNumber = (value: string) => {
    if (error) clear()
    if (waitingForOperand) {
      setDisplay(value)
      setRawInput(value)
      setExpression((current) => `${current}${value}`)
      setWaitingForOperand(false)
      return
    }
    const nextRawInput = rawInput === '0' ? value : `${rawInput}${value}`
    const nextValue = formatInput(nextRawInput)
    if (nextValue === 'Error') {
      showError('Number is too large.')
      return
    }
    setRawInput(nextRawInput)
    setDisplay(nextValue)
    setExpression((current) => current ? `${current}${value}` : nextRawInput)
  }

  const enterDecimal = () => {
    if (error) clear()
    if (waitingForOperand) {
      setDisplay('0.')
      setRawInput('0.')
      setExpression((current) => `${current}0.`)
      setWaitingForOperand(false)
      return
    }
    if (!rawInput.includes('.')) {
      const nextRawInput = rawInput ? `${rawInput}.` : '0.'
      setRawInput(nextRawInput)
      setDisplay(nextRawInput)
      setExpression((current) => current ? `${current}.` : nextRawInput)
    }
  }

  const chooseOperator = (selectedOperator: Operator) => {
    try {
      if (selectedOperator === '−' && !expression && !display && !rawInput) {
        setDisplay('-')
        setRawInput('-')
        setWaitingForOperand(false)
        setError(null)
        return
      }
      if (waitingForOperand) {
        if (!expression.trim()) {
          setExpression(`${display} ${selectedOperator} `)
        } else if (expression.trim().endsWith(')')) {
          setExpression((current) => `${current} ${selectedOperator} `)
        } else if (selectedOperator === '−' && /[+−×÷xʸ]\s$/.test(expression)) {
          setExpression((current) => `${current}− `)
        } else {
          setExpression((current) => current.replace(/[+−×÷xʸ]\s*$/, `${selectedOperator} `))
        }
      } else {
        setExpression((current) => `${current || display} ${selectedOperator} `)
      }
      setDisplay('')
      setRawInput('')
      setWaitingForOperand(true)
      setError(null)
    } catch (calculationError) {
      showError(calculationError instanceof Error ? calculationError.message : 'Operation failed.')
    }
  }

  const calculateResult = async () => {
    if (!expression.trim()) return
    setIsCalculating(true)
    try {
      const result = await calculateWithApi(expression)
      addHistoryItem(expression, result)
      setDisplay(formatNumber(result))
      setRawInput(String(result))
      setExpression('')
      setWaitingForOperand(true)
    } catch (calculationError) {
      showError(calculationError instanceof Error ? calculationError.message : 'Operation failed.')
    } finally {
      setIsCalculating(false)
    }
  }

  const applyUnaryOperation = async (operation: 'percent' | 'squareRoot' | 'sign') => {
    try {
      const value = getValue()
      setIsCalculating(true)
      const result = operation === 'sign'
        ? -value
        : await calculateUnary(operation === 'percent' ? 'percentage' : 'square-root', value)
      const symbol = operation === 'percent' ? '%' : operation === 'squareRoot' ? '√' : '±'
      addHistoryItem(`${symbol}(${value})`, result)
      setDisplay(formatNumber(result))
      setRawInput(String(result))
      setWaitingForOperand(false)
      setError(null)
    } catch (calculationError) {
      showError(calculationError instanceof Error ? calculationError.message : 'Operation failed.')
    } finally {
      setIsCalculating(false)
    }
  }

  const enterParenthesis = (parenthesis: '(' | ')') => {
    if (parenthesis === '(') {
      if (expression === '' || /[+−×÷xʸ]\s*$/.test(expression) || expression.endsWith('(')) {
        setExpression((current) => `${current}(`)
        setDisplay('')
        setRawInput('')
        setWaitingForOperand(true)
      }
      return
    }

    if (/\d$/.test(expression) || expression.endsWith(')')) {
      setExpression((current) => `${current})`)
      setDisplay('')
      setRawInput('')
      setWaitingForOperand(true)
    }
  }

  const removeLastCharacter = () => {
    if (error || waitingForOperand) return
    const nextRawInput = rawInput.length > 1 ? rawInput.slice(0, -1) : '0'
    setRawInput(nextRawInput)
    setDisplay(formatInput(nextRawInput))
  }

  const handleKey = (key: string) => {
    if (isCalculating) return
    if (/^\d$/.test(key)) return enterNumber(key)
    if (key === '.') return enterDecimal()
    if (key === 'C') return clear()
    if (key === '⌫') return removeLastCharacter()
    if (key === '%') return applyUnaryOperation('percent')
    if (key === '√') return applyUnaryOperation('squareRoot')
    if (key === '±') return applyUnaryOperation('sign')
    if (key === '(' || key === ')') return enterParenthesis(key)
    if (key === '=') return void calculateResult()
    if (['+', '−', '×', '÷', 'xʸ'].includes(key)) return chooseOperator(key as Operator)
  }

  const handleKeyRef = useRef(handleKey)
  useEffect(() => {
    handleKeyRef.current = handleKey
  })

  useEffect(() => {
    const handleKeyboard = (event: KeyboardEvent) => {
      const keyMap: Record<string, string> = {
        '*': '×',
        '/': '÷',
        '^': 'xʸ',
        Enter: '=',
        Escape: 'C',
        Backspace: '⌫',
      }
      const key = keyMap[event.key] ?? event.key
      if (/^\d$/.test(key) || ['.', '+', '-', '−', '*', '/', '×', '÷', '^', 'xʸ', '(', ')', '%', 'Enter', '=', 'Escape', 'Backspace'].includes(event.key)) {
        event.preventDefault()
        handleKeyRef.current(key === '-' ? '−' : key)
      }
    }

    window.addEventListener('keydown', handleKeyboard)
    return () => window.removeEventListener('keydown', handleKeyboard)
  }, [])

  return (
    <main className="app-shell">
      <section className="calculator">
        <div className="calculator-main">
          <button
            className="history-toggle"
            onClick={() => setShowHistory((current) => !current)}
            aria-expanded={showHistory}
          >
            {showHistory ? 'Close history' : 'History'}
          </button>
          <div className="display">
            <p className="display-expression">{error ?? expression}</p>
            <strong>{display}</strong>
          </div>

          <div className="keypad">
            {keys.flat().map((key, index) => (
              <button
                className={[
                  key === 'C' && 'clear',
                  key === '⌫' && 'delete',
                  ['+', '−', '×', '÷'].includes(key) && 'operator',
                  key === 'xʸ' && 'power',
                  key === '=' && 'equals',
                  /^\d$/.test(key) && 'digit',
                  key === '.' && 'number',
                  key === '0' && 'zero',
                  !key && 'empty',
                ].filter(Boolean).join(' ')}
                key={`${key}-${index}`}
                onClick={() => handleKey(key)}
                disabled={!key || isCalculating}
              >
                {key}
              </button>
            ))}
          </div>
        </div>

        {showHistory && <button className="history-overlay" aria-label="Close history" onClick={() => setShowHistory(false)} />}
        <aside className={`history-panel${showHistory ? ' is-open' : ''}`}>
          <header className="history-header">
            <h2>History</h2>
            <div className="history-actions">
              <button onClick={clearHistory} disabled={history.length === 0}>Clear</button>
              <button className="history-close" onClick={() => setShowHistory(false)} aria-label="Close history">×</button>
            </div>
          </header>
          <div className="history-list">
            {history.length === 0 && (
              <div className="empty-history">
                <span aria-hidden="true">◷</span>
                <p>Your operations<br />will appear here.</p>
              </div>
            )}
            {history.map((item, index) => (
              <article className="history-item" key={`${item.expression}-${index}`}>
                <p>{item.expression}</p>
                <strong>{item.result}</strong>
              </article>
            ))}
          </div>
        </aside>
      </section>
    </main>
  )
}

export default App
