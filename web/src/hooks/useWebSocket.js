import { useEffect, useRef, useCallback } from 'react'
import i18n from '../i18n'
import { fetchMailboxMails, fetchRecentMails } from '../lib/mailboxApi'
import { loadLegacyShortId, loadSavedTabs, removeFromHistory, saveTabs, upsertHistory } from '../lib/mailboxStorage'
import { notifyNewMail, requestNotificationPermission } from '../lib/mailboxNotifications'
import { normalizeMail } from '../lib/normalizeMail'
import { useToast } from '../components/Toast'
import useMailboxState from './useMailboxState'
import useWebSocketConnection from './useWebSocketConnection'

function normalizeBlacklist(keywordBlacklist) {
  if (Array.isArray(keywordBlacklist)) {
    return keywordBlacklist.map(keyword => keyword.trim().toLowerCase()).filter(Boolean)
  }
  if (typeof keywordBlacklist === 'string') {
    return keywordBlacklist.split(',').map(keyword => keyword.trim().toLowerCase()).filter(Boolean)
  }
  return []
}

function normalizeShortIdValue(shortId) {
  return String(shortId || '').trim().toLowerCase()
}

function isBlacklisted(shortId, blacklist) {
  const lower = normalizeShortIdValue(shortId)
  return blacklist.some(keyword => lower.includes(keyword))
}

export default function useWebSocket(keywordBlacklist) {
  const { state, dispatch, tabs, mails } = useMailboxState()
  const toast = useToast()
  const activeShortIdRef = useRef(state.activeShortId)
  const sendRef = useRef(() => {})
  const loadRecentMailsRef = useRef(null)
  const hasConnectedRef = useRef(false)
  const blacklistRef = useRef([])

  useEffect(() => {
    activeShortIdRef.current = state.activeShortId
  }, [state.activeShortId])

  useEffect(() => {
    blacklistRef.current = normalizeBlacklist(keywordBlacklist)
  }, [keywordBlacklist])

  useEffect(() => {
    if (!hasConnectedRef.current) return
    saveTabs(Array.from(state.mailboxMap.keys()))
  }, [state.mailboxMap])

  const loadMailbox = useCallback(async (shortId) => {
    try {
      const storedMails = await fetchMailboxMails(shortId)
      dispatch({ type: 'merge_mailbox_mails', shortId, mails: storedMails })
    } catch {}
  }, [dispatch])

  const loadRecentMails = useCallback(async () => {
    try {
      const recentMails = await fetchRecentMails()
      dispatch({ type: 'set_recent_mails', mails: recentMails })
    } catch {}
  }, [dispatch])
  loadRecentMailsRef.current = loadRecentMails

  const handleMessage = useCallback((msg) => {
    switch (msg.type) {
      case '_connected': {
        const savedTabs = loadSavedTabs().filter(shortId => !isBlacklisted(shortId, blacklistRef.current))
        const savedSingle = loadLegacyShortId()

        if (savedTabs.length > 0) {
          for (const shortId of savedTabs) {
            sendRef.current({ type: 'subscribe', short_id: shortId })
            void loadMailbox(shortId)
          }
          dispatch({
            type: 'hydrate_mailboxes',
            shortIds: savedTabs,
            activeShortId: savedTabs.includes(activeShortIdRef.current)
              ? activeShortIdRef.current
              : savedTabs[0],
          })
        } else if (savedSingle && !isBlacklisted(savedSingle, blacklistRef.current)) {
          sendRef.current({ type: 'subscribe', short_id: savedSingle })
          dispatch({ type: 'hydrate_mailboxes', shortIds: [savedSingle], activeShortId: savedSingle })
          void loadMailbox(savedSingle)
        } else {
          sendRef.current({ type: 'request_shortid' })
        }
        hasConnectedRef.current = true
        break
      }

      case 'shortid': {
        dispatch({ type: 'receive_shortid', shortId: msg.short_id })
        upsertHistory(msg.short_id)
        void loadMailbox(msg.short_id)
        break
      }

      case 'mail': {
        const mailData = normalizeMail(msg.data)
        dispatch({
          type: 'receive_mail',
          shortId: msg.short_id || activeShortIdRef.current,
          mail: mailData,
        })
        notifyNewMail(mailData)
        loadRecentMailsRef.current?.()
        break
      }

      case 'error':
        console.error('WS error:', msg.message)
        if (msg.message && msg.message.includes('blacklist')) {
          for (const shortId of Array.from(state.mailboxMap.keys())) {
            if (isBlacklisted(shortId, blacklistRef.current)) {
              removeFromHistory(shortId)
            }
          }
          dispatch({ type: 'remove_blacklisted_mailboxes', blacklist: blacklistRef.current })
        }
        break
    }
  }, [dispatch, loadMailbox, state.mailboxMap])

  const { send } = useWebSocketConnection(handleMessage)
  useEffect(() => {
    sendRef.current = send
  }, [send])

  useEffect(() => {
    void loadRecentMails()
    requestNotificationPermission()
  }, [loadRecentMails])

  const subscribeToShortId = useCallback((id) => {
    const normalized = normalizeShortIdValue(id)
    if (!normalized) {
      return
    }
    if (isBlacklisted(normalized, blacklistRef.current)) {
      toast.error(i18n.t('mailbox.blacklisted', { id: normalized }))
      return
    }
    sendRef.current({ type: 'subscribe', short_id: normalized })
    dispatch({ type: 'activate_mailbox', shortId: normalized })
    upsertHistory(normalized)
    void loadMailbox(normalized)
  }, [dispatch, loadMailbox, toast])

  const unsubscribeFromShortId = useCallback((id) => {
    sendRef.current({ type: 'unsubscribe', short_id: id })
    dispatch({ type: 'remove_mailbox', shortId: id })
  }, [dispatch])

  const requestNewShortId = useCallback(() => {
    sendRef.current({ type: 'request_shortid' })
  }, [])

  const clearMails = useCallback(() => {
    dispatch({ type: 'clear_active_mailbox' })
  }, [dispatch])

  const markMailAsRead = useCallback((id) => {
    dispatch({ type: 'mark_mail_read', id })
  }, [dispatch])

  const setSelectedMail = useCallback((selectedMail) => {
    dispatch({ type: 'set_selected_mail', mail: selectedMail })
  }, [dispatch])

  return {
    tabs,
    activeShortId: state.activeShortId,
    setActiveShortId: subscribeToShortId,
    subscribeToShortId,
    unsubscribeFromShortId,
    requestNewShortId,
    mails,
    selectedMail: state.selectedMail,
    setSelectedMail,
    clearMails,
    markMailAsRead,
    recentMails: state.recentMails,
    loadRecentMails,
  }
}
