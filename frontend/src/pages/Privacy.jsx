import React from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import LegalLayout from "../shared/components/LegalLayout.jsx";

export default function Privacy() {
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

  const bottomSlot = (
    <div className="legal-doc-bottom-nav">
      <Link to={backHref}>{backLabel}</Link>
      {(from === "register" || from === "login") && (
        <button type="button" onClick={() => navigate(-1)}>
          Back to previous page
        </button>
      )}
    </div>
  );

  return (
    <LegalLayout title="Privacy Policy" updated="17 April 2026" bottomSlot={bottomSlot}>
      <p className="legal-doc-lead">
        Aatist is a non-profit pilot project supported by the Aalto Sustainability Action
        Booster. Our goal is to provide Aalto University students with a centralized platform
        to showcase their skills, and to help clients discover and connect with student talent.
      </p>

      <h2>1. Information We Collect</h2>
      <p>We may collect the following types of information:</p>
      <ul>
        <li>Personal information (such as name, email address, and profile details)</li>
        <li>Portfolio content voluntarily provided by users</li>
        <li>Basic usage data (such as page visits and interactions)</li>
      </ul>
      <p>We only collect information that is necessary to operate and improve the platform.</p>

      <h2>2. How We Use Your Information</h2>
      <p>We use your data to:</p>
      <ul>
        <li>Enable students to showcase their skills and portfolios</li>
        <li>Help clients discover and contact relevant talent</li>
        <li>Improve the platform experience and functionality</li>
      </ul>
      <p>
        We do not sell your personal data or share it with third parties for commercial
        purposes.
      </p>

      <h2>3. Legal Basis (GDPR)</h2>
      <p>We process personal data based on:</p>
      <ul>
        <li>User consent (when you submit your information)</li>
        <li>Legitimate interest (to operate and improve the platform)</li>
      </ul>

      <h2>4. Data Sharing</h2>
      <ul>
        <li>Student profiles may be visible to clients using the platform</li>
        <li>Contact between users happens voluntarily and directly</li>
      </ul>
      <p>Aatist does not act as an intermediary in contracts or payments.</p>

      <h2>5. Data Storage and Security</h2>
      <p>
        Data is stored on secure servers. We take reasonable technical and organizational
        measures to protect your data.
      </p>
      <p>
        However, as this is an MVP-stage platform, we recommend that users avoid sharing
        sensitive personal information.
      </p>

      <h2>6. User Responsibility</h2>
      <p>Users are responsible for:</p>
      <ul>
        <li>The accuracy of the information they provide</li>
        <li>Managing their own communications and agreements</li>
      </ul>
      <p>
        All agreements, compensation, and tax obligations are the sole responsibility of the
        users involved. Aatist does not act as an employer, agent, or payment intermediary.
      </p>

      <h2>7. Your Rights (EU/Finland)</h2>
      <p>Under the General Data Protection Regulation (GDPR), you have the right to:</p>
      <ul>
        <li>Access your personal data</li>
        <li>Request correction or deletion</li>
        <li>Withdraw your consent at any time</li>
      </ul>
      <p>
        To exercise these rights, please contact us at:{" "}
        <a href="mailto:aatist.fi@gmail.com" className="legal-doc-contact">
          aatist.fi@gmail.com
        </a>
      </p>

      <h2>8. Changes</h2>
      <p>
        We may update this Privacy Policy from time to time as the platform evolves. Continued
        use of the platform indicates acceptance of the updated policy.
      </p>
    </LegalLayout>
  );
}
