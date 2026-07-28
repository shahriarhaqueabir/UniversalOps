import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Safely create a Date object, returning null for invalid inputs.
 * Handles null, undefined, empty strings, and non-parseable values.
 */
export function safeDate(value: string | number | Date | null | undefined): Date | null {
  if (value === null || value === undefined || value === '') return null
  try {
    const d = new Date(value)
    return isNaN(d.getTime()) ? null : d
  } catch {
    return null
  }
}

/**
 * Format a date safely, returning fallback for invalid inputs.
 *
 * @example
 *   formatSafeDate(c.last_run)                          // "7/27/2026, 10:36:00 AM" or ""
 *   formatSafeDate(c.last_run, (d) => d.toLocaleTimeString()) // "10:36:00 AM" or ""
 *   formatSafeDate(v, (d) => format(d, 'HH:mm'))              // "10:36" or ""
 */
export function formatSafeDate(
  value: string | number | Date | null | undefined,
  formatter?: (d: Date) => string,
  fallback = ''
): string {
  const d = safeDate(value)
  if (!d) return fallback
  return formatter ? formatter(d) : d.toLocaleString()
}
