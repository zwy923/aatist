import React from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Box, Typography, Button, Container, Paper } from "@mui/material";

/**
 * Callback page after redirect from Aalto portal.
 * When Aalto SSO integration is complete, this page will receive the auth response.
 * For now, it guides users to use password login if they have an account.
 */
export default function AuthCallback() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const email = searchParams.get("email");

  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "#f5f5f5",
        padding: 2,
      }}
    >
      <Container maxWidth="sm">
        <Paper
          elevation={0}
          sx={{
            padding: 4,
            background: "#ffffff",
            borderRadius: 3,
            border: "1px solid #e0e0e0",
            boxShadow: "0 2px 8px rgba(0,0,0,0.06)",
            textAlign: "center",
          }}
        >
          <Typography variant="h5" fontWeight={600} color="#1a1a1a" gutterBottom>
            Authentication Complete
          </Typography>
          <Typography variant="body1" color="#666" sx={{ mb: 3 }}>
            {email
              ? `You have been redirected back from Aalto portal.`
              : "You have been redirected back from Aalto portal."}
          </Typography>
          <Typography variant="body2" color="#999" sx={{ mb: 3 }}>
            Full Aalto SSO integration is in progress. If you already have an
            account, you can sign in with your password from the login page.
          </Typography>
          <Button
            variant="contained"
            onClick={() => navigate("/auth/login")}
            sx={{
              backgroundColor: "#1976d2",
              "&:hover": { backgroundColor: "#1565c0" },
            }}
          >
            Back to Login
          </Button>
        </Paper>
      </Container>
    </Box>
  );
}
