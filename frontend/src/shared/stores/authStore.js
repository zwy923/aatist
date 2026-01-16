import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

const useAuthStore = create(
    persist(
        (set, get) => ({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            loading: false,

            setAuth: (user, accessToken, refreshToken) => {
                if (refreshToken) {
                    localStorage.setItem('refresh_token', refreshToken);
                }
                set({
                    user,
                    accessToken,
                    refreshToken,
                    isAuthenticated: !!accessToken,
                    loading: false
                });
            },

            updateUser: (userData) => {
                const currentUser = get().user;
                set({ user: { ...currentUser, ...userData } });
            },

            logout: () => {
                set({
                    user: null,
                    accessToken: null,
                    isAuthenticated: false
                });
                // Clear tokens from localStorage if they were stored elsewhere
                localStorage.removeItem('refresh_token');
            },

            setLoading: (loading) => set({ loading }),
        }),
        {
            name: 'auth-storage',
            storage: createJSONStorage(() => localStorage),
            partialize: (state) => ({
                user: state.user,
                accessToken: state.accessToken,
                isAuthenticated: state.isAuthenticated
            }),
        }
    )
);

export default useAuthStore;
