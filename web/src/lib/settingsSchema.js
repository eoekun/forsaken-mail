const defineField = (config) => ({
  defaultValue: '',
  fromApi: value => value ?? '',
  toApi: value => value,
  ...config,
})

function toNumberField(key, min) {
  return (value) => {
    const parsed = Number(value)
    if (!Number.isFinite(parsed)) {
      throw new Error(`${key} must be a number`)
    }
    if (parsed < min) {
      throw new Error(`${key} must be >= ${min}`)
    }
    return parsed
  }
}

const SETTINGS_FIELDS = [
  defineField({
    key: 'mail_host',
    section: 'general',
    type: 'text',
    placeholder: 'example.com,mail.example.com',
  }),
  defineField({
    key: 'site_title',
    section: 'general',
    type: 'text',
    placeholder: 'Tmail',
  }),
  defineField({
    key: 'login_whitelist',
    section: 'security',
    type: 'text',
    placeholder: 'user1@example.com,user2@example.com',
  }),
  defineField({
    key: 'keyword_blacklist',
    section: 'security',
    type: 'text',
    placeholder: 'admin,root,system',
  }),
  defineField({
    key: 'webhook_enabled',
    section: 'notifications',
    type: 'toggle',
    defaultValue: false,
    fromApi: value => Boolean(value),
    toApi: value => Boolean(value),
  }),
  defineField({
    key: 'webhook_service',
    section: 'notifications',
    type: 'select',
    defaultValue: 'dingtalk',
    options: ['dingtalk', 'telegram', 'slack'],
    fromApi: value => value || 'dingtalk',
    toApi: value => value || 'dingtalk',
  }),
  defineField({
    key: 'webhook_config',
    section: 'notifications',
    type: 'textarea',
    placeholder: '{"token":"...","chat_id":"..."}',
    defaultValue: '{}',
    fromApi: value => JSON.stringify(value || {}, null, 2),
    toApi: value => {
      const text = typeof value === 'string' ? value.trim() : ''
      if (!text) {
        return {}
      }
      let parsed
      try {
        parsed = JSON.parse(text)
      } catch {
        throw new Error('webhook_config must be valid JSON')
      }
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
        throw new Error('webhook_config must be a JSON object')
      }
      return parsed
    },
  }),
  defineField({
    key: 'webhook_message',
    section: 'notifications',
    type: 'text',
    placeholder: 'new email received.',
  }),
  defineField({
    key: 'mail_retention_hours',
    section: 'retention',
    type: 'number',
    min: 0,
    defaultValue: 1,
    fromApi: value => Number.isFinite(value) ? value : 1,
    toApi: toNumberField('mail_retention_hours', 0),
  }),
  defineField({
    key: 'mail_max_count',
    section: 'retention',
    type: 'number',
    min: 0,
    defaultValue: 100,
    fromApi: value => Number.isFinite(value) ? value : 100,
    toApi: toNumberField('mail_max_count', 0),
  }),
  defineField({
    key: 'max_mail_size_bytes',
    section: 'retention',
    type: 'number',
    min: 1,
    defaultValue: 1048576,
    fromApi: value => Number.isFinite(value) ? value : 1048576,
    toApi: toNumberField('max_mail_size_bytes', 1),
  }),
  defineField({
    key: 'audit_mail_received',
    section: 'audit',
    type: 'toggle',
    defaultValue: true,
    fromApi: value => Boolean(value),
    toApi: value => Boolean(value),
  }),
  defineField({
    key: 'audit_retention_days',
    section: 'audit',
    type: 'number',
    min: 0,
    defaultValue: 7,
    fromApi: value => Number.isFinite(value) ? value : 7,
    toApi: toNumberField('audit_retention_days', 0),
  }),
  defineField({
    key: 'audit_max_count',
    section: 'audit',
    type: 'number',
    min: 0,
    defaultValue: 5000,
    fromApi: value => Number.isFinite(value) ? value : 5000,
    toApi: toNumberField('audit_max_count', 0),
  }),
]

const fieldMap = Object.fromEntries(SETTINGS_FIELDS.map(field => [field.key, field]))

export const SETTINGS_SECTIONS = [
  { sectionKey: 'general', keys: SETTINGS_FIELDS.filter(field => field.section === 'general').map(field => field.key) },
  { sectionKey: 'security', keys: SETTINGS_FIELDS.filter(field => field.section === 'security').map(field => field.key) },
  { sectionKey: 'notifications', keys: SETTINGS_FIELDS.filter(field => field.section === 'notifications').map(field => field.key) },
  { sectionKey: 'retention', keys: SETTINGS_FIELDS.filter(field => field.section === 'retention').map(field => field.key) },
  { sectionKey: 'audit', keys: SETTINGS_FIELDS.filter(field => field.section === 'audit').map(field => field.key) },
]

export function getSettingField(key) {
  return fieldMap[key]
}

export function createSettingsForm(values = {}) {
  const form = {}
  for (const field of SETTINGS_FIELDS) {
    form[field.key] = field.fromApi(values[field.key])
  }
  return form
}

export function serializeSettingsForm(form) {
  const payload = {}
  for (const field of SETTINGS_FIELDS) {
    payload[field.key] = field.toApi(form[field.key])
  }
  return payload
}
