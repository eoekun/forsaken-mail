import { useState, useEffect } from 'react'

/**
 * Triggers a re-render every `intervalMs` so relative time strings
 * (e.g. "3分钟前") stay up to date while the page is open.
 */
export default function useRelativeTime(intervalMs = 30000) {
  const [, setTick] = useState(0)

  useEffect(() => {
    const id = setInterval(() => setTick(t => t + 1), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs])
}
