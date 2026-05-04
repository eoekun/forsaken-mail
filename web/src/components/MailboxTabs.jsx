import { useTranslation } from 'react-i18next'
import { Plus, X } from 'lucide-react'

export default function MailboxTabs({ tabs, activeShortId, onSelect, onClose, onAdd, host }) {
  const { t } = useTranslation()

  if (tabs.length === 0) return null

  return (
    <div className="flex items-center gap-1 overflow-x-auto pb-1">
      {tabs.map(({ shortId, unreadCount }) => (
        <div
          key={shortId}
          className={`group flex items-center gap-1 rounded-full px-3 py-1.5 text-xs font-medium cursor-pointer transition-all duration-150 shrink-0 ${
            shortId === activeShortId
              ? 'bg-primary text-primary-content shadow-sm'
              : 'bg-base-200 text-base-content/60 hover:bg-base-300'
          }`}
          onClick={() => onSelect(shortId)}
        >
          <span>{shortId}@{host}</span>
          {unreadCount > 0 && (
            <span className={`inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 text-[10px] font-bold rounded-full ${
              shortId === activeShortId
                ? 'bg-primary-content text-primary'
                : 'bg-primary text-primary-content'
            }`}>
              {unreadCount > 99 ? '99+' : unreadCount}
            </span>
          )}
          {tabs.length > 1 && (
            <button
              className={`ml-0.5 p-0.5 rounded-full opacity-0 group-hover:opacity-100 transition-opacity hover:bg-black/10 ${
                shortId === activeShortId ? 'text-primary-content' : 'text-base-content/40'
              }`}
              onClick={(e) => {
                e.stopPropagation()
                onClose(shortId)
              }}
              aria-label="Close"
            >
              <X size={12} />
            </button>
          )}
        </div>
      ))}
      <button
        className="flex items-center gap-1 rounded-full px-2.5 py-1.5 text-xs text-base-content/40 hover:text-base-content/60 hover:bg-base-200 transition-colors shrink-0"
        onClick={onAdd}
        title={t('mailboxTabs.add')}
      >
        <Plus size={14} />
      </button>
    </div>
  )
}
