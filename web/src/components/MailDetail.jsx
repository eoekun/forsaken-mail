import { useEffect } from 'react'
import DOMPurify from 'dompurify'
import { useTranslation } from 'react-i18next'
import { Mail, FileText } from 'lucide-react'
import { apiPut } from '../lib/api'

export default function MailDetail({ mail, onMailRead }) {
  const { t } = useTranslation()

  useEffect(() => {
    if (mail?.id && !mail.is_read) {
      apiPut(`/api/mails/${mail.id}/read`).catch(() => {})
      onMailRead?.(mail.id)
    }
  }, [mail?.id])

  if (!mail) {
    return (
      <div className="card-modern h-full">
        <div className="flex flex-col items-center justify-center h-full py-16">
          <FileText size={36} className="text-base-content/10 mb-3" />
          <p className="text-sm text-base-content/30">{t('mailDetail.empty')}</p>
        </div>
      </div>
    )
  }

  const htmlContent = mail.html || mail.text_body || ''

  return (
    <div className="card-modern">
      <div className="p-5">
        <h2 className="text-lg font-semibold text-base-content mb-3 leading-snug">
          {mail.subject || t('mailDetail.noSubject')}
        </h2>
        <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-base-content/50 mb-4">
          <span className="flex items-center gap-1">
            <Mail size={12} />
            {mail.from}
          </span>
          <span>{t('mailDetail.to')} {mail.to}</span>
          <span className="tabular-nums">{new Date(mail.created_at).toLocaleString()}</span>
        </div>
        <div className="border-t border-base-300/40 pt-4">
          {mail.html ? (
            <div
              className="prose prose-sm max-w-none prose-headings:text-base-content prose-p:text-base-content/80 prose-a:text-primary"
              dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(htmlContent) }}
            />
          ) : (
            <pre className="whitespace-pre-wrap text-sm text-base-content/80 font-sans leading-relaxed">
              {mail.text_body}
            </pre>
          )}
        </div>
      </div>
    </div>
  )
}
