import { useCallback } from 'react';
import useAuthStore from '../../../shared/stores/authStore';
import { authApi } from '../api/auth';

export const useAuth = () => {
    const { user, accessToken, isAuthenticated, loading, setAuth, logout: clearAuth, setLoading } = useAuthStore();

    const login = useCallback(async (credentials) => {
        setLoading(true);
        try {
            const response = await authApi.login(credentials);
            const { user, access_token, refresh_token } = response.data.data;
            setAuth(user, access_token, refresh_token);
            return { success: true };
        } catch (error) {
            return { success: false, error: error.message };
        } finally {
            setLoading(false);
        }
    }, [setAuth, setLoading]);

    const register = useCallback(async (data) => {
        setLoading(true);
        try {
            await authApi.register(data);
            return { success: true };
        } catch (error) {
            return { success: false, error: error.message };
        } finally {
            setLoading(false);
        }
    }, [setLoading]);

    const logout = useCallback(async () => {
        try {
            await authApi.logout();
        } catch (error) {
            console.error('Logout API failed', error);
        } finally {
            clearAuth();
        }
    }, [clearAuth]);

    return {
        user,
        accessToken,
        isAuthenticated,
        loading,
        login,
        register,
        logout,
    };
};

export default useAuth;
