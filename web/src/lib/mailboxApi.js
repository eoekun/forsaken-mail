import { apiGet } from './api'
import { normalizeMail } from './normalizeMail'

export async function fetchMailboxMails(shortId) {
  const mails = await apiGet(`/api/mails?shortId=${encodeURIComponent(shortId)}`)
  if (!Array.isArray(mails)) {
    return []
  }
  return mails.map(normalizeMail)
}

export async function fetchRecentMails() {
  const mails = await apiGet('/api/mails/recent')
  return Array.isArray(mails) ? mails : []
}
