import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Search, Mail, ArrowLeft, X, ChevronLeft, ChevronRight } from 'lucide-react'
import { apiGet } from '../lib/api'
import { formatSender } from '../lib/formatSender'
import useRelativeTime from '../hooks/useRelativeTime'
import { normalizeMail } from '../lib/normalizeMail'
import MailDetail from '../components/MailDetail'
import MailListItem from '../components/MailListItem'

const PAGE_SIZE = 20

export default function RecentMailsPage() {
  const { t } = useTranslation()
  useRelativeTime()

  const [emails, setEmails] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [selectedMail, setSelectedMail] = useState(null)

  // Filter state
  const [filterShortId, setFilterShortId] = useState('')
  const [filterFrom, setFilterFrom] = useState('')
  const [filterQuery, setFilterQuery] = useState('')
  const [searchInput, setSearchInput] = useState('')

  // Filter options
  const [senders, setSenders] = useState([])
  const [recipients, setRecipients] = useState([])

  // Fetch filter options on mount
  useEffect(() => {
    apiGet('/api/mails/filters').then(data => {
      // Deduplicate senders by formatted label
      const raw = data.senders || []
      const labelMap = new Map()
      for (const s of raw) {
        const label = formatSender(s)
        if (!labelMap.has(label)) {
          labelMap.set(label, s)
        }
      }
      setSenders(Array.from(labelMap.entries()).map(([label]) => ({ label, value: label })))
      setRecipients(data.recipients || [])
    }).catch(() => {})
  }, [])

  const fetchMails = useCallback(async (p, shortId, from, q) => {
    setLoading(true)
    try {
      const params = new URLSearchParams({ page: String(p), pageSize: String(PAGE_SIZE) })
      if (shortId) params.set('short_id', shortId)
      if (from) params.set('from', from)
      if (q) params.set('q', q)
      const data = await apiGet(`/api/mails/all?${params}`)
      setEmails(data.emails || [])
      setTotal(data.total || 0)
    } catch {} finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { fetchMails(page, filterShortId, filterFrom, filterQuery) }, [page, filterShortId, filterFrom, filterQuery, fetchMails])

  const handleSearch = useCallback((e) => {
    e.preventDefault()
    setPage(1)
    setFilterQuery(searchInput.trim())
  }, [searchInput])

  const handleClearSearch = useCallback(() => {
    setSearchInput('')
    setFilterQuery('')
    setPage(1)
  }, [])

  const handleShortIdChange = useCallback((value) => {
    setFilterShortId(value)
    setPage(1)
  }, [])

  const handleFromChange = useCallback((value) => {
    setFilterFrom(value)
    setPage(1)
  }, [])

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

        {/* Filters */}
        <div className="flex flex-wrap items-end gap-2 mb-4">
          {/* Account (short_id) dropdown */}
          <div className="flex flex-col gap-1">
            <label className="text-[11px] text-base-content/50 font-medium">{t('recentMails.filterAccount')}</label>
            <select
              className="select select-sm select-bordered min-w-[140px]"
              value={filterShortId}
              onChange={e => handleShortIdChange(e.target.value)}
            >
              <option value="">{t('recentMails.filterAll')}</option>
              {recipients.map(r => <option key={r} value={r}>{r}</option>)}
            </select>
          </div>

          {/* Sender (from) dropdown */}
          <div className="flex flex-col gap-1">
            <label className="text-[11px] text-base-content/50 font-medium">{t('recentMails.filterSender')}</label>
            <select
              className="select select-sm select-bordered min-w-[180px]"
              value={filterFrom}
              onChange={e => handleFromChange(e.target.value)}
            >
              <option value="">{t('recentMails.filterAll')}</option>
              {senders.map(s => <option key={s.value} value={s.value}>{s.label}</option>)}
            </select>
          </div>

          {/* Subject search */}
          <form onSubmit={handleSearch} className="flex-1 min-w-[200px]">
            <div className="flex flex-col gap-1">
              <label className="text-[11px] text-base-content/50 font-medium">{t('recentMails.filterSubject')}</label>
              <div className="relative">
                <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-base-content/30" />
                <input
                  type="text"
                  className="input-modern input-sm w-full pl-10 pr-16"
                  placeholder={t('recentMails.filterSearch')}
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
            </div>
          </form>
        </div>

        <div className="flex gap-4 min-h-0">
          {/* List */}
          <div className={`${selectedMail ? 'hidden lg:block lg:w-2/5' : 'w-full'} card-modern flex flex-col max-h-[calc(100vh-220px)]`}>
            <div className="flex-1 overflow-y-auto min-h-0">
              {loading && emails.length === 0 ? (
                <div className="flex items-center justify-center py-16">
                  <span className="loading loading-spinner text-primary"></span>
                </div>
              ) : emails.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-16">
                  <Mail size={48} className="text-base-content/20 mb-2 opacity-30" />
                  <p className="text-sm text-base-content/30">{(filterShortId || filterFrom || filterQuery) ? t('recentMails.noResults') : t('recentMails.empty')}</p>
                </div>
              ) : (
                <div className="divide-y divide-base-300/40">
                  {emails.map((mail) => {
                    const isSelected = selectedMail?.id === mail.id
                    return (
                      <MailListItem
                        key={mail.id}
                        mail={mail}
                        isSelected={isSelected}
                        badge={mail.short_id}
                        onClick={async () => {
                          try {
                            const full = await apiGet(`/api/mails/${mail.id}`)
                            setSelectedMail(normalizeMail(full))
                          } catch {
                            setSelectedMail(normalizeMail(mail))
                          }
                        }}
                      />
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
