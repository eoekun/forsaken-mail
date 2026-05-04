import { useState, useCallback } from 'react'
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

  const [mobileView, setMobileView] = useState('list')

  const handleSelectMail = useCallback((mail) => {
    setSelectedMail(mail)
    setMobileView('detail')
  }, [setSelectedMail])

  const handleMobileBack = useCallback(() => {
    setMobileView('list')
  }, [])

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

        {/* Mobile layout: toggle between list and detail */}
        <div className="lg:hidden mt-4">
          {mobileView === 'list' ? (
            <MailList
              mails={mails}
              selectedMail={selectedMail}
              onSelect={handleSelectMail}
            />
          ) : (
            <MailDetail
              mail={selectedMail}
              onMailRead={markMailAsRead}
              onBack={handleMobileBack}
            />
          )}
        </div>

        {/* Desktop layout: side-by-side grid */}
        <div className="hidden lg:grid grid-cols-5 gap-4 mt-4">
          <div className="col-span-2">
            <MailList
              mails={mails}
              selectedMail={selectedMail}
              onSelect={setSelectedMail}
            />
          </div>
          <div className="col-span-3">
            <MailDetail mail={selectedMail} onMailRead={markMailAsRead} />
          </div>
        </div>
      </div>
      <HelpModal host={config?.host} />
    </div>
  )
}
