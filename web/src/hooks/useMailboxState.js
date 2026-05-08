import { useReducer } from 'react'

function mergeMailboxMails(existing, incoming) {
  const existingIds = new Set(existing.mails.map(mail => mail.id))
  const freshMails = incoming.filter(mail => !existingIds.has(mail.id))
  const merged = [...existing.mails, ...freshMails]
    .sort((left, right) => new Date(right.created_at) - new Date(left.created_at))

  return {
    mails: merged,
    unreadCount: merged.filter(mail => !mail.is_read).length,
  }
}

function ensureMailbox(mailboxMap, shortId) {
  if (mailboxMap.has(shortId)) {
    return mailboxMap
  }
  const next = new Map(mailboxMap)
  next.set(shortId, { mails: [], unreadCount: 0 })
  return next
}

function removeBlacklistedMailboxes(mailboxMap, blacklist) {
  const next = new Map(mailboxMap)
  let changed = false

  for (const shortId of next.keys()) {
    const lower = shortId.toLowerCase()
    if (blacklist.some(keyword => lower.includes(keyword))) {
      next.delete(shortId)
      changed = true
    }
  }

  return changed ? next : mailboxMap
}

function reducer(state, action) {
  switch (action.type) {
    case 'hydrate_mailboxes': {
      let mailboxMap = state.mailboxMap
      for (const shortId of action.shortIds) {
        mailboxMap = ensureMailbox(mailboxMap, shortId)
      }
      return {
        ...state,
        mailboxMap,
        activeShortId: action.activeShortId || state.activeShortId || action.shortIds[0] || '',
        selectedMail: null,
      }
    }

    case 'activate_mailbox': {
      return {
        ...state,
        mailboxMap: ensureMailbox(state.mailboxMap, action.shortId),
        activeShortId: action.shortId,
        selectedMail: null,
      }
    }

    case 'receive_shortid': {
      return {
        ...state,
        mailboxMap: ensureMailbox(state.mailboxMap, action.shortId),
        activeShortId: action.shortId,
        selectedMail: null,
      }
    }

    case 'merge_mailbox_mails': {
      const next = new Map(state.mailboxMap)
      const existing = next.get(action.shortId) || { mails: [], unreadCount: 0 }
      next.set(action.shortId, mergeMailboxMails(existing, action.mails))
      return {
        ...state,
        mailboxMap: next,
      }
    }

    case 'receive_mail': {
      const next = new Map(state.mailboxMap)
      const existing = next.get(action.shortId) || { mails: [], unreadCount: 0 }
      next.set(action.shortId, {
        mails: [action.mail, ...existing.mails],
        unreadCount: existing.unreadCount + 1,
      })
      return {
        ...state,
        mailboxMap: next,
      }
    }

    case 'remove_mailbox': {
      const next = new Map(state.mailboxMap)
      next.delete(action.shortId)
      const activeShortId = state.activeShortId === action.shortId
        ? (next.keys().next().value || '')
        : state.activeShortId

      return {
        ...state,
        mailboxMap: next,
        activeShortId,
        selectedMail: state.activeShortId === action.shortId ? null : state.selectedMail,
      }
    }

    case 'remove_blacklisted_mailboxes': {
      const mailboxMap = removeBlacklistedMailboxes(state.mailboxMap, action.blacklist)
      if (mailboxMap === state.mailboxMap) {
        return state
      }

      const activeShortId = mailboxMap.has(state.activeShortId)
        ? state.activeShortId
        : (mailboxMap.keys().next().value || '')

      return {
        ...state,
        mailboxMap,
        activeShortId,
        selectedMail: mailboxMap.has(state.activeShortId) ? state.selectedMail : null,
      }
    }

    case 'set_selected_mail':
      return {
        ...state,
        selectedMail: action.mail,
      }

    case 'set_recent_mails':
      return {
        ...state,
        recentMails: action.mails,
      }

    case 'clear_active_mailbox': {
      if (!state.activeShortId) {
        return state
      }
      const next = new Map(state.mailboxMap)
      const existing = next.get(state.activeShortId)
      if (!existing) {
        return state
      }
      next.set(state.activeShortId, { ...existing, mails: [] })
      return {
        ...state,
        mailboxMap: next,
        selectedMail: null,
      }
    }

    case 'mark_mail_read': {
      if (!state.activeShortId) {
        return state
      }

      const next = new Map(state.mailboxMap)
      const existing = next.get(state.activeShortId)
      if (!existing) {
        return state
      }

      const unreadDelta = existing.mails.find(mail => mail.id === action.id && !mail.is_read) ? 1 : 0
      next.set(state.activeShortId, {
        ...existing,
        mails: existing.mails.map(mail => mail.id === action.id ? { ...mail, is_read: true } : mail),
        unreadCount: Math.max(0, existing.unreadCount - unreadDelta),
      })

      return {
        ...state,
        mailboxMap: next,
        selectedMail: state.selectedMail?.id === action.id
          ? { ...state.selectedMail, is_read: true }
          : state.selectedMail,
      }
    }

    default:
      return state
  }
}

const initialState = {
  mailboxMap: new Map(),
  activeShortId: '',
  selectedMail: null,
  recentMails: [],
}

export default function useMailboxState() {
  const [state, dispatch] = useReducer(reducer, initialState)

  return {
    state,
    dispatch,
    tabs: Array.from(state.mailboxMap.entries()).map(([shortId, mailbox]) => ({
      shortId,
      unreadCount: mailbox.unreadCount,
    })),
    mails: state.mailboxMap.get(state.activeShortId)?.mails || [],
  }
}
