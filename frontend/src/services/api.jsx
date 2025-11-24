import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || "http://localhost:8000",
});

// 请求拦截器：自动添加token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器：处理401错误
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem("token");
      localStorage.removeItem("user");
      window.location.reload();
    }
    return Promise.reject(error);
  }
);

// 认证API
export const authAPI = {
  login: async (email, password) => {
    const response = await api.post("/api/v1/auth/login", { email, password });
    return response.data;
  },
  
  register: async (name, email, password) => {
    const response = await api.post("/api/v1/auth/register", {
      name,
      email,
      password,
    });
    return response.data;
  },
  
  getCurrentUser: async () => {
    const response = await api.get("/api/v1/auth/me");
    return response.data;
  },
  
  verifyEmail: async (token) => {
    const response = await api.get("/api/v1/auth/verify", {
      params: { token },
    });
    return response.data;
  },
};

export default api;