import React from "react";
import {
  Avatar,
  Box,
  Button,
  Divider,
  IconButton,
  Menu,
  MenuItem,
  Tooltip,
  Typography,
  Chip,
} from "@mui/material";
import Logout from "@mui/icons-material/Logout";

export default function DashboardHeader({
  navItems,
  onNavClick,
  verificationChip,
  menuAnchorEl,
  isMenuOpen,
  onMenuOpen,
  onMenuClose,
  onLogout,
  userDisplayName,
  userEmail,
  isStudentRole,
  onNavigate,
  variant = "dark",
}) {
  const isLight = variant === "light";
  const textColor = isLight ? "#333" : "text.secondary";
  const textHoverColor = isLight ? "#1a1a1a" : "text.primary";
  const avatarBg = isLight ? "rgba(25, 118, 210, 0.15)" : "rgba(93, 224, 255, 0.2)";
  const avatarColor = isLight ? "#1976d2" : "#5de0ff";
  const borderColor = isLight ? "rgba(0,0,0,0.12)" : "rgba(93, 224, 255, 0.4)";

  return (
    <Box
      sx={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        flexWrap: "wrap",
        gap: 2,
      }}
    >
      <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap" }}>
        {navItems.map((item) => (
          <Button
            key={item.label}
            variant="text"
            onClick={() => onNavClick(item.path)}
            sx={{
              color: textColor,
              fontWeight: 600,
              "&:hover": { color: textHoverColor },
            }}
          >
            {item.label}
          </Button>
        ))}
      </Box>
      <Box sx={{ display: "flex", alignItems: "center", gap: 1.5 }}>
        {verificationChip}
        <Tooltip title="Open profile menu">
          <IconButton
            onClick={onMenuOpen}
            size="small"
            sx={{
              border: `1px solid ${borderColor}`,
              padding: 0.5,
            }}
          >
            <Avatar
              sx={{
                width: 36,
                height: 36,
                bgcolor: avatarBg,
                color: avatarColor,
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
          onClose={onMenuClose}
          anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
          transformOrigin={{ vertical: "top", horizontal: "right" }}
          keepMounted
        >
          <Box sx={{ px: 2, py: 1, maxWidth: 240 }}>
            <Typography variant="subtitle2" fontWeight={600} noWrap>
              {userDisplayName}
            </Typography>
            <Typography variant="caption" color="text.secondary" noWrap>
              {userEmail}
            </Typography>
            <Box mt={1}>{verificationChip}</Box>
          </Box>
          <Divider sx={{ my: 1 }} />
          <MenuItem
            onClick={() => {
              onMenuClose();
              onNavigate("/profile");
            }}
          >
            My profile
          </MenuItem>
          <MenuItem
            onClick={() => {
              onMenuClose();
              onNavigate("/profile?tab=saved");
            }}
          >
            Saved items
          </MenuItem>
          <MenuItem
            onClick={() => {
              onMenuClose();
              onNavigate("/applications");
            }}
          >
            My applications
          </MenuItem>
          {isStudentRole && (
            <MenuItem
              onClick={() => {
                onMenuClose();
                onNavigate("/profile?tab=portfolio");
              }}
            >
              Edit portfolio
            </MenuItem>
          )}
          <MenuItem
            onClick={() => {
              onMenuClose();
              onNavigate("/profile?tab=settings");
            }}
          >
            Settings
          </MenuItem>
          <Divider sx={{ my: 1 }} />
          <MenuItem onClick={onLogout}>
            <Logout fontSize="small" sx={{ mr: 1 }} />
            Logout
          </MenuItem>
        </Menu>
      </Box>
    </Box>
  );
}

