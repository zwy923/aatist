import apiClient from '../../../shared/api/client';

export const opportunitiesApi = {
    getOpportunities: (params, axiosConfig = {}) => apiClient.get('/opportunities', { params, ...axiosConfig }),
    getOpportunityLocations: (axiosConfig = {}) => apiClient.get('/opportunities/locations', axiosConfig),
    getOpportunity: (id, axiosConfig = {}) => apiClient.get(`/opportunities/${id}`, axiosConfig),
    getMyOpportunities: (params) => apiClient.get('/opportunities/me', { params }),
    createOpportunity: (data) => apiClient.post('/opportunities', data),
    applyToOpportunity: (id, data) => apiClient.post(`/opportunities/${id}/apply`, data),
    /** Creator-only: applications for a brief (paginated). */
    getOpportunityApplications: (id, params) =>
        apiClient.get(`/opportunities/${id}/applications`, { params }),
};
