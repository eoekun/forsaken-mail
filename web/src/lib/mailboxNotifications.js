import i18n from '../i18n'

export function requestNotificationPermission() {
  if (!('Notification' in window) || Notification.permission !== 'default') {
    return
  }
  Notification.requestPermission()
}

export function notifyNewMail(mail) {
  if (!('Notification' in window) || Notification.permission !== 'granted') {
    return
  }
  new Notification(i18n.t('notification.newMail', { from: mail.from }))
}
