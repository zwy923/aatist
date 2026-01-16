import { useState, useCallback } from 'react';
import useAuthStore from '../../../shared/stores/authStore';
import { profileApi, portfolioApi } from '../api/profile';

export const useProfile = () => {
    const { user, updateUser } = useAuthStore();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const updateProfile = useCallback(async (data) => {
        setLoading(true);
        setError(null);
        try {
            const response = await profileApi.updateProfile(data);
            const updatedUser = response.data.data;
            updateUser(updatedUser);
            return { success: true, data: updatedUser };
        } catch (err) {
            setError(err.message);
            return { success: false, error: err.message };
        } finally {
            setLoading(false);
        }
    }, [updateUser]);

    const uploadAvatar = useCallback(async (file) => {
        setLoading(true);
        try {
            const response = await profileApi.uploadAvatar(file);
            const { avatar_url } = response.data.data;
            updateUser({ avatar_url });
            return { success: true, avatarUrl: avatar_url };
        } catch (err) {
            return { success: false, error: err.message };
        } finally {
            setLoading(false);
        }
    }, [updateUser]);

    return {
        user,
        loading,
        error,
        updateProfile,
        uploadAvatar,
        // Add other profile/portfolio actions as needed
    };
};

export default useProfile;
