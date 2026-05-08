import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useToast } from './Toast'
import { Save, Play } from 'lucide-react'
import { SETTINGS_SECTIONS, createSettingsForm, getSettingField, serializeSettingsForm } from '../lib/settingsSchema'
import { loadSettings, saveSettings, testWebhook } from '../lib/settingsApi'

function SettingInput({ meta, value, onChange }) {
  const { t } = useTranslation()

  if (meta.type === 'toggle') {
    const isOn = Boolean(value)
    return (
      <button
        type="button"
        className={`btn btn-xs ${isOn ? 'btn-success' : 'btn-ghost'}`}
        onClick={() => onChange(!isOn)}
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
        value={Number.isFinite(value) ? value : ''}
        onChange={e => onChange(e.target.value === '' ? meta.defaultValue : Number(e.target.value))}
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
  const [settings, setSettings] = useState(() => createSettingsForm())
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)

  useEffect(() => {
    loadSettings()
      .then(data => setSettings(createSettingsForm(data)))
      .catch(e => console.error('Failed to load settings:', e))
      .finally(() => setLoading(false))
  }, [])

  const handleSave = async () => {
    setSaving(true)
    try {
      const savedValues = await saveSettings(serializeSettingsForm(settings))
      setSettings(createSettingsForm(savedValues))
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
      const payload = serializeSettingsForm(settings)
      const result = await testWebhook({
        service: payload.webhook_service || 'dingtalk',
        config: JSON.stringify(payload.webhook_config || {}),
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
                const meta = getSettingField(key)
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
