import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { apiGet } from '../lib/api'
import { ChevronLeft, ChevronRight, X } from 'lucide-react'

const EVENTS = ['', 'LOGIN', 'LOGIN_DENIED', 'LOGOUT', 'MAILBOX_CREATE', 'MAIL_RECEIVED', 'MAIL_DROPPED', 'WEBHOOK_SENT', 'CONFIG_CHANGED']

const EVENT_COLORS = {
  LOGIN: 'bg-success/10 text-success',
  LOGIN_DENIED: 'bg-error/10 text-error',
  LOGOUT: 'bg-base-content/10 text-base-content/60',
  MAILBOX_CREATE: 'bg-info/10 text-info',
  MAIL_RECEIVED: 'bg-primary/10 text-primary',
  MAIL_DROPPED: 'bg-warning/10 text-warning',
  WEBHOOK_SENT: 'bg-accent/10 text-accent',
  CONFIG_CHANGED: 'bg-secondary/10 text-secondary',
}

export default function AuditLogTab() {
  const { t } = useTranslation()
  const [logs, setLogs] = useState([])
  const [total, setTotal] = useState(0)
  const [event, setEvent] = useState('')
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(false)
  const [detailLog, setDetailLog] = useState(null)
  const dialogRef = useRef(null)
  const limit = 50

  const fetchLogs = async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams({ offset, limit })
      if (event) params.set('event', event)
      const data = await apiGet(`/api/admin/audit-logs?${params}`)
      setLogs(data.logs || [])
      setTotal(data.total || 0)
    } catch (e) {
      console.error('Failed to fetch audit logs:', e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchLogs() }, [event, offset])

  return (
    <div className="card-modern p-5">
      <div className="flex items-center gap-3 mb-4">
        <select
          className="input-modern input-sm text-sm"
          value={event}
          onChange={e => { setEvent(e.target.value); setOffset(0) }}
        >
          {EVENTS.map(e => <option key={e} value={e}>{e || t('audit.allEvents')}</option>)}
        </select>
        <span className="text-xs text-base-content/40">{t('audit.total', { count: total })}</span>
      </div>
      {loading ? (
        <div className="flex justify-center py-8"><span className="loading loading-spinner text-primary"></span></div>
      ) : (
        <div className="overflow-x-auto -mx-5">
          <table className="table w-full">
            <thead>
              <tr className="text-xs text-base-content/40">
                <th className="font-medium px-5">{t('audit.time')}</th>
                <th className="font-medium">{t('audit.event')}</th>
                <th className="font-medium">{t('audit.email')}</th>
                <th className="font-medium">{t('audit.detail')}</th>
                <th className="font-medium">{t('audit.ip')}</th>
              </tr>
            </thead>
            <tbody>
              {logs.map(log => (
                <tr key={log.id} className="text-sm hover:bg-base-200/40">
                  <td className="whitespace-nowrap text-xs text-base-content/50 px-5">
                    {new Date(log.created_at).toLocaleString()}
                  </td>
                  <td>
                    <span className={`inline-block rounded-full px-2.5 py-0.5 text-[11px] font-medium ${EVENT_COLORS[log.event] || 'bg-base-200 text-base-content/60'}`}>
                      {log.event}
                    </span>
                  </td>
                  <td className="text-xs">{log.email}</td>
                  <td
                    className="max-w-[200px] truncate text-xs text-base-content/60 cursor-pointer hover:text-primary transition-colors"
                    onClick={() => { setDetailLog(log); dialogRef.current?.showModal() }}
                  >
                    {log.detail}
                  </td>
                  <td className="text-xs font-mono text-base-content/50">{log.ip}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <div className="flex items-center justify-between mt-4 pt-4 border-t border-base-300/40">
        <button
          className="btn-modern btn-sm btn-ghost gap-1"
          disabled={offset === 0}
          onClick={() => setOffset(Math.max(0, offset - limit))}
        >
          <ChevronLeft size={14} />
          {t('audit.previous')}
        </button>
        <span className="text-xs text-base-content/40 tabular-nums">
          {t('audit.pagination', { start: offset + 1, end: Math.min(offset + limit, total), total })}
        </span>
        <button
          className="btn-modern btn-sm btn-ghost gap-1"
          disabled={offset + limit >= total}
          onClick={() => setOffset(offset + limit)}
        >
          {t('audit.next')}
          <ChevronRight size={14} />
        </button>
      </div>

      <dialog ref={dialogRef} className="modal modal-bottom sm:modal-middle">
        <div className="modal-box">
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-bold text-sm">{t('audit.detail')}</h3>
            <form method="dialog">
              <button className="btn btn-sm btn-ghost btn-circle"><X size={16} /></button>
            </form>
          </div>
          {detailLog && (
            <div className="space-y-3">
              <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-base-content/50">
                <span>{new Date(detailLog.created_at).toLocaleString()}</span>
                <span className={`inline-block rounded-full px-2 py-0.5 text-[11px] font-medium ${EVENT_COLORS[detailLog.event] || 'bg-base-200 text-base-content/60'}`}>
                  {detailLog.event}
                </span>
                <span>{detailLog.email}</span>
                <span className="font-mono">{detailLog.ip}</span>
              </div>
              <pre className="whitespace-pre-wrap break-all text-sm text-base-content/80 bg-base-200/60 rounded-lg p-3 max-h-[60vh] overflow-y-auto">
                {detailLog.detail || '—'}
              </pre>
            </div>
          )}
        </div>
        <form method="dialog" className="modal-backdrop">
          <button>close</button>
        </form>
      </dialog>
    </div>
  )
}
