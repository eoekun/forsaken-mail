import { apiGet, apiPost, apiPut } from './api'

export async function loadSettings() {
  return apiGet('/api/admin/settings')
}

export async function saveSettings(values) {
  const result = await apiPut('/api/admin/settings', values)
  return result.values || values
}

export async function testWebhook({ service, config, message }) {
  return apiPost('/api/webhook/test', { service, config, message })
}
