import React, { useCallback, useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import ChatBubbleOutlineIcon from "@mui/icons-material/ChatBubbleOutline";
import NotificationsNoneIcon from "@mui/icons-material/NotificationsNone";
import AccountCircleIcon from "@mui/icons-material/AccountCircle";
import { useAuth } from "../features/auth/hooks/useAuth";
import "./Landing.css";

const clamp = (value, min, max) => Math.min(Math.max(value, min), max);
const DISTURB_RADIUS = 170;

const CLIENT_LETTERS = [
  { char: "C", x: 0, y: 72, r: -31 },
  { char: "L", x: 76, y: 40, r: -27 },
  { char: "I", x: 145, y: 25, r: -17 },
  { char: "E", x: 212, y: 6, r: -3 },
  { char: "N", x: 286, y: -10, r: 9 },
  { char: "T", x: 355, y: 6, r: 15 },
];

const STUDENT_LETTERS = [
  { char: "A", x: 0, y: 78, r: -11 },
  { char: "A", x: 76, y: 60, r: -3 },
  { char: "T", x: 146, y: 35, r: 7 },
  { char: "I", x: 220, y: 25, r: 8 },
  { char: "S", x: 286, y: 31, r: 10 },
  { char: "T", x: 355, y: 36, r: 7 },
];

const createLetterPhysics = (letters) =>
  letters.map(() => ({
    x: 0,
    y: 0,
    angle: 0,
    vx: (Math.random() - 0.5) * 0.6,
    vy: 0,
    va: 0,
  }));

function Landing() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const heroRef = useRef(null);
  const letterRefs = useRef({ client: [], student: [] });
  const physicsRef = useRef({
    client: createLetterPhysics(CLIENT_LETTERS),
    student: createLetterPhysics(STUDENT_LETTERS),
  });
  const [activeSide, setActiveSide] = useState("client");
  const [cursor, setCursor] = useState({ x: 0, y: 0, visible: false });

  useEffect(() => {
    let animationId;

    const applyTransforms = () => {
      const groups = [
        { side: "client", letters: CLIENT_LETTERS },
        { side: "student", letters: STUDENT_LETTERS },
      ];

      groups.forEach(({ side, letters }) => {
        const states = physicsRef.current[side];
        const refs = letterRefs.current[side];

        states.forEach((state, index) => {
          state.vy += 0.2;
          state.vx += -state.x * 0.02;
          state.vy += -state.y * 0.024;
          state.va += -state.angle * 0.028;

          state.x += state.vx;
          state.y += state.vy;
          state.angle += state.va;

          state.vx *= 0.9;
          state.vy *= 0.9;
          state.va *= 0.86;

          if (state.y > 86) {
            state.y = 86;
            state.vy *= -0.48;
          }
          if (Math.abs(state.x) > 96) {
            state.x = clamp(state.x, -96, 96);
            state.vx *= -0.46;
          }

          const letterEl = refs[index];
          if (!letterEl) return;
          const base = letters[index];
          letterEl.style.transform = `translate3d(${base.x + state.x}px, ${
            base.y + state.y
          }px, 0) rotate(${base.r + state.angle}deg)`;
        });
      });

      animationId = requestAnimationFrame(applyTransforms);
    };

    applyTransforms();
    return () => cancelAnimationFrame(animationId);
  }, []);

  const disturbWords = useCallback((clientX, clientY, power = 1) => {
    const groups = [
      { side: "client", letters: CLIENT_LETTERS },
      { side: "student", letters: STUDENT_LETTERS },
    ];

    groups.forEach(({ side, letters }) => {
      const refs = letterRefs.current[side];
      const states = physicsRef.current[side];

      letters.forEach((_, index) => {
        const letterEl = refs[index];
        if (!letterEl) return;
        const rect = letterEl.getBoundingClientRect();
        const centerX = rect.left + rect.width / 2;
        const centerY = rect.top + rect.height / 2;
        const dx = centerX - clientX;
        const dy = centerY - clientY;
        const dist = Math.hypot(dx, dy) || 1;

        if (dist > DISTURB_RADIUS) return;
        const force = (1 - dist / DISTURB_RADIUS) * 2.2 * power;
        const state = states[index];
        state.vx += (dx / dist) * force;
        state.vy += (dy / dist) * force - 0.28 * power;
        state.va += ((Math.random() - 0.5) * 0.9 + dx * 0.002) * force;
      });
    });
  }, []);

  const handleHeroPointerMove = (event) => {
    if (event.pointerType === "touch") return;
    const rect = heroRef.current?.getBoundingClientRect();
    if (!rect) return;

    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;
    const inClient = x < rect.width / 2;
    const side = inClient ? "client" : "student";
    setActiveSide(side);
    setCursor({ x, y, visible: true });
    disturbWords(event.clientX, event.clientY, 0.22);
  };

  const handleHeroPointerLeave = () => {
    setCursor((prev) => ({ ...prev, visible: false }));
  };

  const handleHeroPointerDown = (event) => {
    if (event.pointerType === "touch") return;
    disturbWords(event.clientX, event.clientY, 1.15);
  };

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
          <button
            type="button"
            className="icon-button"
            aria-label="Messages"
            onClick={() => navigate("/messages")}
            disabled={!isAuthenticated}
          >
            <ChatBubbleOutlineIcon fontSize="small" />
          </button>
          <button
            type="button"
            className="icon-button"
            aria-label="Notifications"
            disabled={!isAuthenticated}
          >
            <NotificationsNoneIcon fontSize="small" />
          </button>
          <button
            type="button"
            className="icon-button"
            aria-label="Account"
            onClick={() => navigate(isAuthenticated ? "/dashboard" : "/auth/login")}
          >
            <AccountCircleIcon fontSize="small" />
          </button>
        </div>
      </header>

      <section
        className="split-hero"
        ref={heroRef}
        onPointerMove={handleHeroPointerMove}
        onPointerLeave={handleHeroPointerLeave}
        onPointerDown={handleHeroPointerDown}
      >
        <div
          className={`side-indicator ${cursor.visible ? "visible" : ""} ${
            activeSide === "client" ? "for-client" : "for-student"
          }`}
          style={{
            left: `${cursor.x}px`,
            top: `${cursor.y}px`,
          }}
          aria-hidden="true"
        >
          {activeSide === "client" ? "C" : "A"}
        </div>

        <article className="panel client-panel">
          <div className="panel-top">
            <h2>For Client</h2>
            <p>Publish, Find Services</p>
          </div>
          <button
            type="button"
            className="cta-button"
            onClick={() => navigate(isAuthenticated ? "/talents" : "/auth/login/client")}
          >
            <span>Find Services</span>
            <span aria-hidden="true">&rarr;</span>
          </button>
          <div className="hero-word hero-word-client" aria-hidden="true">
            {CLIENT_LETTERS.map((letter, index) => (
              <span
                key={`${letter.char}-${index}`}
                className="hero-letter"
                ref={(el) => {
                  letterRefs.current.client[index] = el;
                }}
              >
                {letter.char}
              </span>
            ))}
          </div>
        </article>

        <article className="panel student-panel">
          <div className="panel-top">
            <h2>For Student</h2>
            <p>Create Portfolio, Find Opportunities</p>
          </div>
          <button
            type="button"
            className="cta-button"
            onClick={() => navigate("/opportunities")}
          >
            <span>Explore Opportunities</span>
            <span aria-hidden="true">&rarr;</span>
          </button>
          <div className="hero-word hero-word-student" aria-hidden="true">
            {STUDENT_LETTERS.map((letter, index) => (
              <span
                key={`${letter.char}-${index}`}
                className="hero-letter"
                ref={(el) => {
                  letterRefs.current.student[index] = el;
                }}
              >
                {letter.char}
              </span>
            ))}
          </div>
        </article>
      </section>
    </main>
  );
}

export default Landing;
