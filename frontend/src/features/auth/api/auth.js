import apiClient from '../../../shared/api/client';

export const authApi = {
    login: (credentials) => apiClient.post('/auth/login', credentials, { skipAuth: true }),
    register: (data) => apiClient.post('/auth/register', data, { skipAuth: true }),
    /** Public route: skipAuth avoids attaching a bad/expired access token. Body revokes refresh in Redis when present. */
    logout: () => {
        const refresh_token =
            typeof localStorage !== 'undefined' ? localStorage.getItem('refresh_token') || '' : '';
        return apiClient.post('/auth/logout', { refresh_token }, { skipAuth: true });
    },
    refresh: () => apiClient.post('/auth/refresh', {}, { withCredentials: true }),
    verifyEmail: (token) => apiClient.get(`/auth/verify?token=${token}`, { skipAuth: true }),
};
