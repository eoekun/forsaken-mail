import { useState } from 'react'
import { apiGet, apiPost } from '../lib/api'

export default function HelpModal({ host }) {
  const [dnsResult, setDnsResult] = useState('')
  const [webhookToken, setWebhookToken] = useState('')
  const [webhookMessage, setWebhookMessage] = useState('')
  const [webhookResult, setWebhookResult] = useState('')
  const [dnsDomain, setDnsDomain] = useState('')

  const runDnsTest = async () => {
    setDnsResult('Testing...')
    try {
      const domain = dnsDomain || host
      const data = await apiGet(`/api/domain-test?domain=${encodeURIComponent(domain)}`)
      setDnsResult(JSON.stringify(data, null, 2))
    } catch (e) {
      setDnsResult(`Error: ${e.message}`)
    }
  }

  const runWebhookTest = async () => {
    if (!webhookToken) {
      setWebhookResult('Please enter a webhook token.')
      return
    }
    setWebhookResult('Sending...')
    try {
      const data = await apiPost('/api/webhook/test', {
        token: webhookToken,
        message: webhookMessage,
      })
      setWebhookResult(JSON.stringify(data, null, 2))
    } catch (e) {
      setWebhookResult(`Error: ${e.message}`)
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
          <h3 className="font-bold text-lg">Help & Testing</h3>

          <div className="divider">DNS Test</div>
          <div className="flex gap-2 mb-2">
            <input
              type="text"
              className="input input-bordered input-sm flex-1"
              placeholder={host}
              value={dnsDomain}
              onChange={e => setDnsDomain(e.target.value)}
            />
            <button className="btn btn-sm btn-primary" onClick={runDnsTest}>Test</button>
          </div>
          <pre className="bg-base-200 p-2 rounded text-xs overflow-auto max-h-40">{dnsResult}</pre>

          <div className="divider">Webhook Test</div>
          <input
            type="text"
            className="input input-bordered input-sm w-full mb-2"
            placeholder="Webhook Token or URL"
            value={webhookToken}
            onChange={e => setWebhookToken(e.target.value)}
          />
          <input
            type="text"
            className="input input-bordered input-sm w-full mb-2"
            placeholder="Test message (optional)"
            value={webhookMessage}
            onChange={e => setWebhookMessage(e.target.value)}
          />
          <button className="btn btn-sm btn-primary mb-2" onClick={runWebhookTest}>Send Test</button>
          <pre className="bg-base-200 p-2 rounded text-xs overflow-auto max-h-40">{webhookResult}</pre>

          <div className="modal-action">
            <form method="dialog">
              <button className="btn">Close</button>
            </form>
          </div>
        </div>
      </dialog>
    </>
  )
}
