import apiClient from '../../../shared/api/client';

const API_BASE_URL = '/community';

const communityApi = {
    getPosts: async (params) => {
        const response = await apiClient.get(`${API_BASE_URL}/posts`, { params });
        return response.data.data;
    },
    getPost: async (id) => {
        const response = await apiClient.get(`${API_BASE_URL}/posts/${id}`);
        return response.data.data;
    },
    createPost: async (data) => {
        const response = await apiClient.post(`${API_BASE_URL}/posts`, data);
        return response.data.data;
    },
    updatePost: async (id, data) => {
        const response = await apiClient.patch(`${API_BASE_URL}/posts/${id}`, data);
        return response.data.data;
    },
    deletePost: async (id) => {
        const response = await apiClient.delete(`${API_BASE_URL}/posts/${id}`);
        return response.data.data;
    },
    likePost: async (id) => {
        const response = await apiClient.post(`${API_BASE_URL}/posts/${id}/like`);
        return response.data.data;
    },
    unlikePost: async (id) => {
        const response = await apiClient.delete(`${API_BASE_URL}/posts/${id}/like`);
        return response.data.data;
    },
    getComments: async (id) => {
        const response = await apiClient.get(`${API_BASE_URL}/posts/${id}/comments`);
        return response.data.data;
    },
    addComment: async (id, data) => {
        const response = await apiClient.post(`${API_BASE_URL}/posts/${id}/comments`, data);
        return response.data.data;
    },
    getMyPosts: async () => {
        const response = await apiClient.get(`${API_BASE_URL}/users/me/posts`);
        return response.data.data;
    }
};

export default communityApi;
