import { clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs) {
  return twMerge(clsx(inputs))
}

/**
 * Format an ISO date string or Date object into a human-readable date.
 * e.g. "15 Jan 2023"
 * Returns "—" for falsy values.
 */
export function formatDate(value) {
  if (!value) return "—"
  const d = new Date(value)
  if (isNaN(d.getTime())) return String(value)
  return d.toLocaleDateString("en-GB", { day: "numeric", month: "short", year: "numeric" })
}

/**
 * Format an ISO datetime string or Date object into a human-readable datetime.
 * e.g. "15 Jan 2023, 14:30"
 * Returns "—" for falsy values.
 */
export function formatDateTime(value) {
  if (!value) return "—"
  const d = new Date(value)
  if (isNaN(d.getTime())) return String(value)
  return d.toLocaleString("en-GB", {
    day: "numeric", month: "short", year: "numeric",
    hour: "2-digit", minute: "2-digit",
  })
}
