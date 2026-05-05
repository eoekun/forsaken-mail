import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Clock, ChevronDown, ChevronRight, KeyRound } from 'lucide-react'

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
            const hasCode = (mail.extracted_codes?.length || 0) > 0
            return (
              <div
                key={mail.id}
                className="px-3 py-2 sm:px-4 sm:py-2.5 flex items-center gap-3 cursor-pointer hover:bg-base-200 transition-colors"
                onClick={() => onOpenMail(mail)}
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2 mb-0.5">
                    <span className="text-xs font-mono text-primary/70 bg-primary/5 px-1.5 py-0.5 rounded shrink-0">
                      {mail.short_id}
                    </span>
                    <span className="text-xs text-base-content/50 truncate">{mail.from_addr || mail.from}</span>
                    {hasCode && <KeyRound size={11} className="text-primary shrink-0" />}
                  </div>
                  <p className="text-xs text-base-content/60 truncate">{mail.subject || t('mailList.noSubject')}</p>
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
