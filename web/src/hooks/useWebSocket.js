import { useState, useEffect, useRef, useCallback } from 'react'
import i18n from '../i18n'

const TABS_STORAGE_KEY = 'mailbox_tabs_v1'

export default function useWebSocket(host) {
  // Map<shortId, {mails: [], unreadCount: number}>
  const [mailboxMap, setMailboxMap] = useState(new Map())
  const [activeShortId, setActiveShortId] = useState('')
  const [selectedMail, setSelectedMail] = useState(null)
  const wsRef = useRef(null)
  const reconnectTimer = useRef(null)
  const delayRef = useRef(1000)
  const activeShortIdRef = useRef(activeShortId)

  useEffect(() => {
    activeShortIdRef.current = activeShortId
  }, [activeShortId])

  // Derive tabs array from mailboxMap
  const tabs = Array.from(mailboxMap.entries()).map(([shortId, data]) => ({
    shortId,
    unreadCount: data.unreadCount,
  }))

  // Get mails for the active tab
  const mails = mailboxMap.get(activeShortId)?.mails || []

  // Save tabs to localStorage whenever they change
  useEffect(() => {
    const shortIds = Array.from(mailboxMap.keys())
    try {
      localStorage.setItem(TABS_STORAGE_KEY, JSON.stringify(shortIds))
    } catch {}
  }, [mailboxMap])

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws`)
    wsRef.current = ws

    ws.onopen = () => {
      delayRef.current = 1000
      // Restore saved tabs
      let savedTabs = []
      try {
        const raw = localStorage.getItem(TABS_STORAGE_KEY)
        savedTabs = raw ? JSON.parse(raw) : []
        if (!Array.isArray(savedTabs)) savedTabs = []
      } catch {}

      const savedSingle = localStorage.getItem('shortid')
      if (savedTabs.length > 0) {
        // Subscribe to all saved tabs
        for (const id of savedTabs) {
          ws.send(JSON.stringify({ type: 'subscribe', short_id: id }))
        }
        setActiveShortId(prev => prev && savedTabs.includes(prev) ? prev : savedTabs[0])
        // Initialize mailboxMap for saved tabs
        setMailboxMap(prev => {
          const next = new Map(prev)
          for (const id of savedTabs) {
            if (!next.has(id)) {
              next.set(id, { mails: [], unreadCount: 0 })
            }
          }
          return next
        })
      } else if (savedSingle) {
        ws.send(JSON.stringify({ type: 'subscribe', short_id: savedSingle }))
        setActiveShortId(savedSingle)
        setMailboxMap(new Map([[savedSingle, { mails: [], unreadCount: 0 }]]))
      } else {
        ws.send(JSON.stringify({ type: 'request_shortid' }))
      }
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        switch (msg.type) {
          case 'shortid': {
            const id = msg.short_id
            setActiveShortId(id)
            setMailboxMap(prev => {
              const next = new Map(prev)
              if (!next.has(id)) {
                next.set(id, { mails: [], unreadCount: 0 })
              }
              return next
            })
            upsertHistory(id)
            break
          }
          case 'mail': {
            const mailData = msg.data
            const targetId = msg.short_id || activeShortIdRef.current
            setMailboxMap(prev => {
              const next = new Map(prev)
              const existing = next.get(targetId) || { mails: [], unreadCount: 0 }
              next.set(targetId, {
                mails: [mailData, ...existing.mails],
                unreadCount: existing.unreadCount + 1,
              })
              return next
            })
            // Browser notification
            if ('Notification' in window && Notification.permission === 'granted') {
              new Notification(i18n.t('notification.newMail', { from: mailData.from }))
            }
            break
          }
          case 'error':
            console.error('WS error:', msg.message)
            break
        }
      } catch (e) {
        console.error('Failed to parse WS message:', e)
      }
    }

    ws.onclose = () => {
      reconnectTimer.current = setTimeout(() => {
        delayRef.current = Math.min(delayRef.current * 2, 30000)
        connect()
      }, delayRef.current)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    connect()
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission()
    }
    return () => {
      clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  const subscribeToShortId = useCallback((id) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'subscribe', short_id: id }))
    }
    setActiveShortId(id)
    setMailboxMap(prev => {
      const next = new Map(prev)
      if (!next.has(id)) {
        next.set(id, { mails: [], unreadCount: 0 })
      }
      return next
    })
    upsertHistory(id)
  }, [])

  const unsubscribeFromShortId = useCallback((id) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'unsubscribe', short_id: id }))
    }
    setMailboxMap(prev => {
      const next = new Map(prev)
      next.delete(id)
      return next
    })
    // If we closed the active tab, switch to the first remaining
    setActiveShortId(prev => {
      if (prev !== id) return prev
      // Can't read new map here, so we'll fix in useEffect
      return ''
    })
    setSelectedMail(null)
  }, [])

  // Fix activeShortId after tab close
  useEffect(() => {
    if (!activeShortId && mailboxMap.size > 0) {
      setActiveShortId(mailboxMap.keys().next().value)
    }
  }, [activeShortId, mailboxMap])

  const requestNewShortId = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'request_shortid' }))
    }
  }, [])

  const clearMails = useCallback(() => {
    setMailboxMap(prev => {
      const next = new Map(prev)
      const existing = next.get(activeShortId)
      if (existing) {
        next.set(activeShortId, { ...existing, mails: [] })
      }
      return next
    })
    setSelectedMail(null)
  }, [activeShortId])

  const markMailAsRead = useCallback((id) => {
    setMailboxMap(prev => {
      const next = new Map(prev)
      const existing = next.get(activeShortId)
      if (existing) {
        next.set(activeShortId, {
          ...existing,
          mails: existing.mails.map(m => m.id === id ? { ...m, is_read: true } : m),
          unreadCount: Math.max(0, existing.unreadCount - (existing.mails.find(m => m.id === id && !m.is_read) ? 1 : 0)),
        })
      }
      return next
    })
    setSelectedMail(prev => prev?.id === id ? { ...prev, is_read: true } : prev)
  }, [activeShortId])

  return {
    shortId: activeShortId,
    tabs,
    activeShortId,
    setActiveShortId: subscribeToShortId,
    subscribeToShortId,
    unsubscribeFromShortId,
    requestNewShortId,
    mails,
    selectedMail,
    setSelectedMail,
    clearMails,
    markMailAsRead,
  }
}

function upsertHistory(shortId) {
  try {
    const raw = localStorage.getItem('shortid_history_v1')
    let list = raw ? JSON.parse(raw) : []
    if (!Array.isArray(list)) list = []
    list = list.filter(id => id !== shortId)
    list.unshift(shortId)
    if (list.length > 6) list = list.slice(0, 6)
    localStorage.setItem('shortid_history_v1', JSON.stringify(list))
  } catch {}
}
