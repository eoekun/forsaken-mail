import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Inbox, KeyRound, Search } from 'lucide-react'

function formatMailTime(dateStr, t) {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diffMs = now - then
  const diffMin = Math.floor(diffMs / 60000)
  const diffHour = Math.floor(diffMs / 3600000)

  if (diffMin < 1) return t('mailList.timeJustNow')
  if (diffHour < 1) return t('mailList.timeMinutesAgo', { count: diffMin })
  if (diffHour < 24) return new Date(dateStr).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  return new Date(dateStr).toLocaleString([], { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

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
    <div className="card-modern">
      <div className="p-3 pb-2 sm:p-4">
        <div className="flex items-center gap-2 mb-2">
          <Inbox size={16} className="text-base-content/40" />
          <span className="text-sm font-medium text-base-content/70">{t('mailList.title')}</span>
          <span className="ml-auto text-xs text-base-content/30">{mails.length}</span>
        </div>
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-base-content/30" />
          <input
            type="text"
            className="input-modern input-sm w-full pl-8 text-xs"
            placeholder={t('mailList.searchPlaceholder')}
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </div>
      </div>
      {filteredMails.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-8">
          <Inbox size={48} className="text-base-content/20 mb-2 opacity-30" />
          <p className="text-sm text-base-content/30">{mails.length === 0 ? t('mailList.empty') : t('mailList.noResults')}</p>
        </div>
      ) : (
        <div className="divide-y divide-base-300/40">
          {filteredMails.map((mail, idx) => {
            const isUnread = !mail.is_read
            const hasCode = (mail.extracted_codes?.length || 0) > 0
            const isSelected = selectedMail === mail
            return (
              <div
                key={mail.id || idx}
                className={`px-3 py-2.5 sm:px-4 sm:py-3 cursor-pointer transition-all duration-150 ${
                  isSelected
                    ? 'bg-primary/5 border-l-3 border-l-primary'
                    : `border-l-3 border-l-transparent hover:bg-base-200 ${isUnread ? 'font-semibold' : 'font-normal opacity-70'}`
                }`}
                onClick={() => onSelect(mail)}
              >
                <div className="flex items-baseline justify-between gap-2 mb-0.5">
                  <span className={`text-sm truncate flex items-center gap-1.5 ${isSelected ? 'font-semibold text-base-content' : isUnread ? 'font-semibold text-base-content' : 'text-base-content/50'}`}>
                    {isUnread && <span className="inline-block w-1.5 h-1.5 rounded-full bg-primary shrink-0" />}
                    {hasCode && <KeyRound size={12} className="text-primary shrink-0" />}
                    <span className="truncate">{mail.from}</span>
                  </span>
                  <span className="text-[11px] text-base-content/30 shrink-0 tabular-nums">
                    {formatMailTime(mail.created_at, t)}
                  </span>
                </div>
                <p className={`text-xs truncate ${isUnread ? 'text-base-content/70 font-medium' : 'text-base-content/50'}`}>
                  {mail.subject || t('mailList.noSubject')}
                </p>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
