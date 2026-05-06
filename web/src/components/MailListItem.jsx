import { useTranslation } from 'react-i18next'
import { KeyRound, Check } from 'lucide-react'
import { formatSender } from '../lib/formatSender'
import { formatRelativeTime, formatAbsoluteTime } from '../lib/formatTime'
import { useCopyToClipboard } from '../hooks/useCopyToClipboard'

export default function MailListItem({ mail, isSelected, onClick, badge, highlight, unread }) {
  const { t } = useTranslation()
  const { copiedId, copy } = useCopyToClipboard()

  const hasCode = (mail.extracted_codes?.length || 0) > 0
  const sender = mail.from || mail.from_addr || ''

  const handleCopyCode = (e) => {
    e.stopPropagation()
    const code = mail.extracted_codes?.[0]
    if (!code) return
    copy(code, mail.id)
  }

  const highlightMatch = (text) => {
    if (!highlight?.trim() || !text) return text
    const idx = text.toLowerCase().indexOf(highlight.toLowerCase())
    if (idx === -1) return text
    return (
      <>
        {text.slice(0, idx)}
        <mark className="bg-warning/30 text-inherit rounded-sm px-0.5">{text.slice(idx, idx + highlight.length)}</mark>
        {text.slice(idx + highlight.length)}
      </>
    )
  }

  const formattedSender = formatSender(sender)

  return (
    <div
      className={`px-3 py-2.5 sm:px-4 sm:py-3 cursor-pointer transition-all duration-150 ${
        isSelected
          ? 'bg-primary/5 border-l-3 border-l-primary'
          : `border-l-3 border-l-transparent hover:bg-base-200 ${unread ? '' : 'opacity-70'}`
      }`}
      onClick={onClick}
    >
      <div className="flex items-baseline justify-between gap-2 mb-0.5">
        <span className={`text-sm truncate flex items-center gap-1.5 ${isSelected ? 'font-semibold text-base-content' : unread ? 'font-semibold text-base-content' : 'text-base-content/50'}`}>
          {unread && <span className="inline-block w-1.5 h-1.5 rounded-full bg-primary shrink-0" />}
          <span className="truncate" title={sender}>
            {highlight ? highlightMatch(formattedSender) : formattedSender}
          </span>
        </span>
        <span className="text-[11px] text-base-content/30 shrink-0" title={formatAbsoluteTime(mail.created_at)}>
          {formatRelativeTime(mail.created_at)}
        </span>
      </div>
      <div className="flex items-center gap-2">
        {badge && (
          <span className="text-[11px] font-mono text-primary/60 bg-primary/5 px-1 py-0.5 rounded shrink-0">
            {badge}
          </span>
        )}
        <p className={`text-xs truncate flex-1 ${unread ? 'text-base-content/70 font-medium' : 'text-base-content/50'}`}>
          {highlight ? highlightMatch(mail.subject || t('mailList.noSubject')) : (mail.subject || t('mailList.noSubject'))}
        </p>
        {hasCode && (
          <button
            onClick={handleCopyCode}
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
}
