import { useState, useEffect, useRef, useCallback } from 'react'
import i18n from '../i18n'
import { apiGet } from '../lib/api'
import { normalizeMail } from '../lib/normalizeMail'
import { useToast } from '../components/Toast'
import useWebSocketConnection from './useWebSocketConnection'

const TABS_STORAGE_KEY = 'mailbox_tabs_v1'

export default function useWebSocket(host, keywordBlacklist) {
  const [mailboxMap, setMailboxMap] = useState(new Map())
  const [activeShortId, setActiveShortId] = useState('')
  const [selectedMail, setSelectedMail] = useState(null)
  const [recentMails, setRecentMails] = useState([])
  const toast = useToast()
  const activeShortIdRef = useRef(activeShortId)
  const loadRecentMailsRef = useRef(null)
  const hasConnectedRef = useRef(false)
  const blacklistRef = useRef([])

  useEffect(() => {
    activeShortIdRef.current = activeShortId
  }, [activeShortId])

  useEffect(() => {
    if (keywordBlacklist) {
      blacklistRef.current = keywordBlacklist.split(',').map(k => k.trim().toLowerCase()).filter(Boolean)
    } else {
      blacklistRef.current = []
    }
  }, [keywordBlacklist])

  // Derive tabs array from mailboxMap
  const tabs = Array.from(mailboxMap.entries()).map(([shortId, data]) => ({
    shortId,
    unreadCount: data.unreadCount,
  }))

  const mails = mailboxMap.get(activeShortId)?.mails || []

  // Save tabs to localStorage
  useEffect(() => {
    if (!hasConnectedRef.current) return
    const shortIds = Array.from(mailboxMap.keys())
    try {
      localStorage.setItem(TABS_STORAGE_KEY, JSON.stringify(shortIds))
    } catch {}
  }, [mailboxMap])

  // Handle WebSocket messages
  const handleMessage = useCallback((msg) => {
    switch (msg.type) {
      case '_connected': {
        // Restore saved tabs on connect
        let savedTabs = []
        try {
          const raw = localStorage.getItem(TABS_STORAGE_KEY)
          savedTabs = raw ? JSON.parse(raw) : []
          if (!Array.isArray(savedTabs)) savedTabs = []
        } catch {}

        const savedSingle = localStorage.getItem('shortid')
        if (savedTabs.length > 0) {
          const validTabs = savedTabs.filter(id => !blacklistRef.current.some(kw => id.toLowerCase().includes(kw)))
          for (const id of validTabs) {
            send({ type: 'subscribe', short_id: id })
          }
          setActiveShortId(prev => prev && validTabs.includes(prev) ? prev : validTabs[0])
          setMailboxMap(prev => {
            const next = new Map(prev)
            for (const id of validTabs) {
              if (!next.has(id)) next.set(id, { mails: [], unreadCount: 0 })
            }
            return next
          })
          for (const id of validTabs) fetchStoredMails(id, setMailboxMap)
        } else if (savedSingle && !blacklistRef.current.some(kw => savedSingle.toLowerCase().includes(kw))) {
          send({ type: 'subscribe', short_id: savedSingle })
          setActiveShortId(savedSingle)
          setMailboxMap(new Map([[savedSingle, { mails: [], unreadCount: 0 }]]))
          fetchStoredMails(savedSingle, setMailboxMap)
        } else {
          send({ type: 'request_shortid' })
        }
        hasConnectedRef.current = true
        break
      }
      case 'shortid': {
        const id = msg.short_id
        setActiveShortId(id)
        setMailboxMap(prev => {
          const next = new Map(prev)
          if (!next.has(id)) next.set(id, { mails: [], unreadCount: 0 })
          return next
        })
        upsertHistory(id)
        fetchStoredMails(id, setMailboxMap)
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
        if ('Notification' in window && Notification.permission === 'granted') {
          new Notification(i18n.t('notification.newMail', { from: mailData.from }))
        }
        loadRecentMailsRef.current?.()
        break
      }
      case 'error':
        console.error('WS error:', msg.message)
        if (msg.message && msg.message.includes('blacklist')) {
          setMailboxMap(prev => {
            const next = new Map(prev)
            let changed = false
            for (const [id] of next) {
              if (blacklistRef.current.some(kw => id.toLowerCase().includes(kw))) {
                next.delete(id)
                removeFromHistory(id)
                changed = true
              }
            }
            return changed ? next : prev
          })
        }
        break
    }
  }, [])

  const { send } = useWebSocketConnection(handleMessage)

  // Load recent mails on mount
  useEffect(() => {
    loadRecentMails()
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission()
    }
  }, [])

  const subscribeToShortId = useCallback((id) => {
    const lower = id.toLowerCase()
    if (blacklistRef.current.some(kw => lower.includes(kw))) {
      toast.error(i18n.t('mailbox.blacklisted', { id }))
      return
    }
    send({ type: 'subscribe', short_id: id })
    setActiveShortId(id)
    setMailboxMap(prev => {
      const next = new Map(prev)
      if (!next.has(id)) next.set(id, { mails: [], unreadCount: 0 })
      return next
    })
    upsertHistory(id)
    fetchStoredMails(id, setMailboxMap)
  }, [send, toast])

  const unsubscribeFromShortId = useCallback((id) => {
    send({ type: 'unsubscribe', short_id: id })
    setMailboxMap(prev => {
      const next = new Map(prev)
      next.delete(id)
      return next
    })
    setActiveShortId(prev => prev === id ? '' : prev)
    setSelectedMail(null)
  }, [send])

  useEffect(() => {
    if (!activeShortId && mailboxMap.size > 0) {
      setActiveShortId(mailboxMap.keys().next().value)
    }
  }, [activeShortId, mailboxMap])

  const loadRecentMails = useCallback(() => {
    apiGet('/api/mails/recent')
      .then(mails => {
        if (Array.isArray(mails)) setRecentMails(mails)
      })
      .catch(() => {})
  }, [])
  loadRecentMailsRef.current = loadRecentMails

  const requestNewShortId = useCallback(() => {
    send({ type: 'request_shortid' })
  }, [send])

  const clearMails = useCallback(() => {
    const sid = activeShortIdRef.current
    setMailboxMap(prev => {
      const next = new Map(prev)
      const existing = next.get(sid)
      if (existing) next.set(sid, { ...existing, mails: [] })
      return next
    })
    setSelectedMail(null)
  }, [])

  const markMailAsRead = useCallback((id) => {
    const sid = activeShortIdRef.current
    setMailboxMap(prev => {
      const next = new Map(prev)
      const existing = next.get(sid)
      if (existing) {
        next.set(sid, {
          ...existing,
          mails: existing.mails.map(m => m.id === id ? { ...m, is_read: true } : m),
          unreadCount: Math.max(0, existing.unreadCount - (existing.mails.find(m => m.id === id && !m.is_read) ? 1 : 0)),
        })
      }
      return next
    })
    setSelectedMail(prev => prev?.id === id ? { ...prev, is_read: true } : prev)
  }, [])

  return {
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
    recentMails,
    loadRecentMails,
  }
}

function fetchStoredMails(shortId, setMailboxMap) {
  apiGet(`/api/mails?shortId=${encodeURIComponent(shortId)}&reextract=true`)
    .then(mails => {
      if (!Array.isArray(mails) || mails.length === 0) return
      setMailboxMap(prev => {
        const next = new Map(prev)
        const existing = next.get(shortId) || { mails: [], unreadCount: 0 }
        const existingIds = new Set(existing.mails.map(m => m.id))
        const newMails = mails.filter(m => !existingIds.has(m.id)).map(normalizeMail)
        if (newMails.length === 0) return prev
        const merged = [...existing.mails, ...newMails].sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
        const unreadCount = merged.filter(m => !m.is_read).length
        next.set(shortId, { mails: merged, unreadCount })
        return next
      })
    })
    .catch(() => {})
}

function removeFromHistory(shortId) {
  try {
    const raw = localStorage.getItem(TABS_STORAGE_KEY)
    let list = raw ? JSON.parse(raw) : []
    if (!Array.isArray(list)) list = []
    list = list.filter(id => id !== shortId)
    localStorage.setItem(TABS_STORAGE_KEY, JSON.stringify(list))

    const raw2 = localStorage.getItem('shortid_history_v1')
    let list2 = raw2 ? JSON.parse(raw2) : []
    if (!Array.isArray(list2)) list2 = []
    list2 = list2.filter(id => id !== shortId)
    localStorage.setItem('shortid_history_v1', JSON.stringify(list2))
  } catch {}
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
