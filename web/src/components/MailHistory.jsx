import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'

export default function MailHistory({ host, activeShortId, onSelect }) {
  const [history, setHistory] = useState([])
  const { t } = useTranslation()

  useEffect(() => {
    try {
      const raw = localStorage.getItem('shortid_history_v1')
      const list = raw ? JSON.parse(raw) : []
      if (Array.isArray(list)) setHistory(list)
    } catch {}
  }, [activeShortId])

  if (history.length === 0) return null

  return (
    <div className="flex flex-wrap gap-2 mt-3">
      <span className="text-xs text-base-content/50 self-center">{t('mailHistory.recent')}</span>
      {history.map(id => (
        <div key={id} className="join">
          <button
            className={`btn btn-xs join-item ${id === activeShortId ? 'btn-primary' : 'btn-ghost'}`}
            onClick={() => onSelect(id)}
          >
            {id}@{host}
          </button>
        </div>
      ))}
    </div>
  )
}
