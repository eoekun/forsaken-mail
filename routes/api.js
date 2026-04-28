/**
 * Created by Hongcai Deng on 2015/12/28.
 */

'use strict';

let express = require('express');
let router = express.Router();
let dns = require('dns').promises;
let config = require('../modules/config');
let sendDingtalkNotification = require('../modules/dingtalk');

function normalizeDomain(value) {
  if (typeof value !== 'string') {
    return '';
  }
  let domain = value.trim().toLowerCase();
  domain = domain.replace(/^https?:\/\//, '');
  domain = domain.replace(/\/.*$/, '');
  domain = domain.replace(/:\d+$/, '');
  return /^[a-z0-9.-]+$/.test(domain) ? domain : '';
}

router.get('/', function(req, res) {
  res.end();
});

router.get('/config', function(req, res) {
  res.json({
    host: config.host || req.hostname,
    siteTitle: config.siteTitle
  });
});

router.get('/domain-test', async function(req, res) {
  const targetDomain = normalizeDomain(req.query.domain) || config.host || req.hostname;
  if (!targetDomain) {
    res.status(400).json({
      ok: false,
      message: 'Invalid domain.'
    });
    return;
  }

  try {
    const mxRecords = await dns.resolveMx(targetDomain);
    mxRecords.sort((a, b) => a.priority - b.priority);

    const detailedRecords = await Promise.all(mxRecords.map(async function(record) {
      let ipv4 = [];
      let ipv6 = [];
      let error = null;
      try {
        ipv4 = await dns.resolve4(record.exchange);
      } catch (err) {
        if (err && err.code !== 'ENODATA' && err.code !== 'ENOTFOUND' && err.code !== 'ESERVFAIL') {
          error = err.code || err.message;
        }
      }
      try {
        ipv6 = await dns.resolve6(record.exchange);
      } catch (err) {
        if (!error && err && err.code !== 'ENODATA' && err.code !== 'ENOTFOUND' && err.code !== 'ESERVFAIL') {
          error = err.code || err.message;
        }
      }
      return {
        priority: record.priority,
        exchange: record.exchange,
        ipv4: ipv4,
        ipv6: ipv6,
        hasAddress: (ipv4.length + ipv6.length) > 0,
        error: error
      };
    }));

    const hasDeliverableMx = detailedRecords.some(function(item) {
      return item.hasAddress;
    });

    res.json({
      ok: hasDeliverableMx,
      domain: targetDomain,
      mx: detailedRecords,
      message: hasDeliverableMx
        ? 'DNS looks good for mail delivery.'
        : 'MX found, but no resolvable IP for MX host(s).'
    });
  } catch (err) {
    res.status(200).json({
      ok: false,
      domain: targetDomain,
      mx: [],
      message: 'Failed to resolve MX records.',
      error: err && (err.code || err.message)
    });
  }
});

router.post('/webhook/test', async function(req, res) {
  const token = req.body && typeof req.body.token === 'string' ? req.body.token.trim() : '';
  const message = req.body && typeof req.body.message === 'string' ? req.body.message.trim() : '';
  if (!token) {
    res.status(400).json({
      ok: false,
      message: 'Webhook token is required.'
    });
    return;
  }

  try {
    const result = await sendDingtalkNotification.sendTestNotification(token, message);
    res.status(result.ok ? 200 : 400).json(result);
  } catch (err) {
    res.status(502).json({
      ok: false,
      message: 'Webhook request failed.',
      error: err && err.message ? err.message : String(err)
    });
  }
});

module.exports = router;
