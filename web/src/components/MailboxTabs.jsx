import { useState, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { useToast } from './Toast'
import { Plus, X, Copy, Pencil, RefreshCw, Check } from 'lucide-react'
import { useCopyToClipboard } from '../hooks/useCopyToClipboard'

const SHORTID_REGEX = /^[a-z0-9._\-+]{1,64}$/

export default function MailboxTabs({ tabs, activeShortId, host, hosts, onSelect, onClose, onAdd, onSetShortId }) {
  const { t } = useTranslation()
  const toast = useToast()
  const { copiedId, copy } = useCopyToClipboard()
  const [editing, setEditing] = useState(false)
  const [editValue, setEditValue] = useState('')
  const [refreshing, setRefreshing] = useState(false)
  const [selectedDomain, setSelectedDomain] = useState(host)
  const inputRef = useRef(null)

  const activeHost = hosts && hosts.length > 1 ? selectedDomain : host
  const address = activeShortId ? `${activeShortId}@${activeHost}` : ''
  const copied = copiedId === address

  const handleCopy = useCallback(async () => {
    if (!address) return
    try {
      await navigator.clipboard.writeText(address)
      copy(address, address)
      toast.success(t('mailbox.copied'))
    } catch {}
  }, [address, copy, toast, t])

  const handleRefresh = useCallback(() => {
    setRefreshing(true)
    onAdd()
    setTimeout(() => setRefreshing(false), 1000)
  }, [onAdd])

  const handleEdit = useCallback(() => {
    setEditValue(activeShortId || '')
    setEditing(true)
    requestAnimationFrame(() => inputRef.current?.focus())
  }, [activeShortId])

  const handleSubmit = useCallback((e) => {
    e.preventDefault()
    const normalized = editValue.trim().toLowerCase()
    if (SHORTID_REGEX.test(normalized)) {
      onSetShortId(normalized)
    }
    setEditing(false)
  }, [editValue, onSetShortId])

  if (tabs.length === 0) return null

  return (
    <div className="sticky top-14 z-10 bg-base-200/90 backdrop-blur-sm -mx-4 sm:-mx-6 px-4 sm:px-6">
      {/* Address bar */}
      <div className="flex items-center gap-2 py-2 border-b border-base-300/40">
        {editing ? (
          <form onSubmit={handleSubmit} className="flex items-center gap-2 flex-1 min-w-0">
            <input
              ref={inputRef}
              type="text"
              className="input-modern input-xs flex-1 font-mono min-w-0"
              value={editValue}
              onChange={e => setEditValue(e.target.value)}
              placeholder={t('mailbox.customShortId')}
              pattern="[a-z0-9._\-+]{1,64}"
            />
            {hosts && hosts.length > 1 ? (
              <select
                className="select select-xs select-ghost font-mono text-xs shrink-0"
                value={selectedDomain}
                onChange={e => setSelectedDomain(e.target.value)}
              >
                {hosts.map(d => <option key={d} value={d}>{d}</option>)}
              </select>
            ) : (
              <span className="text-xs text-base-content/60 font-mono shrink-0">@{activeHost}</span>
            )}
            <button type="submit" className="btn btn-xs btn-primary shrink-0">{t('mailbox.set')}</button>
            <button type="button" className="btn btn-xs btn-ghost shrink-0" onClick={() => setEditing(false)}>{t('mailbox.cancel')}</button>
          </form>
        ) : (
          <>
            <span className="text-sm sm:text-base font-mono font-semibold text-base-content truncate">
              {address || <span className="text-base-content/30 font-sans font-normal">{t('mailbox.connecting')}</span>}
            </span>
            {address && (
              <div className="flex items-center gap-0.5 shrink-0 ml-auto">
                <button
                  className={`btn btn-xs btn-ghost ${copied ? 'text-success' : ''}`}
                  onClick={handleCopy}
                  title="Copy"
                >
                  {copied ? <Check size={13} /> : <Copy size={13} />}
                </button>
                <button className="btn btn-xs btn-ghost" onClick={handleEdit} title="Edit">
                  <Pencil size={13} />
                </button>
                <button className="btn btn-xs btn-ghost" onClick={handleRefresh} title="New address">
                  <RefreshCw size={13} className={refreshing ? 'animate-spin' : ''} />
                </button>
              </div>
            )}
          </>
        )}
      </div>

      {/* Tabs row */}
      <div className="flex items-center gap-1 overflow-x-auto scrollbar-hide py-1.5">
        {tabs.map(({ shortId, unreadCount }) => (
          <div
            key={shortId}
            className={`group flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium cursor-pointer transition-all duration-150 shrink-0 border-b-2 ${
              shortId === activeShortId
                ? 'border-primary bg-primary/5 text-primary'
                : 'border-transparent text-base-content/60 hover:bg-base-300/60'
            }`}
            onClick={() => onSelect(shortId)}
          >
            {unreadCount > 0 && (
              <span className="w-2 h-2 rounded-full bg-red-500 shrink-0" />
            )}
            <span>{shortId}</span>
            {tabs.length > 1 && (
              <button
                className={`ml-0.5 p-0.5 rounded-full opacity-0 group-hover:opacity-100 transition-opacity hover:bg-black/10 ${
                  shortId === activeShortId ? 'text-primary' : 'text-base-content/40'
                }`}
                onClick={(e) => {
                  e.stopPropagation()
                  onClose(shortId)
                }}
                aria-label="Close"
              >
                <X size={12} />
              </button>
            )}
          </div>
        ))}
        <button
          className="flex items-center gap-1 px-2.5 py-1.5 text-xs text-base-content/40 hover:text-base-content/60 hover:bg-base-300/60 transition-colors shrink-0"
          onClick={onAdd}
          title={t('mailboxTabs.add')}
        >
          <Plus size={14} />
        </button>
        {tabs.length > 1 && (
          <button
            className="flex items-center gap-1 ml-auto px-2.5 py-1.5 text-xs text-base-content/40 hover:text-error transition-colors shrink-0"
            onClick={() => tabs.forEach(({ shortId }) => onClose(shortId))}
            title={t('mailboxTabs.closeAll')}
          >
            <X size={14} />
            <span className="hidden sm:inline">{t('mailboxTabs.closeAll')}</span>
          </button>
        )}
      </div>
    </div>
  )
}
