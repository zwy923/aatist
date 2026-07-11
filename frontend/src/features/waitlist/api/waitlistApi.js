import apiClient from '../../../shared/api/client';

export const waitlistApi = {
    join: (data) => apiClient.post('/waitlist', data, { skipAuth: true }),
    submitConsent: (id, data) => apiClient.patch(`/waitlist/${id}/consent`, data, { skipAuth: true }),
    logPageView: (data) => apiClient.post('/waitlist/pageview', data, { skipAuth: true }),
};
