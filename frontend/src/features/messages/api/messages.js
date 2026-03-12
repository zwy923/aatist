import apiClient from '../../../shared/api/client';

export const messagesApi = {
  /** 获取当前用户的会话列表 */
  getConversations: (params) =>
    apiClient.get('/conversations', { params }),

  /** 获取指定会话的消息历史 */
  getMessages: (conversationId, params) =>
    apiClient.get(`/conversations/${encodeURIComponent(conversationId)}/messages`, { params }),

  /** 发起新会话，返回 conversation_id（首次发消息前调用） */
  startConversation: (otherUserId) =>
    apiClient.post('/conversations/start', { other_user_id: otherUserId }),
};
