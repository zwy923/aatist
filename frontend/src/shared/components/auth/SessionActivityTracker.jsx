import { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import useAuthStore from "../../stores/authStore";
import {
  ACTIVITY_THROTTLE_MS,
  AUTH_ABSOLUTE_SESSION_MS,
  AUTH_IDLE_TIMEOUT_MS,
  SESSION_CHECK_INTERVAL_MS,
} from "../../config/authSessionPolicy";

function throttle(fn, ms) {
  let last = 0;
  return (...args) => {
    const now = Date.now();
    if (now - last >= ms) {
      last = now;
      fn(...args);
    }
  };
}

/**
 * Enforces idle timeout and absolute browser session length.
 * Backend still revokes via refresh/access JWT expiry.
 */
export default function SessionActivityTracker() {
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const accessToken = useAuthStore((s) => s.accessToken);
  const sessionStartedAt = useAuthStore((s) => s.sessionStartedAt);
  const ensureSessionAnchor = useAuthStore((s) => s.ensureSessionAnchor);

  const lastActivityRef = useRef(Date.now());

  useEffect(() => {
    if (!isAuthenticated || !accessToken) return;

    ensureSessionAnchor();
    lastActivityRef.current = Date.now();

    const bump = () => {
      lastActivityRef.current = Date.now();
    };
    const throttledBump = throttle(bump, ACTIVITY_THROTTLE_MS);
    const events = ["mousedown", "keydown", "touchstart", "scroll", "click"];
    events.forEach((ev) => window.addEventListener(ev, throttledBump, { passive: true }));

    const onVis = () => {
      if (document.visibilityState === "visible") bump();
    };
    document.addEventListener("visibilitychange", onVis);

    const tick = () => {
      const state = useAuthStore.getState();
      if (!state.isAuthenticated || !state.accessToken) return;

      const now = Date.now();
      const start = state.sessionStartedAt;
      if (start != null && now - start > AUTH_ABSOLUTE_SESSION_MS) {
        state.logout();
        navigate("/auth/login?reason=session_expired", { replace: true });
        return;
      }
      if (now - lastActivityRef.current > AUTH_IDLE_TIMEOUT_MS) {
        state.logout();
        navigate("/auth/login?reason=idle_timeout", { replace: true });
      }
    };

    const interval = setInterval(tick, SESSION_CHECK_INTERVAL_MS);

    return () => {
      events.forEach((ev) => window.removeEventListener(ev, throttledBump));
      document.removeEventListener("visibilitychange", onVis);
      clearInterval(interval);
    };
  }, [isAuthenticated, accessToken, navigate, ensureSessionAnchor]);

  return null;
}
