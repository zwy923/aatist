import React, { createContext, useCallback, useContext, useMemo } from 'react';
import { useChatWebSocket } from './useChatWebSocket';
import { useAuth } from '../auth/hooks/useAuth';
import { useChatUnreadStore } from '../../shared/stores/chatUnreadStore';

const ChatContext = createContext(null);

export function ChatProvider({ children }) {
  const { accessToken, user } = useAuth();
  const myId = user?.id ?? user?.user_id;
  const ws = useChatWebSocket(accessToken, myId);
  const lastSeenByUser = useChatUnreadStore((s) => s.lastSeenByUser);
  const setLastSeen = useChatUnreadStore((s) => s.setLastSeen);

  const markConversationAsRead = useCallback(
    (conversationId, count) => {
      if (!myId || !conversationId) return;
      setLastSeen(myId, conversationId, count);
    },
    [myId, setLastSeen]
  );

  const getLastSeen = useCallback(
    (conversationId) => {
      const byUser = lastSeenByUser[myId] || {};
      return byUser[conversationId] || 0;
    },
    [myId, lastSeenByUser]
  );

  const totalUnreadCount = useMemo(() => {
    const byUser = lastSeenByUser[myId] || {};
    let total = 0;
    Object.entries(ws.messagesByConversation || {}).forEach(([convId, msgs]) => {
      const list = msgs || [];
      const seen = byUser[convId] || 0;
      const count = list.length;
      const diff = count >= seen ? count - seen : count;
      if (diff > 0) total += diff;
    });
    return total;
  }, [myId, lastSeenByUser, ws.messagesByConversation]);

  const value = useMemo(
    () => ({
      ...ws,
      totalUnreadCount,
      markConversationAsRead,
      getLastSeen,
    }),
    [ws, totalUnreadCount, markConversationAsRead, getLastSeen]
  );

  return <ChatContext.Provider value={value}>{children}</ChatContext.Provider>;
}

export function useChat() {
  const ctx = useContext(ChatContext);
  return ctx;
}
