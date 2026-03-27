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
            return { success: false, error: error.message, code: error.code };
        } finally {
            setLoading(false);
        }
    }, [setAuth, setLoading]);

    const register = useCallback(async (data) => {
        setLoading(true);
        try {
            const response = await authApi.register(data);
            const { user, access_token, refresh_token } = response.data?.data || {};
            if (user && access_token) {
                setAuth(user, access_token, refresh_token);
                return { success: true, autoLogin: true };
            }
            return { success: true, autoLogin: false };
        } catch (error) {
            return { success: false, error: error.message };
        } finally {
            setLoading(false);
        }
    }, [setAuth, setLoading]);

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
