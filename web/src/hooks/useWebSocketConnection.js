import { useRef, useCallback, useEffect } from 'react'

/**
 * Manages WebSocket connection lifecycle with exponential backoff reconnect.
 * @param {function} onMessage - Callback for incoming messages. Receives { type, ... } objects.
 *   A special { type: '_connected' } message is sent on each successful connection.
 * @returns {{ send: function }} - Send data over the WebSocket.
 */
export default function useWebSocketConnection(onMessage) {
  const wsRef = useRef(null)
  const reconnectTimer = useRef(null)
  const delayRef = useRef(1000)
  const onMessageRef = useRef(onMessage)

  useEffect(() => {
    onMessageRef.current = onMessage
  }, [onMessage])

  const send = useCallback((data) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(typeof data === 'string' ? data : JSON.stringify(data))
    }
  }, [])

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws`)
    wsRef.current = ws

    ws.onopen = () => {
      delayRef.current = 1000
      onMessageRef.current({ type: '_connected' })
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        onMessageRef.current(msg)
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
    return () => {
      clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  return { send }
}
