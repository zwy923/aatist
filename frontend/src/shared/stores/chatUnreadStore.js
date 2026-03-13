import { create } from 'zustand';
import { persist } from 'zustand/middleware';

const STORAGE_KEY = 'chat_last_seen';

/**
 * Persists last-seen message count per conversation per user.
 * Storage key includes user id so we don't mix data across users.
 */
const storageKey = (userId) => `${STORAGE_KEY}_${userId || 'anon'}`;

export const useChatUnreadStore = create(
  persist(
    (set, get) => ({
      lastSeenByUser: {}, // { [userId]: { [conversationId]: count } }

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
        set((state) => {
          const byUser = { ...state.lastSeenByUser };
          const byConv = { ...(byUser[userId] || {}) };
          delete byConv[conversationId];
          byUser[userId] = byConv;
          return { lastSeenByUser: byUser };
        });
      },
    }),
    {
      name: 'chat-unread',
      partialize: (s) => ({ lastSeenByUser: s.lastSeenByUser }),
    }
  )
);
