import DOMPurify from 'dompurify'

export default function MailDetail({ mail }) {
  if (!mail) {
    return (
      <div className="card bg-base-100 shadow-md">
        <div className="card-body">
          <h3 className="card-title text-sm">Mail Detail</h3>
          <p className="text-base-content/50 text-sm">Select an email to view its content.</p>
        </div>
      </div>
    )
  }

  const htmlContent = mail.html || mail.text_body || ''

  return (
    <div className="card bg-base-100 shadow-md">
      <div className="card-body">
        <h3 className="card-title">{mail.subject || '(no subject)'}</h3>
        <div className="text-sm text-base-content/70 mb-2">
          <span>From: {mail.from}</span>
          <span className="ml-4">To: {mail.to}</span>
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
