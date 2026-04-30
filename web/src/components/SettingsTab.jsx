import { useState, useEffect } from 'react'
import { apiGet, apiPut } from '../lib/api'

const EDITABLE_KEYS = [
  'mail_host', 'site_title', 'allowed_emails', 'keyword_blacklist',
  'dingtalk_webhook_token', 'dingtalk_webhook_message',
  'mail_retention_hours', 'mail_max_count', 'max_mail_size_bytes',
  'audit_retention_days', 'audit_max_count',
]

export default function SettingsTab() {
  const [settings, setSettings] = useState({})
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [toast, setToast] = useState('')

  useEffect(() => {
    apiGet('/api/admin/settings')
      .then(data => setSettings(data))
      .catch(e => console.error('Failed to load settings:', e))
      .finally(() => setLoading(false))
  }, [])

  const handleSave = async () => {
    setSaving(true)
    try {
      const updates = {}
      for (const key of EDITABLE_KEYS) {
        if (key in settings) updates[key] = settings[key]
      }
      await apiPut('/api/admin/settings', updates)
      setToast('Settings saved!')
      setTimeout(() => setToast(''), 3000)
    } catch (e) {
      setToast(`Error: ${e.message}`)
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div className="flex justify-center"><span className="loading loading-spinner"></span></div>

  return (
    <div className="card bg-base-100 shadow-md">
      <div className="card-body">
        {toast && <div className="alert alert-success mb-4"><span>{toast}</span></div>}
        <div className="grid gap-4">
          {EDITABLE_KEYS.map(key => (
            <div key={key} className="form-control">
              <label className="label"><span className="label-text">{key}</span></label>
              <input
                type="text"
                className="input input-bordered input-sm"
                value={settings[key] || ''}
                onChange={e => setSettings(prev => ({ ...prev, [key]: e.target.value }))}
              />
            </div>
          ))}
        </div>
        <div className="mt-4">
          <button className={`btn btn-primary ${saving ? 'loading' : ''}`} onClick={handleSave} disabled={saving}>
            {saving ? 'Saving...' : 'Save Settings'}
          </button>
        </div>
      </div>
    </div>
  )
}
