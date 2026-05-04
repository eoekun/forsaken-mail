import { useState, useEffect, useRef, useCallback } from 'react'
import i18n from '../i18n'

export default function useWebSocket(host) {
  const [shortId, setShortId] = useState('')
  const [mails, setMails] = useState([])
  const [selectedMail, setSelectedMail] = useState(null)
  const wsRef = useRef(null)
  const reconnectTimer = useRef(null)
  const delayRef = useRef(1000)

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws`)
    wsRef.current = ws

    ws.onopen = () => {
      delayRef.current = 1000
      const saved = localStorage.getItem('shortid')
      if (saved) {
        ws.send(JSON.stringify({ type: 'set_shortid', short_id: saved }))
      } else {
        ws.send(JSON.stringify({ type: 'request_shortid' }))
      }
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        switch (msg.type) {
          case 'shortid':
            setShortId(msg.short_id)
            localStorage.setItem('shortid', msg.short_id)
            upsertHistory(msg.short_id)
            break
          case 'mail':
            setMails(prev => [msg.data, ...prev])
            // Browser notification
            if ('Notification' in window && Notification.permission === 'granted') {
              new Notification(i18n.t('notification.newMail', { from: msg.data.from }))
            }
            break
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
  }, [])

  useEffect(() => {
    connect()
    // Request notification permission
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission()
    }
    return () => {
      clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  const sendShortId = useCallback((id) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'set_shortid', short_id: id }))
      setMails([])
      setSelectedMail(null)
    }
  }, [])

  const requestNewShortId = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'request_shortid' }))
      setMails([])
      setSelectedMail(null)
    }
  }, [])

  const clearMails = useCallback(() => {
    setMails([])
    setSelectedMail(null)
  }, [])

  const markMailAsRead = useCallback((id) => {
    setMails(prev => prev.map(m => m.id === id ? { ...m, is_read: true } : m))
    setSelectedMail(prev => prev?.id === id ? { ...prev, is_read: true } : prev)
  }, [])

  return {
    shortId,
    setShortId: sendShortId,
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
