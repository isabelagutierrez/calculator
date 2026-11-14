import { formatNumber } from './format'

export type Operator = '+' | '-' | '*' | '/' | '^'
export type Token =
  | { kind: 'number'; value: string }
  | { kind: 'operator'; value: Operator }
  | { kind: 'paren'; value: '(' | ')' }

export type HistoryItem = { id: string; expression: string; result: string }
export interface CalculatorState {
  tokens: Token[]
  current: string
  currentIsResult: boolean
  error: string | null
  hint: string | null
  history: HistoryItem[]
  pending: boolean
}

export type CalculatorAction =
  | { type: 'digit'; value: string }
  | { type: 'decimal' }
  | { type: 'operator'; value: Operator }
  | { type: 'sign' }
  | { type: 'backspace' }
  | { type: 'openParen' }
  | { type: 'closeParen' }
  | { type: 'clear' }
  | { type: 'hint'; value: string }
  | { type: 'pending'; value: boolean }
  | { type: 'result'; expression: string; value: number }
  | { type: 'unaryResult'; expression: string; value: number }
  | { type: 'clearHistory' }
  | { type: 'error'; value: string }

export const initialState: CalculatorState = {
  tokens: [], current: '', currentIsResult: false, error: null, hint: null, history: [], pending: false,
}

let idCounter = 0
const historyId = () => `history-${Date.now()}-${idCounter++}`
const last = (state: CalculatorState) => state.tokens[state.tokens.length - 1]
const isOperator = (token: Token | undefined): token is Extract<Token, { kind: 'operator' }> => token?.kind === 'operator'
const hasOpenParen = (tokens: Token[]) => tokens.reduce((count, token) => count + (token.value === '(' ? 1 : token.value === ')' ? -1 : 0), 0) > 0

function resetInput(state: CalculatorState): CalculatorState {
  return { ...state, tokens: [], current: '', currentIsResult: false, error: null, hint: null }
}

function confirmCurrent(state: CalculatorState): Token[] {
  if (!state.current) return state.tokens
  const value = state.current.endsWith('.') ? state.current.slice(0, -1) : state.current
  return [...state.tokens, { kind: 'number', value: value || '0' }]
}

export function serializeTokens(tokens: Token[], mode: 'display' | 'api'): string {
  return tokens.map((token) => {
    if (token.kind === 'operator') return mode === 'display' ? ({ '*': '×', '/': '÷', '-': '−', '^': 'xʸ', '+': '+' }[token.value]) : token.value
    if (token.kind === 'paren') return token.value
    return token.value.startsWith('-') ? `(${token.value})` : token.value
  }).join(mode === 'api' ? ' ' : ' ')
}

export function expressionTokens(state: CalculatorState, closeParentheses = false): Token[] {
  let tokens = confirmCurrent(state)
  if (closeParentheses) {
    const open = tokens.reduce((count, token) => count + (token.value === '(' ? 1 : token.value === ')' ? -1 : 0), 0)
    tokens = [...tokens, ...Array.from({ length: Math.max(0, open) }, (): Token => ({ kind: 'paren', value: ')' }))]
  }
  return tokens
}

function addHistory(state: CalculatorState, expression: string, value: number): CalculatorState {
  return {
    ...state,
    history: [{ id: historyId(), expression: `${expression} =`, result: formatNumber(value) }, ...state.history].slice(0, 20),
  }
}

export function reducer(state: CalculatorState, action: CalculatorAction): CalculatorState {
  if (state.pending && !['pending', 'result', 'unaryResult', 'error', 'clearHistory'].includes(action.type)) return state
  if (action.type === 'clearHistory') return { ...state, history: [] }
  if (action.type === 'pending') return { ...state, pending: action.value }
  if (action.type === 'clear') return { ...resetInput(state), history: state.history }
  if (action.type === 'error') return { ...state, error: action.value, hint: null, pending: false }
  if (action.type === 'hint') return { ...state, hint: action.value, error: null }
  if (state.error && !['digit', 'decimal', 'openParen', 'sign', 'backspace'].includes(action.type)) return state
  if (action.type === 'digit') {
    const base = state.error ? resetInput(state) : state
    if (last(base)?.kind === 'paren' && last(base)?.value === ')') return base
    const current = base.currentIsResult ? '' : base.current
    const value = current === '0' || current === '-0' ? (current.startsWith('-') ? `-${action.value}` : action.value) : `${current}${action.value}`
    if (value.replace(/[-.]/g, '').length > 15) return { ...base, hint: null }
    return { ...base, current: value, currentIsResult: false, hint: null }
  }
  if (action.type === 'decimal') {
    const base = state.error || state.currentIsResult ? resetInput(state) : state
    if (last(base)?.value === ')') return base
    if (base.current.includes('.')) return base
    return { ...base, current: base.current ? `${base.current}.` : '0.', currentIsResult: false, hint: null }
  }
  if (action.type === 'sign') {
    const base = state.error ? resetInput(state) : state
    if (last(base)?.value === ')') return base
    return { ...base, current: base.current ? (base.current.startsWith('-') ? base.current.slice(1) : `-${base.current}`) : '-0', currentIsResult: false, hint: null }
  }
  if (action.type === 'operator') {
    if (state.error) return state
    let tokens = state.tokens
    if (state.current) tokens = confirmCurrent(state)
    const previous = tokens[tokens.length - 1]
    if (isOperator(previous)) tokens = [...tokens.slice(0, -1), { kind: 'operator', value: action.value }]
    else if (previous?.kind === 'paren' && previous.value === '(') return state
    else if (previous?.kind === 'paren' && previous.value === ')') tokens = [...tokens, { kind: 'operator', value: action.value }]
    else if (!tokens.length) tokens = [{ kind: 'number', value: '0' }, { kind: 'operator', value: action.value }]
    else tokens = [...tokens, { kind: 'operator', value: action.value }]
    return { ...state, tokens, current: '', currentIsResult: false, hint: null }
  }
  if (action.type === 'openParen') {
    const base = state.error || (state.currentIsResult && state.tokens.length === 0) ? resetInput(state) : state
    const previous = last(base)
    if (base.current || (previous && previous.kind === 'paren' && previous.value === ')')) return base
    if (!base.tokens.length || isOperator(previous) || previous?.value === '(') {
      return { ...base, tokens: [...base.tokens, { kind: 'paren', value: '(' }], hint: null }
    }
    return base
  }
  if (action.type === 'closeParen') {
    if (!hasOpenParen(state.tokens) || (!state.current && last(state)?.value !== ')')) return state
    const tokens = confirmCurrent(state)
    return { ...state, tokens: [...tokens, { kind: 'paren', value: ')' }], current: '', currentIsResult: false, hint: null }
  }
  if (action.type === 'backspace') {
    if (state.error) return resetInput(state)
    if (state.currentIsResult) return state
    if (state.current) {
      const current = state.current.slice(0, -1)
      return { ...state, current: current === '-' ? '' : current, hint: null }
    }
    const tokens = state.tokens.slice(0, -1)
    const previous = tokens[tokens.length - 1]
    if (previous?.kind === 'number') return { ...state, tokens: tokens.slice(0, -1), current: previous.value, hint: null }
    return { ...state, tokens, hint: null }
  }
  if (action.type === 'result') {
    const next = addHistory({ ...state, tokens: [], current: String(action.value), currentIsResult: true, error: null, hint: null, pending: false }, action.expression, action.value)
    return next
  }
  if (action.type === 'unaryResult') {
    return addHistory({ ...state, current: String(action.value), currentIsResult: true, error: null, hint: null, pending: false }, action.expression, action.value)
  }
  return state
}
