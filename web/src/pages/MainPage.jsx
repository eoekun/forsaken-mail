import { useState, useCallback } from 'react'
import Navbar from '../components/Navbar'
import MailboxAddress from '../components/MailboxAddress'
import MailboxTabs from '../components/MailboxTabs'
import MailHistory from '../components/MailHistory'
import MailList from '../components/MailList'
import MailDetail from '../components/MailDetail'
import RecentMails from '../components/RecentMails'
import HelpModal from '../components/HelpModal'
import useWebSocket from '../hooks/useWebSocket'
import { useAuth } from '../App'

export default function MainPage() {
  const { config } = useAuth()
  const {
    shortId, requestNewShortId,
    tabs, activeShortId, setActiveShortId, subscribeToShortId, unsubscribeFromShortId,
    mails, selectedMail, setSelectedMail, clearMails, markMailAsRead,
    recentMails, loadRecentMails,
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
    // Normalize API field names to match MailDetail expectations
    setSelectedMail({
      ...mail,
      from: mail.from || mail.from_addr,
      to: mail.to || mail.to_addr,
      html: mail.html || mail.html_body,
      text: mail.text || mail.text_body,
    })
    setMobileView('detail')
  }, [subscribeToShortId, setSelectedMail])

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
        <RecentMails
          recentMails={recentMails}
          onOpenMail={handleOpenRecentMail}
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
