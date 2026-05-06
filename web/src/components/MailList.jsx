import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Inbox, Search, X } from 'lucide-react'
import MailListItem from './MailListItem'

export default function MailList({ mails, selectedMail, onSelect }) {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')

  const filteredMails = useMemo(() => {
    if (!search.trim()) return mails
    const q = search.toLowerCase()
    return mails.filter(m =>
      (m.from || '').toLowerCase().includes(q) ||
      (m.subject || '').toLowerCase().includes(q)
    )
  }, [mails, search])

  return (
    <div className="card-modern h-full flex flex-col">
      <div className="p-3 pb-2 sm:p-4 shrink-0">
        <div className="flex items-center gap-2 mb-2">
          <Inbox size={16} className="text-base-content/40" />
          <span className="text-sm font-medium text-base-content/70">{t('mailList.title')}</span>
          <span className="ml-auto text-xs text-base-content/30">{mails.length}</span>
        </div>
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-base-content/30" />
          <input
            type="text"
            className="input-modern input-sm w-full pl-8 pr-7 text-xs"
            placeholder={t('mailList.searchPlaceholder')}
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
          {search && (
            <button
              className="absolute right-2 top-1/2 -translate-y-1/2 text-base-content/30 hover:text-base-content/60 transition-colors"
              onClick={() => setSearch('')}
            >
              <X size={13} />
            </button>
          )}
        </div>
      </div>
      <div className="flex-1 overflow-y-auto min-h-0">
        {filteredMails.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-8">
            <Inbox size={48} className="text-base-content/20 mb-2 opacity-30" />
            <p className="text-sm text-base-content/30">{mails.length === 0 ? t('mailList.empty') : t('mailList.noResults')}</p>
          </div>
        ) : (
          <div className="divide-y divide-base-300/40">
            {filteredMails.map((mail, idx) => (
              <MailListItem
                key={mail.id || idx}
                mail={mail}
                isSelected={selectedMail === mail}
                onClick={() => onSelect(mail)}
                unread={!mail.is_read}
                highlight={search}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
