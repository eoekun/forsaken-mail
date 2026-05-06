/**
 * Relative time display (e.g. "刚刚", "3分钟前", "2小时前").
 * Always returns a short Chinese relative string.
 */
export function formatRelativeTime(dateStr) {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diffSec = Math.floor((now - then) / 1000)

  if (diffSec < 60) return '刚刚'
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return `${diffMin}分钟前`
  const diffHour = Math.floor(diffMin / 60)
  if (diffHour < 24) return `${diffHour}小时前`
  const diffDay = Math.floor(diffHour / 24)
  if (diffDay < 30) return `${diffDay}天前`
  const diffMonth = Math.floor(diffDay / 30)
  if (diffMonth < 12) return `${diffMonth}个月前`
  return `${Math.floor(diffMonth / 12)}年前`
}

/**
 * Full absolute time string for tooltip / detail view.
 * Returns "YYYY/MM/DD HH:mm:ss".
 */
export function formatAbsoluteTime(dateStr) {
  const d = new Date(dateStr)
  const y = d.getFullYear()
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mi = String(d.getMinutes()).padStart(2, '0')
  const ss = String(d.getSeconds()).padStart(2, '0')
  return `${y}/${mm}/${dd} ${hh}:${mi}:${ss}`
}
