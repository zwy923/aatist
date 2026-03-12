import React, { useMemo, useState } from "react";
import {
  Avatar,
  Badge,
  Box,
  Container,
  Divider,
  IconButton,
  Menu,
  MenuItem,
  Tooltip,
  Typography,
} from "@mui/material";
import ChatBubbleOutlineIcon from "@mui/icons-material/ChatBubbleOutline";
import NotificationsNoneIcon from "@mui/icons-material/NotificationsNone";
import LogoutIcon from "@mui/icons-material/Logout";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../../features/auth/hooks/useAuth";
import { useChat } from "../../features/messages/ChatProvider";
import "../../pages/Landing.css";

const NAV_LINKS = [
  { label: "Home", path: "/" },
  { label: "Hire Talent", path: "/talents" },
  { label: "Opportunities", path: "/opportunities" },
];

const isLinkActive = (pathname, linkPath) => {
  if (linkPath === "/") return pathname === "/";
  return pathname === linkPath || pathname.startsWith(`${linkPath}/`);
};

const PageLayout = ({ children, maxWidth = "xl" }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout, isAuthenticated } = useAuth();
  const chat = useChat();
  const unreadCount = chat?.totalUnreadCount ?? 0;
  const [menuAnchorEl, setMenuAnchorEl] = useState(null);
  const isMenuOpen = Boolean(menuAnchorEl);

  const userDisplayName = useMemo(
    () => user?.name || user?.username || user?.email || "User",
    [user]
  );

  const handleMenuOpen = (event) => setMenuAnchorEl(event.currentTarget);
  const handleMenuClose = () => setMenuAnchorEl(null);

  const handleLogout = async () => {
    await logout();
    handleMenuClose();
    navigate("/auth/login");
  };

  return (
    <Box sx={{ minHeight: "100vh", background: "#fff" }}>
      <header className="landing-header">
        <Link to="/" className="brand" aria-label="Aatist Home">
          <span className="brand-icon">A</span>
          <span className="brand-text">atist</span>
        </Link>

        <nav className="landing-nav" aria-label="Primary">
          {NAV_LINKS.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className={`nav-link ${
                isLinkActive(location.pathname, item.path) ? "active" : ""
              }`}
            >
              {item.label}
            </Link>
          ))}
        </nav>

        <div className="nav-actions">
          <Tooltip title={unreadCount > 0 ? `${unreadCount} unread messages` : "Messages"}>
            <button
              type="button"
              className="icon-button"
              aria-label={unreadCount > 0 ? `${unreadCount} unread messages` : "Messages"}
              onClick={() => navigate("/messages")}
              disabled={!isAuthenticated}
            >
              <Badge badgeContent={unreadCount > 0 ? unreadCount : 0} color="error" max={99}>
                <ChatBubbleOutlineIcon fontSize="small" />
              </Badge>
            </button>
          </Tooltip>
          <button
            type="button"
            className="icon-button"
            aria-label="Notifications"
            onClick={() => navigate("/dashboard")}
            disabled={!isAuthenticated}
          >
            <NotificationsNoneIcon fontSize="small" />
          </button>
          <Tooltip title="Open account menu">
            <IconButton
              onClick={handleMenuOpen}
              size="small"
              className="icon-button"
              aria-label="Account"
            >
              <Avatar
                sx={{
                  width: 26,
                  height: 26,
                  bgcolor: "transparent",
                  color: "inherit",
                  fontSize: "0.95rem",
                  fontWeight: 700,
                }}
              >
                {userDisplayName?.[0]?.toUpperCase?.() || "U"}
              </Avatar>
            </IconButton>
          </Tooltip>
          <Menu
            anchorEl={menuAnchorEl}
            open={isMenuOpen}
            onClose={handleMenuClose}
            anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
            transformOrigin={{ vertical: "top", horizontal: "right" }}
            keepMounted
          >
            <Box sx={{ px: 2, py: 1, maxWidth: 260 }}>
              <Typography variant="subtitle2" fontWeight={600} noWrap>
                {userDisplayName}
              </Typography>
              <Typography variant="caption" color="text.secondary" noWrap>
                {user?.email}
              </Typography>
            </Box>
            <Divider sx={{ my: 1 }} />
            <MenuItem
              onClick={() => {
                handleMenuClose();
                navigate("/dashboard");
              }}
            >
              Dashboard
            </MenuItem>
            <MenuItem
              onClick={() => {
                handleMenuClose();
                navigate("/profile");
              }}
            >
              My profile
            </MenuItem>
            <MenuItem
              onClick={() => {
                handleMenuClose();
                navigate("/profile?tab=saved");
              }}
            >
              Saved items
            </MenuItem>
            <MenuItem
              onClick={() => {
                handleMenuClose();
                navigate("/profile?tab=settings");
              }}
            >
              Settings
            </MenuItem>
            <Divider sx={{ my: 1 }} />
            <MenuItem onClick={handleLogout}>
              <LogoutIcon fontSize="small" sx={{ mr: 1 }} />
              Logout
            </MenuItem>
          </Menu>
        </div>
      </header>

      <Box sx={{ py: 4 }}>
        <Container maxWidth={maxWidth}>{children}</Container>
      </Box>
    </Box>
  );
};

export default PageLayout;
