import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { authAPI } from "../services/api";
import "./Verify.css";

const STATUS = {
  LOADING: "loading",
  SUCCESS: "success",
  ERROR: "error",
};

export default function Verify() {
  const [status, setStatus] = useState(STATUS.LOADING);
  const [message, setMessage] = useState("正在验证邮箱链接…");
  const [hint, setHint] = useState("我们正在为你激活 Aatist 账户。");
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const lastVerifiedTokenRef = useRef(null);

  useEffect(() => {
    const token = searchParams.get("token");
    if (!token) {
      setStatus(STATUS.ERROR);
      setMessage("未提供验证 token。");
      setHint("请返回首页并重新请求验证邮件。");
      return;
    }

    if (lastVerifiedTokenRef.current === token) {
      return;
    }
    lastVerifiedTokenRef.current = token;

    setStatus(STATUS.LOADING);
    setMessage("正在验证邮箱链接…");
    setHint("请稍候，这通常只需要几秒钟。");

    authAPI
      .verifyEmail(token)
      .then((res) => {
        setStatus(STATUS.SUCCESS);
        setMessage(res.message || "邮箱验证成功！欢迎回到 Aatist。");
        setHint("我们会自动跳转到登录页面，或点击下方按钮立即登录。");
      })
      .catch((err) => {
        setStatus(STATUS.ERROR);
        setMessage(
          err.response?.data?.message ||
            err.response?.data?.detail ||
            "验证失败：链接无效或已过期。"
        );
        setHint("请重新请求验证邮件，或联系支持团队获取帮助。");
      });
  }, [searchParams]);

  useEffect(() => {
    if (status === STATUS.SUCCESS) {
      const timer = setTimeout(() => {
        navigate("/?modal=login", { replace: true });
      }, 3500);
      return () => clearTimeout(timer);
    }
  }, [status, navigate]);

  const buttonLabel =
    status === STATUS.SUCCESS ? "前往登录" : "返回首页";

  const isLoading = status === STATUS.LOADING;

  const handlePrimaryAction = () => {
    if (status === STATUS.SUCCESS) {
      navigate("/?modal=login", { replace: true });
    } else {
      navigate("/");
    }
  };

  const iconContent = useMemo(() => {
    if (status === STATUS.LOADING) return null;
    return status === STATUS.SUCCESS ? "✓" : "✗";
  }, [status]);

  return (
    <div className="verify-container">
      <div className="verify-content">
        {status === STATUS.LOADING ? (
          <div className="verify-spinner" aria-label="Loading" />
        ) : (
          <div
            className={`verify-icon ${
              status === STATUS.SUCCESS ? "success" : "error"
            }`}
          >
            {iconContent}
          </div>
        )}

        <h2 className="verify-status">{message}</h2>
        <p className="verify-hint">{hint}</p>

        <button
          className="verify-button"
          onClick={handlePrimaryAction}
          disabled={isLoading}
        >
          {isLoading ? "处理中…" : buttonLabel}
        </button>
      </div>
    </div>
  );
}

