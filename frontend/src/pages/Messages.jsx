import React, { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { Link as RouterLink, useNavigate, useSearchParams } from "react-router-dom";
import {
  Avatar,
  Badge,
  Box,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Divider,
  IconButton,
  InputBase,
  List,
  ListItem,
  ListItemAvatar,
  ListItemButton,
  ListItemText,
  Menu,
  MenuItem,
  Paper,
  Stack,
  Typography,
  useTheme,
  Tooltip,
  TextField,
  Button,
  Link,
  Snackbar,
} from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import AddIcon from "@mui/icons-material/Add";
import AttachFileIcon from "@mui/icons-material/AttachFile";
import SendIcon from "@mui/icons-material/Send";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import DeleteIcon from "@mui/icons-material/Delete";
import PageLayout from "../shared/components/PageLayout";
import { useAuth } from "../features/auth/hooks/useAuth";
import { useChat } from "../features/messages/ChatProvider";
import { conversationId } from "../features/messages/useChatWebSocket";
import { useChatUnreadStore, unreadMapFromConversations } from "../shared/stores/chatUnreadStore";
import { messagesApi } from "../features/messages/api/messages";
import { formatChatPreview, parseChatPayload } from "../features/messages/chatPayload";
import { talentDisplayName } from "../shared/utils/displayName";
import { profileApi } from "../features/profile/api/profile";
import apiClient from "../shared/api/client";

const MAX_CHAT_ATTACHMENT_BYTES = 100 * 1024 * 1024;

function ChatMessageContent({ text, fromMe }) {
  const parsed = parseChatPayload(text);
  if (parsed.type === "text") {
    return (
      <Typography variant="body2" sx={{ whiteSpace: "pre-wrap", wordBreak: "break-word" }}>
        {parsed.text}
      </Typography>
    );
  }
  const isImg = parsed.mime.startsWith("image/");
  const linkColor = fromMe ? "rgba(255,255,255,0.95)" : "primary.main";
  return (
    <Stack spacing={1}>
      {parsed.caption ? (
        <Typography variant="body2" sx={{ whiteSpace: "pre-wrap", wordBreak: "break-word" }}>
          {parsed.caption}
        </Typography>
      ) : null}
      {isImg ? (
        <Box
          component="a"
          href={parsed.url}
          target="_blank"
          rel="noopener noreferrer"
          sx={{ display: "block", maxWidth: 280 }}
        >
          <Box
            component="img"
            src={parsed.url}
            alt={parsed.name}
            sx={{ maxWidth: "100%", borderRadius: 2, display: "block" }}
          />
        </Box>
      ) : (
        <Link
          href={parsed.url}
          target="_blank"
          rel="noopener noreferrer"
          underline="hover"
          sx={{
            color: linkColor,
            fontWeight: 600,
            wordBreak: "break-all",
            fontSize: 14,
          }}
        >
          📎 {parsed.name}
        </Link>
      )}
    </Stack>
  );
}

const OnlineBadge = (props) => (
  <Badge
    overlap="circular"
    variant="dot"
    anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
    sx={{
      "& .MuiBadge-badge": {
        backgroundColor: "#3ddc97",
        color: "#3ddc97",
        boxShadow: "0 0 0 2px #ffffff",
      },
    }}
    {...props}
  />
);

const MessageBubble = ({ fromMe, text, timestamp }) => {
  const theme = useTheme();
  const align = fromMe ? "flex-end" : "flex-start";
  const bg = fromMe
    ? "linear-gradient(135deg, #2563eb, #60a5fa)"
    : "#f8fafc";
  const color = fromMe ? "#f9fafb" : theme.palette.text.primary;

  return (
    <Stack spacing={0.5} alignItems={align}>
      <Box
        sx={{
          maxWidth: "70%",
          px: 2,
          py: 1.5,
          borderRadius: fromMe ? "18px 18px 4px 18px" : "18px 18px 18px 4px",
          background: bg,
          color,
          boxShadow: fromMe
            ? "0 10px 30px rgba(37,99,235,0.45)"
            : "0 8px 20px rgba(15,23,42,0.08)",
          border: fromMe
            ? "1px solid rgba(191,219,254,0.35)"
            : "1px solid #e2e8f0",
        }}
      >
        <ChatMessageContent text={text} fromMe={fromMe} />
      </Box>
      {timestamp && (
        <Typography
          variant="caption"
          sx={{ color: "text.secondary", mt: 0.25 }}
        >
          {timestamp}
        </Typography>
      )}
    </Stack>
  );
};

const MessagesPage = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { user, isAuthenticated } = useAuth();
  const myId = user?.id ?? user?.user_id;
  const roleLower = user?.role?.toLowerCase?.();
  const isStudentRole = roleLower === "student" || roleLower === "alumni";
  /** 项目议价提示仅面向客户（组织方），学生端不显示 */
  const showClientProjectBanner = !isStudentRole;
  const chat = useChat();
  const {
    connectionStatus = "closed",
    sendMessage = () => false,
    getMessages = () => [],
    isUserOnline = () => false,
    sendTyping = () => {},
    isTyping = () => false,
    messagesByConversation = {},
    markConversationAsRead = () => {},
  } = chat || {};
  const lastSeenByUser = useChatUnreadStore((s) => s.lastSeenByUser);
  const lastSeenCount = (lastSeenByUser[myId] || {});
  const removeConversationFromStore = useChatUnreadStore((s) => s.removeConversation);

  const [search, setSearch] = useState("");
  const [conversationsFromApi, setConversationsFromApi] = useState([]);
  const [loadingConversations, setLoadingConversations] = useState(false);
  const [newConvDialogOpen, setNewConvDialogOpen] = useState(false);
  const [userSearchQuery, setUserSearchQuery] = useState("");
  const [userSearchResults, setUserSearchResults] = useState([]);
  const [userSearchLoading, setUserSearchLoading] = useState(false);
  const [historyByConversation, setHistoryByConversation] = useState({});
  const [loadingHistory, setLoadingHistory] = useState(false);
  const [selectedId, setSelectedId] = useState(null);
  const [drafts, setDrafts] = useState({});
  const [moreMenuAnchor, setMoreMenuAnchor] = useState(null);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const chatFileInputRef = useRef(null);
  const messagesScrollRef = useRef(null);
  const [attachUploading, setAttachUploading] = useState(false);
  const [toast, setToast] = useState({ open: false, message: "" });

  const loadConversations = useCallback(async () => {
    if (!isAuthenticated) return;
    setLoadingConversations(true);
    try {
      const res = await messagesApi.getConversations({ limit: 50 });
      const list = res.data?.data?.conversations || [];
      const enriched = list.map((c) => {
        const name = c.other_user_name || `User ${c.other_user_id}`;
        const subtitle = c.organization_name || "";
        return {
          id: c.conversation_id,
          conversation_id: c.conversation_id,
          otherUserId: c.other_user_id,
          otherUserAvatar: c.other_user_avatar,
          name,
          subtitle,
          lastMessage: formatChatPreview(c.last_message || ""),
          updatedAt: c.last_at ? new Date(c.last_at).toLocaleDateString(undefined, { month: "short", day: "numeric" }) : "",
          updatedAtTs: c.last_at ? new Date(c.last_at).getTime() : 0,
          unread: c.unread_count ?? 0,
        };
      });
      setConversationsFromApi(enriched);
      useChatUnreadStore.getState().setServerUnreadMap(unreadMapFromConversations(enriched));
      setSelectedId((prev) => {
        if (enriched.length === 0) return prev;
        const hasPrev = enriched.some((c) => c.id === prev);
        return hasPrev ? prev : enriched[0].id;
      });
    } catch (e) {
      console.warn("Load conversations failed", e);
    } finally {
      setLoadingConversations(false);
    }
  }, [isAuthenticated]);

  const startConversationWithUser = useCallback(async (otherUserId) => {
    if (!isAuthenticated || !myId || !otherUserId) return;
    setLoadingConversations(true);
    try {
      const res = await messagesApi.startConversation(otherUserId);
      const convId = res.data?.data?.conversation_id;
      if (!convId) return;
      let name = `User ${otherUserId}`;
      let subtitle = "";
      try {
        const u = await profileApi.getPublicProfile(otherUserId);
        const d = u.data?.data;
        if (d) {
          name = talentDisplayName(d) || d.username || name;
          subtitle = d.organization_name || d.school || subtitle;
        }
      } catch (_) {}
      const newConv = {
        id: convId,
        conversation_id: convId,
        otherUserId: Number(otherUserId),
        name,
        subtitle,
        lastMessage: "",
        updatedAt: "",
        updatedAtTs: 0,
        unread: 0,
      };
      setConversationsFromApi((prev) => {
        const exists = prev.some((c) => c.conversation_id === convId);
        if (exists) return prev;
        return [newConv, ...prev];
      });
      setSelectedId(convId);
      try {
        const raw = sessionStorage.getItem("aatist_chat_prefill");
        if (raw) {
          const p = JSON.parse(raw);
          if (Number(p.forUser) === Number(otherUserId) && typeof p.text === "string") {
            setDrafts((prev) => ({ ...prev, [convId]: p.text }));
            sessionStorage.removeItem("aatist_chat_prefill");
          }
        }
      } catch (_) {
        /* ignore */
      }
    } catch (e) {
      console.warn("Start conversation failed", e);
    } finally {
      setLoadingConversations(false);
    }
  }, [isAuthenticated, myId]);

  useEffect(() => {
    loadConversations();
  }, [loadConversations]);

  const searchUsersForNewConv = useCallback(async () => {
    if (!isAuthenticated) return;
    setUserSearchLoading(true);
    try {
      const res = await profileApi.searchUsers({
        q: userSearchQuery.trim() || undefined,
        limit: 30,
      });
      const list = res.data?.data || [];
      const filtered = list.filter((u) => u.id && Number(u.id) !== Number(myId));
      setUserSearchResults(filtered);
    } catch (e) {
      console.warn("Search users failed", e);
      setUserSearchResults([]);
    } finally {
      setUserSearchLoading(false);
    }
  }, [isAuthenticated, userSearchQuery, myId]);

  useEffect(() => {
    if (!newConvDialogOpen) return;
    const timer = setTimeout(() => searchUsersForNewConv(), 300);
    return () => clearTimeout(timer);
  }, [newConvDialogOpen, userSearchQuery, searchUsersForNewConv]);

  const handleOpenNewConvDialog = () => {
    setUserSearchQuery("");
    setUserSearchResults([]);
    setNewConvDialogOpen(true);
  };

  const handleSelectUserForNewConv = async (otherUserId) => {
    setNewConvDialogOpen(false);
    await startConversationWithUser(Number(otherUserId));
  };

  const handleOpenDeleteConfirm = () => {
    setMoreMenuAnchor(null);
    setDeleteConfirmOpen(true);
  };

  const targetUserId = searchParams.get("user");
  useEffect(() => {
    if (targetUserId && isAuthenticated) {
      startConversationWithUser(Number(targetUserId));
    }
  }, [targetUserId, isAuthenticated, startConversationWithUser]);

  const conversationList = useMemo(() => {
    const byID = new Map(
      conversationsFromApi.map((c) => [c.conversation_id || c.id, { ...c }])
    );

    Object.entries(messagesByConversation).forEach(([convID, msgs]) => {
      if (!Array.isArray(msgs) || msgs.length === 0) return;
      const last = msgs[msgs.length - 1];
      const lastText = last?.text || last?.content || "";
      const lastAtRaw = last?.createdAt || last?.created_at || new Date().toISOString();
      const lastAt = new Date(lastAtRaw);
      const formatted = Number.isNaN(lastAt.getTime())
        ? ""
        : lastAt.toLocaleDateString(undefined, { month: "short", day: "numeric" });
      const timestamp = Number.isNaN(lastAt.getTime()) ? 0 : lastAt.getTime();

      if (!byID.has(convID)) {
        const [a, b] = convID.split("_").map((x) => Number(x));
        const otherUserID =
          Number.isFinite(a) && Number.isFinite(b)
            ? (a === Number(myId) ? b : a)
            : null;

        byID.set(convID, {
          id: convID,
          conversation_id: convID,
          otherUserId: otherUserID,
          name: otherUserID ? `User ${otherUserID}` : "Unknown user",
          subtitle: "",
          lastMessage: formatChatPreview(lastText),
          updatedAt: formatted,
          updatedAtTs: timestamp,
          unread: 0,
        });
        return;
      }

      const prev = byID.get(convID);
      byID.set(convID, {
        ...prev,
        lastMessage: lastText ? formatChatPreview(lastText) : prev.lastMessage,
        updatedAt: formatted || prev.updatedAt,
        updatedAtTs: timestamp || prev.updatedAtTs || 0,
      });
    });

    return Array.from(byID.values()).sort((a, b) => (b.updatedAtTs || 0) - (a.updatedAtTs || 0));
  }, [conversationsFromApi, messagesByConversation, myId]);

  useEffect(() => {
    if (conversationList.length > 0 && selectedId == null) {
      setSelectedId(conversationList[0].id);
    }
  }, [conversationList.length, selectedId]);

  const loadedHistoryRef = useRef({});
  const loadHistory = useCallback(async (convId) => {
    if (!convId || loadedHistoryRef.current[convId]) return;
    loadedHistoryRef.current[convId] = true;
    setLoadingHistory(true);
    try {
      const res = await messagesApi.getMessages(convId, { limit: 50 });
      const list = res.data?.data?.messages || [];
      const normalized = list.map((m) => ({
        id: String(m.id),
        from_user_id: String(m.from_user_id),
        fromMe:
          myId != null &&
          Number(m.from_user_id) === Number(myId),
        text: m.content,
        createdAt: m.created_at,
      }));
      setHistoryByConversation((prev) => ({ ...prev, [convId]: normalized }));
    } catch (e) {
      console.warn("Load history failed", e);
      delete loadedHistoryRef.current[convId];
    } finally {
      setLoadingHistory(false);
    }
  }, [myId]);

  const conversations = useMemo(() => {
    if (!search.trim()) return conversationList;
    const term = search.toLowerCase();
    return conversationList.filter(
      (c) =>
        (c.name || "").toLowerCase().includes(term) ||
        (c.subtitle || "").toLowerCase().includes(term)
    );
  }, [search, conversationList]);

  const activeConversation = useMemo(
    () => conversationList.find((c) => c.id === selectedId) || null,
    [conversationList, selectedId]
  );

  const activeConversationId = useMemo(
    () =>
      myId != null && activeConversation?.otherUserId != null
        ? conversationId(myId, activeConversation.otherUserId)
        : activeConversation?.id ?? "",
    [myId, activeConversation]
  );

  const handleDeleteConversation = useCallback(async () => {
    if (!activeConversationId || !isAuthenticated) return;
    setDeleting(true);
    try {
      await messagesApi.deleteConversation(activeConversationId);
      removeConversationFromStore(myId, activeConversationId);
      setHistoryByConversation((prev) => {
        const next = { ...prev };
        delete next[activeConversationId];
        return next;
      });
      await loadConversations();
      setDeleteConfirmOpen(false);
    } catch (e) {
      console.warn("Delete conversation failed", e);
    } finally {
      setDeleting(false);
    }
  }, [activeConversationId, isAuthenticated, myId, removeConversationFromStore, loadConversations]);

  useEffect(() => {
    if (activeConversationId) loadHistory(activeConversationId);
  }, [activeConversationId, loadHistory]);

  /** 从服务端重新拉取当前会话历史（WebSocket 未送达时的兜底，与轮询配合） */
  const refetchActiveHistory = useCallback(async () => {
    const convId = activeConversationId;
    if (!convId || !isAuthenticated) return;
    try {
      const res = await messagesApi.getMessages(convId, { limit: 50 });
      const list = res.data?.data?.messages || [];
      const normalized = list.map((m) => ({
        id: String(m.id),
        from_user_id: String(m.from_user_id),
        fromMe: myId != null && Number(m.from_user_id) === Number(myId),
        text: m.content,
        createdAt: m.created_at,
      }));
      setHistoryByConversation((prev) => ({ ...prev, [convId]: normalized }));
      loadedHistoryRef.current[convId] = true;
    } catch (e) {
      console.warn("Refetch chat history failed", e);
    }
  }, [activeConversationId, isAuthenticated, myId]);

  useEffect(() => {
    if (!isAuthenticated) return;
    const tick = () => {
      if (document.visibilityState !== "visible") return;
      loadConversations();
      refetchActiveHistory();
    };
    const onVis = () => {
      if (document.visibilityState === "visible") tick();
    };
    document.addEventListener("visibilitychange", onVis);
    const interval = setInterval(tick, 10000);
    return () => {
      document.removeEventListener("visibilitychange", onVis);
      clearInterval(interval);
    };
  }, [isAuthenticated, loadConversations, refetchActiveHistory]);

  const wsMessages = getMessages(activeConversationId);
  const historyMessages = activeConversationId ? (historyByConversation[activeConversationId] || []) : [];
  const mergedMessages = useMemo(() => {
    const seen = new Set();
    const out = [];
    const dedupeKey = (m) =>
      `${m.from_user_id ?? (m.fromMe ? myId : '')}|${(m.text || m.content || '').slice(0, 80)}|${Math.floor(new Date(m.createdAt || m.created_at || 0).getTime() / 10000)}`;
    historyMessages.forEach((m) => {
      const key = dedupeKey(m);
      if (seen.has(key)) return;
      seen.add(key);
      seen.add(String(m.id));
      out.push({ ...m, text: m.text ?? m.content, createdAt: m.createdAt ?? m.created_at });
    });
    wsMessages.forEach((m) => {
      if (seen.has(m.id) || seen.has(m.temp_id)) return;
      const key = dedupeKey(m);
      if (seen.has(key)) return;
      seen.add(key);
      seen.add(m.id || m.temp_id);
      out.push({ ...m, text: m.text ?? m.content, createdAt: m.createdAt ?? m.created_at });
    });
    return out.sort((a, b) => new Date(a.createdAt || 0) - new Date(b.createdAt || 0));
  }, [historyMessages, wsMessages, myId]);

  const messages = mergedMessages;

  const scrollChatToBottom = useCallback(() => {
    const node = messagesScrollRef.current;
    if (!node) return;
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        const el = messagesScrollRef.current;
        if (el) el.scrollTop = el.scrollHeight;
      });
    });
  }, []);

  /** 进入会话 / 历史加载完成后滚到最新消息；双 rAF 等待布局与图片 */
  useLayoutEffect(() => {
    if (!activeConversationId || loadingHistory) return;
    scrollChatToBottom();
  }, [activeConversationId, loadingHistory, scrollChatToBottom]);

  /** 同一会话内新消息：仅当用户已在底部附近时才跟随滚动 */
  useLayoutEffect(() => {
    if (!activeConversationId || loadingHistory) return;
    const node = messagesScrollRef.current;
    if (!node || mergedMessages.length === 0) return;
    const threshold = 120;
    const nearBottom =
      node.scrollHeight - node.scrollTop - node.clientHeight < threshold;
    if (nearBottom) scrollChatToBottom();
  }, [mergedMessages.length, activeConversationId, loadingHistory, scrollChatToBottom]);

  useEffect(() => {
    if (!activeConversationId || !mergedMessages.length) return;
    markConversationAsRead(activeConversationId, mergedMessages.length);
    useChatUnreadStore.getState().setServerUnreadForConversation(activeConversationId, 0);
  }, [activeConversationId, mergedMessages.length, markConversationAsRead]);

  useEffect(() => {
    if (!activeConversationId || !isAuthenticated) return;
    messagesApi.markConversationAsRead(activeConversationId).catch(() => {});
    useChatUnreadStore.getState().setServerUnreadForConversation(activeConversationId, 0);
  }, [activeConversationId, isAuthenticated]);

  const unreadByConversation = useMemo(() => {
    const out = {};
    const allConvIds = new Set([
      ...Object.keys(messagesByConversation || {}),
      ...Object.keys(historyByConversation || {}),
      ...conversationList.map((c) => c.conversation_id || c.id).filter(Boolean),
    ]);
    allConvIds.forEach((convId) => {
      const total =
        convId === activeConversationId
          ? mergedMessages.length
          : (messagesByConversation[convId] || []).length +
            (historyByConversation[convId] || []).length;
      const seen = lastSeenCount[convId] || 0;
      const diff = Math.max(0, total - seen);
      if (diff > 0) out[convId] = diff;
    });
    return out;
  }, [
    messagesByConversation,
    historyByConversation,
    conversationList,
    activeConversationId,
    mergedMessages.length,
    lastSeenCount,
  ]);

  const handleSend = () => {
    const text = drafts[selectedId]?.trim();
    if (!text) return;
    if (activeConversationId && sendMessage(activeConversationId, text, `t-${Date.now()}`, null)) {
      setDrafts((prev) => ({ ...prev, [selectedId]: "" }));
    }
  };

  const handleChatFileChange = async (event) => {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file || !activeConversationId || !isAuthenticated) return;
    if (file.size > MAX_CHAT_ATTACHMENT_BYTES) {
      setToast({ open: true, message: "File must be 100MB or smaller." });
      return;
    }
    setAttachUploading(true);
    try {
      const formData = new FormData();
      formData.append("file", file);
      const res = await apiClient.post("/files/upload?type=other", formData);
      const data = res.data?.data;
      const url = data?.url;
      if (!url) throw new Error("Upload did not return a file URL.");
      const caption = (drafts[selectedId] || "").trim();
      const tempId = `t-${Date.now()}`;
      const sent = sendMessage(activeConversationId, caption, tempId, {
        url,
        filename: data?.filename || file.name,
        mime: data?.content_type || file.type || "application/octet-stream",
      });
      if (sent) {
        setDrafts((prev) => ({ ...prev, [selectedId]: "" }));
      } else {
        setToast({ open: true, message: "Could not send — check your chat connection." });
      }
    } catch (err) {
      const wrapped = err?.response?.data?.error?.message;
      const msg =
        typeof wrapped === "string" && wrapped.trim()
          ? wrapped
          : err?.message || "Upload failed.";
      setToast({ open: true, message: msg });
    } finally {
      setAttachUploading(false);
    }
  };

  const handleKeyDown = (event) => {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      handleSend();
    }
  };

  const onlineForConv = activeConversation?.otherUserId != null
    ? isUserOnline(activeConversation.otherUserId)
    : false;

  const typingTimeoutRef = useRef(null);

  const notifyTyping = useCallback(() => {
    if (!isAuthenticated || !activeConversationId) return;
    sendTyping(activeConversationId, true);
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }
    typingTimeoutRef.current = setTimeout(() => {
      sendTyping(activeConversationId, false);
    }, 3000);
  }, [activeConversationId, isAuthenticated, sendTyping]);

  if (!isAuthenticated) {
    return (
      <PageLayout maxWidth="xl" variant="light">
        <Box sx={{ py: 8, textAlign: "center" }}>
          <Typography variant="h6" color="text.secondary" gutterBottom>
            请先登录以查看消息
          </Typography>
          <Button variant="contained" onClick={() => navigate("/auth/login")}>
            去登录
          </Button>
        </Box>
      </PageLayout>
    );
  }

  return (
    <>
    <PageLayout maxWidth="xl" variant="light">
      <Stack
        direction={{ xs: "column", md: "row" }}
        spacing={3}
        sx={{ height: { md: "75vh", xs: "auto" } }}
      >
        {/* 左侧会话列表 */}
        <Paper
          elevation={0}
          sx={{
            width: { xs: "100%", md: 340 },
            flexShrink: 0,
            borderRadius: 3,
            border: "1px solid #e5e7eb",
            background: "#ffffff",
            display: "flex",
            flexDirection: "column",
            overflow: "hidden",
          }}
        >
          <Box sx={{ p: 2.5, pb: 1.5 }}>
            <Box sx={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 1, mb: 0.5 }}>
              <Typography variant="h6" fontWeight={600}>
                Messages
              </Typography>
              <Stack direction="row" spacing={1} alignItems="center">
                {isAuthenticated && (
                  <Tooltip title="New conversation">
                    <IconButton
                      size="small"
                      onClick={handleOpenNewConvDialog}
                      sx={{ color: "primary.main" }}
                    >
                      <AddIcon />
                    </IconButton>
                  </Tooltip>
                )}
                {isAuthenticated && (
                <Tooltip
                  title={
                    connectionStatus === "open"
                      ? "Connected"
                      : connectionStatus === "connecting"
                        ? "Connecting…"
                        : connectionStatus === "reconnecting"
                          ? "Reconnecting…"
                          : connectionStatus === "error"
                            ? "Connection error"
                            : "Disconnected"
                  }
                >
                  <Box
                    sx={{
                      width: 10,
                      height: 10,
                      borderRadius: "50%",
                      bgcolor:
                        connectionStatus === "open"
                          ? "#22c55e"
                          : connectionStatus === "connecting" || connectionStatus === "reconnecting"
                            ? "#eab308"
                            : connectionStatus === "error"
                              ? "#ef4444"
                              : "#94a3b8",
                      boxShadow: (theme) =>
                        connectionStatus === "open"
                          ? `0 0 8px ${theme.palette.success.main}80`
                          : connectionStatus === "connecting" || connectionStatus === "reconnecting"
                            ? `0 0 8px #eab30880`
                            : connectionStatus === "error"
                              ? `0 0 8px ${theme.palette.error.main}80`
                              : "none",
                      animation:
                        connectionStatus === "connecting" || connectionStatus === "reconnecting"
                          ? "pulse 1.5s ease-in-out infinite"
                          : "none",
                      "@keyframes pulse": {
                        "0%, 100%": { opacity: 1 },
                        "50%": { opacity: 0.5 },
                      },
                    }}
                  />
                </Tooltip>
                )}
              </Stack>
            </Box>
            <Typography variant="body2" color="text.secondary">
              Follow up with your clients and collaborators
            </Typography>
            <Paper
              sx={{
                mt: 2,
                px: 1.5,
                py: 0.75,
                display: "flex",
                alignItems: "center",
                gap: 1,
                borderRadius: 999,
                backgroundColor: "#f8fafc",
                border: "1px solid #e2e8f0",
              }}
              variant="outlined"
            >
              <SearchIcon sx={{ fontSize: 18, color: "text.secondary" }} />
              <InputBase
                placeholder="Search by name or organization"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                sx={{ fontSize: 14, flex: 1 }}
              />
            </Paper>
          </Box>

          <Divider sx={{ borderColor: "#e5e7eb" }} />

          <Box sx={{ flex: 1, overflow: "auto", p: 1.5 }}>
            {loadingConversations ? (
              <Stack alignItems="center" justifyContent="center" py={4}>
                <CircularProgress size={28} />
              </Stack>
            ) : (
            <List dense disablePadding>
              {conversations.length === 0 && (
                <Box sx={{ px: 2, py: 4, textAlign: "center" }}>
                  <Typography variant="body2" color="text.secondary">
                    No conversations yet. Open a user profile and click message to start chatting.
                  </Typography>
                </Box>
              )}
              {conversations.map((c) => {
                const isActive = c.id === selectedId;
                const showOnline = isUserOnline(c.otherUserId);

                const profilePath = c.otherUserId != null ? `/users/${c.otherUserId}` : null;
                const avatarNode = (
                  <Avatar
                    src={c.otherUserAvatar || undefined}
                    sx={{
                      width: 40,
                      height: 40,
                      bgcolor: isActive ? "primary.main" : "primary.dark",
                      fontSize: 18,
                    }}
                  >
                    {!c.otherUserAvatar && c.name
                      .split(" ")
                      .map((x) => x[0])
                      .join("")}
                  </Avatar>
                );

                const avatarWrap =
                  profilePath != null ? (
                    <IconButton
                      component={RouterLink}
                      to={profilePath}
                      onClick={(e) => e.stopPropagation()}
                      size="small"
                      sx={{ p: 0 }}
                      aria-label={`View ${c.name || "user"} profile`}
                    >
                      {showOnline ? <OnlineBadge>{avatarNode}</OnlineBadge> : avatarNode}
                    </IconButton>
                  ) : showOnline ? (
                    <OnlineBadge>{avatarNode}</OnlineBadge>
                  ) : (
                    avatarNode
                  );

                return (
                  <ListItem
                    key={c.id}
                    disablePadding
                    sx={{ mb: 0.5 }}
                  >
                    <ListItemButton
                      onClick={() => setSelectedId(c.id)}
                      sx={{
                        borderRadius: 2.5,
                        px: 1.5,
                        py: 1.25,
                        backgroundColor: isActive
                          ? "rgba(25,118,210,0.12)"
                          : "transparent",
                        "&:hover": {
                          backgroundColor: "rgba(25,118,210,0.08)",
                        },
                        transition: "background-color 0.15s ease",
                      }}
                    >
                      <ListItemAvatar sx={{ minWidth: 52 }}>{avatarWrap}</ListItemAvatar>
                      <ListItemText
                        primary={
                          <Box
                            component="span"
                            sx={{
                              display: "flex",
                              alignItems: "center",
                              justifyContent: "space-between",
                              gap: 1,
                              width: "100%",
                            }}
                          >
                            <Typography component="span" variant="subtitle2" noWrap sx={{ fontWeight: 600, flex: 1 }}>
                              {c.name}
                            </Typography>
                            <Typography component="span" variant="caption" sx={{ color: "text.secondary", flexShrink: 0 }}>
                              {c.updatedAt}
                            </Typography>
                          </Box>
                        }
                        secondary={[c.subtitle, c.lastMessage].filter(Boolean).join(" · ") || "No messages yet"}
                        primaryTypographyProps={{ component: "div", sx: { fontWeight: 600 } }}
                        secondaryTypographyProps={{ component: "div", variant: "caption", sx: { color: "text.secondary", mt: 0.25 } }}
                      />
                      {(() => {
                        const convKey = c.conversation_id || c.id;
                        const apiUnread = typeof c.unread === "number" ? c.unread : 0;
                        const localUnread = unreadByConversation[convKey] || 0;
                        const unread = Math.max(apiUnread, localUnread);
                        return unread > 0 ? (
                          <Chip
                            size="small"
                            color="primary"
                            label={unread}
                            sx={{
                              height: 18,
                              minWidth: 22,
                              fontSize: 11,
                              borderRadius: 999,
                              ml: 1,
                            }}
                          />
                        ) : null;
                      })()}
                    </ListItemButton>
                  </ListItem>
                );
              })}
            </List>
            )}
          </Box>
        </Paper>

        {/* 右侧聊天区域 */}
        <Paper
          elevation={0}
          sx={{
            flex: 1,
            borderRadius: 3,
            border: "1px solid #e5e7eb",
            background: "#ffffff",
            display: "flex",
            flexDirection: "column",
            overflow: "hidden",
            minHeight: { xs: 480, md: "auto" },
          }}
        >
          {activeConversation ? (
            <>
              {/* 顶部标题栏 */}
              <Box
                sx={{
                  px: 3,
                  py: 2,
                  borderBottom: "1px solid #e5e7eb",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  gap: 2,
                  background: "#ffffff",
                }}
              >
                <Stack direction="row" spacing={2} alignItems="center" sx={{ minWidth: 0, flex: 1 }}>
                  {activeConversation.otherUserId != null ? (
                    <IconButton
                      component={RouterLink}
                      to={`/users/${activeConversation.otherUserId}`}
                      sx={{ p: 0 }}
                      aria-label={`View ${activeConversation.name || "user"} profile`}
                    >
                      <OnlineBadge invisible={!onlineForConv}>
                        <Avatar
                          src={activeConversation.otherUserAvatar || undefined}
                          sx={{ width: 44, height: 44 }}
                        >
                          {!activeConversation.otherUserAvatar && activeConversation.name
                            .split(" ")
                            .map((x) => x[0])
                            .join("")}
                        </Avatar>
                      </OnlineBadge>
                    </IconButton>
                  ) : (
                    <OnlineBadge invisible={!onlineForConv}>
                      <Avatar
                        src={activeConversation.otherUserAvatar || undefined}
                        sx={{ width: 44, height: 44 }}
                      >
                        {!activeConversation.otherUserAvatar && activeConversation.name
                          .split(" ")
                          .map((x) => x[0])
                          .join("")}
                      </Avatar>
                    </OnlineBadge>
                  )}
                  <Box
                    component={activeConversation.otherUserId != null ? RouterLink : "div"}
                    {...(activeConversation.otherUserId != null
                      ? { to: `/users/${activeConversation.otherUserId}` }
                      : {})}
                    sx={{
                      minWidth: 0,
                      textDecoration: "none",
                      color: "inherit",
                      cursor: activeConversation.otherUserId != null ? "pointer" : "default",
                      "&:hover":
                        activeConversation.otherUserId != null
                          ? { opacity: 0.88 }
                          : undefined,
                    }}
                  >
                    <Typography variant="subtitle1" fontWeight={600}>
                      {activeConversation.name}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {activeConversation.subtitle}
                    </Typography>
                  </Box>
                </Stack>
                <Tooltip title="More actions">
                  <IconButton
                    size="small"
                    sx={{ color: "text.secondary" }}
                    onClick={(e) => setMoreMenuAnchor(e.currentTarget)}
                  >
                    <MoreVertIcon />
                  </IconButton>
                </Tooltip>
                  <Menu
                    anchorEl={moreMenuAnchor}
                    open={Boolean(moreMenuAnchor)}
                    onClose={() => setMoreMenuAnchor(null)}
                    anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
                    transformOrigin={{ vertical: "top", horizontal: "right" }}
                  >
                    <MenuItem
                      onClick={handleOpenDeleteConfirm}
                      sx={{ color: "error.main" }}
                    >
                      <DeleteIcon fontSize="small" sx={{ mr: 1 }} />
                      Delete conversation
                    </MenuItem>
                  </Menu>
                </Stack>
              </Box>

              {/* 顶部提示条：项目议价文案仅客户可见；学生仍可见「正在输入」 */}
              {(showClientProjectBanner || isTyping(activeConversationId)) && (
                <Box
                  sx={{
                    px: 3,
                    py: 1.5,
                    background: "rgba(25,118,210,0.06)",
                    borderBottom: "1px solid #e5e7eb",
                  }}
                >
                  <Stack direction="row" spacing={1} alignItems="center" justifyContent="space-between">
                    {showClientProjectBanner ? (
                      <Typography variant="body2" sx={{ color: "#334155" }}>
                        You are discussing a project with{" "}
                        <strong>{activeConversation.name}</strong>. Agree on a price
                        and timeline here.
                      </Typography>
                    ) : (
                      <span />
                    )}
                    {isTyping(activeConversationId) && (
                      <Typography
                        variant="body2"
                        sx={{ color: "#334155", fontStyle: "italic" }}
                      >
                        对方正在输入…
                      </Typography>
                    )}
                  </Stack>
                </Box>
              )}

              {/* 消息列表 */}
              <Box
                ref={messagesScrollRef}
                sx={{
                  flex: 1,
                  overflow: "auto",
                  px: 3,
                  py: 2.5,
                  display: "flex",
                  flexDirection: "column",
                  gap: 1.5,
                }}
              >
                {loadingHistory && historyMessages.length === 0 ? (
                  <Stack alignItems="center" justifyContent="center" py={4}>
                    <CircularProgress size={28} />
                  </Stack>
                ) : (
                <>
                {messages.map((m) => (
                  <MessageBubble
                    key={m.id || m.temp_id}
                    fromMe={m.fromMe}
                    text={m.text}
                    timestamp={
                      m.createdAt
                        ? (m.createdAt.length > 10
                          ? new Date(m.createdAt).toLocaleString(undefined, {
                              month: "short",
                              day: "numeric",
                              hour: "2-digit",
                              minute: "2-digit",
                            })
                          : m.createdAt)
                        : ""
                    }
                  />
                ))}
                </>
                )}
              </Box>

              {/* 输入区 */}
              <Box
                sx={{
                  borderTop: "1px solid #e5e7eb",
                  px: 2.5,
                  py: 1.75,
                  background: "#ffffff",
                }}
              >
                <Stack direction="row" spacing={1.5} alignItems="flex-end">
                  <input
                    ref={chatFileInputRef}
                    type="file"
                    hidden
                    onChange={handleChatFileChange}
                  />
                  <Tooltip
                    title={
                      attachUploading
                        ? "Uploading…"
                        : connectionStatus !== "open"
                          ? "Attach a file (uploads now; sends when chat is connected)"
                          : "Attach a file"
                    }
                  >
                    <span>
                      <IconButton
                        size="small"
                        disabled={
                          !isAuthenticated ||
                          !activeConversationId ||
                          attachUploading
                        }
                        sx={{
                          color: "text.secondary",
                          bgcolor: "#f8fafc",
                          borderRadius: 2,
                        }}
                        onClick={() => chatFileInputRef.current?.click()}
                      >
                        {attachUploading ? (
                          <CircularProgress size={18} thickness={5} />
                        ) : (
                          <AttachFileIcon fontSize="small" />
                        )}
                      </IconButton>
                    </span>
                  </Tooltip>
                  <TextField
                    multiline
                    minRows={1}
                    maxRows={4}
                    fullWidth
                    placeholder="Write a message..."
                    value={drafts[selectedId] || ""}
                    onChange={(e) => {
                      const value = e.target.value;
                      setDrafts((prev) => ({
                        ...prev,
                        [selectedId]: value,
                      }));
                      notifyTyping();
                    }}
                    onKeyDown={handleKeyDown}
                    variant="outlined"
                    disabled={!isAuthenticated}
                    sx={{
                      "& .MuiOutlinedInput-root": {
                        borderRadius: 3,
                        backgroundColor: "#ffffff",
                        "& fieldset": {
                          borderColor: "#d1d5db",
                        },
                        "&:hover fieldset": {
                          borderColor: "#93c5fd",
                        },
                        "&.Mui-focused fieldset": {
                          borderColor: "primary.main",
                          boxShadow: "0 0 0 1px rgba(59,130,246,0.35)",
                        },
                      },
                    }}
                  />
                  <Button
                    variant="contained"
                    endIcon={<SendIcon sx={{ ml: -0.5 }} />}
                    onClick={handleSend}
                    disabled={!isAuthenticated || !activeConversationId}
                    sx={{
                      borderRadius: 999,
                      px: 2.5,
                      py: 1,
                      textTransform: "none",
                      fontWeight: 600,
                    }}
                  >
                    Send
                  </Button>
                </Stack>
              </Box>
            </>
          ) : (
            <Box
              sx={{
                flex: 1,
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                justifyContent: "center",
                p: 4,
                gap: 2,
              }}
            >
              <Typography variant="h6" fontWeight={600}>
                Select a conversation to get started
              </Typography>
              <Typography variant="body2" color="text.secondary" align="center">
                Choose a thread from the left to review your messages and
                continue the discussion.
              </Typography>
            </Box>
          )}
        </Paper>
      </Stack>
    </PageLayout>

    <Dialog
      open={newConvDialogOpen}
      onClose={() => setNewConvDialogOpen(false)}
      maxWidth="sm"
      fullWidth
      PaperProps={{ sx: { borderRadius: 3 } }}
    >
      <DialogTitle sx={{ pb: 1 }}>New conversation</DialogTitle>
      <DialogContent>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Search for a user to start chatting
        </Typography>
        <TextField
          fullWidth
          placeholder="Search by name, skills, or field..."
          value={userSearchQuery}
          onChange={(e) => setUserSearchQuery(e.target.value)}
          autoFocus
          InputProps={{
            startAdornment: <SearchIcon sx={{ mr: 1, color: "text.secondary" }} />,
          }}
          sx={{ mb: 2 }}
        />
        <Box sx={{ maxHeight: 320, overflow: "auto" }}>
          {userSearchLoading ? (
            <Stack alignItems="center" py={4}>
              <CircularProgress size={28} />
            </Stack>
          ) : userSearchResults.length === 0 ? (
            <Typography variant="body2" color="text.secondary" sx={{ py: 4, textAlign: "center" }}>
              {userSearchQuery.trim() ? "No users found. Try a different search." : "Type to search for users"}
            </Typography>
          ) : (
            <List dense disablePadding>
              {userSearchResults.map((u) => (
                <ListItem key={u.id} disablePadding>
                  <ListItemButton
                    onClick={() => handleSelectUserForNewConv(u.id)}
                    sx={{ borderRadius: 2, py: 1.5 }}
                  >
                    <ListItemAvatar>
                      <Avatar sx={{ width: 40, height: 40 }}>
                        {(talentDisplayName(u) || u.name || "?").charAt(0)}
                      </Avatar>
                    </ListItemAvatar>
                    <ListItemText
                      primary={talentDisplayName(u) || `User ${u.id}`}
                      secondary={u.faculty || u.major || u.organization_name || " "}
                      primaryTypographyProps={{ fontWeight: 600 }}
                      secondaryTypographyProps={{ variant: "caption" }}
                    />
                  </ListItemButton>
                </ListItem>
              ))}
            </List>
          )}
        </Box>
      </DialogContent>
    </Dialog>

    <Dialog
      open={deleteConfirmOpen}
      onClose={() => !deleting && setDeleteConfirmOpen(false)}
      maxWidth="xs"
      fullWidth
      PaperProps={{ sx: { borderRadius: 3 } }}
    >
      <DialogTitle>Delete conversation</DialogTitle>
      <DialogContent>
        <DialogContentText>
          Delete this conversation and all messages? This cannot be undone.
        </DialogContentText>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={() => setDeleteConfirmOpen(false)} disabled={deleting}>
          Cancel
        </Button>
        <Button
          variant="contained"
          color="error"
          onClick={handleDeleteConversation}
          disabled={deleting}
          startIcon={deleting ? <CircularProgress size={16} color="inherit" /> : <DeleteIcon />}
        >
          {deleting ? "Deleting..." : "Delete"}
        </Button>
      </DialogActions>
    </Dialog>

    <Snackbar
      open={toast.open}
      autoHideDuration={6000}
      onClose={() => setToast((t) => ({ ...t, open: false }))}
      message={toast.message}
      anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
    />
    </>
  );
};

export default MessagesPage;

