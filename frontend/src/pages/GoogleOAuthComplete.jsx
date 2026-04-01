import React, { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Box, CircularProgress, Typography, Button, Paper, Container } from "@mui/material";
import apiClient from "../shared/api/client";
import useAuthStore from "../shared/stores/authStore";

const errorMessages = {
  denied: "Google sign-in was cancelled.",
  state_invalid: "Sign-in session expired. Please try again.",
  email_conflict: "This email already has an account. Sign in with your password or use the matching account type.",
  not_configured: "Google sign-in is not available.",
  token_exchange_failed: "Could not complete sign-in with Google. Try again.",
  no_id_token: "Google did not return a valid identity token.",
  invalid_id_token: "Could not verify Google sign-in.",
  server_error: "Something went wrong. Please try again later.",
  missing_code: "Sign-in was interrupted. Please try again.",
  oauth_state_invalid: "Sign-in session expired. Please try again.",
};

export default function GoogleOAuthComplete() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [error, setError] = useState(null);

  useEffect(() => {
    const qErr = searchParams.get("error");
    if (qErr) {
      setError(errorMessages[qErr] || `Sign-in failed (${qErr}).`);
      return;
    }

    const hash = (window.location.hash || "").replace(/^#/, "");
    const params = new URLSearchParams(hash);
    const access = params.get("access_token");
    const refresh = params.get("refresh_token");

    if (!access || !refresh) {
      setError("Missing tokens. Return to login and try Google sign-in again.");
      return;
    }

    if (refresh) {
      localStorage.setItem("refresh_token", refresh);
    }
    setAuth(null, access, refresh);

    let cancelled = false;
    (async () => {
      try {
        const res = await apiClient.get("/users/me");
        const user = res.data?.data;
        if (cancelled) return;
        if (!user) {
          setError("Could not load your profile.");
          return;
        }
        setAuth(user, access, refresh);
        navigate("/talents", { replace: true });
      } catch {
        if (!cancelled) {
          setError("Could not load your profile. Try signing in again.");
          useAuthStore.getState().logout();
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [navigate, searchParams, setAuth]);

  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        bgcolor: "#f5f7fb",
        p: 2,
      }}
    >
      <Container maxWidth="sm">
        <Paper elevation={0} sx={{ p: 4, borderRadius: 3, border: "1px solid #e5e7eb", textAlign: "center" }}>
          {error ? (
            <>
              <Typography variant="h6" fontWeight={600} color="#1a1a1a" gutterBottom>
                Google sign-in
              </Typography>
              <Typography color="text.secondary" sx={{ mb: 3 }}>
                {error}
              </Typography>
              <Button variant="contained" onClick={() => navigate("/auth/login", { replace: true })}>
                Back to login
              </Button>
            </>
          ) : (
            <>
              <CircularProgress size={36} sx={{ mb: 2 }} />
              <Typography color="text.secondary">Completing sign-in…</Typography>
            </>
          )}
        </Paper>
      </Container>
    </Box>
  );
}
