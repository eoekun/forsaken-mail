/**
 * Formats a sender address for display.
 * - "Display Name <email>" → "Display Name"
 * - long-bounce@em7877.tm.openai.com → "openai.com"
 * - user@example.com → "user@example.com" (short enough)
 */
export function formatSender(from) {
  if (!from) return ''

  // Extract display name: "Name <email>" → "Name"
  const nameMatch = from.match(/^"?(.+?)"?\s*<[^>]+>$/)
  if (nameMatch) {
    const name = nameMatch[1].trim()
    if (name && name !== from) return name
  }

  // Extract email from "Name <email>" format
  const emailMatch = from.match(/<([^>]+)>/)
  const email = emailMatch ? emailMatch[1] : from

  // If short enough, show as-is
  if (email.length <= 30) return email

  // For long addresses, extract meaningful domain
  const atIndex = email.lastIndexOf('@')
  if (atIndex === -1) return truncate(email, 30)

  const local = email.slice(0, atIndex)
  const domain = email.slice(atIndex + 1)

  // Strip common relay subdomains: em7877.tm.openai.com → openai.com
  // tm1.openai.com → openai.com
  const domainParts = domain.split('.')
  const meaningfulDomain = extractMeaningfulDomain(domainParts)

  // If the local part is a recognizable name, show it
  const cleanLocal = cleanLocalPart(local)
  if (cleanLocal && cleanLocal.length <= 20) {
    return `${cleanLocal}@${meaningfulDomain}`
  }

  return meaningfulDomain
}

/**
 * Extracts the meaningful domain from a potentially nested subdomain.
 * em7877.tm.openai.com → openai.com
 * tm1.openai.com → openai.com
 * bounce.mail.example.com → example.com
 */
function extractMeaningfulDomain(parts) {
  if (parts.length <= 2) return parts.join('.')

  // Skip known relay/bounce subdomains
  const skipPrefixes = /^(em\d+|tm\d*|bounce|mail|relay|mta|mx|smtp|send|no-?reply|mailer)$/

  // Find the first non-relay subdomain from the right
  for (let i = parts.length - 2; i >= 0; i--) {
    if (!skipPrefixes.test(parts[i])) {
      // Return this level + TLD (up to 2 parts from here)
      const start = Math.max(0, i)
      return parts.slice(start).join('.')
    }
  }

  // Fallback: return last 2 parts
  return parts.slice(-2).join('.')
}

/**
 * Cleans up the local part of an email address.
 * bounces+20216706-bd67-luck=eoekun.top → ""
 * bounce+63ed33.b462b7-luck=eoekun.top → ""
 * noreply → "noreply"
 * support → "support"
 */
function cleanLocalPart(local) {
  // Remove bounce/relay prefixes
  if (/^(bounce|bounces|mailer-daemon)[+@]/i.test(local)) return ''
  // Remove long hash-like prefixes
  if (/^[a-f0-9]{8,}/i.test(local)) return ''
  // Remove addresses with +extension that's very long
  const plusIdx = local.indexOf('+')
  if (plusIdx > 0 && local.length - plusIdx > 20) {
    return local.slice(0, plusIdx)
  }
  return local
}

function truncate(str, max) {
  if (str.length <= max) return str
  return str.slice(0, max - 1) + '…'
}
