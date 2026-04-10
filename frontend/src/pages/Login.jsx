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
} from "@mui/material";
import { useAuth } from "../features/auth/hooks/useAuth";
import { getGoogleOAuthStartURL } from "../shared/utils/googleOAuth";
import "./Landing.css";
import "./AuthLogin.css";

const CLIENT_LETTERS = [
  { char: "C", x: 20, y: 80, r: -20 },
  { char: "L", x: 100, y: 60, r: 22 },
  { char: "I", x: 220, y: 20, r: -38 },
  { char: "E", x: 340, y: 90, r: -5 },
  { char: "N", x: 440, y: -30, r: 28 },
  { char: "T", x: 540, y: 70, r: -10 },
];

const STUDENT_LETTERS = [
  { char: "A", x: 20, y: 70, r: 10 },
  { char: "A", x: 120, y: 40, r: -32 },
  { char: "T", x: 240, y: 10, r: 24 },
  { char: "I", x: 360, y: 50, r: -3 },
  { char: "S", x: 460, y: 80, r: -18 },
  { char: "T", x: 560, y: 30, r: 5, scale: 1.15 },
];

export default function Login() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { login, loading } = useAuth();
  const [error, setError] = useState("");
  const [mode, setMode] = useState("client");
  const [remember, setRemember] = useState(false);
  const [termsChecked, setTermsChecked] = useState(false);
  const [form, setForm] = useState({ email: "", password: "" });

  useEffect(() => {
    const modeFromUrl = searchParams.get("mode");
    if (modeFromUrl === "client" || modeFromUrl === "student") {
      setMode(modeFromUrl);
    }
  }, [searchParams]);

  const isStudentMode = mode === "student";

  const sessionNotice = useMemo(() => {
    const r = searchParams.get("reason");
    if (r === "session_expired") {
      return "Your session has expired. Please sign in again.";
    }
    if (r === "idle_timeout") {
      return "You were signed out after a period of inactivity.";
    }
    return "";
  }, [searchParams]);

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
    if (isStudentMode && !termsChecked) {
      setError("Please confirm you are an Aalto student and agree to Terms & Privacy Policy.");
      return;
    }

    const result = await login({
      email: form.email.trim(),
      password: form.password,
      login_type: isStudentMode ? "student" : "client",
    });

    if (!result.success) {
      setError(result.code === "USER_NOT_REGISTERED" ? "该邮箱尚未注册，请先注册" : (result.error || "Sign in failed, please try again."));
      return;
    }

    if (!remember) {
      localStorage.removeItem("refresh_token");
    }
    navigate("/talents");
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
          <Link to="/auth/register" className="nav-btn nav-btn-signup">
            Sign up
          </Link>
          <Link to="/auth/login" className="nav-btn nav-btn-login">
            Log in
          </Link>
        </div>
      </header>

      <section className="login-auth-hero">
        <div className="login-auth-split-left">
          <div className="login-hero-word login-hero-word-client" aria-hidden="true">
            {CLIENT_LETTERS.map((letter, index) => (
              <span
                key={`c-${index}`}
                className="login-hero-letter"
                style={{ transform: `translate(${letter.x}px, ${letter.y}px) rotate(${letter.r}deg)` }}
              >
                {letter.char}
              </span>
            ))}
          </div>
        </div>
        <div className="login-auth-split-right">
          <div className="login-hero-word login-hero-word-student" aria-hidden="true">
            {STUDENT_LETTERS.map((letter, index) => (
              <span
                key={`s-${index}`}
                className={`login-hero-letter ${letter.scale ? "login-hero-letter-large" : ""}`}
                style={{ transform: `translate(${letter.x}px, ${letter.y}px) rotate(${letter.r}deg)` }}
              >
                {letter.char}
              </span>
            ))}
          </div>
        </div>

        <div className="login-card">
          <h1>Welcome Back !</h1>

          <div className="login-mode-toggle" role="tablist" aria-label="Login role">
            <button
              type="button"
              className={`login-tab ${mode === "client" ? "login-tab-client active" : ""}`}
              onClick={() => switchMode("client")}
            >
              Client
            </button>
            <button
              type="button"
              className={`login-tab ${mode === "student" ? "login-tab-student active" : ""}`}
              onClick={() => switchMode("student")}
            >
              Student
            </button>
          </div>

          {sessionNotice && (
            <Alert severity="info" sx={{ mt: 2 }}>
              {sessionNotice}
            </Alert>
          )}
          {error && <Alert severity="error" sx={{ mt: 2 }}>{error}</Alert>}

          <Box component="form" className="login-form" onSubmit={onSubmit}>
            <label>{isStudentMode ? "Aalto Email" : "Email Address"}</label>
            <TextField
              placeholder={isStudentMode ? "your.name@aalto.fi" : "you@company.com"}
              type="email"
              value={form.email}
              onChange={(e) => setForm((prev) => ({ ...prev, email: e.target.value }))}
              required
              fullWidth
              size="small"
            />

            <div className="password-row">
              <label>Password</label>
              {!isStudentMode && (
                <Link to="/auth/forgot-password" className="forgot-link">Forgot Password?</Link>
              )}
            </div>
            <TextField
              placeholder={isStudentMode ? "(8+ characters)" : "••••••••"}
              type="password"
              value={form.password}
              onChange={(e) => setForm((prev) => ({ ...prev, password: e.target.value }))}
              required
              fullWidth
              size="small"
            />

            {isStudentMode ? (
              <FormControlLabel
                control={
                  <Checkbox
                    checked={termsChecked}
                    onChange={(e) => setTermsChecked(e.target.checked)}
                    size="small"
                  />
                }
                label={
                  <span>
                    I confirm that I am an Aalto student and agree to{" "}
                    <Link
                      to="/terms"
                      state={{ from: "login", loginMode: mode }}
                      className="login-terms-link"
                    >
                      Terms & Privacy Policy
                    </Link>
                  </span>
                }
                className="login-checkbox"
              />
            ) : (
              <FormControlLabel
                control={
                  <Checkbox
                    checked={remember}
                    onChange={(e) => setRemember(e.target.checked)}
                    size="small"
                  />
                }
                label="Remember me for 30 days"
                className="login-checkbox"
              />
            )}

            <Button type="submit" className="login-submit" disabled={loading}>
              {loading ? <CircularProgress size={18} color="inherit" /> : "Log in"}
            </Button>
          </Box>

          {!isStudentMode && (
            <>
              <div className="login-divider">
                <span>Or continue with</span>
              </div>
              <div className="social-row">
                <button
                  type="button"
                  className="social-btn"
                  aria-label="Google"
                  onClick={() => {
                    window.location.href = getGoogleOAuthStartURL();
                  }}
                >
                  <span className="social-google">G</span>
                </button>
                <button type="button" className="social-btn" aria-label="Apple">
                  <svg className="apple-logo" viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
                    <path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09l.01-.01zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.42.42 2.43-2.43 4.66-3.74 4.42z"/>
                  </svg>
                </button>
              </div>
            </>
          )}

          <footer className="login-footer">
            Don&apos;t have an account? <Link to="/auth/register">Sign up</Link>
          </footer>
        </div>
      </section>
    </main>
  );
}
