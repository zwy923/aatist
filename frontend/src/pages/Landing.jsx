import { Link, useNavigate } from "react-router-dom";
import ChatBubbleOutlineIcon from "@mui/icons-material/ChatBubbleOutline";
import NotificationsNoneIcon from "@mui/icons-material/NotificationsNone";
import AccountCircleIcon from "@mui/icons-material/AccountCircle";
import { useAuth } from "../features/auth/hooks/useAuth";
import "./Landing.css";

function Landing() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();

  return (
    <main className="landing-page">
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
          {isAuthenticated ? (
            <>
              <button
                type="button"
                className="icon-button"
                aria-label="Messages"
                onClick={() => navigate("/messages")}
              >
                <ChatBubbleOutlineIcon fontSize="small" />
              </button>
              <button
                type="button"
                className="icon-button"
                aria-label="Notifications"
              >
                <NotificationsNoneIcon fontSize="small" />
              </button>
              <button
                type="button"
                className="icon-button"
                aria-label="Account"
                onClick={() => navigate("/talents")}
              >
                <AccountCircleIcon fontSize="small" />
              </button>
            </>
          ) : (
            <>
              <Link to="/auth/login" className="nav-btn nav-btn-login">
                Log in
              </Link>
              <Link to="/auth/register" className="nav-btn nav-btn-signup">
                Sign up
              </Link>
            </>
          )}
        </div>
      </header>

      <section className="aurora-hero">
        <div className="aurora-bg" aria-hidden="true">
          <div className="aurora-orb aurora-orb-1" />
          <div className="aurora-orb aurora-orb-2" />
          <div className="aurora-orb aurora-orb-3" />
          <div className="aurora-orb aurora-orb-4" />
          <div className="aurora-orb aurora-orb-5" />
        </div>
        <div className="hero-content">
          <div className="hero-badge">
            <svg className="badge-icon" viewBox="0 0 16 16" fill="none" aria-hidden="true">
              <path d="M8 13.5C8 13.5 2 9.5 2 5.5C2 3.567 3.567 2 5.5 2C6.48 2 7.37 2.4 8 3.07C8.63 2.4 9.52 2 10.5 2C12.433 2 14 3.567 14 5.5C14 9.5 8 13.5 8 13.5Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round"/>
            </svg>
            Verified Aalto talent network
          </div>
          <h1 className="hero-title-main">Inside Aalto.</h1>
          <p className="hero-title-gradient">Real talent, real work.</p>
          <p className="hero-description">
            Connect with verified Aalto creative students for your next project.<br />
            Or showcase your skills and get hired.
          </p>
          <div className="hero-ctas">
            <button
              type="button"
              className="cta-find"
              onClick={() =>
                navigate(isAuthenticated ? "/talents" : "/auth/register?mode=client")
              }
            >
              Find Talent <span aria-hidden="true">→</span>
            </button>
            <button
              type="button"
              className="cta-join"
              onClick={() => navigate("/opportunities")}
            >
              Join as Student <span aria-hidden="true">→</span>
            </button>
          </div>
        </div>
      </section>

      <section className="how-it-works">
        <div className="how-it-works-container">
          <h2 className="how-it-works-main-title">Two sides. One platform.</h2>

          <div className="how-it-works-grid">
            <div className="how-it-works-column client">
              <h3 className="column-title">For Clients</h3>
              <div className="step">
                <span className="step-num">01</span>
                <div className="step-content">
                  <h4>Browse talent</h4>
                  <p>Browse verified Aalto students with portfolios that fit.</p>
                </div>
              </div>
              <div className="step">
                <span className="step-num">02</span>
                <div className="step-content">
                  <h4>Post an opportunity</h4>
                  <p>Describe your needs, scope and budget in minutes.</p>
                </div>
              </div>
              <div className="step">
                <span className="step-num">03</span>
                <div className="step-content">
                  <h4>Connect &amp; collaborate</h4>
                  <p>Reach out directly to students and work out the details together.</p>
                </div>
              </div>
            </div>

            <div className="how-it-works-column student">
              <h3 className="column-title">For Aalto Students</h3>
              <div className="step">
                <span className="step-num">01</span>
                <div className="step-content">
                  <h4>Verify with Aalto</h4>
                  <p>Sign up using your Aalto email account in seconds.</p>
                </div>
              </div>
              <div className="step">
                <span className="step-num">02</span>
                <div className="step-content">
                  <h4>Build your profile</h4>
                  <p>Showcase your portfolio, skills, and services.</p>
                </div>
              </div>
              <div className="step">
                <span className="step-num">03</span>
                <div className="step-content">
                  <h4>Get hired</h4>
                  <p>Receive project invites and grow your real-world experience.</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="trust-section">
        <div className="trust-container">
          <h2 className="trust-main-title">Built for trust and quality.</h2>
          <div className="trust-grid">
            <div className="trust-card">
              <h4>Verified Aalto Identity</h4>
              <p>
                Every student is verified through Aalto University, so you know exactly who you&apos;re hiring.
              </p>
            </div>
            <div className="trust-card">
              <h4>Transparent Portfolios</h4>
              <p>
                Real work, real reviews. Browse complete portfolios before you reach out.
              </p>
            </div>
            <div className="trust-card">
              <h4>Project-Based Work</h4>
              <p>
                Built for short, scoped engagements — no long contracts or overhead.
              </p>
            </div>
          </div>
        </div>
      </section>

      <footer className="landing-footer" aria-label="Legal">
        <p className="landing-footer-line">
          <Link to="/privacy">Privacy Policy</Link>
          <span className="landing-footer-sep"> · </span>
          <Link to="/terms">Terms of Use</Link>
        </p>
        <p className="landing-footer-meta">Last updated: 17 April 2026</p>
      </footer>
    </main>
  );
}

export default Landing;
