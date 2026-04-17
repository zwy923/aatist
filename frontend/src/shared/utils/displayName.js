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

/**
 * Deduplicate display segments (e.g. school + faculty) when the API repeats the same string.
 */
export function dedupeStringsIgnoreCase(parts) {
  if (!Array.isArray(parts)) return [];
  const seen = new Set();
  const out = [];
  for (const p of parts) {
    const t = String(p ?? "").trim();
    if (!t) continue;
    const k = t.toLowerCase();
    if (seen.has(k)) continue;
    seen.add(k);
    out.push(t);
  }
  return out;
}
