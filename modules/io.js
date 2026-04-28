/**
 * Created by Hongcai Deng on 2015/12/28.
 */

'use strict';

const mailin = require('./mailin');
const config = require('./config');
const sendDingtalkNotification = require('./dingtalk');

let onlines = new Map();
const shortidExp = /^[a-z0-9._\-\+]{1,64}$/;
const firstNamePool = [
  'alex', 'mike', 'tom', 'jack', 'leo', 'sam', 'eric', 'lucas', 'liam', 'noah',
  'emma', 'olivia', 'sophia', 'mia', 'ava', 'lily', 'grace', 'ella', 'zoe', 'nina'
];
const tagPool = [
  'mail', 'inbox', 'user', 'note', 'cloud', 'river', 'forest', 'stone', 'ocean', 'field',
  'sun', 'moon', 'star', 'leaf', 'bird', 'fox', 'wolf', 'lake', 'hill', 'wind'
];

function normalizeShortId(id) {
  if (typeof id !== 'string') {
    return null;
  }
  const normalized = id.trim().toLowerCase();
  if (!shortidExp.test(normalized)) {
    return null;
  }
  return normalized;
}

function checkShortIdMatchBlackList(id) {
  const keywordBlackList = config.keywordBlackList;
  if (keywordBlackList && keywordBlackList.length > 0) {
    for (let i = 0; i < keywordBlackList.length; i++) {
      const keyword = keywordBlackList[i];
      if (typeof keyword !== 'string' || keyword.length === 0) {
        continue;
      }
      if (id.includes(keyword.toLowerCase())) {
        return true;
      }
    }
  } 
  return false;
}

function assignGeneratedShortId(socket) {
  onlines.delete(socket.shortid);
  socket.shortid = generateReadableShortId();
  onlines.set(socket.shortid, socket); // add incomming connection to online table
  socket.emit('shortid', socket.shortid);
}

function randomPick(pool) {
  return pool[Math.floor(Math.random() * pool.length)];
}

function generateCandidate() {
  const name = randomPick(firstNamePool);
  const tag = randomPick(tagPool);
  const suffix = Math.floor(100 + Math.random() * 900);
  return name + tag + suffix;
}

function generateReadableShortId() {
  for (let i = 0; i < 20; i++) {
    const candidate = generateCandidate();
    if (!shortidExp.test(candidate)) {
      continue;
    }
    if (checkShortIdMatchBlackList(candidate)) {
      continue;
    }
    if (onlines.has(candidate)) {
      continue;
    }
    return candidate;
  }

  // Extremely rare fallback path.
  return 'mailuser' + Date.now();
}

module.exports = function(io) {
  mailin.on('message', function(connection, data) {
    sendDingtalkNotification(data);

    let to = (data && data.headers && typeof data.headers.to === 'string')
      ? data.headers.to.toLowerCase()
      : '';
    if (!to) {
      return;
    }
    let exp = /[\w\._\-\+]+@[\w\._\-\+]+/i;
    if(exp.test(to)) {
      let matches = to.match(exp);
      let shortid = matches[0].substring(0, matches[0].indexOf('@'));
      if(onlines.has(shortid)) {
        onlines.get(shortid).emit('mail', data);
      }
    }
  });

  io.on('connection', socket => {
    socket.on('request shortid', function() {
      assignGeneratedShortId(socket);
    });

    socket.on('set shortid', function(id) {
      const normalizedId = normalizeShortId(id);
      if (!normalizedId || checkShortIdMatchBlackList(normalizedId)) {
        // fallback to generated shortid if invalid or match keyword blacklist
        assignGeneratedShortId(socket);
        return;
      }
      onlines.delete(socket.shortid);
      socket.shortid = normalizedId;
      onlines.set(socket.shortid, socket);
      socket.emit('shortid', socket.shortid);
    })
    
    socket.on('disconnect', socket => {
      onlines.delete(socket.shortid);
    });
  });
};
