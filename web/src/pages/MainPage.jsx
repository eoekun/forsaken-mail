import { useState, useCallback } from 'react'
import Navbar from '../components/Navbar'
import MailboxTabs from '../components/MailboxTabs'
import MailList from '../components/MailList'
import MailDetail from '../components/MailDetail'
import HelpModal from '../components/HelpModal'
import useWebSocket from '../hooks/useWebSocket'
import { normalizeMail } from '../lib/normalizeMail'
import { useAuth } from '../App'

export default function MainPage() {
  const { config } = useAuth()
  const {
    requestNewShortId,
    tabs, activeShortId, setActiveShortId, subscribeToShortId, unsubscribeFromShortId,
    mails, selectedMail, setSelectedMail, markMailAsRead,
    recentMails,
  } = useWebSocket(config?.host)

  const [mobileView, setMobileView] = useState('list')

  const handleSelectMail = useCallback((mail) => {
    setSelectedMail(mail)
    setMobileView('detail')
  }, [setSelectedMail])

  const handleMobileBack = useCallback(() => {
    setMobileView('list')
  }, [])

  const handleOpenRecentMail = useCallback((mail) => {
    subscribeToShortId(mail.short_id || (mail.to_addr || mail.to || '').split('@')[0])
    setSelectedMail(normalizeMail(mail))
    setMobileView('detail')
  }, [subscribeToShortId, setSelectedMail])

  return (
    <div className="min-h-screen bg-base-200 flex flex-col">
      <Navbar recentMails={recentMails} onOpenRecentMail={handleOpenRecentMail} />
      <div className="max-w-7xl w-full mx-auto px-4 sm:px-6 flex-1 flex flex-col min-h-0 pb-6">
        <MailboxTabs
          tabs={tabs}
          activeShortId={activeShortId}
          host={config?.host}
          onSelect={setActiveShortId}
          onClose={unsubscribeFromShortId}
          onAdd={requestNewShortId}
          onSetShortId={subscribeToShortId}
        />

        {/* Mobile layout: toggle between list and detail */}
        <div className="lg:hidden flex-1 min-h-0 mt-3">
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
        <div className="hidden lg:grid grid-cols-5 gap-4 flex-1 min-h-0 mt-3">
          <div className="col-span-2 min-h-0">
            <MailList
              mails={mails}
              selectedMail={selectedMail}
              onSelect={setSelectedMail}
            />
          </div>
          <div className="col-span-3 min-h-0">
            <MailDetail mail={selectedMail} onMailRead={markMailAsRead} />
          </div>
        </div>
      </div>
      <HelpModal host={config?.host} />
    </div>
  )
}
