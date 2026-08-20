// Formatting helpers shared by the UI. Kept dependency-free and pure so they
// are easy to unit test.

/**
 * Turns a `YYYY-MM-DD` date string into a long, human-readable label in the
 * viewer's locale, e.g. "Monday, August 17, 2026".
 */
export function formatDate(date: string): string {
  return new Date(`${date}T00:00:00`).toLocaleDateString(undefined, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}
