/**
 * DingTalk robot webhook notifier.
 */

'use strict';

const https = require('https');
const config = require('./config');

const DINGTALK_HOST = 'oapi.dingtalk.com';
const DINGTALK_DEFAULT_PATH = '/robot/send';
const DINGTALK_TEST_MESSAGE = 'Forsaken-Mail test message.';
const MAX_PREVIEW_LENGTH = 200;
const MAX_MESSAGE_LENGTH = 1800;

function buildWebhookTarget(tokenOrUrl) {
  const normalizedValue = typeof tokenOrUrl === 'string' ? tokenOrUrl.trim() : '';
  if (!normalizedValue) {
    return null;
  }

  if (/^https?:\/\//i.test(normalizedValue)) {
    try {
      const parsed = new URL(normalizedValue);
      if (!parsed.hostname) {
        return null;
      }
      if (parsed.protocol !== 'https:') {
        return null;
      }
      return {
        hostname: parsed.hostname,
        port: parsed.port ? parseInt(parsed.port, 10) : undefined,
        path: (parsed.pathname || '/') + (parsed.search || ''),
        protocol: 'https:'
      };
    } catch (err) {
      return null;
    }
  }

  return {
    hostname: DINGTALK_HOST,
    path: DINGTALK_DEFAULT_PATH + '?access_token=' + encodeURIComponent(normalizedValue),
    protocol: 'https:'
  };
}

function postDingtalkText(tokenOrUrl, text) {
  const target = buildWebhookTarget(tokenOrUrl);
  if (!target) {
    return Promise.resolve({
      ok: false,
      message: 'Webhook token/url is empty or invalid.',
      data: null
    });
  }
  const payload = JSON.stringify({
    msgtype: 'text',
    text: {
      content: text
    }
  });

  return new Promise((resolve, reject) => {
    const req = https.request({
      protocol: target.protocol || 'https:',
      hostname: target.hostname,
      port: target.port || 443,
      method: 'POST',
      path: target.path,
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(payload)
      }
    }, function(res) {
      let raw = '';
      res.on('data', function(chunk) {
        raw += chunk;
      });
      res.on('end', function() {
        let parsed = null;
        try {
          parsed = raw ? JSON.parse(raw) : null;
        } catch (err) {
          parsed = null;
        }
        const success = !!(parsed && parsed.errcode === 0);
        resolve({
          ok: success,
          message: success ? 'ok' : 'DingTalk returned non-zero errcode.',
          data: parsed,
          statusCode: res.statusCode
        });
      });
    });

    req.setTimeout(10000, function() {
      req.destroy(new Error('Request timeout.'));
    });

    req.on('error', function(err) {
      reject(err);
    });

    req.write(payload);
    req.end();
  });
}

function sanitizeSingleLine(value, fallback) {
  if (typeof value !== 'string') {
    return fallback;
  }
  const normalized = value.replace(/\s+/g, ' ').trim();
  return normalized || fallback;
}

function buildTextPreview(data) {
  const rawText = data && typeof data.text === 'string' ? data.text : '';
  const normalized = rawText.replace(/\s+/g, ' ').trim();
  if (!normalized) {
    return '';
  }
  if (normalized.length <= MAX_PREVIEW_LENGTH) {
    return normalized;
  }
  return normalized.slice(0, MAX_PREVIEW_LENGTH) + '...';
}

function buildMailMessage(data) {
  const from = sanitizeSingleLine(data && data.headers ? data.headers.from : '', 'unknown');
  const to = sanitizeSingleLine(data && data.headers ? data.headers.to : '', 'unknown');
  const subject = sanitizeSingleLine(data && data.headers ? data.headers.subject : '', '(no subject)');
  const date = sanitizeSingleLine(data && data.headers ? data.headers.date : '', new Date().toISOString());
  const preview = buildTextPreview(data);
  const title = sanitizeSingleLine(config.dingtalkWebhookMessage, 'Forsaken-Mail: new email received.');

  const lines = [
    title,
    'From: ' + from,
    'To: ' + to,
    'Subject: ' + subject,
    'Date: ' + date
  ];

  if (preview) {
    lines.push('Preview: ' + preview);
  }

  const message = lines.join('\n');
  if (message.length <= MAX_MESSAGE_LENGTH) {
    return message;
  }
  return message.slice(0, MAX_MESSAGE_LENGTH) + '...';
}

function sendDingtalkNotification(mailData) {
  const tokenOrUrl = config.dingtalkWebhookToken;
  if (!tokenOrUrl) {
    return;
  }
  const text = buildMailMessage(mailData);
  postDingtalkText(tokenOrUrl, text)
    .then(function(result) {
      if (!result.ok) {
        console.error('DingTalk webhook returned non-ok result:', JSON.stringify(result));
      }
    })
    .catch(function(err) {
      console.error('DingTalk webhook request failed:', err.message);
    });
}

module.exports = sendDingtalkNotification;
module.exports.sendTestNotification = function(token, text) {
  const message = typeof text === 'string' && text.trim()
    ? text.trim()
    : DINGTALK_TEST_MESSAGE;
  return postDingtalkText(token, message);
};
