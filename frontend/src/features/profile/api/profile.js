import apiClient from '../../../shared/api/client';

export const profileApi = {
    getMyProfile: () => apiClient.get('/users/me'),
    getServices: () => apiClient.get('/users/me/services'),
    createService: (data) => apiClient.post('/users/me/services', data),
    updateService: (id, data) => apiClient.patch(`/users/me/services/${id}`, data),
    deleteService: (id) => apiClient.delete(`/users/me/services/${id}`),
    updateProfile: (data) => apiClient.patch('/users/me', data),
    uploadAvatar: (file) => {
        const formData = new FormData();
        formData.append('avatar', file);
        return apiClient.post('/users/me/avatar', formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
        });
    },
    uploadProfileBanner: (file) => {
        const formData = new FormData();
        formData.append('banner', file);
        return apiClient.post('/users/me/banner', formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
        });
    },
    changePassword: (currentPassword, newPassword) =>
        apiClient.patch('/users/me/password', { current_password: currentPassword, new_password: newPassword }),
    getSavedItems: (params) => apiClient.get('/users/me/saved', { params }),
    saveItem: (type, targetId) => apiClient.post('/users/me/saved', { item_type: type, item_id: targetId }),
    removeSavedItem: (savedItemId) => apiClient.delete(`/users/me/saved/${savedItemId}`),
    removeSavedItemByTarget: (type, targetId) =>
        apiClient.delete('/users/me/saved', { params: { type, targetId } }),
    getMyApplications: () => apiClient.get('/users/me/applications'),
    searchUsers: (params) => apiClient.get('/users/search', { params }),
    getPublicProfile: (id) => apiClient.get(`/users/${id}`),
};

export const portfolioApi = {
    getPublicPortfolios: (params) => apiClient.get('/portfolio', { params }),
    getPortfolioById: (id) => apiClient.get(`/portfolio/${id}`),
    getUserPortfolio: (id) => apiClient.get(`/users/${id}/portfolio`),
    getMyPortfolio: () => apiClient.get('/users/me/portfolio'),
    uploadProjectCover: (file) => {
        const formData = new FormData();
        formData.append('file', file);
        return apiClient.post('/files/upload?type=project_cover', formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
        });
    },
    createPortfolioItem: (data) => apiClient.post('/users/me/portfolio', data),
    updatePortfolioItem: (id, data) => apiClient.patch(`/users/me/portfolio/${id}`, data),
    deletePortfolioItem: (id) => apiClient.delete(`/users/me/portfolio/${id}`),
};
