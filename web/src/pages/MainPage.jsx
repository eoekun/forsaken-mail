import Navbar from '../components/Navbar'
import MailboxAddress from '../components/MailboxAddress'
import MailHistory from '../components/MailHistory'
import MailList from '../components/MailList'
import MailDetail from '../components/MailDetail'
import HelpModal from '../components/HelpModal'
import useWebSocket from '../hooks/useWebSocket'
import { useAuth } from '../App'

export default function MainPage() {
  const { config } = useAuth()
  const {
    shortId, setShortId, requestNewShortId,
    mails, selectedMail, setSelectedMail, clearMails,
  } = useWebSocket(config?.host)

  return (
    <div className="min-h-screen bg-base-200">
      <Navbar />
      <div className="container mx-auto p-4 max-w-6xl">
        <MailboxAddress
          shortId={shortId}
          host={config?.host}
          onRefresh={requestNewShortId}
          onSetShortId={setShortId}
        />
        <MailHistory
          host={config?.host}
          activeShortId={shortId}
          onSelect={setShortId}
        />
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-4 mt-4">
          <div className="lg:col-span-2">
            <MailList
              mails={mails}
              selectedMail={selectedMail}
              onSelect={setSelectedMail}
            />
          </div>
          <div className="lg:col-span-3">
            <MailDetail mail={selectedMail} />
          </div>
        </div>
      </div>
      <HelpModal host={config?.host} />
    </div>
  )
}
