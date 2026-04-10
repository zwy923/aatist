import { useCallback, useEffect, useRef, useState } from 'react';
import { CHAT_FILE_PREFIX } from './chatPayload';
import { useChatUnreadStore } from '../../shared/stores/chatUnreadStore';

const API_WS_BASE = (() => {
  const u = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';
  const base = u.replace(/^http/, 'ws').replace(/\/api\/v1\/?$/, '');
  return `${base}/api/v1`;
})();

/**
 * 协议：conversation_id = "userID1_userID2"（数字升序）
 * 服务端下推：{ type: "message" | "message_sent", id, conversation_id, from_user_id, content, created_at, temp_id? }
 * 客户端发送：{ type: "message", conversation_id, content, temp_id? }
 */
export function useChatWebSocket(accessToken, currentUserId) {
  const [connectionStatus, setConnectionStatus] = useState('closed'); // closed | connecting | open | error
  const [messagesByConversation, setMessagesByConversation] = useState({});
  const [onlineUserIds, setOnlineUserIds] = useState([]);
  const [typingByConversation, setTypingByConversation] = useState({});
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const shouldReconnectRef = useRef(true);
  const connectionIdRef = useRef(0);
  const currentUserIdStr = currentUserId != null ? String(currentUserId) : null;

  const appendMessage = useCallback((conversationId, msg) => {
    setMessagesByConversation((prev) => {
      const list = prev[conversationId] || [];
      if (list.some((m) => m.id === msg.id || m.temp_id === msg.temp_id)) return prev;
      return { ...prev, [conversationId]: [...list, msg] };
    });
  }, []);

  const handleMessage = useCallback(
    (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === 'message' || data.type === 'message_sent') {
          const conversationId = data.conversation_id;
          if (conversationId) {
            const fromId = data.from_user_id != null ? String(data.from_user_id) : '';
            const fromMe = fromId === currentUserIdStr;
            appendMessage(conversationId, {
              id: data.id || `msg-${Date.now()}`,
              from_user_id: fromId,
              fromMe,
              text: data.content,
              createdAt: data.created_at || new Date().toISOString(),
              temp_id: data.temp_id,
            });
            if (data.type === 'message' && !fromMe && fromId) {
              useChatUnreadStore.getState().bumpServerUnreadForConversation(conversationId);
            }
          }
        }
        if (data.type === 'online' && Array.isArray(data.online_user_ids)) {
          setOnlineUserIds(data.online_user_ids);
        }
        if (data.type === 'typing' && data.conversation_id) {
          const convId = data.conversation_id;
          const isOtherTyping =
            data.from_user_id &&
            data.from_user_id !== currentUserIdStr &&
            !!data.is_typing;
          setTypingByConversation((prev) => ({
            ...prev,
            [convId]: isOtherTyping,
          }));
        }
      } catch (e) {
        console.warn('chat ws parse error', e);
      }
    },
    [appendMessage, currentUserIdStr]
  );

  const handleMessageRef = useRef(handleMessage);
  handleMessageRef.current = handleMessage;

  const connect = useCallback(() => {
    if (!accessToken || !currentUserIdStr) {
      setConnectionStatus('closed');
      return;
    }
    if (wsRef.current?.readyState === WebSocket.OPEN) return;
    shouldReconnectRef.current = true;

    const url = `${API_WS_BASE}/ws?token=${encodeURIComponent(accessToken)}`;
    setConnectionStatus('connecting');
    connectionIdRef.current += 1;
    const thisConnectionId = connectionIdRef.current;
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      if (connectionIdRef.current === thisConnectionId) setConnectionStatus('open');
    };
    ws.onclose = () => {
      if (connectionIdRef.current !== thisConnectionId) return;
      wsRef.current = null;
      if (shouldReconnectRef.current) {
        setConnectionStatus('reconnecting');
        reconnectTimeoutRef.current = setTimeout(() => {
          if (accessToken && currentUserIdStr) connect();
        }, 5000);
      } else {
        setConnectionStatus('closed');
      }
    };
    ws.onerror = () => {
      if (connectionIdRef.current === thisConnectionId) setConnectionStatus('error');
    };
    ws.onmessage = (e) => handleMessageRef.current(e);
  }, [accessToken, currentUserIdStr]);

  const disconnect = useCallback(() => {
    shouldReconnectRef.current = false;
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setConnectionStatus('closed');
  }, []);

  useEffect(() => {
    connect();
    return () => {
      shouldReconnectRef.current = false;
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      setConnectionStatus('connecting');
    };
  }, [connect]);

  // 从后台切回前台时重连（笔记本休眠、移动端杀进程后 WS 常已死，仅靠轮询不够及时）
  useEffect(() => {
    const onVis = () => {
      if (document.visibilityState !== 'visible') return;
      if (!accessToken || !currentUserIdStr) return;
      const ws = wsRef.current;
      if (!ws || ws.readyState === WebSocket.CLOSED || ws.readyState === WebSocket.CLOSING) {
        connect();
      }
    };
    document.addEventListener('visibilitychange', onVis);
    return () => document.removeEventListener('visibilitychange', onVis);
  }, [accessToken, currentUserIdStr, connect]);

  const sendMessage = useCallback(
    (conversationId, content, tempId, attachment) => {
      const ws = wsRef.current;
      const trimmed = (content || '').trim();
      const hasFile =
        attachment &&
        typeof attachment.url === 'string' &&
        attachment.url.trim() !== '';
      if ((!trimmed && !hasFile) || !currentUserIdStr) return false;
      if (!ws || ws.readyState !== WebSocket.OPEN) return false;

      let payload = trimmed;
      if (hasFile) {
        const filePayload = {
          url: attachment.url.trim(),
          name: (attachment.filename || attachment.name || 'file').trim() || 'file',
          mime: (attachment.mime || attachment.content_type || 'application/octet-stream').trim(),
        };
        if (trimmed) filePayload.t = trimmed;
        payload = CHAT_FILE_PREFIX + JSON.stringify(filePayload);
      }
      if (payload.length > 60000) return false;

      // optimistic append for better UX, deduped by temp_id after server ack
      const optimisticTempID = tempId || `t-${Date.now()}`;
      appendMessage(conversationId, {
        id: optimisticTempID,
        temp_id: optimisticTempID,
        from_user_id: currentUserIdStr,
        fromMe: true,
        text: payload,
        createdAt: new Date().toISOString(),
      });
      ws.send(
        JSON.stringify({
          type: 'message',
          conversation_id: conversationId,
          content: payload,
          temp_id: optimisticTempID,
        })
      );
      return true;
    },
    [appendMessage, currentUserIdStr]
  );

  const getMessages = useCallback(
    (conversationId) => {
      return messagesByConversation[conversationId] || [];
    },
    [messagesByConversation]
  );

  const isUserOnline = useCallback(
    (userId) => {
      const id = userId != null ? String(userId) : '';
      return onlineUserIds.includes(id);
    },
    [onlineUserIds]
  );

  const sendTyping = useCallback((conversationId, isTyping) => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(
      JSON.stringify({
        type: 'typing',
        conversation_id: conversationId,
        is_typing: !!isTyping,
      })
    );
  }, []);

  const isTyping = useCallback(
    (conversationId) => !!typingByConversation[conversationId],
    [typingByConversation]
  );

  return {
    connectionStatus,
    sendMessage,
    getMessages,
    messagesByConversation,
    onlineUserIds,
    isUserOnline,
    sendTyping,
    isTyping,
    connect,
    disconnect,
  };
}

/**
 * 生成与后端一致的 conversation_id（两用户 ID 升序拼接）
 */
export function conversationId(userId1, userId2) {
  if (userId1 == null || userId2 == null) return '';
  const a = Number(userId1);
  const b = Number(userId2);
  if (a > b) return `${b}_${a}`;
  return `${a}_${b}`;
}
