import { create } from 'zustand';
import { persist } from 'zustand/middleware';

const STORAGE_KEY = 'chat_last_seen';

/**
 * Persists last-seen message count per conversation per user.
 * Storage key includes user id so we don't mix data across users.
 */
const storageKey = (userId) => `${STORAGE_KEY}_${userId || 'anon'}`;

/**
 * Build { [conversationId]: unreadCount } from GET /conversations response rows.
 */
export function unreadMapFromConversations(list) {
  const map = {};
  if (!Array.isArray(list)) return map;
  list.forEach((c) => {
    const id = c?.conversation_id != null ? String(c.conversation_id) : c?.id != null ? String(c.id) : '';
    if (!id) return;
    const n = Number(c.unread_count ?? c.unread ?? 0);
    map[id] = Number.isFinite(n) && n > 0 ? n : 0;
  });
  return map;
}

export const useChatUnreadStore = create(
  persist(
    (set, get) => ({
      lastSeenByUser: {}, // { [userId]: { [conversationId]: count } }
      /** Server-reported unread per conversation (not persisted; refreshed by polling). */
      serverUnreadByConversation: {},

      setServerUnreadMap: (map) => {
        set({ serverUnreadByConversation: map && typeof map === 'object' ? { ...map } : {} });
      },

      setServerUnreadForConversation: (conversationId, count) => {
        if (!conversationId) return;
        const id = String(conversationId);
        const n = Math.max(0, Number(count) || 0);
        set((state) => {
          const next = { ...state.serverUnreadByConversation };
          if (n === 0) delete next[id];
          else next[id] = n;
          return { serverUnreadByConversation: next };
        });
      },

      bumpServerUnreadForConversation: (conversationId) => {
        if (!conversationId) return;
        const id = String(conversationId);
        set((state) => ({
          serverUnreadByConversation: {
            ...state.serverUnreadByConversation,
            [id]: (Number(state.serverUnreadByConversation[id]) || 0) + 1,
          },
        }));
      },

      setLastSeen: (userId, conversationId, count) => {
        if (!userId || !conversationId) return;
        set((state) => {
          const byUser = { ...state.lastSeenByUser };
          const byConv = { ...(byUser[userId] || {}) };
          const prev = byConv[conversationId] || 0;
          if (count > prev) {
            byConv[conversationId] = count;
            byUser[userId] = byConv;
            return { lastSeenByUser: byUser };
          }
          return state;
        });
      },

      getLastSeen: (userId, conversationId) => {
        const byUser = get().lastSeenByUser;
        const byConv = byUser[userId] || {};
        return byConv[conversationId] || 0;
      },

      removeConversation: (userId, conversationId) => {
        if (!userId || !conversationId) return;
        const cid = String(conversationId);
        set((state) => {
          const byUser = { ...state.lastSeenByUser };
          const byConv = { ...(byUser[userId] || {}) };
          delete byConv[conversationId];
          delete byConv[cid];
          byUser[userId] = byConv;
          const serverNext = { ...state.serverUnreadByConversation };
          delete serverNext[cid];
          return { lastSeenByUser: byUser, serverUnreadByConversation: serverNext };
        });
      },

      clearServerUnreadState: () => set({ serverUnreadByConversation: {} }),
    }),
    {
      name: 'chat-unread',
      partialize: (s) => ({ lastSeenByUser: s.lastSeenByUser }),
    }
  )
);
