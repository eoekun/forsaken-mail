import { useEffect, useState, useMemo } from 'react'
import DOMPurify from 'dompurify'
import { useTranslation } from 'react-i18next'
import { useToast } from './Toast'
import { Mail, Copy, Check, ExternalLink, ArrowLeft, KeyRound } from 'lucide-react'
import { apiPut } from '../lib/api'
import { formatSender } from '../lib/formatSender'
import { useCopyToClipboard } from '../hooks/useCopyToClipboard'

export default function MailDetail({ mail, onMailRead, onBack }) {
  const { t } = useTranslation()
  const toast = useToast()
  const { copiedId: copiedCode, copy } = useCopyToClipboard()

  useEffect(() => {
    if (mail?.id && !mail.is_read) {
      apiPut(`/api/mails/${mail.id}/read`).catch(() => {})
      onMailRead?.(mail.id)
    }
  }, [mail?.id])

  const copyCode = (code) => {
    copy(code, code)
    toast.success(t('mailDetail.codeCopied'))
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
        <div className="mb-4">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-0.5 sm:gap-2">
            <span className="text-base sm:text-lg font-semibold text-base-content truncate">{formatSender(mail.from)}</span>
            <span className="text-xs sm:text-sm text-base-content/50 shrink-0 tabular-nums">
              {new Date(mail.created_at).toLocaleString()}
            </span>
          </div>
          {formatSender(mail.from) !== mail.from && (
            <p className="text-xs text-base-content/40 font-mono truncate mt-0.5" title={mail.from}>
              {t('mailDetail.from')}: {mail.from}
            </p>
          )}
          <p className="text-xs sm:text-sm text-base-content/40 truncate">{t('mailDetail.to')} {mail.to}</p>
          <h2 className="text-sm sm:text-base font-medium text-base-content mt-1">
            {mail.subject || t('mailDetail.noSubject')}
          </h2>
        </div>

        {codes.length > 0 && (
          <div className="mb-4 rounded-xl overflow-hidden border border-emerald-200 dark:border-emerald-800/60">
            <div className="bg-gradient-to-r from-emerald-50 to-teal-50 dark:from-emerald-950/40 dark:to-teal-950/40 px-3.5 py-2 flex items-center gap-2">
              <KeyRound size={14} className="text-emerald-600 dark:text-emerald-400" />
              <span className="text-xs font-semibold text-emerald-700 dark:text-emerald-400 tracking-wide uppercase">{t('mailDetail.verificationCodes')}</span>
            </div>
            <div className="bg-base-100 px-3.5 py-3 flex flex-wrap gap-2">
              {codes.map((code) => (
                <button
                  key={code}
                  onClick={() => copyCode(code)}
                  className={`group inline-flex items-center gap-2 px-3 py-2 rounded-lg border transition-all cursor-pointer ${
                    copiedCode === code
                      ? 'bg-emerald-100 dark:bg-emerald-900/30 border-emerald-400 dark:border-emerald-600'
                      : 'bg-emerald-50/80 dark:bg-emerald-950/30 border-emerald-200 dark:border-emerald-800/50 hover:border-emerald-400 dark:hover:border-emerald-600 hover:shadow-sm'
                  }`}
                >
                  <span className="text-xl sm:text-2xl font-mono font-bold text-emerald-700 dark:text-emerald-300 tracking-[0.15em] select-all">{code}</span>
                  {copiedCode === code ? (
                    <Check size={15} className="text-emerald-600 dark:text-emerald-400" />
                  ) : (
                    <Copy size={15} className="text-emerald-400 dark:text-emerald-600 opacity-0 group-hover:opacity-100 transition-opacity" />
                  )}
                </button>
              ))}
            </div>
            <div className="bg-emerald-50/50 dark:bg-emerald-950/20 px-3.5 py-1.5 border-t border-emerald-100 dark:border-emerald-800/30">
              <p className="text-[11px] text-emerald-600/60 dark:text-emerald-400/50">{t('mailDetail.clickToCopy')}</p>
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
      {onBack && (
        <button
          className="lg:hidden fixed bottom-4 left-4 right-4 btn btn-primary btn-sm gap-2 z-30 shadow-lg"
          onClick={onBack}
        >
          <ArrowLeft size={16} />
          {t('mailDetail.back')}
        </button>
      )}
    </div>
  )
}
