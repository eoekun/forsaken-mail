import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { apiGet, apiPost } from '../lib/api'

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
        className="btn btn-circle btn-sm fixed bottom-4 right-4"
        onClick={() => document.getElementById('help_modal').showModal()}
      >
        ?
      </button>
      <dialog id="help_modal" className="modal">
        <div className="modal-box max-w-2xl">
          <h3 className="font-bold text-lg">{t('help.title')}</h3>

          <div className="divider">{t('help.dnsTest')}</div>
          <div className="flex gap-2 mb-2">
            <input
              type="text"
              className="input input-bordered input-sm flex-1"
              placeholder={host}
              value={dnsDomain}
              onChange={e => setDnsDomain(e.target.value)}
            />
            <button className="btn btn-sm btn-primary" onClick={runDnsTest}>{t('help.test')}</button>
          </div>
          <pre className="bg-base-200 p-2 rounded text-xs overflow-auto max-h-40">{dnsResult}</pre>

          <div className="divider">{t('help.webhookTest')}</div>
          <input
            type="text"
            className="input input-bordered input-sm w-full mb-2"
            placeholder={t('help.webhookToken')}
            value={webhookToken}
            onChange={e => setWebhookToken(e.target.value)}
          />
          <input
            type="text"
            className="input input-bordered input-sm w-full mb-2"
            placeholder={t('help.testMessage')}
            value={webhookMessage}
            onChange={e => setWebhookMessage(e.target.value)}
          />
          <button className="btn btn-sm btn-primary mb-2" onClick={runWebhookTest}>{t('help.sendTest')}</button>
          <pre className="bg-base-200 p-2 rounded text-xs overflow-auto max-h-40">{webhookResult}</pre>

          <div className="divider">{t('help.testEmail')}</div>
          <input
            type="text"
            className="input input-bordered input-sm w-full mb-2"
            placeholder={t('help.qqEmail')}
            value={testEmailSender}
            onChange={e => setTestEmailSender(e.target.value)}
          />
          <input
            type="text"
            className="input input-bordered input-sm w-full mb-2"
            placeholder={t('help.qqAuthCode')}
            value={testEmailAuthCode}
            onChange={e => setTestEmailAuthCode(e.target.value)}
          />
          <input
            type="text"
            className="input input-bordered input-sm w-full mb-2"
            placeholder={t('help.recipientShortId', { host })}
            value={testEmailShortID}
            onChange={e => setTestEmailShortID(e.target.value)}
          />
          <button className="btn btn-sm btn-primary mb-2" onClick={runTestEmail}>{t('help.sendTestEmail')}</button>
          <pre className="bg-base-200 p-2 rounded text-xs overflow-auto max-h-40">{testEmailResult}</pre>

          <div className="modal-action">
            <form method="dialog">
              <button className="btn">{t('help.close')}</button>
            </form>
          </div>
        </div>
      </dialog>
    </>
  )
}
