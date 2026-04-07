import React from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import "./AuthRegister.css";

export default function Terms() {
  const location = useLocation();
  const navigate = useNavigate();
  const from = location.state?.from;
  const registerMode = location.state?.registerMode;
  const loginMode = location.state?.loginMode;

  const registerBackTo =
    registerMode === "client" || registerMode === "student"
      ? `/auth/register?mode=${registerMode}`
      : "/auth/register";

  const loginBackTo =
    loginMode === "client" || loginMode === "student"
      ? `/auth/login?mode=${loginMode}`
      : "/auth/login";

  const isFromLogin = from === "login";
  const backHref = isFromLogin ? loginBackTo : registerBackTo;
  const backLabel = isFromLogin ? "Back to sign in" : "Back to registration";

  return (
    <main className="register-auth-page">
      <header className="register-header">
        <Link to="/" className="register-brand" aria-label="Aatist Home">
          <span className="register-brand-icon">A</span>
          <span className="register-brand-text">atist</span>
        </Link>
        <nav className="register-nav">
          <Link to="/talents" className="register-nav-link">Hire Talent</Link>
          <Link to="/opportunities" className="register-nav-link">Opportunities</Link>
          <Link to="/about" className="register-nav-link">About</Link>
        </nav>
      </header>
      <section className="about-hero" style={{ minHeight: "calc(100vh - 56px)", padding: "48px 24px", background: "#95e3ba", textAlign: "center" }}>
        <h1 style={{ fontSize: "2.5rem", color: "#1a1a1a", marginBottom: "16px" }}>Terms & Privacy Policy</h1>
        <p style={{ color: "#5f6d7b", fontSize: "1.125rem", maxWidth: 600, margin: "0 auto" }}>
          Please review our terms of service and privacy policy. This page will be updated with full content.
        </p>
        <div style={{ marginTop: 24, display: "flex", flexDirection: "column", alignItems: "center", gap: 12 }}>
          <Link to={backHref} style={{ color: "#1976d2", fontWeight: 600 }}>{backLabel}</Link>
          {(from === "register" || from === "login") && (
            <button
              type="button"
              onClick={() => navigate(-1)}
              style={{
                background: "none",
                border: "none",
                color: "#5f6d7b",
                textDecoration: "underline",
                cursor: "pointer",
                fontSize: "0.95rem",
              }}
            >
              Back to previous page
            </button>
          )}
        </div>
      </section>
    </main>
  );
}
