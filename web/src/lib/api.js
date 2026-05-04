import i18n from '../i18n'

export async function apiFetch(url, options = {}) {
  const headers = new Headers(options.headers)
  headers.set('Accept-Language', i18n.language)

  const res = await fetch(url, {
    ...options,
    headers,
    credentials: 'same-origin',
  })

  if (res.status === 401) {
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }

  return res.json()
}

export async function apiGet(url) {
  return apiFetch(url)
}

export async function apiPost(url, data) {
  return apiFetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
}

export async function apiPut(url, data) {
  return apiFetch(url, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
}

export async function apiDelete(url) {
  return apiFetch(url, {
    method: 'DELETE',
  })
}
