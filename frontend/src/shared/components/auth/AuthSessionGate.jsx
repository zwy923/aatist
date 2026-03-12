import React, { useEffect, useState } from "react";
import { Box, CircularProgress } from "@mui/material";
import apiClient from "../../api/client";
import useAuthStore from "../../stores/authStore";

export default function AuthSessionGate({ children }) {
  const accessToken = useAuthStore((s) => s.accessToken);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const updateUser = useAuthStore((s) => s.updateUser);
  const logout = useAuthStore((s) => s.logout);
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    let cancelled = false;

    const verifySession = async () => {
      // No local auth state -> render app immediately
      if (!isAuthenticated || !accessToken) {
        if (!cancelled) setChecking(false);
        return;
      }

      try {
        const res = await apiClient.get("/users/me");
        const user = res?.data?.data;
        if (!user?.id) {
          throw new Error("invalid session user");
        }
        updateUser(user);
      } catch (err) {
        // Token exists locally but backend no longer recognizes the user/session
        logout();
      } finally {
        if (!cancelled) setChecking(false);
      }
    };

    verifySession();

    return () => {
      cancelled = true;
    };
  }, [isAuthenticated, accessToken, updateUser, logout]);

  if (checking) {
    return (
      <Box
        sx={{
          minHeight: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          bgcolor: "#f5f7fb",
        }}
      >
        <CircularProgress size={32} />
      </Box>
    );
  }

  return children;
}
