import React, { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Checkbox,
  CircularProgress,
  FormControlLabel,
  TextField,
  Typography,
} from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import NotificationsNoneIcon from "@mui/icons-material/NotificationsNone";
import AccountCircleIcon from "@mui/icons-material/AccountCircle";
import { useAuth } from "../features/auth/hooks/useAuth";
import "./Landing.css";
import "./AuthLogin.css";

export default function Login() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { login, loading } = useAuth();
  const [error, setError] = useState("");
  const [mode, setMode] = useState("client");
  const [remember, setRemember] = useState(false);
  const [form, setForm] = useState({ email: "", password: "" });

  useEffect(() => {
    const modeFromUrl = searchParams.get("mode");
    if (modeFromUrl === "client" || modeFromUrl === "student") {
      setMode(modeFromUrl);
    }
  }, [searchParams]);

  const isStudentMode = mode === "student";
  const emailPlaceholder = useMemo(
    () => (isStudentMode ? "you@aalto.fi" : "you@company.com"),
    [isStudentMode]
  );

  const switchMode = (nextMode) => {
    setMode(nextMode);
    setSearchParams({ mode: nextMode });
    setError("");
  };

  const onSubmit = async (event) => {
    event.preventDefault();
    setError("");

    if (!form.email.trim() || !form.password) {
      setError("Email and password are required.");
      return;
    }
    if (isStudentMode && !form.email.trim().toLowerCase().endsWith("@aalto.fi")) {
      setError("Student login requires @aalto.fi email.");
      return;
    }

    const result = await login({
      email: form.email.trim(),
      password: form.password,
      login_type: isStudentMode ? "student" : "client",
    });

    if (!result.success) {
      setError(result.error || "Sign in failed, please try again.");
      return;
    }

    if (!remember) {
      localStorage.removeItem("refresh_token");
    }
    navigate("/dashboard");
  };

  return (
    <main className="login-auth-page">
      <header className="landing-header">
        <Link to="/" className="brand" aria-label="Aatist Home">
          <span className="brand-icon">A</span>
          <span className="brand-text">atist</span>
        </Link>
        <nav className="landing-nav" aria-label="Primary">
          <Link to="/" className="nav-link active">
            Home
          </Link>
          <Link to="/talents" className="nav-link">
            Hire Talent
          </Link>
          <Link to="/opportunities" className="nav-link">
            Opportunities
          </Link>
        </nav>
        <div className="nav-actions">
          <button type="button" className="icon-button" disabled aria-label="Messages">
            <SearchIcon fontSize="small" />
          </button>
          <button type="button" className="icon-button" disabled aria-label="Notifications">
            <NotificationsNoneIcon fontSize="small" />
          </button>
          <button
            type="button"
            className="icon-button"
            aria-label="Account"
            onClick={() => navigate("/auth/login")}
          >
            <AccountCircleIcon fontSize="small" />
          </button>
        </div>
      </header>

      <section className="login-auth-hero">
        <div className="login-card">
          <h1>Welcome Back</h1>
          <p>Sign in to continue to Aatist</p>

          <div className="login-mode-toggle" role="tablist" aria-label="Login role">
            <button
              type="button"
              className={mode === "client" ? "active" : ""}
              onClick={() => switchMode("client")}
            >
              Client
            </button>
            <button
              type="button"
              className={mode === "student" ? "active" : ""}
              onClick={() => switchMode("student")}
            >
              Student
            </button>
          </div>

          {error && <Alert severity="error">{error}</Alert>}

          <Box component="form" className="login-form" onSubmit={onSubmit}>
            <label>Email Address</label>
            <TextField
              placeholder={emailPlaceholder}
              type="email"
              value={form.email}
              onChange={(e) => setForm((prev) => ({ ...prev, email: e.target.value }))}
              required
              fullWidth
            />

            <div className="password-row">
              <label>Password</label>
              <Link to="/auth/forgot-password">Forgot password?</Link>
            </div>
            <TextField
              placeholder="••••••••"
              type="password"
              value={form.password}
              onChange={(e) => setForm((prev) => ({ ...prev, password: e.target.value }))}
              required
              fullWidth
            />

            <FormControlLabel
              control={
                <Checkbox
                  checked={remember}
                  onChange={(e) => setRemember(e.target.checked)}
                  size="small"
                />
              }
              label="Remember me for 30 days"
              sx={{ color: "#6c7787", mt: 0.5 }}
            />

            <Button type="submit" className="login-submit" disabled={loading}>
              {loading ? <CircularProgress size={18} color="inherit" /> : "Sign in"}
            </Button>
          </Box>

          <div className="login-divider">
            <span>Or continue with</span>
          </div>

          <div className="social-row">
            <button type="button">G</button>
            {isStudentMode && <button type="button">A!</button>}
          </div>

          <footer className="login-footer">
            Don&apos;t have an account? <Link to="/auth/register">Sign up now</Link>
          </footer>
        </div>
      </section>
    </main>
  );
}
