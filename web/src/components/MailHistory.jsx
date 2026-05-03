import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { History } from 'lucide-react'

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
    <div className="flex items-center gap-2 mt-3 flex-wrap">
      <History size={13} className="text-base-content/30 shrink-0" />
      <span className="text-xs text-base-content/30">{t('mailHistory.recent')}</span>
      {history.map(id => (
        <button
          key={id}
          className={`rounded-full px-3 py-1 text-xs font-medium transition-all duration-150 ${
            id === activeShortId
              ? 'bg-primary text-primary-content shadow-sm'
              : 'bg-base-200 text-base-content/60 hover:bg-base-300'
          }`}
          onClick={() => onSelect(id)}
        >
          {id}
        </button>
      ))}
    </div>
  )
}
