import Navbar from '../components/Navbar'
import MailboxAddress from '../components/MailboxAddress'
import MailboxTabs from '../components/MailboxTabs'
import MailHistory from '../components/MailHistory'
import MailList from '../components/MailList'
import MailDetail from '../components/MailDetail'
import HelpModal from '../components/HelpModal'
import useWebSocket from '../hooks/useWebSocket'
import { useAuth } from '../App'

export default function MainPage() {
  const { config } = useAuth()
  const {
    shortId, requestNewShortId,
    tabs, activeShortId, setActiveShortId, subscribeToShortId, unsubscribeFromShortId,
    mails, selectedMail, setSelectedMail, clearMails, markMailAsRead,
  } = useWebSocket(config?.host)

  return (
    <div className="min-h-screen bg-base-200">
      <Navbar />
      <div className="max-w-7xl mx-auto px-4 sm:px-6 pb-24">
        <MailboxAddress
          shortId={shortId}
          host={config?.host}
          onRefresh={requestNewShortId}
          onSetShortId={subscribeToShortId}
        />
        <MailHistory
          host={config?.host}
          activeShortId={activeShortId}
          onSelect={subscribeToShortId}
        />
        <div className="mt-3">
          <MailboxTabs
            tabs={tabs}
            activeShortId={activeShortId}
            host={config?.host}
            onSelect={setActiveShortId}
            onClose={unsubscribeFromShortId}
            onAdd={requestNewShortId}
          />
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-4 mt-4">
          <div className="lg:col-span-2">
            <MailList
              mails={mails}
              selectedMail={selectedMail}
              onSelect={setSelectedMail}
            />
          </div>
          <div className="lg:col-span-3">
            <MailDetail mail={selectedMail} onMailRead={markMailAsRead} />
          </div>
        </div>
      </div>
      <HelpModal host={config?.host} />
    </div>
  )
}
