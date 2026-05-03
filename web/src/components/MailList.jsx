import { useTranslation } from 'react-i18next'

export default function MailList({ mails, selectedMail, onSelect }) {
  const { t } = useTranslation()

  if (mails.length === 0) {
    return (
      <div className="card bg-base-100 shadow-md">
        <div className="card-body">
          <h3 className="card-title text-sm">{t('mailList.title')}</h3>
          <p className="text-base-content/50 text-sm">{t('mailList.empty')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="card bg-base-100 shadow-md">
      <div className="card-body p-0">
        <div className="overflow-x-auto">
          <table className="table table-sm">
            <thead>
              <tr>
                <th>{t('mailList.from')}</th>
                <th>{t('mailList.subject')}</th>
                <th>{t('mailList.time')}</th>
              </tr>
            </thead>
            <tbody>
              {mails.map((mail, idx) => (
                <tr
                  key={mail.id || idx}
                  className={`cursor-pointer hover:bg-base-200 ${selectedMail === mail ? 'active' : ''}`}
                  onClick={() => onSelect(mail)}
                >
                  <td className="max-w-[120px] truncate">{mail.from}</td>
                  <td className="max-w-[150px] truncate">{mail.subject || t('mailList.noSubject')}</td>
                  <td className="whitespace-nowrap">{new Date(mail.created_at).toLocaleTimeString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
