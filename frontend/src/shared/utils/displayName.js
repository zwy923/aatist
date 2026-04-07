/**
 * Public-facing name for students in search, chat, and profile cards.
 * Prefer API field preferred_name; fall back to legal name.
 */
export function talentDisplayName(user) {
  if (!user) return "";
  const pref = String(user.preferred_name ?? user.preferredName ?? "").trim();
  const legal = String(user.name ?? "").trim();
  return pref || legal || "";
}
