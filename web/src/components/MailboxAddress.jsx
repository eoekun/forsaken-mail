import { useState, useRef } from 'react'

const SHORTID_REGEX = /^[a-z0-9._\-+]{1,64}$/

export default function MailboxAddress({ shortId, host, onRefresh, onSetShortId }) {
  const [editing, setEditing] = useState(false)
  const [editValue, setEditValue] = useState('')
  const [copied, setCopied] = useState(false)
  const inputRef = useRef(null)

  const address = shortId ? `${shortId}@${host}` : ''

  const handleCopy = async () => {
    if (!address) return
    try {
      await navigator.clipboard.writeText(address)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {}
  }

  const handleEdit = () => {
    setEditValue(shortId || '')
    setEditing(true)
    setTimeout(() => inputRef.current?.focus(), 0)
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
    <div className="card bg-base-100 shadow-md p-4 mt-4">
      <div className="flex items-center gap-2 flex-wrap">
        <span className="text-sm font-medium">Your temp email:</span>
        {editing ? (
          <form onSubmit={handleSubmit} className="flex items-center gap-2 flex-1">
            <input
              ref={inputRef}
              type="text"
              className="input input-bordered input-sm flex-1"
              value={editValue}
              onChange={e => setEditValue(e.target.value)}
              placeholder="Enter custom short ID"
              pattern="[a-z0-9._\-+]{1,64}"
            />
            <span className="text-sm">@{host}</span>
            <button type="submit" className="btn btn-sm btn-primary">Set</button>
            <button type="button" className="btn btn-sm btn-ghost" onClick={() => setEditing(false)}>Cancel</button>
          </form>
        ) : (
          <>
            <span className="text-lg font-mono">{address || 'Connecting...'}</span>
            <button className="btn btn-sm btn-ghost" onClick={handleCopy} title="Copy">
              {copied ? '✓' : '📋'}
            </button>
            <button className="btn btn-sm btn-ghost" onClick={handleEdit} title="Edit">✏️</button>
            <button className="btn btn-sm btn-ghost" onClick={onRefresh} title="New address">🔄</button>
          </>
        )}
      </div>
    </div>
  )
}
