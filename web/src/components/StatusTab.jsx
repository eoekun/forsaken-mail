import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { apiGet } from '../lib/api'

export default function StatusTab() {
  const { t } = useTranslation()
  const [status, setStatus] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    apiGet('/api/admin/status')
      .then(data => setStatus(data))
      .catch(e => console.error('Failed to load status:', e))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="flex justify-center py-12"><span className="loading loading-spinner text-primary"></span></div>

  if (!status) return <div className="alert alert-error rounded-lg">{t('status.failed')}</div>

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
      {Object.entries(status).map(([key, value]) => (
        <div key={key} className="card-modern p-4">
          <div className="text-[11px] font-medium text-base-content/40 uppercase tracking-wider mb-1">{key}</div>
          <div className="text-base font-mono font-medium text-base-content">{String(value)}</div>
        </div>
      ))}
    </div>
  )
}
