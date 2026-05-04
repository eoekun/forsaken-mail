import { useEffect, useState } from 'react'
import DOMPurify from 'dompurify'
import { useTranslation } from 'react-i18next'
import { Mail, FileText, Copy, Check, ExternalLink } from 'lucide-react'
import { apiPut } from '../lib/api'

export default function MailDetail({ mail, onMailRead }) {
  const { t } = useTranslation()
  const [copiedCode, setCopiedCode] = useState(null)

  useEffect(() => {
    if (mail?.id && !mail.is_read) {
      apiPut(`/api/mails/${mail.id}/read`).catch(() => {})
      onMailRead?.(mail.id)
    }
  }, [mail?.id])

  const copyCode = (code) => {
    navigator.clipboard.writeText(code).then(() => {
      setCopiedCode(code)
      setTimeout(() => setCopiedCode(null), 2000)
    })
  }

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
  const codes = mail.extracted_codes || []
  const links = mail.extracted_links || []

  return (
    <div className="card-modern">
      <div className="p-5">
        {codes.length > 0 && (
          <div className="mb-4 p-3 bg-primary/5 border border-primary/20 rounded-lg">
            <p className="text-xs font-medium text-primary/70 mb-2">{t('mailDetail.verificationCodes')}</p>
            <div className="flex flex-wrap gap-2">
              {codes.map((code) => (
                <button
                  key={code}
                  onClick={() => copyCode(code)}
                  className="inline-flex items-center gap-2 px-3 py-1.5 bg-base-100 border border-primary/30 rounded-md hover:bg-primary/10 transition-colors cursor-pointer"
                >
                  <span className="text-xl font-mono font-bold text-primary tracking-wider">{code}</span>
                  {copiedCode === code ? (
                    <Check size={14} className="text-success" />
                  ) : (
                    <Copy size={14} className="text-base-content/40" />
                  )}
                </button>
              ))}
            </div>
          </div>
        )}

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

        {links.length > 0 && (
          <div className="mt-4 pt-4 border-t border-base-300/40">
            <p className="text-xs font-medium text-base-content/50 mb-2">{t('mailDetail.extractedLinks')}</p>
            <div className="flex flex-col gap-1">
              {links.map((link) => (
                <a
                  key={link}
                  href={link}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 text-xs text-primary/80 hover:text-primary truncate"
                >
                  <ExternalLink size={12} className="shrink-0" />
                  <span className="truncate">{link}</span>
                </a>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
