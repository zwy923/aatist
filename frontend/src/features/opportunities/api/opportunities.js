import apiClient from '../../../shared/api/client';

export const opportunitiesApi = {
    getOpportunities: (params) => apiClient.get('/opportunities', { params }),
    getOpportunity: (id) => apiClient.get(`/opportunities/${id}`),
    getMyOpportunities: (params) => apiClient.get('/opportunities/me', { params }),
    createOpportunity: (data) => apiClient.post('/opportunities', data),
    applyToOpportunity: (id, data) => apiClient.post(`/opportunities/${id}/apply`, data),
};
