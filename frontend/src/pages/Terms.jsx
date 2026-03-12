import React from "react";
import { Link } from "react-router-dom";
import "./AuthRegister.css";

export default function Terms() {
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
        <Link to="/auth/register" style={{ display: "inline-block", marginTop: 24, color: "#1976d2", fontWeight: 600 }}>Back to Register</Link>
      </section>
    </main>
  );
}
