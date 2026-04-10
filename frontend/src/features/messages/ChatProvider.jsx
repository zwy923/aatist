import React, { createContext, useCallback, useContext, useEffect, useMemo } from 'react';
import { useChatWebSocket } from './useChatWebSocket';
import { useAuth } from '../auth/hooks/useAuth';
import { messagesApi } from './api/messages';
import { useChatUnreadStore, unreadMapFromConversations } from '../../shared/stores/chatUnreadStore';

const ChatContext = createContext(null);

export function ChatProvider({ children }) {
  const { accessToken, user } = useAuth();
  const myId = user?.id ?? user?.user_id;
  const ws = useChatWebSocket(accessToken, myId);
  const lastSeenByUser = useChatUnreadStore((s) => s.lastSeenByUser);
  const serverUnreadByConversation = useChatUnreadStore((s) => s.serverUnreadByConversation);
  const setLastSeen = useChatUnreadStore((s) => s.setLastSeen);
  const setServerUnreadMap = useChatUnreadStore((s) => s.setServerUnreadMap);

  /** 全局轮询未读：导航栏红点依赖服务端 unread_count，不能只在 /messages 页拉取 */
  useEffect(() => {
    if (!accessToken || !myId) return;
    let cancelled = false;
    const syncUnreadFromServer = async () => {
      try {
        const res = await messagesApi.getConversations({ limit: 50 });
        if (cancelled) return;
        const list = res.data?.data?.conversations || [];
        setServerUnreadMap(unreadMapFromConversations(list));
      } catch {
        /* ignore */
      }
    };
    syncUnreadFromServer();
    const interval = setInterval(syncUnreadFromServer, 10000);
    const onVis = () => {
      if (document.visibilityState === "visible") syncUnreadFromServer();
    };
    document.addEventListener("visibilitychange", onVis);
    return () => {
      cancelled = true;
      clearInterval(interval);
      document.removeEventListener("visibilitychange", onVis);
    };
  }, [accessToken, myId, setServerUnreadMap]);

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
    const wsMsgs = ws.messagesByConversation || {};
    const ids = new Set([
      ...Object.keys(serverUnreadByConversation || {}),
      ...Object.keys(wsMsgs),
    ]);
    let total = 0;
    ids.forEach((convId) => {
      const apiN = Number(serverUnreadByConversation[convId]) || 0;
      const list = wsMsgs[convId] || [];
      const seen = byUser[convId] || 0;
      const wsUnread = Math.max(0, list.length - seen);
      total += Math.max(apiN, wsUnread);
    });
    return total;
  }, [myId, lastSeenByUser, ws.messagesByConversation, serverUnreadByConversation]);

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
