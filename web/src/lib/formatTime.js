export function formatMailTime(dateStr, t) {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diffMs = now - then
  const diffMin = Math.floor(diffMs / 60000)
  const diffHour = Math.floor(diffMs / 3600000)

  if (diffMin < 1) return t('mailList.timeJustNow')
  if (diffHour < 1) return t('mailList.timeMinutesAgo', { count: diffMin })
  if (diffHour < 24) return new Date(dateStr).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  return new Date(dateStr).toLocaleString([], { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

/** Fixed-width time format (MM/DD HH:mm) for consistent column alignment. */
export function formatMailTimeFixed(dateStr) {
  const d = new Date(dateStr)
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mi = String(d.getMinutes()).padStart(2, '0')
  return `${mm}/${dd} ${hh}:${mi}`
}
