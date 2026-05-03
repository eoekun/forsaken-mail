import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import Navbar from '../components/Navbar'
import AuditLogTab from '../components/AuditLogTab'
import SettingsTab from '../components/SettingsTab'
import StatusTab from '../components/StatusTab'

export default function AdminPage() {
  const { t } = useTranslation()
  const [tab, setTab] = useState('audit')

  return (
    <div className="min-h-screen bg-base-200">
      <Navbar />
      <div className="container mx-auto p-4 max-w-4xl">
        <div className="tabs tabs-boxed mb-4">
          <button className={`tab ${tab === 'audit' ? 'tab-active' : ''}`} onClick={() => setTab('audit')}>{t('admin.auditLogs')}</button>
          <button className={`tab ${tab === 'settings' ? 'tab-active' : ''}`} onClick={() => setTab('settings')}>{t('admin.settings')}</button>
          <button className={`tab ${tab === 'status' ? 'tab-active' : ''}`} onClick={() => setTab('status')}>{t('admin.status')}</button>
        </div>
        {tab === 'audit' && <AuditLogTab />}
        {tab === 'settings' && <SettingsTab />}
        {tab === 'status' && <StatusTab />}
      </div>
    </div>
  )
}
