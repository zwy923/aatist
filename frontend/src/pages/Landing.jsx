import React, { useEffect, useRef, useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { Box, Button, Typography } from "@mui/material";
import { useAuth } from "../features/auth/hooks/useAuth";

// Constants
const PARTICLE_TEXT =
  "Aatist.fi aims to become the bridge between campus creativity and real-world impact — a place where students learn, earn, and create together.";
const CANVAS_HEIGHT = 280;
const PARTICLE_GAP = 2;
const MOUSE_RADIUS = 100;

// Custom hook for title floating animation
function useTitleFloat(ref) {
  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    let x = 0;
    let y = 0;
    let vx = 0.2;
    let vy = 0.15;
    let animationId;

    const tick = () => {
      x += vx;
      y += vy;
      if (x > 3 || x < -3) vx *= -1;
      if (y > 2 || y < -2) vy *= -1;
      el.style.transform = `translate(${x}px, ${y}px)`;
      animationId = requestAnimationFrame(tick);
    };

    tick();

    return () => {
      if (animationId) {
        cancelAnimationFrame(animationId);
      }
    };
  }, [ref]);
}

// Custom hook for particle animation
function useParticleAnimation(canvasRef) {
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    let animationId;
    let particles = [];
    const mouse = { x: null, y: null, radius: MOUSE_RADIUS };

    const initCanvas = () => {
      const ctx = canvas.getContext("2d", { willReadFrequently: true });
      const DPR = window.devicePixelRatio || 1;
      const CSS_W = window.innerWidth;
      const CSS_H = CANVAS_HEIGHT;

      canvas.width = Math.floor(CSS_W * DPR);
      canvas.height = Math.floor(CSS_H * DPR);
      canvas.style.width = CSS_W + "px";
      canvas.style.height = CSS_H + "px";

      ctx.setTransform(DPR, 0, 0, DPR, 0, 0);

      // Render text and extract particles
      const fontSize = CSS_W < 600 ? 18 : 22;
      ctx.font = `bold ${fontSize}px Inter, system-ui, sans-serif`;
      ctx.fillStyle = "white";

      const words = PARTICLE_TEXT.split(" ");
      const maxWidth = Math.min(CSS_W * 0.86, 900);
      const lines = [];
      let line = "";

      for (let i = 0; i < words.length; i++) {
        const test = line + words[i] + " ";
        if (ctx.measureText(test).width > maxWidth && i > 0) {
          lines.push(line.trim());
          line = words[i] + " ";
        } else {
          line = test;
        }
      }
      if (line.trim()) lines.push(line.trim());

      const lineHeight = Math.round(fontSize * 1.3);
      const totalH = lines.length * lineHeight;
      const startY = (CSS_H - totalH) / 2;

      lines.forEach((l, i) => {
        const lw = ctx.measureText(l).width;
        const x = (CSS_W - lw) / 2;
        const y = startY + i * lineHeight;
        ctx.fillText(l, x, y);
      });

      const img = ctx.getImageData(0, 0, canvas.width, canvas.height);
      ctx.clearRect(0, 0, CSS_W, CSS_H);

      // Generate particles
      const step = Math.max(1, Math.floor(PARTICLE_GAP * DPR));
      particles = [];

      for (let y = 0; y < img.height; y += step) {
        for (let x = 0; x < img.width; x += step) {
          const i = (y * img.width + x) * 4;
          if (img.data[i + 3] > 128) {
            particles.push({
              x: Math.random() * CSS_W,
              y: Math.random() * CSS_H,
              originX: x / DPR,
              originY: y / DPR,
              vx: 0,
              vy: 0,
              alpha: 0,
            });
          }
        }
      }

      return { ctx, CSS_W, CSS_H, particles: particles };
    };

    let canvasState = initCanvas();
    let { ctx, CSS_W, CSS_H } = canvasState;
    particles = canvasState.particles;

    const fixY = (eY) => eY - (window.innerHeight - CSS_H) / 2;

    const handleMouseMove = (e) => {
      mouse.x = e.clientX;
      mouse.y = fixY(e.clientY);
    };

    const handleTouchMove = (e) => {
      const t = e.touches[0];
      mouse.x = t.clientX;
      mouse.y = fixY(t.clientY);
    };

    const render = () => {
      ctx.clearRect(0, 0, CSS_W, CSS_H);
      for (const p of particles) {
        const dx = (mouse.x ?? -9999) - p.x;
        const dy = (mouse.y ?? -9999) - p.y;
        const dist = Math.hypot(dx, dy);
        if (dist < mouse.radius) {
          const f = (mouse.radius - dist) / mouse.radius;
          p.vx -= (dx / dist) * f * 3;
          p.vy -= (dy / dist) * f * 3;
        } else {
          p.vx += (p.originX - p.x) * 0.015;
          p.vy += (p.originY - p.y) * 0.015;
        }
        p.x += p.vx;
        p.y += p.vy;
        p.vx *= 0.9;
        p.vy *= 0.9;
        if (p.alpha < 1) p.alpha += 0.02;

        ctx.fillStyle = `rgba(255,255,255,${p.alpha})`;
        ctx.shadowBlur = 5;
        ctx.shadowColor = "rgba(0,128,255,0.6)";
        ctx.fillRect(p.x, p.y, 2.2, 2.2);
      }
      animationId = requestAnimationFrame(render);
    };

    const handleResize = () => {
      if (animationId) {
        cancelAnimationFrame(animationId);
      }
      canvasState = initCanvas();
      ctx = canvasState.ctx;
      CSS_W = canvasState.CSS_W;
      CSS_H = canvasState.CSS_H;
      particles = canvasState.particles;
      animationId = requestAnimationFrame(render);
    };

    render();

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("touchmove", handleTouchMove, { passive: true });
    window.addEventListener("resize", handleResize);

    return () => {
      if (animationId) {
        cancelAnimationFrame(animationId);
      }
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("touchmove", handleTouchMove);
      window.removeEventListener("resize", handleResize);
    };
  }, [canvasRef]);
}

export default function Landing() {
  const textRef = useRef(null);
  const canvasRef = useRef(null);
  const navigate = useNavigate();
  const { user, logout, isAuthenticated } = useAuth();

  useTitleFloat(textRef);
  useParticleAnimation(canvasRef);

  const handleJoinClick = useCallback(() => {
    navigate("/auth/register");
  }, [navigate]);

  const handleDashboardClick = useCallback(() => {
    navigate("/dashboard");
  }, [navigate]);

  const userDisplayName = useMemo(
    () => user?.name || user?.email || "User",
    [user]
  );

  return (
    <Box
      sx={{
        position: "relative",
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
        alignItems: "center",
        textAlign: "center",
        padding: 2,
        overflow: "hidden",
        background: "radial-gradient(ellipse at top left, #101820, #050505)",
      }}
    >
      {/* Background glow effect */}
      <Box
        sx={{
          position: "absolute",
          width: { xs: 300, md: 600 },
          height: { xs: 300, md: 600 },
          background: "radial-gradient(circle, rgba(0, 128, 255, 0.4), transparent 70%)",
          filter: "blur(160px)",
          zIndex: 0,
          animation: "pulse 8s infinite alternate ease-in-out",
          "@keyframes pulse": {
            "0%": { transform: "scale(1) translate(0, 0)", opacity: 0.8 },
            "100%": { transform: "scale(1.2) translate(30px, -30px)", opacity: 0.6 },
          },
        }}
      />

      {/* Overlay */}
      <Box
        sx={{
          position: "absolute",
          inset: 0,
          background: "radial-gradient(circle at center, rgba(255,255,255,0.03), transparent 70%)",
          pointerEvents: "none",
          zIndex: 1,
        }}
      />

      {/* Top navigation bar */}
      <Box
        sx={{
          position: "absolute",
          top: { xs: "1rem", md: "2rem" },
          right: { xs: "1rem", md: "2rem" },
          zIndex: 10,
          display: "flex",
          alignItems: "center",
          gap: 2,
        }}
      >
        {isAuthenticated ? (
          <>
            <Typography
              variant="body1"
              sx={{
                color: "white",
                fontSize: { xs: "0.9rem", md: "1rem" },
                textShadow: "0 0 10px rgba(0, 150, 255, 0.6)",
              }}
            >
              Welcome, {userDisplayName}!
            </Typography>
            <Button
              variant="outlined"
              onClick={handleDashboardClick}
              sx={{
                borderColor: "rgba(93, 224, 255, 0.3)",
                color: "#5de0ff",
                "&:hover": {
                  borderColor: "rgba(93, 224, 255, 0.5)",
                  background: "rgba(93, 224, 255, 0.1)",
                },
              }}
            >
              Dashboard
            </Button>
            <Button
              variant="outlined"
              onClick={logout}
              sx={{
                borderColor: "rgba(255, 255, 255, 0.3)",
                color: "white",
                "&:hover": {
                  borderColor: "rgba(255, 255, 255, 0.5)",
                  background: "rgba(255, 255, 255, 0.1)",
                },
              }}
            >
              Sign out
            </Button>
          </>
        ) : (
          <Button
            variant="contained"
            onClick={handleJoinClick}
            sx={{
              padding: { xs: "0.6rem 1.2rem", md: "0.75rem 1.5rem" },
              background: "linear-gradient(135deg, #007bff 0%, #7f5dff 100%)",
              color: "white",
              fontSize: { xs: "0.9rem", md: "1rem" },
              fontWeight: 600,
              boxShadow: "0 4px 12px rgba(0, 128, 255, 0.4)",
              "&:hover": {
                transform: "translateY(-2px)",
                boxShadow: "0 6px 16px rgba(0, 128, 255, 0.5)",
                background: "linear-gradient(135deg, #0066cc 0%, #6b4dd9 100%)",
              },
              transition: "all 0.2s",
            }}
          >
            Join Aatist
          </Button>
        )}
      </Box>

      {/* Main title */}
      <Typography
        ref={textRef}
        variant="h1"
        sx={{
          position: "relative",
          fontSize: { xs: "2.8rem", md: "4rem" },
          fontWeight: 800,
          textShadow: "0 0 25px rgba(0, 150, 255, 0.6)",
          animation: "fadeIn 2s ease-out",
          zIndex: 2,
          color: "white",
          "@keyframes fadeIn": {
            from: { opacity: 0, transform: "translateY(10px)" },
            to: { opacity: 1, transform: "translateY(0)" },
          },
        }}
      >
        Aatist.fi
      </Typography>

      {/* Particle canvas */}
      <Box
        component="canvas"
        ref={canvasRef}
        sx={{
          position: "relative",
          width: "100%",
          height: { xs: 180, md: 280 },
          marginTop: 2,
          zIndex: 2,
          pointerEvents: "none",
        }}
      />
    </Box>
  );
}
