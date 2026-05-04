import { useEffect, useState, useRef, useMemo } from 'react'
import DOMPurify from 'dompurify'
import { useTranslation } from 'react-i18next'
import { useToast } from './Toast'
import { Mail, FileText, Copy, Check, ExternalLink, ArrowLeft } from 'lucide-react'
import { apiPut } from '../lib/api'

export default function MailDetail({ mail, onMailRead, onBack }) {
  const { t } = useTranslation()
  const toast = useToast()
  const [copiedCode, setCopiedCode] = useState(null)
  const copiedTimerRef = useRef(null)

  useEffect(() => {
    if (mail?.id && !mail.is_read) {
      apiPut(`/api/mails/${mail.id}/read`).catch(() => {})
      onMailRead?.(mail.id)
    }
  }, [mail?.id])

  useEffect(() => {
    return () => clearTimeout(copiedTimerRef.current)
  }, [])

  const copyCode = (code) => {
    navigator.clipboard.writeText(code).then(() => {
      setCopiedCode(code)
      toast.success(t('mailDetail.codeCopied'))
      clearTimeout(copiedTimerRef.current)
      copiedTimerRef.current = setTimeout(() => setCopiedCode(null), 2000)
    })
  }

  const htmlContent = mail?.html || mail?.text_body || ''
  const sanitizedHtml = useMemo(() => DOMPurify.sanitize(htmlContent), [htmlContent])

  if (!mail) {
    return (
      <div className="card-modern h-full">
        <div className="flex flex-col items-center justify-center h-full py-16">
          <Mail size={64} className="text-base-content/10 mb-3 opacity-20" />
          <p className="text-sm text-base-content/30">{t('mailDetail.empty')}</p>
        </div>
      </div>
    )
  }

  const codes = mail.extracted_codes || []
  const links = mail.extracted_links || []

  return (
    <div className="card-modern">
      <div className="p-3 sm:p-5">
        {onBack && (
          <button
            className="btn btn-sm btn-ghost mb-3 lg:hidden"
            onClick={onBack}
          >
            <ArrowLeft size={16} />
            {t('mailDetail.back')}
          </button>
        )}

        <div className="mb-4">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-0.5 sm:gap-2">
            <span className="text-base sm:text-lg font-semibold text-base-content truncate">{mail.from}</span>
            <span className="text-xs sm:text-sm text-base-content/50 shrink-0 tabular-nums">
              {new Date(mail.created_at).toLocaleString()}
            </span>
          </div>
          <p className="text-xs sm:text-sm text-base-content/40 truncate">{t('mailDetail.to')} {mail.to}</p>
          <h2 className="text-sm sm:text-base font-medium text-base-content mt-1">
            {mail.subject || t('mailDetail.noSubject')}
          </h2>
        </div>

        {codes.length > 0 && (
          <div className="mb-4 p-2.5 sm:p-3 bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-lg">
            <p className="text-xs font-medium text-emerald-700 dark:text-emerald-400 mb-2">{t('mailDetail.verificationCodes')}</p>
            <div className="flex flex-wrap gap-1.5 sm:gap-2">
              {codes.map((code) => (
                <button
                  key={code}
                  onClick={() => copyCode(code)}
                  className="inline-flex items-center gap-1.5 sm:gap-2 px-2 py-1 sm:px-3 sm:py-1.5 bg-base-100 border border-emerald-300 dark:border-emerald-700 rounded-md hover:bg-emerald-100 dark:hover:bg-emerald-900/40 transition-colors cursor-pointer"
                >
                  <span className="text-lg sm:text-2xl font-mono font-bold text-emerald-700 dark:text-emerald-400 tracking-wider">{code}</span>
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

        <div className="border-t border-base-300/40 pt-3 sm:pt-4">
          {mail.html ? (
            <div
              className="prose prose-sm max-w-none prose-headings:text-base-content prose-p:text-base-content/80 prose-a:text-primary break-words"
              dangerouslySetInnerHTML={{ __html: sanitizedHtml }}
            />
          ) : (
            <pre className="whitespace-pre-wrap break-all text-xs sm:text-sm text-base-content/80 font-sans leading-relaxed">
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
                  className="inline-flex items-center gap-2 px-3 py-2 rounded-lg bg-base-200/60 hover:bg-base-200 text-xs text-primary/80 hover:text-primary truncate transition-colors"
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
