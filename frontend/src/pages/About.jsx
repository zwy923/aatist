import React from "react";
import { Link } from "react-router-dom";
import "./AuthRegister.css";

export default function About() {
  return (
    <main className="about-page">
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
        <h1 style={{ fontSize: "2.5rem", color: "#1a1a1a", marginBottom: "16px" }}>About Aatist</h1>
        <p style={{ color: "#5f6d7b", fontSize: "1.125rem" }}>Connecting Aalto talent with opportunities.</p>
      </section>
    </main>
  );
}
