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

  if (loading) return <div className="flex justify-center"><span className="loading loading-spinner"></span></div>

  if (!status) return <div className="alert alert-error">{t('status.failed')}</div>

  return (
    <div className="card bg-base-100 shadow-md">
      <div className="card-body">
        <div className="grid grid-cols-2 gap-4">
          {Object.entries(status).map(([key, value]) => (
            <div key={key}>
              <div className="text-sm text-base-content/70">{key}</div>
              <div className="text-lg font-mono">{String(value)}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
