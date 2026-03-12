import axios from 'axios';
import useAuthStore from '../stores/authStore';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

const apiClient = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Request Interceptor
apiClient.interceptors.request.use(
    (config) => {
        // Check if auth is explicitly disabled for this request
        if (config.skipAuth) {
            return config;
        }

        const token = useAuthStore.getState().accessToken;
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => Promise.reject(error)
);

// Response Interceptor
let isRefreshing = false;
let failedQueue = [];

const getApiErrorMeta = (error) => {
    const payload = error?.response?.data || {};
    const wrapped = payload?.error || {};
    const code = wrapped.code || payload.code || null;
    const message = wrapped.message || payload.message || error.message || 'An unexpected error occurred';
    return { code, message, payload };
};

const isMeEndpoint = (url = '') => {
    if (!url) return false;
    return url.includes('/users/me');
};

const processQueue = (error, token = null) => {
    failedQueue.forEach((prom) => {
        if (error) {
            prom.reject(error);
        } else {
            prom.resolve(token);
        }
    });
    failedQueue = [];
};

apiClient.interceptors.response.use(
    (response) => response,
    async (error) => {
        const originalRequest = error.config;
        const status = error.response?.status;
        const { code, message, payload } = getApiErrorMeta(error);

        // Handle 401 Unauthorized
        if (status === 401 && !originalRequest._retry) {
            if (isRefreshing) {
                return new Promise((resolve, reject) => {
                    failedQueue.push({ resolve, reject });
                })
                    .then((token) => {
                        originalRequest.headers.Authorization = `Bearer ${token}`;
                        return apiClient(originalRequest);
                    })
                    .catch((err) => Promise.reject(err));
            }

            originalRequest._retry = true;
            isRefreshing = true;

            try {
                // Attempt to refresh token
                const refreshToken = localStorage.getItem('refresh_token');
                const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
                    refresh_token: refreshToken
                }, {
                    withCredentials: true // Important for httpOnly cookies
                });

                const { access_token, refresh_token: newRefreshToken, user: newUser } = response.data.data;
                const currentUser = useAuthStore.getState().user;

                // Update store
                useAuthStore.getState().setAuth(newUser || currentUser, access_token, newRefreshToken);

                processQueue(null, access_token);
                isRefreshing = false;

                // Retry original request
                originalRequest.headers.Authorization = `Bearer ${access_token}`;
                return apiClient(originalRequest);
            } catch (refreshError) {
                processQueue(refreshError, null);
                isRefreshing = false;

                // Refresh failed, logout user
                useAuthStore.getState().logout();
                return Promise.reject(refreshError);
            }
        }

        // Session invalidation guard:
        // - /users/me returns USER_NOT_FOUND after backend data reset
        // - or token-related 401s that cannot be refreshed
        const isTokenIssue =
            code === 'INVALID_TOKEN' ||
            code === 'TOKEN_EXPIRED' ||
            code === 'UNAUTHORIZED';
        const isCurrentUserMissing = status === 404 && code === 'USER_NOT_FOUND' && isMeEndpoint(originalRequest?.url);
        if (isCurrentUserMissing || (status === 401 && isTokenIssue)) {
            useAuthStore.getState().logout();
        }

        // Unified Error Handling
        const appError = new Error(message);
        appError.status = status;
        appError.code = code;
        appError.data = payload;

        return Promise.reject(appError);
    }
);

export default apiClient;
