import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { apiGet, apiPut } from '../lib/api'
import { Save, Check } from 'lucide-react'

const SETTING_SECTIONS = [
  {
    sectionKey: 'general',
    keys: ['mail_host', 'site_title'],
  },
  {
    sectionKey: 'security',
    keys: ['allowed_emails', 'keyword_blacklist'],
  },
  {
    sectionKey: 'notifications',
    keys: ['dingtalk_webhook_token', 'dingtalk_webhook_message'],
  },
  {
    sectionKey: 'retention',
    keys: ['mail_retention_hours', 'mail_max_count', 'max_mail_size_bytes'],
  },
  {
    sectionKey: 'audit',
    keys: ['audit_retention_days', 'audit_max_count'],
  },
]

const ALL_KEYS = SETTING_SECTIONS.flatMap(s => s.keys)

export default function SettingsTab() {
  const { t } = useTranslation()
  const [settings, setSettings] = useState({})
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [toast, setToast] = useState('')
  const toastTimerRef = useRef(null)

  useEffect(() => {
    return () => clearTimeout(toastTimerRef.current)
  }, [])

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
      for (const key of ALL_KEYS) {
        if (key in settings) updates[key] = settings[key]
      }
      await apiPut('/api/admin/settings', updates)
      setToast(t('settings.saved'))
      clearTimeout(toastTimerRef.current)
      toastTimerRef.current = setTimeout(() => setToast(''), 3000)
    } catch (e) {
      setToast(t('settings.error', { message: e.message }))
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div className="flex justify-center py-12"><span className="loading loading-spinner text-primary"></span></div>

  return (
    <div className="card-modern p-5">
      {toast && (
        <div className="flex items-center gap-2 mb-4 px-4 py-2.5 rounded-lg bg-success/10 text-success text-sm">
          <Check size={15} />
          <span>{toast}</span>
        </div>
      )}
      <div className="space-y-6">
        {SETTING_SECTIONS.map(({ sectionKey, keys }) => (
          <div key={sectionKey}>
            <h3 className="text-sm font-semibold text-base-content/70 mb-3 pb-1.5 border-b border-base-300/40">
              {t(`settings.sections.${sectionKey}`)}
            </h3>
            <div className="space-y-4">
              {keys.map(key => (
                <div key={key}>
                  <label className="block text-xs font-medium text-base-content/60 mb-1">
                    {t(`settings.labels.${key}`, { defaultValue: key })}
                  </label>
                  <input
                    type="text"
                    className="input-modern input-sm w-full"
                    value={settings[key] || ''}
                    onChange={e => setSettings(prev => ({ ...prev, [key]: e.target.value }))}
                  />
                  <p className="text-[11px] text-base-content/30 mt-1">
                    {t(`settings.descriptions.${key}`, { defaultValue: '' })}
                  </p>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
      <div className="mt-6 pt-4 border-t border-base-300/40">
        <button
          className={`btn-modern btn-sm btn-primary gap-2 ${saving ? 'loading' : ''}`}
          onClick={handleSave}
          disabled={saving}
        >
          {!saving && <Save size={14} />}
          {saving ? t('settings.saving') : t('settings.save')}
        </button>
      </div>
    </div>
  )
}
