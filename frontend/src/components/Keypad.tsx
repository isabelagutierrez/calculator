const keys = [
  ['C', '⌫', '%', '÷'],
  ['√', 'xʸ', '(', ')'],
  ['7', '8', '9', '×'],
  ['4', '5', '6', '−'],
  ['1', '2', '3', '+'],
  ['±', '0', '.', '='],
]

const labels: Record<string, string> = {
  C: 'Clear', '⌫': 'Backspace', '%': 'Percent', '÷': 'Divide', '√': 'Square root',
  'xʸ': 'Power', '(': 'Open parenthesis', ')': 'Close parenthesis', '×': 'Multiply',
  '−': 'Subtract', '+': 'Add', '±': 'Toggle sign', '.': 'Decimal point', '=': 'Equals',
}

export function Keypad({ onKey, disabled }: { onKey: (key: string) => void; disabled: boolean }) {
  return <div className="keypad">{keys.flat().map((key) => (
    <button
      key={key}
      className={[key === 'C' && 'clear', key === '⌫' && 'delete', ['+', '−', '×', '÷'].includes(key) && 'operator', key === 'xʸ' && 'power', key === '=' && 'equals', /^\d$/.test(key) && 'digit', key === '.' && 'number'].filter(Boolean).join(' ')}
      aria-label={labels[key]}
      onClick={() => onKey(key)}
      disabled={disabled}
    >{key}</button>
  ))}</div>
}
