import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { apiGet } from '../lib/api'
import { RefreshCw, CheckCircle, AlertTriangle } from 'lucide-react'

function formatValue(value) {
  if (typeof value === 'number') {
    return value.toLocaleString()
  }
  return String(value)
}

function getStatusColor(key, value) {
  if (typeof value === 'string') {
    const lower = value.toLowerCase()
    if (lower === 'ok' || lower === 'healthy' || lower === 'running' || lower === 'connected') {
      return 'success'
    }
    if (lower === 'error' || lower === 'failed' || lower === 'disconnected') {
      return 'error'
    }
  }
  return null
}

function StatusCard({ label, value, t }) {
  const statusColor = getStatusColor(label, value)
  const borderColor = statusColor === 'success'
    ? 'border-l-success'
    : statusColor === 'error'
      ? 'border-l-error'
      : 'border-l-transparent'

  return (
    <div className={`card-modern p-4 border-l-3 ${borderColor}`}>
      <div className="text-[11px] font-medium text-base-content/40 uppercase tracking-wider mb-1">
        {label.replace(/_/g, ' ')}
      </div>
      <div className="flex items-center gap-2">
        {statusColor === 'success' && <CheckCircle size={14} className="text-success shrink-0" />}
        {statusColor === 'error' && <AlertTriangle size={14} className="text-error shrink-0" />}
        <div className={`text-base font-mono font-medium ${
          statusColor === 'success' ? 'text-success' : statusColor === 'error' ? 'text-error' : 'text-base-content'
        }`}>
          {formatValue(value)}
        </div>
      </div>
    </div>
  )
}

export default function StatusTab() {
  const { t } = useTranslation()
  const [status, setStatus] = useState(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const fetchStatus = useCallback(async () => {
    try {
      const data = await apiGet('/api/admin/status')
      setStatus(data)
    } catch (e) {
      console.error('Failed to load status:', e)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [])

  useEffect(() => {
    fetchStatus()
  }, [fetchStatus])

  const handleRefresh = () => {
    setRefreshing(true)
    fetchStatus()
  }

  if (loading) return <div className="flex justify-center py-12"><span className="loading loading-spinner text-primary"></span></div>

  if (!status) return <div className="alert alert-error rounded-lg">{t('status.failed')}</div>

  return (
    <div>
      <div className="flex justify-end mb-3">
        <button
          className="btn btn-sm btn-ghost gap-1.5"
          onClick={handleRefresh}
          disabled={refreshing}
        >
          <RefreshCw size={14} className={refreshing ? 'animate-spin' : ''} />
          {refreshing ? t('status.refreshing') : t('status.refresh')}
        </button>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {Object.entries(status).map(([key, value]) => (
          <StatusCard key={key} label={key} value={value} t={t} />
        ))}
      </div>
    </div>
  )
}
