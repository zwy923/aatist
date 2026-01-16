import apiClient from '../../../shared/api/client';

export const profileApi = {
    getMyProfile: () => apiClient.get('/users/me'),
    updateProfile: (data) => apiClient.patch('/users/me', data),
    uploadAvatar: (file) => {
        const formData = new FormData();
        formData.append('avatar', file);
        return apiClient.post('/users/me/avatar', formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
        });
    },
    changePassword: (currentPassword, newPassword) =>
        apiClient.patch('/users/me/password', { current_password: currentPassword, new_password: newPassword }),
    getAvailability: () => apiClient.get('/users/me/availability'),
    updateAvailability: (data) => apiClient.patch('/users/me/availability', data),
    getSavedItems: (params) => apiClient.get('/users/me/saved', { params }),
    saveItem: (type, targetId) => apiClient.post('/users/me/saved', { type, targetId }),
    removeSavedItem: (savedItemId) => apiClient.delete(`/users/me/saved/${savedItemId}`),
    removeSavedItemByTarget: (type, targetId) =>
        apiClient.delete('/users/me/saved', { params: { type, targetId } }),
    getMyApplications: () => apiClient.get('/users/me/applications'),
    searchUsers: (params) => apiClient.get('/users/search', { params }),
    getPublicProfile: (id) => apiClient.get(`/users/${id}`),
};

export const portfolioApi = {
    getPublicPortfolios: (params) => apiClient.get('/portfolio', { params }),
    getUserPortfolio: (id) => apiClient.get(`/users/${id}/portfolio`),
    getMyPortfolio: () => apiClient.get('/users/me/portfolio'),
    createPortfolioItem: (data) => apiClient.post('/users/me/portfolio', data),
    updatePortfolioItem: (id, data) => apiClient.patch(`/users/me/portfolio/${id}`, data),
    deletePortfolioItem: (id) => apiClient.delete(`/users/me/portfolio/${id}`),
};
