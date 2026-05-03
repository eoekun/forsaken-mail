import DOMPurify from 'dompurify'
import { useTranslation } from 'react-i18next'

export default function MailDetail({ mail }) {
  const { t } = useTranslation()

  if (!mail) {
    return (
      <div className="card bg-base-100 shadow-md">
        <div className="card-body">
          <h3 className="card-title text-sm">{t('mailDetail.title')}</h3>
          <p className="text-base-content/50 text-sm">{t('mailDetail.empty')}</p>
        </div>
      </div>
    )
  }

  const htmlContent = mail.html || mail.text_body || ''

  return (
    <div className="card bg-base-100 shadow-md">
      <div className="card-body">
        <h3 className="card-title">{mail.subject || t('mailDetail.noSubject')}</h3>
        <div className="text-sm text-base-content/70 mb-2">
          <span>{t('mailDetail.from')} {mail.from}</span>
          <span className="ml-4">{t('mailDetail.to')} {mail.to}</span>
          <span className="ml-4">{new Date(mail.created_at).toLocaleString()}</span>
        </div>
        <div className="divider my-1"></div>
        {mail.html ? (
          <div
            className="prose max-w-none"
            dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(htmlContent) }}
          />
        ) : (
          <pre className="whitespace-pre-wrap text-sm">{mail.text_body}</pre>
        )}
      </div>
    </div>
  )
}
