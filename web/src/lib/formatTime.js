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
