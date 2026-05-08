const TABS_STORAGE_KEY = 'mailbox_tabs_v1'
const SHORTID_STORAGE_KEY = 'shortid'
const HISTORY_STORAGE_KEY = 'shortid_history_v1'

function readList(key) {
  try {
    const raw = localStorage.getItem(key)
    const list = raw ? JSON.parse(raw) : []
    return Array.isArray(list) ? list : []
  } catch {
    return []
  }
}

export function loadSavedTabs() {
  return readList(TABS_STORAGE_KEY)
}

export function saveTabs(shortIds) {
  try {
    localStorage.setItem(TABS_STORAGE_KEY, JSON.stringify(shortIds))
  } catch {}
}

export function loadLegacyShortId() {
  return localStorage.getItem(SHORTID_STORAGE_KEY) || ''
}

export function removeFromHistory(shortId) {
  try {
    const tabs = readList(TABS_STORAGE_KEY).filter(id => id !== shortId)
    localStorage.setItem(TABS_STORAGE_KEY, JSON.stringify(tabs))

    const history = readList(HISTORY_STORAGE_KEY).filter(id => id !== shortId)
    localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(history))
  } catch {}
}

export function upsertHistory(shortId) {
  try {
    let list = readList(HISTORY_STORAGE_KEY).filter(id => id !== shortId)
    list.unshift(shortId)
    if (list.length > 6) {
      list = list.slice(0, 6)
    }
    localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(list))
  } catch {}
}
