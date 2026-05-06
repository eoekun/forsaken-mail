import { useState, useRef, useEffect, useCallback } from 'react'

export function useCopyToClipboard(resetDelay = 2000) {
  const [copiedId, setCopiedId] = useState(null)
  const timerRef = useRef(null)

  useEffect(() => () => clearTimeout(timerRef.current), [])

  const copy = useCallback((text, id) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopiedId(id)
      clearTimeout(timerRef.current)
      timerRef.current = setTimeout(() => setCopiedId(null), resetDelay)
    }).catch(() => {})
  }, [resetDelay])

  return { copiedId, copy }
}
