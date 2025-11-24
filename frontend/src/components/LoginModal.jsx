import React, { useState } from "react";
import { authAPI } from "../services/api";
import { useUser } from "../store/userStore";
import "./LoginModal.css";

export default function LoginModal({ onClose }) {
  const [isLogin, setIsLogin] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [registerSuccess, setRegisterSuccess] = useState(false);
  const [registeredEmail, setRegisteredEmail] = useState("");
  
  // 登录表单
  const [loginEmail, setLoginEmail] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  
  // 注册表单
  const [registerName, setRegisterName] = useState("");
  const [registerEmail, setRegisterEmail] = useState("");
  const [registerPassword, setRegisterPassword] = useState("");
  
  const { login } = useUser();

  const validateEmail = (email) => {
    return email.endsWith("@aalto.fi");
  };

  const handleLogin = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    if (!validateEmail(loginEmail)) {
      setError("仅支持Aalto学生邮箱（@aalto.fi）");
      setLoading(false);
      return;
    }

    try {
      const response = await authAPI.login(loginEmail, loginPassword);
      login(response.user, response.access_token);
      onClose();
    } catch (err) {
      setError(err.response?.data?.detail || "登录失败，请检查邮箱和密码");
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    if (!validateEmail(registerEmail)) {
      setError("仅支持Aalto学生邮箱（@aalto.fi）");
      setLoading(false);
      return;
    }

    if (registerPassword.length < 6) {
      setError("密码长度至少为6位");
      setLoading(false);
      return;
    }

    try {
      const response = await authAPI.register(
        registerName,
        registerEmail,
        registerPassword
      );
      // 注册成功，显示验证提示
      setRegisterSuccess(true);
      setRegisteredEmail(registerEmail);
      setError("");
      // 清空表单
      setRegisterName("");
      setRegisterEmail("");
      setRegisterPassword("");
    } catch (err) {
      setError(err.response?.data?.detail || "注册失败，请重试");
      setRegisterSuccess(false);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <button className="modal-close" onClick={onClose}>×</button>
        
        <h2>{isLogin ? "登录" : "注册"}</h2>
        
        {error && <div className="error-message">{error}</div>}
        {registerSuccess && (
          <div className="success-message">
            <p>✓ 注册成功！</p>
            <p>我们已向 <strong>{registeredEmail}</strong> 发送了验证邮件。</p>
            <p>请检查您的邮箱并点击验证链接以激活账户。</p>
            <p style={{ fontSize: "0.85rem", marginTop: "0.5rem", opacity: 0.8 }}>
              验证链接24小时内有效
            </p>
          </div>
        )}
        
        {isLogin ? (
          <form onSubmit={handleLogin}>
            <div className="form-group">
              <label>邮箱</label>
              <input
                type="email"
                value={loginEmail}
                onChange={(e) => setLoginEmail(e.target.value)}
                placeholder="your.name@aalto.fi"
                required
              />
            </div>
            <div className="form-group">
              <label>密码</label>
              <input
                type="password"
                value={loginPassword}
                onChange={(e) => setLoginPassword(e.target.value)}
                placeholder="请输入密码"
                required
              />
            </div>
            <button type="submit" disabled={loading} className="submit-btn">
              {loading ? "登录中..." : "登录"}
            </button>
            <p className="switch-mode">
              还没有账号？{" "}
              <span onClick={() => setIsLogin(false)}>立即注册</span>
            </p>
          </form>
        ) : !registerSuccess ? (
          <form onSubmit={handleRegister}>
            <div className="form-group">
              <label>姓名</label>
              <input
                type="text"
                value={registerName}
                onChange={(e) => setRegisterName(e.target.value)}
                placeholder="请输入姓名"
                required
              />
            </div>
            <div className="form-group">
              <label>邮箱</label>
              <input
                type="email"
                value={registerEmail}
                onChange={(e) => setRegisterEmail(e.target.value)}
                placeholder="your.name@aalto.fi"
                required
              />
            </div>
            <div className="form-group">
              <label>密码</label>
              <input
                type="password"
                value={registerPassword}
                onChange={(e) => setRegisterPassword(e.target.value)}
                placeholder="至少6位字符"
                required
                minLength={6}
              />
            </div>
            <button type="submit" disabled={loading} className="submit-btn">
              {loading ? "注册中..." : "注册"}
            </button>
            <p className="switch-mode">
              已有账号？{" "}
              <span onClick={() => setIsLogin(true)}>立即登录</span>
            </p>
          </form>
        ) : (
          <div>
            <button
              onClick={() => {
                setRegisterSuccess(false);
                setIsLogin(true);
              }}
              className="submit-btn"
            >
              返回登录
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

