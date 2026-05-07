import i18n from '../i18n'

/**
 * Relative time display (e.g. "just now", "3m ago", "2h ago").
 * Uses i18n for localized strings.
 */
export function formatRelativeTime(dateStr) {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diffSec = Math.floor((now - then) / 1000)

  const isZh = i18n.language === 'zh'

  if (diffSec < 60) return isZh ? '刚刚' : 'just now'
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return isZh ? `${diffMin}分钟前` : `${diffMin}m ago`
  const diffHour = Math.floor(diffMin / 60)
  if (diffHour < 24) return isZh ? `${diffHour}小时前` : `${diffHour}h ago`
  const diffDay = Math.floor(diffHour / 24)
  if (diffDay < 30) return isZh ? `${diffDay}天前` : `${diffDay}d ago`
  const diffMonth = Math.floor(diffDay / 30)
  if (diffMonth < 12) return isZh ? `${diffMonth}个月前` : `${diffMonth}mo ago`
  const years = Math.floor(diffMonth / 12)
  return isZh ? `${years}年前` : `${years}y ago`
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
