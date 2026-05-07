import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { apiGet, apiPut, apiPost } from '../lib/api'
import { useToast } from './Toast'
import { Save, Play } from 'lucide-react'

// Input type metadata for each setting key.
const INPUT_META = {
  mail_host:              { type: 'text', placeholder: 'example.com,mail.example.com' },
  site_title:             { type: 'text', placeholder: 'Tmail' },
  login_whitelist:        { type: 'text', placeholder: 'user1@example.com,user2@example.com' },
  keyword_blacklist:      { type: 'text', placeholder: 'admin,root,system' },
  webhook_enabled:        { type: 'toggle' },
  webhook_service:        { type: 'select', options: ['dingtalk', 'telegram', 'slack'] },
  webhook_config:         { type: 'textarea', placeholder: '{"token":"...","chat_id":"..."}' },
  webhook_message:        { type: 'text', placeholder: 'new email received.' },
  mail_retention_hours:   { type: 'number', min: 0 },
  mail_max_count:         { type: 'number', min: 0 },
  max_mail_size_bytes:    { type: 'number', min: 0 },
  audit_mail_received:    { type: 'toggle' },
  audit_retention_days:   { type: 'number', min: 0 },
  audit_max_count:        { type: 'number', min: 0 },
}

const SETTING_SECTIONS = [
  { sectionKey: 'general',      keys: ['mail_host', 'site_title'] },
  { sectionKey: 'security',     keys: ['login_whitelist', 'keyword_blacklist'] },
  { sectionKey: 'notifications', keys: ['webhook_enabled', 'webhook_service', 'webhook_config', 'webhook_message'] },
  { sectionKey: 'retention',    keys: ['mail_retention_hours', 'mail_max_count', 'max_mail_size_bytes'] },
  { sectionKey: 'audit',        keys: ['audit_mail_received', 'audit_retention_days', 'audit_max_count'] },
]

const ALL_KEYS = SETTING_SECTIONS.flatMap(s => s.keys)

function SettingInput({ meta, value, onChange }) {
  const { t } = useTranslation()

  if (meta.type === 'toggle') {
    const isOn = value === '1'
    return (
      <button
        type="button"
        className={`btn btn-xs ${isOn ? 'btn-success' : 'btn-ghost'}`}
        onClick={() => onChange(isOn ? '0' : '1')}
      >
        {isOn ? t('settings.on') : t('settings.off')}
      </button>
    )
  }

  if (meta.type === 'select') {
    return (
      <select
        className="select select-sm select-bordered w-full"
        value={value || ''}
        onChange={e => onChange(e.target.value)}
      >
        {meta.options.map(opt => (
          <option key={opt} value={opt}>{opt}</option>
        ))}
      </select>
    )
  }

  if (meta.type === 'textarea') {
    return (
      <textarea
        className="textarea textarea-bordered textarea-sm w-full font-mono text-xs"
        rows={3}
        value={value || ''}
        onChange={e => onChange(e.target.value)}
        placeholder={meta.placeholder}
      />
    )
  }

  if (meta.type === 'number') {
    return (
      <input
        type="number"
        className="input-modern input-sm w-full"
        value={value || ''}
        onChange={e => onChange(e.target.value)}
        min={meta.min}
        placeholder={meta.placeholder}
      />
    )
  }

  return (
    <input
      type="text"
      className="input-modern input-sm w-full"
      value={value || ''}
      onChange={e => onChange(e.target.value)}
      placeholder={meta.placeholder}
    />
  )
}

export default function SettingsTab() {
  const { t } = useTranslation()
  const toast = useToast()
  const [settings, setSettings] = useState({})
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)

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
      toast.success(t('settings.saved'))
    } catch (e) {
      toast.error(t('settings.error', { message: e.message }))
    } finally {
      setSaving(false)
    }
  }

  const handleTestWebhook = async () => {
    setTesting(true)
    try {
      const result = await apiPost('/api/webhook/test', {
        service: settings.webhook_service || 'dingtalk',
        config: settings.webhook_config || '{}',
        message: '',
      })
      if (result.ok) {
        toast.success(t('settings.testSuccess'))
      } else {
        toast.error(result.message || t('settings.testFailed'))
      }
    } catch (e) {
      toast.error(t('settings.error', { message: e.message }))
    } finally {
      setTesting(false)
    }
  }

  if (loading) return <div className="flex justify-center py-12"><span className="loading loading-spinner text-primary"></span></div>

  return (
    <div className="card-modern p-5">
      <div className="space-y-6">
        {SETTING_SECTIONS.map(({ sectionKey, keys }) => (
          <div key={sectionKey}>
            <h3 className="text-sm font-semibold text-base-content/70 mb-3 pb-1.5 border-b border-base-300/40">
              {t(`settings.sections.${sectionKey}`)}
            </h3>
            <div className="space-y-4">
              {keys.map(key => {
                const meta = INPUT_META[key] || { type: 'text' }
                return (
                  <div key={key}>
                    <label className="block text-xs font-medium text-base-content/60 mb-1">
                      {t(`settings.labels.${key}`, { defaultValue: key })}
                    </label>
                    <SettingInput
                      meta={meta}
                      value={settings[key]}
                      onChange={val => setSettings(prev => ({ ...prev, [key]: val }))}
                    />
                    <p className="text-[11px] text-base-content/30 mt-1">
                      {t(`settings.descriptions.${key}`, { defaultValue: '' })}
                    </p>
                  </div>
                )
              })}
            </div>
          </div>
        ))}
      </div>
      <div className="mt-6 pt-4 border-t border-base-300/40 flex items-center gap-2">
        <button
          className={`btn-modern btn-sm btn-primary gap-2 ${saving ? 'loading' : ''}`}
          onClick={handleSave}
          disabled={saving}
        >
          {!saving && <Save size={14} />}
          {saving ? t('settings.saving') : t('settings.save')}
        </button>
        <button
          className={`btn btn-sm btn-ghost gap-2 ${testing ? 'loading' : ''}`}
          onClick={handleTestWebhook}
          disabled={testing}
        >
          {!testing && <Play size={14} />}
          {testing ? t('settings.testing') : t('settings.testWebhook')}
        </button>
      </div>
    </div>
  )
}
