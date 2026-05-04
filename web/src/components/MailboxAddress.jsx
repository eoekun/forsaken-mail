import { useState, useRef, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useToast } from './Toast'
import { Copy, Pencil, RefreshCw, Check } from 'lucide-react'

const SHORTID_REGEX = /^[a-z0-9._\-+]{1,64}$/

export default function MailboxAddress({ shortId, host, onRefresh, onSetShortId }) {
  const [editing, setEditing] = useState(false)
  const [editValue, setEditValue] = useState('')
  const [copied, setCopied] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const inputRef = useRef(null)
  const copiedTimerRef = useRef(null)
  const { t } = useTranslation()
  const toast = useToast()

  const address = shortId ? `${shortId}@${host}` : ''

  useEffect(() => {
    return () => clearTimeout(copiedTimerRef.current)
  }, [])

  const handleCopy = async () => {
    if (!address) return
    try {
      await navigator.clipboard.writeText(address)
      setCopied(true)
      toast.success(t('mailbox.copied'))
      clearTimeout(copiedTimerRef.current)
      copiedTimerRef.current = setTimeout(() => setCopied(false), 2000)
    } catch {}
  }

  const handleRefresh = () => {
    setRefreshing(true)
    onRefresh()
    setTimeout(() => setRefreshing(false), 1000)
  }

  const handleEdit = () => {
    setEditValue(shortId || '')
    setEditing(true)
    requestAnimationFrame(() => inputRef.current?.focus())
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    const normalized = editValue.trim().toLowerCase()
    if (SHORTID_REGEX.test(normalized)) {
      onSetShortId(normalized)
    }
    setEditing(false)
  }

  return (
    <div className="card-modern p-5 mt-4">
      <div className="text-xs font-medium text-base-content/40 uppercase tracking-wider mb-2">
        {t('mailbox.yourTempEmail')}
      </div>
      {editing ? (
        <form onSubmit={handleSubmit} className="flex items-center gap-2">
          <input
            ref={inputRef}
            type="text"
            className="input-modern input-sm flex-1 font-mono"
            value={editValue}
            onChange={e => setEditValue(e.target.value)}
            placeholder={t('mailbox.customShortId')}
            pattern="[a-z0-9._\-+]{1,64}"
          />
          <span className="text-sm text-base-content/60 font-mono">@{host}</span>
          <button type="submit" className="btn-modern btn-sm btn-primary">{t('mailbox.set')}</button>
          <button type="button" className="btn-modern btn-sm btn-ghost" onClick={() => setEditing(false)}>{t('mailbox.cancel')}</button>
        </form>
      ) : (
        <div className="bg-primary/5 border border-primary/20 rounded-lg p-4">
          <div className="flex items-center gap-2 flex-wrap">
            {address ? (
              <span className="text-2xl sm:text-3xl font-mono font-semibold text-base-content tracking-wide break-all">
                {address}
              </span>
            ) : (
              <span className="text-lg text-base-content/30">{t('mailbox.connecting')}</span>
            )}
            {address && (
              <button
                className={`btn btn-sm btn-ghost transition-colors ${copied ? 'text-success' : ''}`}
                onClick={handleCopy}
                title="Copy"
              >
                {copied ? <Check size={16} /> : <Copy size={16} />}
                {copied && <span className="text-xs">{t('mailbox.copied')}</span>}
              </button>
            )}
          </div>
          <div className="flex items-center gap-1 mt-2">
            <button className="btn btn-sm btn-ghost" onClick={handleEdit} title="Edit">
              <Pencil size={15} />
            </button>
            <button className="btn btn-sm btn-ghost" onClick={handleRefresh} title="New address">
              <RefreshCw size={15} className={refreshing ? 'animate-spin' : ''} />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
