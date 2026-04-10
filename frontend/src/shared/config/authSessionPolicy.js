/**
 * Browser session policy (defense in depth; backend JWT still authoritative).
 * Override via env: VITE_AUTH_IDLE_MINUTES, VITE_AUTH_SESSION_DAYS
 */

function parsePositiveInt(raw, fallback) {
  const n = Number.parseInt(String(raw), 10);
  return Number.isFinite(n) && n > 0 ? n : fallback;
}

/** No user input for 45 minutes → logout (shared / lab machines). */
export const AUTH_IDLE_TIMEOUT_MS =
  parsePositiveInt(import.meta.env.VITE_AUTH_IDLE_MINUTES, 45) * 60 * 1000;

/**
 * Max time logged in on this browser without signing in again.
 * Align with backend refresh_days (default 14); caps re-login even if refresh token were longer.
 */
export const AUTH_ABSOLUTE_SESSION_MS =
  parsePositiveInt(import.meta.env.VITE_AUTH_SESSION_DAYS, 14) * 24 * 60 * 60 * 1000;

/** How often we evaluate idle / absolute expiry */
export const SESSION_CHECK_INTERVAL_MS = 30 * 1000;

/** Throttle recording user activity (avoid main-thread work on every mousemove) */
export const ACTIVITY_THROTTLE_MS = 20 * 1000;
