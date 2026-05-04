import { useTranslation } from 'react-i18next'
import { Inbox } from 'lucide-react'

export default function MailList({ mails, selectedMail, onSelect }) {
  const { t } = useTranslation()

  if (mails.length === 0) {
    return (
      <div className="card-modern">
        <div className="p-5">
          <div className="flex items-center gap-2 mb-3">
            <Inbox size={16} className="text-base-content/40" />
            <span className="text-sm font-medium text-base-content/70">{t('mailList.title')}</span>
          </div>
          <div className="flex flex-col items-center justify-center py-8">
            <Inbox size={32} className="text-base-content/15 mb-2" />
            <p className="text-sm text-base-content/30">{t('mailList.empty')}</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="card-modern">
      <div className="p-4 pb-2">
        <div className="flex items-center gap-2">
          <Inbox size={16} className="text-base-content/40" />
          <span className="text-sm font-medium text-base-content/70">{t('mailList.title')}</span>
          <span className="ml-auto text-xs text-base-content/30">{mails.length}</span>
        </div>
      </div>
      <div className="divide-y divide-base-300/40">
        {mails.map((mail, idx) => {
          const isUnread = !mail.is_read
          return (
            <div
              key={mail.id || idx}
              className={`px-4 py-3 cursor-pointer transition-colors duration-100 ${
                selectedMail === mail
                  ? 'bg-primary/5 border-l-2 border-l-primary'
                  : 'border-l-2 border-l-transparent hover:bg-base-200/60'
              }`}
              onClick={() => onSelect(mail)}
            >
              <div className="flex items-baseline justify-between gap-2 mb-0.5">
                <span className={`text-sm truncate ${selectedMail === mail ? 'font-medium text-base-content' : isUnread ? 'font-semibold text-base-content' : 'text-base-content/50'}`}>
                  {isUnread && <span className="inline-block w-1.5 h-1.5 rounded-full bg-primary mr-1.5 align-middle" />}
                  {mail.from}
                </span>
                <span className="text-[11px] text-base-content/30 shrink-0 tabular-nums">
                  {new Date(mail.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>
              <p className={`text-xs truncate ${isUnread ? 'text-base-content/70 font-medium' : 'text-base-content/50'}`}>
                {mail.subject || t('mailList.noSubject')}
              </p>
            </div>
          )
        })}
      </div>
    </div>
  )
}
