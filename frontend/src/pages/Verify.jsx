import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { authAPI } from "../services/api";
import "./Verify.css";

export default function Verify() {
  const [message, setMessage] = useState("验证中...");
  const [isSuccess, setIsSuccess] = useState(false);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  useEffect(() => {
    const token = searchParams.get("token");
    if (token) {
      authAPI
        .verifyEmail(token)
        .then((res) => {
          setMessage(res.message || "邮箱验证成功！您现在可以登录了。");
          setIsSuccess(true);
        })
        .catch((err) => {
          setMessage(
            err.response?.data?.detail || "验证失败，链接无效或已过期。"
          );
          setIsSuccess(false);
        });
    } else {
      setMessage("未提供验证token。");
      setIsSuccess(false);
    }
  }, [searchParams]);

  return (
    <div className="verify-container">
      <div className="verify-content">
        <div className={`verify-icon ${isSuccess ? "success" : "error"}`}>
          {isSuccess ? "✓" : "✗"}
        </div>
        <h2>{message}</h2>
        <button
          className="verify-button"
          onClick={() => navigate("/")}
        >
          返回首页
        </button>
      </div>
    </div>
  );
}

