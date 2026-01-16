import apiClient from '../../../shared/api/client';

export const eventsApi = {
    getEvents: (params) => apiClient.get('/events', { params }),
    getEvent: (id) => apiClient.get(`/events/${id}`),
    registerForEvent: (id) => apiClient.post(`/events/${id}/register`),
};
