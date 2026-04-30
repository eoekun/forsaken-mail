import { useState } from 'react'
import Navbar from '../components/Navbar'
import AuditLogTab from '../components/AuditLogTab'
import SettingsTab from '../components/SettingsTab'
import StatusTab from '../components/StatusTab'

export default function AdminPage() {
  const [tab, setTab] = useState('audit')

  return (
    <div className="min-h-screen bg-base-200">
      <Navbar />
      <div className="container mx-auto p-4 max-w-4xl">
        <div className="tabs tabs-boxed mb-4">
          <button className={`tab ${tab === 'audit' ? 'tab-active' : ''}`} onClick={() => setTab('audit')}>Audit Logs</button>
          <button className={`tab ${tab === 'settings' ? 'tab-active' : ''}`} onClick={() => setTab('settings')}>Settings</button>
          <button className={`tab ${tab === 'status' ? 'tab-active' : ''}`} onClick={() => setTab('status')}>Status</button>
        </div>
        {tab === 'audit' && <AuditLogTab />}
        {tab === 'settings' && <SettingsTab />}
        {tab === 'status' && <StatusTab />}
      </div>
    </div>
  )
}
