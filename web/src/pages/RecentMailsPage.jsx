import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Search, Mail, ArrowLeft, KeyRound, Check, X, ChevronLeft, ChevronRight } from 'lucide-react'
import { apiGet } from '../lib/api'
import { formatSender } from '../lib/formatSender'
import { formatRelativeTime, formatAbsoluteTime } from '../lib/formatTime'
import { useCopyToClipboard } from '../hooks/useCopyToClipboard'
import useRelativeTime from '../hooks/useRelativeTime'
import MailDetail from '../components/MailDetail'

const PAGE_SIZE = 20

export default function RecentMailsPage() {
  const { t } = useTranslation()
  const { copiedId, copy } = useCopyToClipboard()
  useRelativeTime()

  const [emails, setEmails] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [query, setQuery] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [selectedMail, setSelectedMail] = useState(null)

  const fetchMails = useCallback(async (p, q) => {
    setLoading(true)
    try {
      const params = new URLSearchParams({ page: String(p), pageSize: String(PAGE_SIZE) })
      if (q) params.set('q', q)
      const data = await apiGet(`/api/mails/all?${params}`)
      setEmails(data.emails || [])
      setTotal(data.total || 0)
    } catch {} finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { fetchMails(page, query) }, [page, query, fetchMails])

  const handleSearch = useCallback((e) => {
    e.preventDefault()
    setPage(1)
    setQuery(searchInput.trim())
  }, [searchInput])

  const handleClearSearch = useCallback(() => {
    setSearchInput('')
    setQuery('')
    setPage(1)
  }, [])

  const handleCopyCode = useCallback((e, mail) => {
    e.stopPropagation()
    const code = mail.extracted_codes?.[0]
    if (!code) return
    copy(code, mail.id)
  }, [copy])

  const totalPages = Math.ceil(total / PAGE_SIZE)

  return (
    <div className="min-h-screen bg-base-200">
      <div className="max-w-5xl w-full mx-auto px-4 sm:px-6 py-4">
        {/* Header */}
        <div className="flex items-center gap-3 mb-4">
          <Link to="/" className="btn btn-ghost btn-sm btn-circle">
            <ArrowLeft size={18} />
          </Link>
          <Mail size={20} className="text-primary" />
          <h1 className="text-lg font-semibold">{t('recentMails.pageTitle')}</h1>
          <span className="text-sm text-base-content/40">{total}</span>
        </div>

        {/* Search */}
        <form onSubmit={handleSearch} className="mb-4">
          <div className="relative">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-base-content/30" />
            <input
              type="text"
              className="input-modern input-sm w-full pl-10 pr-20"
              placeholder={t('recentMails.searchPlaceholder')}
              value={searchInput}
              onChange={e => setSearchInput(e.target.value)}
            />
            <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
              {searchInput && (
                <button type="button" className="btn btn-xs btn-ghost btn-circle" onClick={handleClearSearch}>
                  <X size={14} />
                </button>
              )}
              <button type="submit" className="btn btn-xs btn-primary">{t('recentMails.search')}</button>
            </div>
          </div>
        </form>

        <div className="flex gap-4 min-h-0">
          {/* List */}
          <div className={`${selectedMail ? 'hidden lg:block lg:w-2/5' : 'w-full'} card-modern flex flex-col max-h-[calc(100vh-180px)]`}>
            <div className="flex-1 overflow-y-auto min-h-0">
              {loading && emails.length === 0 ? (
                <div className="flex items-center justify-center py-16">
                  <span className="loading loading-spinner text-primary"></span>
                </div>
              ) : emails.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-16">
                  <Mail size={48} className="text-base-content/20 mb-2 opacity-30" />
                  <p className="text-sm text-base-content/30">{query ? t('recentMails.noResults') : t('recentMails.empty')}</p>
                </div>
              ) : (
                <div className="divide-y divide-base-300/40">
                  {emails.map((mail) => {
                    const hasCode = (mail.extracted_codes?.length || 0) > 0
                    const isSelected = selectedMail?.id === mail.id
                    return (
                      <div
                        key={mail.id}
                        className={`px-4 py-3 cursor-pointer transition-all duration-150 ${
                          isSelected
                            ? 'bg-primary/5 border-l-3 border-l-primary'
                            : 'border-l-3 border-l-transparent hover:bg-base-200'
                        }`}
                        onClick={() => setSelectedMail(mail)}
                      >
                        <div className="flex items-baseline justify-between gap-2 mb-0.5">
                          <span className={`text-sm truncate ${isSelected ? 'font-semibold text-base-content' : 'text-base-content/70'}`}>
                            {formatSender(mail.from_addr || mail.from)}
                          </span>
                          <span className="text-[11px] text-base-content/30 shrink-0" title={formatAbsoluteTime(mail.created_at)}>
                            {formatRelativeTime(mail.created_at)}
                          </span>
                        </div>
                        <div className="flex items-center gap-2">
                          <span className="text-[11px] font-mono text-primary/60 bg-primary/5 px-1 py-0.5 rounded shrink-0">
                            {mail.short_id}
                          </span>
                          <p className="text-xs text-base-content/50 truncate flex-1">{mail.subject || t('mailList.noSubject')}</p>
                          {hasCode && (
                            <button
                              onClick={(e) => handleCopyCode(e, mail)}
                              className={`shrink-0 inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] font-mono font-semibold transition-all cursor-pointer ${
                                copiedId === mail.id
                                  ? 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400'
                                  : 'bg-primary/10 text-primary hover:bg-primary/20'
                              }`}
                              title={t('mailDetail.clickToCopy')}
                            >
                              {copiedId === mail.id ? <Check size={10} /> : <KeyRound size={10} />}
                              <span className="tracking-wider">{mail.extracted_codes[0]}</span>
                            </button>
                          )}
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-between px-4 py-2 border-t border-base-300/40 shrink-0">
                <button
                  className="btn btn-xs btn-ghost"
                  disabled={page <= 1}
                  onClick={() => setPage(p => p - 1)}
                >
                  <ChevronLeft size={14} />
                </button>
                <span className="text-xs text-base-content/50">{page} / {totalPages}</span>
                <button
                  className="btn btn-xs btn-ghost"
                  disabled={page >= totalPages}
                  onClick={() => setPage(p => p + 1)}
                >
                  <ChevronRight size={14} />
                </button>
              </div>
            )}
          </div>

          {/* Detail */}
          {selectedMail && (
            <div className={`${selectedMail ? 'w-full lg:w-3/5' : 'hidden'} min-h-0`}>
              <MailDetail
                mail={selectedMail}
                onMailRead={(id) => {
                  setSelectedMail(prev => prev?.id === id ? { ...prev, is_read: true } : prev)
                  setEmails(prev => prev.map(m => m.id === id ? { ...m, is_read: true } : m))
                }}
                onBack={() => setSelectedMail(null)}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
