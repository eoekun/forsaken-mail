import { useState, useRef, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Clock, ChevronDown, ChevronRight, KeyRound, Copy, Check } from 'lucide-react'

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

export default function RecentMails({ recentMails, onOpenMail }) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(false)
  const [copiedId, setCopiedId] = useState(null)
  const copiedTimerRef = useRef(null)

  useEffect(() => {
    return () => clearTimeout(copiedTimerRef.current)
  }, [])

  const handleCopyCode = (e, mail) => {
    e.stopPropagation()
    const code = mail.extracted_codes?.[0]
    if (!code) return
    navigator.clipboard.writeText(code).then(() => {
      setCopiedId(mail.id)
      clearTimeout(copiedTimerRef.current)
      copiedTimerRef.current = setTimeout(() => setCopiedId(null), 2000)
    })
  }

  if (!recentMails || recentMails.length === 0) return null

  return (
    <div className="card-modern mt-3">
      <button
        className="flex items-center gap-2 w-full px-3 py-2.5 sm:px-4 text-left hover:bg-base-200 transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <Clock size={14} className="text-base-content/40" />
        <span className="text-sm font-medium text-base-content/70">{t('recentMails.title')}</span>
        <span className="ml-1 text-xs text-base-content/30">({recentMails.length})</span>
        <span className="ml-auto text-base-content/30">
          {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        </span>
      </button>
      {expanded && (
        <div className="divide-y divide-base-300/40 border-t border-base-300/40">
          {recentMails.map((mail) => {
            const codes = mail.extracted_codes || []
            const hasCode = codes.length > 0
            const recipient = mail.short_id || (mail.to_addr || mail.to || '').split('@')[0]
            const sender = mail.from_addr || mail.from || ''
            return (
              <div
                key={mail.id}
                className="px-3 py-2 sm:px-4 sm:py-2.5 flex items-center gap-3 cursor-pointer hover:bg-base-200 transition-colors"
                onClick={() => onOpenMail(mail)}
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2 mb-0.5">
                    {recipient && (
                      <span className="text-xs font-mono text-primary/70 bg-primary/5 px-1.5 py-0.5 rounded shrink-0">
                        {recipient}
                      </span>
                    )}
                    <span className="text-xs text-base-content/50 truncate">{sender}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <p className="text-xs text-base-content/60 truncate flex-1">{mail.subject || t('mailList.noSubject')}</p>
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
                        <span className="tracking-wider">{codes[0]}</span>
                      </button>
                    )}
                  </div>
                </div>
                <span className="text-[11px] text-base-content/30 shrink-0 tabular-nums">
                  {formatMailTime(mail.created_at, t)}
                </span>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
