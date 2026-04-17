import React from "react";
import { Link } from "react-router-dom";
import "../../pages/AuthRegister.css";
import "../../styles/LegalDoc.css";

/**
 * Shared shell for Privacy / Terms: same header as register/about flows.
 */
export default function LegalLayout({ title, updated, children, bottomSlot = null }) {
  return (
    <div className="legal-doc-page">
      <header className="register-header">
        <Link to="/" className="register-brand" aria-label="Aatist Home">
          <span className="register-brand-icon">A</span>
          <span className="register-brand-text">atist</span>
        </Link>
        <nav className="register-nav">
          <Link to="/talents" className="register-nav-link">
            Hire Talent
          </Link>
          <Link to="/opportunities" className="register-nav-link">
            Opportunities
          </Link>
          <Link to="/about" className="register-nav-link">
            About
          </Link>
        </nav>
      </header>
      <main className="legal-doc-main">
        <article className="legal-doc-article">
          <h1 className="legal-doc-h1">{title}</h1>
          <p className="legal-doc-updated">Last updated: {updated}</p>
          {children}
        </article>
        {bottomSlot}
      </main>
    </div>
  );
}
