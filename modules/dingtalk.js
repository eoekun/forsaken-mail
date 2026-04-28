/**
 * DingTalk robot webhook notifier.
 */

'use strict';

const https = require('https');
const config = require('./config');

const DINGTALK_HOST = 'oapi.dingtalk.com';
const FIXED_TEXT = 'Forsaken-Mail: new email received.';

function postDingtalkText(token, text) {
  const normalizedToken = typeof token === 'string' ? token.trim() : '';
  if (!normalizedToken) {
    return Promise.resolve({
      ok: false,
      message: 'Webhook token is empty.',
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
      hostname: DINGTALK_HOST,
      port: 443,
      method: 'POST',
      path: '/robot/send?access_token=' + encodeURIComponent(normalizedToken),
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

    req.on('error', function(err) {
      reject(err);
    });

    req.write(payload);
    req.end();
  });
}

function sendDingtalkNotification() {
  const token = config.dingtalkWebhookToken;
  if (!token) {
    return;
  }
  postDingtalkText(token, FIXED_TEXT)
    .catch(function(err) {
      console.error('DingTalk webhook request failed:', err.message);
    });
}

module.exports = sendDingtalkNotification;
module.exports.sendTestNotification = function(token, text) {
  const message = typeof text === 'string' && text.trim()
    ? text.trim()
    : 'Forsaken-Mail test message.';
  return postDingtalkText(token, message);
};
