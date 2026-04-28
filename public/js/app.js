/**
 * Created by Hongcai Deng on 2015/12/29.
 */

$(function(){
  var mailDomain = location.hostname;
  var siteTitle = 'Forsaken Mail';
  var HISTORY_KEY = 'shortid_history_v1';
  var HISTORY_LIMIT = 6;
  var canUseLocalStorage = ('localStorage' in window);
  var shortidExp = /^[a-z0-9._\-\+]{1,64}$/;

  $('.ui.modal')
    .modal()
  ;

  var clipboard = new Clipboard('.copyable, .history-copy');

  var $customShortId = $('#customShortid');
  var $shortId = $('#shortid');
  var $customTheme = 'check';
  var $placeholder_old = '请等待分配临时邮箱';
  var $placeholder_new = '请输入不带后缀邮箱账号';
  $customShortId.on('click',function() {
    var self = $(this);
    $shortId.prop('disabled', false);
    if(self.hasClass('edit')) {
      $shortId.val('');
      self.removeClass('edit');
      self.toggleClass($customTheme);
      $shortId.prop('placeholder', $placeholder_new);
    } else {
      $shortId.prop('disabled', true);
      self.removeClass('check');
      self.toggleClass('edit');
      $shortId.prop('placeholder',$placeholder_old);
      var $mailUser = $shortId.val();
      var mailaddress = $mailUser + '@' + mailDomain;
      setMailAddress($mailUser);
      $shortId.val(mailaddress);
      window.location.reload();
    }
  });
  
  
  var $maillist = $('#maillist');

  $maillist.on('click', 'tr', function() {
    var mail = $(this).data('mail');
    $('#mailcard .header').text(mail.headers.subject || '无主题');
    $('#mailcard .content:last').html(mail.html);
    $('#mailcard i').click(function() {
      $('#raw').modal('show');
    });
    $('#raw .header').text('RAW');
    $('#raw .content').html($('<pre>').html($('<code>').addClass('language-json').html(JSON.stringify(mail, null, 2))));
    Prism.highlightAll();
  });

  var socket = io();

  function normalizeShortId(id) {
    if (typeof id !== 'string') {
      return '';
    }
    var normalized = id.trim().toLowerCase();
    return shortidExp.test(normalized) ? normalized : '';
  }

  function getHistoryList() {
    if (!canUseLocalStorage) {
      return [];
    }
    var raw = localStorage.getItem(HISTORY_KEY);
    if (!raw) {
      return [];
    }
    try {
      var parsed = JSON.parse(raw);
      if (!Array.isArray(parsed)) {
        return [];
      }
      return parsed.map(normalizeShortId).filter(Boolean);
    } catch (err) {
      return [];
    }
  }

  function saveHistoryList(list) {
    if (!canUseLocalStorage) {
      return;
    }
    localStorage.setItem(HISTORY_KEY, JSON.stringify(list));
  }

  function renderHistoryList(activeShortId) {
    var list = getHistoryList();
    var $history = $('#mailHistoryList');
    $history.empty();
    if (list.length === 0) {
      $history.append($('<span>').addClass('ui small grey text').text('暂无历史记录'));
      return;
    }

    list.forEach(function(shortid) {
      var mailaddress = shortid + '@' + mailDomain;
      var $item = $('<div>').addClass('ui mini buttons mail-history-item');
      var $switch = $('<button>')
        .addClass('ui button history-switch')
        .toggleClass('primary', shortid === activeShortId)
        .attr('data-shortid', shortid)
        .text(mailaddress);
      var $copy = $('<button>')
        .addClass('ui icon button history-copy')
        .attr('data-clipboard-text', mailaddress)
        .attr('title', '复制邮箱地址')
        .append($('<i>').addClass('copy icon'));
      $item.append($switch).append($copy);
      $history.append($item);
    });
  }

  function upsertHistory(shortid) {
    var normalized = normalizeShortId(shortid);
    if (!normalized) {
      return;
    }
    var list = getHistoryList().filter(function(item) {
      return item !== normalized;
    });
    list.unshift(normalized);
    if (list.length > HISTORY_LIMIT) {
      list = list.slice(0, HISTORY_LIMIT);
    }
    saveHistoryList(list);
    renderHistoryList(normalized);
  }

  renderHistoryList(normalizeShortId(canUseLocalStorage ? localStorage.getItem('shortid') : ''));

  function setSiteTitle(title) {
    if (typeof title !== 'string' || !title.trim()) {
      return;
    }
    siteTitle = title.trim();
    document.title = siteTitle;
    if (canUseLocalStorage) {
      localStorage.setItem('siteTitle', siteTitle);
    }
    $('#siteTitleText').text(siteTitle);
  }

  var setMailAddress = function(id) {
    var normalizedId = normalizeShortId(id);
    if (!normalizedId) {
      return;
    }
    if (canUseLocalStorage) {
      localStorage.setItem('shortid', normalizedId);
    }
    var mailaddress = normalizedId + '@' + mailDomain;
    $('#shortid').val(mailaddress).parent().siblings('button').find('.mail').attr('data-clipboard-text', mailaddress);
    upsertHistory(normalizedId);
  };

  $.get('/api/config', function(data) {
    if (data && typeof data.siteTitle === 'string') {
      setSiteTitle(data.siteTitle);
    }
    if (data && typeof data.host === 'string' && data.host.trim()) {
      mailDomain = data.host.trim().toLowerCase();
      $('#domainToTest').val(mailDomain);
      if(('localStorage' in window)) {
        var currentShortId = localStorage.getItem('shortid');
        if(currentShortId) {
          setMailAddress(currentShortId);
        } else {
          renderHistoryList('');
        }
      }
    }
  });

  $('#openHelp').on('click', function() {
    $('#helpModal').modal('show');
  });

  $('#runDomainTest').on('click', function() {
    var domain = $('#domainToTest').val().trim() || mailDomain;
    $('#domainTestResult').text('测试中...');
    $.get('/api/domain-test', { domain: domain })
      .done(function(resp) {
        $('#domainTestResult').text(JSON.stringify(resp, null, 2));
      })
      .fail(function(xhr) {
        var errResp = (xhr && xhr.responseJSON) ? xhr.responseJSON : { ok: false, message: '请求失败' };
        $('#domainTestResult').text(JSON.stringify(errResp, null, 2));
      });
  });

  $('#runWebhookTest').on('click', function() {
    var token = $('#webhookToken').val().trim();
    var message = $('#webhookMessage').val().trim();
    if (!token) {
      $('#webhookTestResult').text('请先输入 Webhook Token。');
      return;
    }
    $('#webhookTestResult').text('发送中...');
    $.ajax({
      url: '/api/webhook/test',
      method: 'POST',
      contentType: 'application/json',
      data: JSON.stringify({
        token: token,
        message: message
      })
    })
      .done(function(resp) {
        $('#webhookTestResult').text(JSON.stringify(resp, null, 2));
      })
      .fail(function(xhr) {
        var errResp = (xhr && xhr.responseJSON) ? xhr.responseJSON : { ok: false, message: '请求失败' };
        $('#webhookTestResult').text(JSON.stringify(errResp, null, 2));
      });
  });

  $('#refreshShortid').click(function() {
    socket.emit('request shortid', true);
  });

  $('#mailHistoryList').on('click', '.history-switch', function() {
    var selectedId = $(this).attr('data-shortid');
    if (!selectedId) {
      return;
    }
    setMailAddress(selectedId);
    socket.emit('set shortid', selectedId);
  });

  socket.on('connect', function() {
    if(canUseLocalStorage) {
      var shortid = localStorage.getItem('shortid');
      if(!shortid) {
        socket.emit('request shortid', true);
      }
      else {
        socket.emit('set shortid', shortid);
      }
    } else {
      socket.emit('request shortid', true);
    }
  });

  socket.on('shortid', function(id) {
    setMailAddress(id);
  });

  socket.on('mail', function(mail) {
    if(('Notification' in window)) {
      if(Notification.permission === 'granted') {
        new Notification('New mail from ' + mail.headers.from);
      }
      else if(Notification.permission !== 'denied') {
        Notification.requestPermission(function(permission) {
          if(permission === 'granted') {
            new Notification('New mail from ' + mail.headers.from);
          }
        })
      }
    }
    var $tr = $('<tr>').data('mail', mail);
    $tr
      .append($('<td>').text(mail.headers.from))
      .append($('<td>').text(mail.headers.subject || '无主题'))
      .append($('<td>').text((new Date(mail.headers.date)).toLocaleTimeString()));
    $maillist.prepend($tr);
  });
});
