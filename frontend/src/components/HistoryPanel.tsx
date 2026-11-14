import type { HistoryItem } from '../calculator/state'

export function HistoryPanel({ history, open, onClear, onClose }: { history: HistoryItem[]; open: boolean; onClear: () => void; onClose: () => void }) {
  return <>
    {open && <button className="history-overlay" aria-label="Close history" onClick={onClose} />}
    <aside className={`history-panel${open ? ' is-open' : ''}`} id="history-panel">
      <header className="history-header"><h2>History</h2><div className="history-actions">
        <button onClick={onClear} disabled={!history.length}>Clear</button>
        <button className="history-close" onClick={onClose} aria-label="Close history">×</button>
      </div></header>
      <div className="history-list">
        {!history.length && <div className="empty-history"><span aria-hidden="true">◷</span><p>Your operations<br />will appear here.</p></div>}
        {history.map((item) => <article className="history-item" key={item.id}><p>{item.expression}</p><strong>{item.result}</strong></article>)}
      </div>
    </aside>
  </>
}
