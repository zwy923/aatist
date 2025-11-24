import React, { useEffect, useRef, useState } from "react";
import "./styles/global.css";
import { useUser } from "./store/userStore";
import LoginModal from "./components/LoginModal";

export default function App() {
  const textRef = useRef(null);
  const canvasRef = useRef(null);
  const { user, logout, isAuthenticated } = useUser();
  const [showLoginModal, setShowLoginModal] = useState(false);

  // 标题轻微浮动
  useEffect(() => {
    const el = textRef.current;
    let x = 0, y = 0, vx = 0.2, vy = 0.15;
    const tick = () => {
      x += vx; y += vy;
      if (x > 3 || x < -3) vx *= -1;
      if (y > 2 || y < -2) vy *= -1;
      el.style.transform = `translate(${x}px, ${y}px)`;
      requestAnimationFrame(tick);
    };
    tick();
  }, []);

  useEffect(() => {
    const canvas = canvasRef.current;
    const ctx = canvas.getContext("2d", { willReadFrequently: true });
    const DPR = window.devicePixelRatio || 1;

    const CSS_W = window.innerWidth;
    const CSS_H = 280;

    // 🔧 画布大小：bitmap 用设备像素，视图用 CSS 像素
    canvas.width  = Math.floor(CSS_W * DPR);
    canvas.height = Math.floor(CSS_H * DPR);
    canvas.style.width  = CSS_W + "px";
    canvas.style.height = CSS_H + "px";

    // 之后所有绘制与坐标使用“CSS 像素”
    ctx.setTransform(DPR, 0, 0, DPR, 0, 0); // 等价于 scale(DPR, DPR)

    const text = "Aatist.fi aims to become the bridge between campus creativity and real-world impact — a place where students learn, earn, and create together.";

    // ---- 自动换行 + 每行独立居中 ----
    const fontSize = CSS_W < 600 ? 18 : 22;
    ctx.font = `bold ${fontSize}px Inter, system-ui, sans-serif`;
    ctx.fillStyle = "white";

    const words = text.split(" ");
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

    // 先把文字真正画出来到位图（CSS 坐标）
    lines.forEach((l, i) => {
      const lw = ctx.measureText(l).width;
      const x = (CSS_W - lw) / 2;       // 每行独立居中（左起点）
      const y = startY + i * lineHeight;
      ctx.fillText(l, x, y);
    });

    // 从位图读取像素（⚠️ 用设备像素尺寸）
    const img = ctx.getImageData(0, 0, canvas.width, canvas.height);

    // 清空画布，准备粒子渲染（CSS 坐标系）
    ctx.clearRect(0, 0, CSS_W, CSS_H);

    // ---- 生成粒子：将设备像素坐标 ➜ CSS 坐标（除以 DPR）----
    const gap = 2;                       // CSS 像素间隔
    const step = Math.max(1, Math.floor(gap * DPR)); // 设备像素步长
    const particles = [];

    for (let y = 0; y < img.height; y += step) {
      for (let x = 0; x < img.width; x += step) {
        const i = (y * img.width + x) * 4;
        if (img.data[i + 3] > 128) {
          particles.push({
            // 初始随机位置（CSS）
            x: Math.random() * CSS_W,
            y: Math.random() * CSS_H,
            // 目标位置（把位图坐标转回 CSS）
            originX: x / DPR,
            originY: y / DPR,
            vx: 0, vy: 0, alpha: 0
          });
        }
      }
    }

    const mouse = { x: null, y: null, radius: 100 };
    const fixY = (eY) => eY - (window.innerHeight - CSS_H) / 2;

    const onMove = (e) => { mouse.x = e.clientX; mouse.y = fixY(e.clientY); };
    const onTouch = (e) => {
      const t = e.touches[0];
      mouse.x = t.clientX; mouse.y = fixY(t.clientY);
    };
    window.addEventListener("mousemove", onMove);
    window.addEventListener("touchmove", onTouch, { passive: true });

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
        p.x += p.vx; p.y += p.vy;
        p.vx *= 0.9; p.vy *= 0.9;
        if (p.alpha < 1) p.alpha += 0.02;

        ctx.fillStyle = `rgba(255,255,255,${p.alpha})`;
        ctx.shadowBlur = 5;
        ctx.shadowColor = "rgba(0,128,255,0.6)";
        ctx.fillRect(p.x, p.y, 2.2, 2.2);
      }
      requestAnimationFrame(render);
    };
    render();

    // 监听窗口尺寸变化，简单做法：刷新重建
    const onResize = () => window.location.reload();
    window.addEventListener("resize", onResize);

    return () => {
      window.removeEventListener("mousemove", onMove);
      window.removeEventListener("touchmove", onTouch);
      window.removeEventListener("resize", onResize);
    };
  }, []);

  return (
    <div className="container">
      <div className="glow"></div>
      <div className="overlay"></div>
      
      {/* 顶部导航栏 */}
      <div style={{
        position: "absolute",
        top: "2rem",
        right: "2rem",
        zIndex: 10,
        display: "flex",
        alignItems: "center",
        gap: "1rem"
      }}>
        {isAuthenticated ? (
          <>
            <div style={{
              color: "white",
              fontSize: "1rem",
              textShadow: "0 0 10px rgba(0, 150, 255, 0.6)"
            }}>
              欢迎, {user?.name}!
            </div>
            <button
              onClick={logout}
              style={{
                padding: "0.5rem 1rem",
                background: "rgba(255, 255, 255, 0.1)",
                border: "1px solid rgba(255, 255, 255, 0.3)",
                borderRadius: "8px",
                color: "white",
                cursor: "pointer",
                fontSize: "0.9rem",
                transition: "all 0.2s"
              }}
              onMouseEnter={(e) => {
                e.target.style.background = "rgba(255, 255, 255, 0.2)";
              }}
              onMouseLeave={(e) => {
                e.target.style.background = "rgba(255, 255, 255, 0.1)";
              }}
            >
              退出
            </button>
          </>
        ) : (
          <button
            onClick={() => setShowLoginModal(true)}
            style={{
              padding: "0.75rem 1.5rem",
              background: "linear-gradient(135deg, #007bff 0%, #0056b3 100%)",
              border: "none",
              borderRadius: "8px",
              color: "white",
              cursor: "pointer",
              fontSize: "1rem",
              fontWeight: 600,
              boxShadow: "0 4px 12px rgba(0, 128, 255, 0.4)",
              transition: "all 0.2s"
            }}
            onMouseEnter={(e) => {
              e.target.style.transform = "translateY(-2px)";
              e.target.style.boxShadow = "0 6px 16px rgba(0, 128, 255, 0.5)";
            }}
            onMouseLeave={(e) => {
              e.target.style.transform = "translateY(0)";
              e.target.style.boxShadow = "0 4px 12px rgba(0, 128, 255, 0.4)";
            }}
          >
            登录
          </button>
        )}
      </div>
      
      <h1 ref={textRef} className="main-text">Aatist.fi</h1>
      <canvas ref={canvasRef} className="particle-canvas" />
      
      {showLoginModal && (
        <LoginModal onClose={() => setShowLoginModal(false)} />
      )}
    </div>
  );
}
