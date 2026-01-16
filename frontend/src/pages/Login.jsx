import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import {
  Box,
  Button,
  TextField,
  Typography,
  Alert,
  CircularProgress,
  Container,
  Stack,
  Paper,
} from "@mui/material";
import { useAuth } from "../features/auth/hooks/useAuth";

export default function Login() {
  const navigate = useNavigate();
  const { login, loading } = useAuth();
  const [error, setError] = useState("");

  const handleLogin = async (e) => {
    e.preventDefault();
    setError("");

    const form = e.currentTarget;
    const email = form.email.value;
    const password = form.password.value;

    const result = await login({ email, password });
    if (result.success) {
      navigate("/dashboard");
    } else {
      setError(result.error || "Sign-in failed, please try again.");
    }
  };

  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "radial-gradient(ellipse at top left, #101820, #050505)",
        padding: 2,
      }}
    >
      <Container maxWidth="sm">
        <Paper
          elevation={0}
          sx={{
            padding: { xs: 3, md: 5 },
            background: "rgba(7, 12, 30, 0.96)",
            borderRadius: 3,
            border: "1px solid rgba(93, 224, 255, 0.25)",
            position: "relative",
            overflow: "hidden",
            "&::before": {
              content: '""',
              position: "absolute",
              inset: 0,
              pointerEvents: "none",
              background:
                "radial-gradient(circle at 20% 20%, rgba(93,224,255,0.15), transparent 45%)",
            },
          }}
        >
          <Stack spacing={3} sx={{ position: "relative" }}>
            <Stack spacing={1}>
              <Typography
                variant="h4"
                fontWeight={700}
                sx={{ color: "text.primary" }}
              >
                Sign in to Aatist
              </Typography>
              <Typography variant="body1" color="text.secondary">
                Access your studio, review briefs, and keep shipping.
              </Typography>
            </Stack>

            {error && (
              <Alert severity="error" variant="outlined">
                {error}
              </Alert>
            )}

            <Box component="form" onSubmit={handleLogin}>
              <Stack spacing={3}>
                <TextField
                  name="email"
                  type="email"
                  label="Email"
                  placeholder="you@aatist.fi"
                  fullWidth
                  required
                  autoComplete="email"
                  sx={{
                    "& .MuiOutlinedInput-root": {
                      color: "text.primary",
                    },
                  }}
                />
                <TextField
                  name="password"
                  type="password"
                  label="Password"
                  placeholder="Enter your password"
                  fullWidth
                  required
                  autoComplete="current-password"
                  sx={{
                    "& .MuiOutlinedInput-root": {
                      color: "text.primary",
                    },
                  }}
                />
                <Button
                  type="submit"
                  variant="contained"
                  size="large"
                  disabled={loading}
                  fullWidth
                  endIcon={
                    loading ? <CircularProgress size={18} color="inherit" /> : undefined
                  }
                  sx={{
                    background: "linear-gradient(135deg, #007bff 0%, #7f5dff 100%)",
                    "&:hover": {
                      background: "linear-gradient(135deg, #0066cc 0%, #6b4dd9 100%)",
                    },
                  }}
                >
                  {loading ? "Signing in..." : "Sign in"}
                </Button>
                <Typography variant="body2" color="text.secondary" textAlign="center">
                  New here?{" "}
                  <Link
                    to="/auth/register"
                    style={{
                      color: "#5de0ff",
                      textDecoration: "none",
                      fontWeight: 600,
                    }}
                  >
                    Create an account
                  </Link>
                </Typography>
                <Button
                  component={Link}
                  to="/"
                  variant="text"
                  size="small"
                  sx={{
                    alignSelf: "center",
                    textTransform: "none",
                    color: "text.secondary",
                  }}
                >
                  Back to home
                </Button>
              </Stack>
            </Box>
          </Stack>
        </Paper>
      </Container>
    </Box>
  );
}

