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
}) {
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
              color: "text.secondary",
              fontWeight: 600,
              "&:hover": { color: "text.primary" },
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
              border: "1px solid rgba(93, 224, 255, 0.4)",
              padding: 0.5,
            }}
          >
            <Avatar
              sx={{
                width: 36,
                height: 36,
                bgcolor: "rgba(93, 224, 255, 0.2)",
                color: "#5de0ff",
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
              onNavigate("/saved");
            }}
          >
            Saved items
          </MenuItem>
          {isStudentRole && (
            <MenuItem
              onClick={() => {
                onMenuClose();
                onNavigate("/portfolio/edit");
              }}
            >
              Edit portfolio
            </MenuItem>
          )}
          <MenuItem
            onClick={() => {
              onMenuClose();
              onNavigate("/settings");
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

