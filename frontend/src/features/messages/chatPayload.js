/** Prefix for chat messages that carry a file reference (stored as DB content text). */
export const CHAT_FILE_PREFIX = "__AATIST_FILE__";

/**
 * @returns {{ type: 'text', text: string } | { type: 'file', url: string, name: string, mime: string, caption: string }}
 */
export function parseChatPayload(raw) {
  const s = typeof raw === "string" ? raw : "";
  if (!s.startsWith(CHAT_FILE_PREFIX)) {
    return { type: "text", text: s };
  }
  try {
    const o = JSON.parse(s.slice(CHAT_FILE_PREFIX.length));
    if (o && typeof o.url === "string") {
      return {
        type: "file",
        url: o.url,
        name: o.name || "Attachment",
        mime: o.mime || "application/octet-stream",
        caption: typeof o.t === "string" ? o.t : "",
      };
    }
  } catch {
    /* ignore */
  }
  return { type: "text", text: s };
}

/** Short label for conversation list previews. */
export function formatChatPreview(raw) {
  const p = parseChatPayload(raw);
  if (p.type === "file") {
    if (p.caption) return p.caption;
    return `📎 ${p.name}`;
  }
  return p.text || "";
}
