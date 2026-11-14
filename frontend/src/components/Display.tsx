import { formatNumber } from '../calculator/format'
import type { CalculatorState } from '../calculator/state'
import { serializeTokens } from '../calculator/state'

export function Display({ state }: { state: CalculatorState }) {
  const value = state.error ? 'Error' : state.current
    ? state.currentIsResult ? formatNumber(Number(state.current)) : state.current
    : state.tokens.length ? serializeTokens(state.tokens, 'display') : '0'
  return (
    <div className="display">
      <p className="display-expression">{state.error ?? state.hint ?? serializeTokens(state.tokens, 'display')}</p>
      <strong aria-live="polite" style={{ fontSize: value.length > 12 ? 'clamp(1.45rem, 7vw, 2.8rem)' : undefined }}>{value}</strong>
    </div>
  )
}
