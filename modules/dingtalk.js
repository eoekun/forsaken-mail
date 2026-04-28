/**
 * DingTalk robot webhook notifier.
 */

'use strict';

const https = require('https');
const config = require('./config');

const DINGTALK_HOST = 'oapi.dingtalk.com';
const FIXED_TEXT = 'Forsaken-Mail: new email received.';

function sendDingtalkNotification() {
  const token = config.dingtalkWebhookToken;
  if (!token) {
    return;
  }

  const payload = JSON.stringify({
    msgtype: 'text',
    text: {
      content: FIXED_TEXT
    }
  });

  const req = https.request({
    hostname: DINGTALK_HOST,
    port: 443,
    method: 'POST',
    path: '/robot/send?access_token=' + encodeURIComponent(token),
    headers: {
      'Content-Type': 'application/json',
      'Content-Length': Buffer.byteLength(payload)
    }
  }, function(res) {
    // Drain response to keep socket healthy.
    res.on('data', function() {});
  });

  req.on('error', function(err) {
    console.error('DingTalk webhook request failed:', err.message);
  });

  req.write(payload);
  req.end();
}

module.exports = sendDingtalkNotification;
