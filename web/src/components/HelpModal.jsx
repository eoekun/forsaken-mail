import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { apiGet, apiPost } from '../lib/api'
import { HelpCircle, Globe, Webhook, Mail, X } from 'lucide-react'

export default function HelpModal({ host }) {
  const { t } = useTranslation()
  const [dnsResult, setDnsResult] = useState('')
  const [webhookToken, setWebhookToken] = useState('')
  const [webhookMessage, setWebhookMessage] = useState('')
  const [webhookResult, setWebhookResult] = useState('')
  const [dnsDomain, setDnsDomain] = useState('')
  const [testEmailSender, setTestEmailSender] = useState('')
  const [testEmailAuthCode, setTestEmailAuthCode] = useState('')
  const [testEmailShortID, setTestEmailShortID] = useState('')
  const [testEmailResult, setTestEmailResult] = useState('')

  const runDnsTest = async () => {
    setDnsResult(t('help.testing'))
    try {
      const domain = dnsDomain || host
      const data = await apiGet(`/api/domain-test?domain=${encodeURIComponent(domain)}`)
      setDnsResult(JSON.stringify(data, null, 2))
    } catch (e) {
      setDnsResult(t('help.error', { message: e.message }))
    }
  }

  const runWebhookTest = async () => {
    if (!webhookToken) {
      setWebhookResult(t('help.enterWebhookToken'))
      return
    }
    setWebhookResult(t('help.sending'))
    try {
      const data = await apiPost('/api/webhook/test', {
        token: webhookToken,
        message: webhookMessage,
      })
      setWebhookResult(JSON.stringify(data, null, 2))
    } catch (e) {
      setWebhookResult(t('help.error', { message: e.message }))
    }
  }

  const runTestEmail = async () => {
    if (!testEmailSender || !testEmailAuthCode) {
      setTestEmailResult(t('help.enterSenderAndCode'))
      return
    }
    setTestEmailResult(t('help.sendingViaQQ'))
    try {
      const data = await apiPost('/api/test-email', {
        sender_email: testEmailSender,
        auth_code: testEmailAuthCode,
        short_id: testEmailShortID || 'test',
      })
      setTestEmailResult(JSON.stringify(data, null, 2))
    } catch (e) {
      setTestEmailResult(t('help.error', { message: e.message }))
    }
  }

  return (
    <>
      <button
        className="btn btn-circle btn-sm fixed bottom-6 right-6 bg-base-100 border border-base-300/60 shadow-lg hover:shadow-xl text-base-content/50 hover:text-primary z-40"
        onClick={() => document.getElementById('help_modal').showModal()}
      >
        <HelpCircle size={18} />
      </button>
      <dialog id="help_modal" className="modal">
        <div className="modal-box max-w-2xl rounded-2xl bg-base-100 p-0 overflow-hidden">
          <div className="flex items-center justify-between px-6 pt-5 pb-3">
            <h3 className="font-semibold text-lg">{t('help.title')}</h3>
            <form method="dialog">
              <button className="btn btn-sm btn-circle btn-ghost">
                <X size={16} />
              </button>
            </form>
          </div>

          <div className="px-6 pb-6 space-y-6">
            {/* DNS Test */}
            <section>
              <div className="flex items-center gap-2 mb-3">
                <Globe size={15} className="text-primary" />
                <h4 className="text-sm font-medium">{t('help.dnsTest')}</h4>
              </div>
              <div className="flex gap-2 mb-2">
                <input
                  type="text"
                  className="input-modern input-sm flex-1"
                  placeholder={host}
                  value={dnsDomain}
                  onChange={e => setDnsDomain(e.target.value)}
                />
                <button className="btn-modern btn-sm btn-primary" onClick={runDnsTest}>{t('help.test')}</button>
              </div>
              {dnsResult && (
                <pre className="bg-base-200/80 p-3 rounded-lg text-xs overflow-auto max-h-40 font-mono text-base-content/70">{dnsResult}</pre>
              )}
            </section>

            {/* Webhook Test */}
            <section>
              <div className="flex items-center gap-2 mb-3">
                <Webhook size={15} className="text-primary" />
                <h4 className="text-sm font-medium">{t('help.webhookTest')}</h4>
              </div>
              <input
                type="text"
                className="input-modern input-sm w-full mb-2"
                placeholder={t('help.webhookToken')}
                value={webhookToken}
                onChange={e => setWebhookToken(e.target.value)}
              />
              <input
                type="text"
                className="input-modern input-sm w-full mb-2"
                placeholder={t('help.testMessage')}
                value={webhookMessage}
                onChange={e => setWebhookMessage(e.target.value)}
              />
              <button className="btn-modern btn-sm btn-primary mb-2" onClick={runWebhookTest}>{t('help.sendTest')}</button>
              {webhookResult && (
                <pre className="bg-base-200/80 p-3 rounded-lg text-xs overflow-auto max-h-40 font-mono text-base-content/70">{webhookResult}</pre>
              )}
            </section>

            {/* Test Email */}
            <section>
              <div className="flex items-center gap-2 mb-3">
                <Mail size={15} className="text-primary" />
                <h4 className="text-sm font-medium">{t('help.testEmail')}</h4>
              </div>
              <input
                type="text"
                className="input-modern input-sm w-full mb-2"
                placeholder={t('help.qqEmail')}
                value={testEmailSender}
                onChange={e => setTestEmailSender(e.target.value)}
              />
              <input
                type="text"
                className="input-modern input-sm w-full mb-2"
                placeholder={t('help.qqAuthCode')}
                value={testEmailAuthCode}
                onChange={e => setTestEmailAuthCode(e.target.value)}
              />
              <input
                type="text"
                className="input-modern input-sm w-full mb-2"
                placeholder={t('help.recipientShortId', { host })}
                value={testEmailShortID}
                onChange={e => setTestEmailShortID(e.target.value)}
              />
              <button className="btn-modern btn-sm btn-primary mb-2" onClick={runTestEmail}>{t('help.sendTestEmail')}</button>
              {testEmailResult && (
                <pre className="bg-base-200/80 p-3 rounded-lg text-xs overflow-auto max-h-40 font-mono text-base-content/70">{testEmailResult}</pre>
              )}
            </section>
          </div>
        </div>
        <form method="dialog" className="modal-backdrop">
          <button>close</button>
        </form>
      </dialog>
    </>
  )
}
