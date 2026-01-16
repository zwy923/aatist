import apiClient from '../../../shared/api/client';

export const opportunitiesApi = {
    getOpportunities: (params) => apiClient.get('/opportunities', { params }),
    getOpportunity: (id) => apiClient.get(`/opportunities/${id}`),
    applyToOpportunity: (id, data) => apiClient.post(`/opportunities/${id}/applications`, data),
};
