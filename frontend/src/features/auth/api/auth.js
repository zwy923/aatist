import apiClient from '../../../shared/api/client';

export const authApi = {
    login: (credentials) => apiClient.post('/auth/login', credentials, { skipAuth: true }),
    register: (data) => apiClient.post('/auth/register', data, { skipAuth: true }),
    logout: () => apiClient.post('/auth/logout'),
    refresh: () => apiClient.post('/auth/refresh', {}, { withCredentials: true }),
    verifyEmail: (token) => apiClient.get(`/auth/verify?token=${token}`, { skipAuth: true }),
};
