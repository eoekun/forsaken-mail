import { apiGet } from './api'

function normalizeStringList(value) {
  if (Array.isArray(value)) {
    return value.map(item => String(item).trim()).filter(Boolean)
  }
  if (typeof value === 'string') {
    return value.split(',').map(item => item.trim()).filter(Boolean)
  }
  return []
}

export async function loadPublicConfig() {
  const data = await apiGet('/api/config')
  const hosts = normalizeStringList(data.hosts && data.hosts.length > 0 ? data.hosts : [data.host])

  return {
    host: hosts[0] || '',
    hosts,
    siteTitle: data.site_title || '',
    authMode: data.auth_mode || 'oauth',
    keywordBlacklist: normalizeStringList(data.keyword_blacklist),
    email: data.email || '',
  }
}
