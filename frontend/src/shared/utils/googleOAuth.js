/** Full URL to start Google OAuth (backend redirects to Google, then to callback). */
export function getGoogleOAuthStartURL() {
  const base = import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1";
  return `${base.replace(/\/$/, "")}/auth/oauth/google`;
}
