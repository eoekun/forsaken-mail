import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Inbox, KeyRound, Search, Check, X } from 'lucide-react'
import { formatMailTime } from '../lib/formatTime'
import { formatSender } from '../lib/formatSender'
import { useCopyToClipboard } from '../hooks/useCopyToClipboard'

export default function MailList({ mails, selectedMail, onSelect }) {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const { copiedId, copy } = useCopyToClipboard()

  const handleCopyCode = (e, mail) => {
    e.stopPropagation()
    const code = mail.extracted_codes?.[0]
    if (!code) return
    copy(code, mail.id)
  }

  const filteredMails = useMemo(() => {
    if (!search.trim()) return mails
    const q = search.toLowerCase()
    return mails.filter(m =>
      (m.from || '').toLowerCase().includes(q) ||
      (m.subject || '').toLowerCase().includes(q)
    )
  }, [mails, search])

  const highlightMatch = (text, query) => {
    if (!query.trim() || !text) return text
    const idx = text.toLowerCase().indexOf(query.toLowerCase())
    if (idx === -1) return text
    return (
      <>
        {text.slice(0, idx)}
        <mark className="bg-warning/30 text-inherit rounded-sm px-0.5">{text.slice(idx, idx + query.length)}</mark>
        {text.slice(idx + query.length)}
      </>
    )
  }

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
                      <span className="truncate" title={mail.from}>{search ? highlightMatch(formatSender(mail.from), search) : formatSender(mail.from)}</span>
                    </span>
                    <span className="text-[11px] text-base-content/30 shrink-0 tabular-nums">
                      {formatMailTime(mail.created_at, t)}
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <p className={`text-xs truncate flex-1 ${isUnread ? 'text-base-content/70 font-medium' : 'text-base-content/50'}`}>
                      {search ? highlightMatch(mail.subject || t('mailList.noSubject'), search) : (mail.subject || t('mailList.noSubject'))}
                    </p>
                    {hasCode && (
                      <button
                        onClick={(e) => handleCopyCode(e, mail)}
                        className={`shrink-0 inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] font-mono font-semibold transition-all cursor-pointer ${
                          copiedId === mail.id
                            ? 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400'
                            : 'bg-primary/10 text-primary hover:bg-primary/20'
                        }`}
                        title={t('mailDetail.clickToCopy')}
                      >
                        {copiedId === mail.id ? <Check size={10} /> : <KeyRound size={10} />}
                        <span className="tracking-wider">{mail.extracted_codes[0]}</span>
                      </button>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
