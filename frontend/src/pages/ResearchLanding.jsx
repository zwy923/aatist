import { useEffect, useState } from "react";
import { waitlistApi } from "../features/waitlist/api/waitlistApi";
import { getTrackingFields } from "../features/waitlist/utils/tracking";
import talentMascot1 from "../assets/mascots/mascot-talent-1.png";
import talentMascot2 from "../assets/mascots/mascot-talent-2.png";
import talentMascot3 from "../assets/mascots/mascot-talent-3.png";
import opportunityMascot1 from "../assets/mascots/mascot-opportunity-1.png";
import opportunityMascot2 from "../assets/mascots/mascot-opportunity-2.png";
import opportunityMascot3 from "../assets/mascots/mascot-opportunity-3.png";
import partyMascot from "../assets/mascots/mascot-party.png";
import "./ResearchLanding.css";

const COPY = {
  talent: {
    tabLabel: "Looking for talents",
    heading: (
      <span>
        Discover trusted <strong>Aalto art and design students</strong> for
        your next project
      </span>
    ),
  },
  opportunity: {
    tabLabel: "Looking for opportunities",
    heading: (
      <span>
        As <strong>Aalto art and design students</strong>, you can share
        portfolio and find your job opportunities here
      </span>
    ),
  },
};

const CARD_COPY = [
  {
    title: "What services can you find here?",
    body: "Design · Photography · Tutoring · Event support · Exhibition setup · Social media · Consulting · and more.",
    footer: "If a student can do it, you can find it here.",
  },
  {
    title: "Why Aatist?",
    body: "We're building the first student-led platform dedicated entirely to Aalto student talent, making it easy to discover portfolios and connect with students across disciplines.",
  },
  {
    title: "Why real people still matter in the AI era?",
    body: "Experience, judgment, and perspective still matter. Some things still need real people in the real world. Not just generating content, but delivering solutions.",
  },
];

const MASCOTS_BY_TAB = {
  talent: [talentMascot1, talentMascot2, talentMascot3],
  opportunity: [opportunityMascot1, opportunityMascot2, opportunityMascot3],
};

function ResearchLanding() {
  const [tab, setTab] = useState("talent");
  const [email, setEmail] = useState("");
  const [step, setStep] = useState("form");
  const [entryId, setEntryId] = useState(null);
  const [consent, setConsent] = useState(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const mascots = MASCOTS_BY_TAB[tab];

  useEffect(() => {
    waitlistApi.logPageView({ tab: "talent", ...getTrackingFields() }).catch(() => {
      // Best-effort: tracking must never block the page.
    });
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (submitting) return;
    setError("");
    setSubmitting(true);
    try {
      const { data } = await waitlistApi.join({
        email,
        interest: tab,
        ...getTrackingFields(),
      });
      setEntryId(data.id);
      setStep("success");
    } catch (err) {
      setError(
        err?.data?.error || err?.message || "Something went wrong. Please try again."
      );
    } finally {
      setSubmitting(false);
    }
  };

  const handleConsent = async (value) => {
    if (!entryId || consent) return;
    setConsent(value);
    try {
      await waitlistApi.submitConsent(entryId, { consent: value });
    } catch {
      // Best-effort: the UI already reflects the visitor's choice.
    }
  };

  return (
    <main className="research-landing">
      <div className="research-card">
        {step === "form" ? (
          <>
            <div className="tab-toggle" role="tablist" aria-label="I am">
              <button
                type="button"
                role="tab"
                aria-selected={tab === "talent"}
                className={`tab-btn ${tab === "talent" ? "active" : ""}`}
                onClick={() => setTab("talent")}
              >
                Looking for talents
              </button>
              <button
                type="button"
                role="tab"
                aria-selected={tab === "opportunity"}
                className={`tab-btn ${tab === "opportunity" ? "active" : ""}`}
                onClick={() => setTab("opportunity")}
              >
                Looking for opportunities
              </button>
            </div>

            <h1 className="research-heading">{COPY[tab].heading}</h1>
            <p className="research-subtext">
              We are launching soon. Join the waitlist now!
            </p>

            <form className="waitlist-form" onSubmit={handleSubmit}>
              <input
                type="email"
                required
                placeholder="Enter your email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="waitlist-input"
                aria-label="Email address"
              />
              <button
                type="submit"
                className="waitlist-submit"
                disabled={submitting}
              >
                {submitting ? "Joining…" : "Join waitlist"}{" "}
                <span aria-hidden="true">→</span>
              </button>
            </form>
            {error && <p className="research-error">{error}</p>}

            <div className="research-cards">
              {CARD_COPY.map((card, i) => (
                <div className="card-slot" key={card.title}>
                  <div className="card-mascot-frame" aria-hidden="true">
                    <img src={mascots[i]} alt="" className="card-mascot" />
                  </div>
                  <div className="research-card-item">
                    <div className="card-item-head">
                      <h3>{card.title}</h3>
                      <span className="card-sparkle" aria-hidden="true">
                        ✦
                      </span>
                    </div>
                    <p>{card.body}</p>
                    {card.footer && (
                      <p className="card-footer-note">{card.footer}</p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </>
        ) : (
          <div className="research-success">
            <img
              src={partyMascot}
              alt=""
              aria-hidden="true"
              className="success-mascots"
            />
            <h2>Thank you! You&apos;re in the waitlist!</h2>
            <p className="success-copy">
              If you wish to participate further in our product development
              process, you can give us consent to contact via email.
            </p>

            {consent ? (
              <div className={`consent-status ${consent}`}>
                <span aria-hidden="true">✓</span>
                {consent === "agreed" ? "I Agreed" : "I declined"}
              </div>
            ) : (
              <div className="consent-actions">
                <button
                  type="button"
                  className="consent-btn decline"
                  onClick={() => handleConsent("declined")}
                >
                  I don&apos;t want to participate
                </button>
                <button
                  type="button"
                  className="consent-btn agree"
                  onClick={() => handleConsent("agreed")}
                >
                  I agree
                </button>
              </div>
            )}
          </div>
        )}

        <footer className="research-footer">
          <span>Supported by</span>
          <span className="footer-badge">
            <span className="footer-dots" aria-hidden="true">
              ●●●
            </span>
            <span className="footer-badge-text">
              SUSTAINABILITY
              <br />
              ACTION BOOSTER
            </span>
          </span>
        </footer>
      </div>
    </main>
  );
}

export default ResearchLanding;
