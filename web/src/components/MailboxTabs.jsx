import { useTranslation } from 'react-i18next'
import { Plus, X } from 'lucide-react'

export default function MailboxTabs({ tabs, activeShortId, onSelect, onClose, onAdd, host }) {
  const { t } = useTranslation()

  if (tabs.length === 0) return null

  return (
    <div className="sticky top-14 z-10 bg-base-200/90 backdrop-blur-sm -mx-4 sm:-mx-6 px-4 sm:px-6 py-2">
      <div className="flex items-center gap-1 overflow-x-auto scrollbar-hide pb-1">
        {tabs.map(({ shortId, unreadCount }) => (
          <div
            key={shortId}
            className={`group flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium cursor-pointer transition-all duration-150 shrink-0 border-b-2 ${
              shortId === activeShortId
                ? 'border-primary bg-primary/5 text-primary'
                : 'border-transparent text-base-content/60 hover:bg-base-300/60'
            }`}
            onClick={() => onSelect(shortId)}
          >
            {unreadCount > 0 && (
              <span className="w-2 h-2 rounded-full bg-red-500 shrink-0" />
            )}
            <span>{shortId}@{host}</span>
            {tabs.length > 1 && (
              <button
                className={`ml-0.5 p-0.5 rounded-full opacity-0 group-hover:opacity-100 transition-opacity hover:bg-black/10 ${
                  shortId === activeShortId ? 'text-primary' : 'text-base-content/40'
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
          className="flex items-center gap-1 px-2.5 py-1.5 text-xs text-base-content/40 hover:text-base-content/60 hover:bg-base-300/60 transition-colors shrink-0"
          onClick={onAdd}
          title={t('mailboxTabs.add')}
        >
          <Plus size={14} />
        </button>
        {tabs.length > 1 && (
          <button
            className="flex items-center gap-1 ml-auto px-2.5 py-1.5 text-xs text-base-content/40 hover:text-error transition-colors shrink-0"
            onClick={() => tabs.forEach(({ shortId }) => onClose(shortId))}
            title={t('mailboxTabs.closeAll')}
          >
            <X size={14} />
            <span className="hidden sm:inline">{t('mailboxTabs.closeAll')}</span>
          </button>
        )}
      </div>
    </div>
  )
}
