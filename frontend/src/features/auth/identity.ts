import type { User } from '@/types/auth'

/** The name to show for a user: their display name, else the email local-part. */
export function displayNameOf(user: Pick<User, 'displayName' | 'email'>): string {
  const name = user.displayName?.trim()
  if (name) return name
  return user.email.split('@')[0]
}

/** One- or two-letter initials for an avatar, derived from name or email. */
export function initialsOf(user: Pick<User, 'displayName' | 'email'>): string {
  const source = displayNameOf(user)
  const parts = source.split(/[\s._-]+/).filter(Boolean)
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase()
  }
  return source.slice(0, 2).toUpperCase()
}
