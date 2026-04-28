/**
 * Created by Hongcai Deng on 2015/12/28.
 */

'use strict';

function parsePort(value, fallback) {
  if (typeof value !== 'string') {
    return fallback;
  }
  const port = parseInt(value, 10);
  if (isNaN(port) || port < 1 || port > 65535) {
    return fallback;
  }
  return port;
}

function normalizeMailHost(host, fallback) {
  if (typeof host !== 'string') {
    return fallback;
  }
  let normalized = host.trim().toLowerCase();
  if (!normalized) {
    return fallback;
  }
  normalized = normalized.replace(/^https?:\/\//, '');
  normalized = normalized.replace(/\/.*$/, '');
  normalized = normalized.replace(/:\d+$/, '');
  return /^[a-z0-9.-]+$/.test(normalized) ? normalized : fallback;
}

function parseKeywordList(value, fallback) {
  if (typeof value !== 'string') {
    return fallback;
  }
  return value
    .split(',')
    .map((keyword) => keyword.trim().toLowerCase())
    .filter((keyword) => keyword);
}

function parseToken(value) {
  if (typeof value !== 'string') {
    return '';
  }
  return value.trim();
}

function parseText(value, fallback) {
  if (typeof value !== 'string') {
    return fallback;
  }
  const normalized = value.trim();
  return normalized || fallback;
}

function normalizeSiteTitle(value, fallback) {
  if (typeof value !== 'string') {
    return fallback;
  }
  const normalized = value.trim();
  if (!normalized) {
    return fallback;
  }
  return normalized;
}

const DEFAULTS = {
  mailHost: 'disposable.dhc-app.com',
  mailinHost: '0.0.0.0',
  mailinPort: 25,
  siteTitle: 'Forsaken Mail',
  dingtalkWebhookMessage: 'Forsaken-Mail: new email received.',
  keywordBlackList: [
    'admin',
    'postmaster',
    'system',
    'webmaster',
    'administrator',
    'hostmaster',
    'service',
    'server',
    'root'
  ]
};

const config = {
  mailin: {
    host: process.env.MAILIN_HOST || DEFAULTS.mailinHost,
    port: parsePort(process.env.MAILIN_PORT, DEFAULTS.mailinPort),
    disableWebhook: true
  },
  host: normalizeMailHost(process.env.MAIL_HOST, DEFAULTS.mailHost),
  siteTitle: normalizeSiteTitle(process.env.SITE_TITLE, DEFAULTS.siteTitle),
  keywordBlackList: parseKeywordList(process.env.KEYWORD_BLACKLIST, DEFAULTS.keywordBlackList),
  dingtalkWebhookToken: parseToken(process.env.DINGTALK_WEBHOOK_TOKEN),
  dingtalkWebhookMessage: parseText(process.env.DINGTALK_WEBHOOK_MESSAGE, DEFAULTS.dingtalkWebhookMessage)
};

module.exports = config;
