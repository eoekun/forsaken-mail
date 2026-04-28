/**
 * Created by Hongcai Deng on 2015/12/28.
 */

'use strict';

const { EventEmitter } = require('events');
const { SMTPServer } = require('smtp-server');
const { simpleParser } = require('mailparser');
const config = require('./config');

const mailin = new EventEmitter();

function formatAddress(address) {
  if (!address || !Array.isArray(address.value) || address.value.length === 0) {
    return '';
  }
  return address.value
    .map(function(item) {
      if (!item) {
        return '';
      }
      if (item.name && item.address) {
        return item.name + ' <' + item.address + '>';
      }
      return item.address || '';
    })
    .filter(Boolean)
    .join(', ');
}

function envelopeToHeaderTo(session) {
  const recipients = session && session.envelope && Array.isArray(session.envelope.rcptTo)
    ? session.envelope.rcptTo
    : [];
  return recipients
    .map(function(item) {
      return item && item.address ? item.address : '';
    })
    .filter(Boolean)
    .join(', ');
}

function buildMailPayload(parsed, session) {
  const fromText = formatAddress(parsed.from);
  const toText = formatAddress(parsed.to) || envelopeToHeaderTo(session);
  const subject = typeof parsed.subject === 'string' ? parsed.subject : '';
  const date = parsed.date instanceof Date ? parsed.date.toISOString() : new Date().toISOString();

  return {
    headers: {
      from: fromText,
      to: toText,
      subject: subject,
      date: date
    },
    subject: subject,
    from: parsed.from || null,
    to: parsed.to || null,
    date: date,
    text: typeof parsed.text === 'string' ? parsed.text : '',
    html: typeof parsed.html === 'string'
      ? parsed.html
      : '<pre>' + (typeof parsed.text === 'string' ? parsed.text : '') + '</pre>'
  };
}

const server = new SMTPServer({
  disabledCommands: ['AUTH'],
  secure: false,
  onData: function(stream, session, callback) {
    simpleParser(stream)
      .then(function(parsed) {
        const payload = buildMailPayload(parsed, session);
        mailin.emit('message', session, payload);
        callback();
      })
      .catch(function(err) {
        callback(err);
      });
  }
});

server.on('error', function(err) {
  mailin.emit('error', err);
  console.error(err && err.stack ? err.stack : err);
});

server.listen(config.mailin.port, config.mailin.host, function() {
  console.info('info: SMTP server listening on port ' + config.mailin.port);
});

module.exports = mailin;
