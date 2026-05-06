export function normalizeMail(m) {
  return {
    ...m,
    from: m.from || m.from_addr,
    to: m.to || m.to_addr,
    html: m.html || m.html_body,
    text: m.text || m.text_body,
  }
}
