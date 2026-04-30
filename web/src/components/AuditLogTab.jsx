import { useState, useEffect } from 'react'
import { apiGet } from '../lib/api'

const EVENTS = ['', 'LOGIN', 'LOGIN_DENIED', 'LOGOUT', 'MAILBOX_CREATE', 'MAIL_RECEIVED', 'MAIL_DROPPED', 'WEBHOOK_SENT', 'CONFIG_CHANGED']

export default function AuditLogTab() {
  const [logs, setLogs] = useState([])
  const [total, setTotal] = useState(0)
  const [event, setEvent] = useState('')
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(false)
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
    <div className="card bg-base-100 shadow-md">
      <div className="card-body">
        <div className="flex gap-2 mb-4">
          <select className="select select-bordered select-sm" value={event} onChange={e => { setEvent(e.target.value); setOffset(0) }}>
            {EVENTS.map(e => <option key={e} value={e}>{e || 'All Events'}</option>)}
          </select>
          <span className="text-sm self-center text-base-content/70">Total: {total}</span>
        </div>
        {loading ? (
          <div className="flex justify-center"><span className="loading loading-spinner"></span></div>
        ) : (
          <div className="overflow-x-auto">
            <table className="table table-xs">
              <thead>
                <tr><th>Time</th><th>Event</th><th>Email</th><th>Detail</th><th>IP</th></tr>
              </thead>
              <tbody>
                {logs.map(log => (
                  <tr key={log.id}>
                    <td className="whitespace-nowrap">{new Date(log.created_at).toLocaleString()}</td>
                    <td><span className="badge badge-sm">{log.event}</span></td>
                    <td>{log.email}</td>
                    <td className="max-w-[200px] truncate">{log.detail}</td>
                    <td>{log.ip}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        <div className="flex justify-between mt-4">
          <button className="btn btn-sm" disabled={offset === 0} onClick={() => setOffset(Math.max(0, offset - limit))}>Previous</button>
          <span className="text-sm self-center">{offset + 1} - {Math.min(offset + limit, total)} of {total}</span>
          <button className="btn btn-sm" disabled={offset + limit >= total} onClick={() => setOffset(offset + limit)}>Next</button>
        </div>
      </div>
    </div>
  )
}
