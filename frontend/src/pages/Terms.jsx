import React from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import LegalLayout from "../shared/components/LegalLayout.jsx";

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
    <LegalLayout title="Terms of Use" updated="17 April 2026" bottomSlot={bottomSlot}>
      <p className="legal-doc-lead">
        Aatist is a non-profit pilot project supported by the Aalto Sustainability Action
        Booster. The platform is designed to help Aalto University students showcase their
        skills and connect with external clients.
      </p>
      <p>By using this platform, you agree to the following terms.</p>

      <h2>1. Nature of the Platform</h2>
      <p>
        Aatist is provided as a pilot (MVP-stage) platform and is currently free of charge.
      </p>
      <p>The platform functions solely as a discovery and connection tool between students and clients.</p>
      <p>Aatist does NOT:</p>
      <ul>
        <li>Act as a contracting party</li>
        <li>Facilitate or process payments</li>
        <li>Guarantee project outcomes or user performance</li>
      </ul>

      <h2>2. User Responsibility</h2>
      <p>All interactions, agreements, and collaborations are conducted directly between users.</p>
      <p>Users are solely responsible for:</p>
      <ul>
        <li>Negotiating project terms</li>
        <li>Managing communication</li>
        <li>Handling payments</li>
        <li>Complying with applicable tax and legal obligations</li>
      </ul>
      <p>
        Aatist is not responsible for any disputes, losses, or damages arising from user
        interactions.
      </p>

      <h2>3. Eligibility</h2>
      <p>
        The platform is intended primarily for Aalto University students and individuals or
        organizations seeking to collaborate with them.
      </p>
      <p>We reserve the right to remove or restrict access to any user who misuses the platform.</p>

      <h2>4. Content and Profiles</h2>
      <p>Users are responsible for the content they provide, including:</p>
      <ul>
        <li>Accuracy of profile information</li>
        <li>Ownership of portfolio materials</li>
      </ul>
      <p>
        By submitting content, users grant Aatist the right to display this information on the
        platform for the purpose of facilitating connections.
      </p>

      <h2>5. No Fees</h2>
      <p>At this stage, Aatist does not charge any fees for use of the platform.</p>
      <p>We reserve the right to introduce new features or pricing models in the future.</p>

      <h2>6. Limitation of Liability</h2>
      <p>The platform is provided &quot;as is&quot; without warranties of any kind.</p>
      <p>Aatist is not liable for:</p>
      <ul>
        <li>The quality, safety, or legality of services provided by users</li>
        <li>The accuracy of user-generated content</li>
        <li>Any direct or indirect damages resulting from platform use</li>
      </ul>

      <h2>7. Changes to the Service</h2>
      <p>
        We may modify, suspend, or discontinue the platform at any time as part of the pilot
        phase.
      </p>

      <h2>8. Changes to Terms</h2>
      <p>
        We may update these Terms of Use from time to time. Continued use of the platform
        indicates acceptance of the updated terms.
      </p>

      <h2>9. Contact</h2>
      <p>If you have any questions regarding these Terms, please contact us at:</p>
      <p>
        Email:{" "}
        <a href="mailto:aatist.fi@gmail.com" className="legal-doc-contact">
          aatist.fi@gmail.com
        </a>
      </p>
    </LegalLayout>
  );
}
