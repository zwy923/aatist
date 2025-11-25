import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1",
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

// 响应拦截器：处理401错误和token刷新
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    
    // 如果是401错误且不是刷新token请求，尝试刷新token
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      const refreshToken = localStorage.getItem("refresh_token");
      if (refreshToken) {
        try {
          const response = await axios.post(
            `${import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1"}/auth/refresh`,
            { refresh_token: refreshToken },
            {
              headers: {
                // 刷新token时不需要Authorization header
              }
            }
          );
          
          if (response.data.success && response.data.data) {
            const { access_token, refresh_token: newRefreshToken } = response.data.data;
            localStorage.setItem("token", access_token);
            if (newRefreshToken) {
              localStorage.setItem("refresh_token", newRefreshToken);
            }
            
            originalRequest.headers.Authorization = `Bearer ${access_token}`;
            return api(originalRequest);
          }
        } catch (refreshError) {
          // 刷新失败，清除token并跳转登录
          localStorage.removeItem("token");
          localStorage.removeItem("refresh_token");
          localStorage.removeItem("user");
          window.location.href = "/auth/login";
          return Promise.reject(refreshError);
        }
      }
    }
    
    // 其他401错误，清除token
    if (error.response?.status === 401) {
      localStorage.removeItem("token");
      localStorage.removeItem("refresh_token");
      localStorage.removeItem("user");
      if (window.location.pathname !== "/auth/login" && window.location.pathname !== "/auth/register") {
        window.location.href = "/auth/login";
      }
    }
    
    return Promise.reject(error);
  }
);

// 认证API
export const authAPI = {
  login: async (email, password) => {
    const response = await api.post("/auth/login", { email, password });
    // 响应格式: { success: true, data: { user, access_token, refresh_token } }
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
  
  register: async (payload) => {
    // payload格式: { name, email, password, role (required), profile (required) }
    // profile格式: { studentId?, school?, faculty?, organizationName?, contactTitle? }
    try {
      const response = await api.post("/auth/register", payload);
      // 响应格式: { success: true, data: { user, access_token, refresh_token } }
      if (response.data.success && response.data.data) {
        return response.data.data;
      }
      return response.data;
    } catch (error) {
      // 确保错误被正确抛出，不要在这里处理，让调用方处理
      throw error;
    }
  },
  
  refreshToken: async (refreshToken) => {
    const response = await api.post("/auth/refresh", { refresh_token: refreshToken });
    // 响应格式: { success: true, data: { access_token, refresh_token } }
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
  
  logout: async () => {
    const response = await api.post("/auth/logout");
    return response.data;
  },
  
  getCurrentUser: async () => {
    const response = await api.get("/users/me");
    // 响应格式: { success: true, data: { ...user } }
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
  
  verifyEmail: async (token) => {
    const response = await api.get("/auth/verify", {
      params: { token },
    });
    return response.data;
  },
};

// Opportunities API
export const opportunitiesAPI = {
  getOpportunities: async (params = {}) => {
    const response = await api.get("/opportunities", { params });
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
  
  getOpportunity: async (id) => {
    const response = await api.get(`/opportunities/${id}`);
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
};

// Events API
export const eventsAPI = {
  getEvents: async (params = {}) => {
    const response = await api.get("/events", { params });
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
  
  getEvent: async (id) => {
    const response = await api.get(`/events/${id}`);
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
};

// Portfolio API
export const portfolioAPI = {
  getMyPortfolio: async () => {
    const response = await api.get("/users/me/portfolio");
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
  
  getUserPortfolio: async (userId) => {
    const response = await api.get(`/users/${userId}/portfolio`);
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
  
  getProject: async (projectId) => {
    const response = await api.get(`/portfolio/${projectId}`);
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
};

// Users API
export const usersAPI = {
  getUser: async (userId) => {
    const response = await api.get(`/users/${userId}`);
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
  
  getUserSummary: async (userId) => {
    const response = await api.get(`/users/${userId}/summary`);
    if (response.data.success && response.data.data) {
      return response.data.data;
    }
    return response.data;
  },
};

export default api;