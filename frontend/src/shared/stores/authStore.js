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
            /** Wall-clock start of this sign-in (ms); not reset on access-token refresh */
            sessionStartedAt: null,

            /**
             * @param {object} [meta]
             * @param {boolean} [meta.tokenRefresh] - true when tokens come from /auth/refresh (do not reset session clock)
             */
            setAuth: (user, accessToken, refreshToken, meta = {}) => {
                if (refreshToken) {
                    localStorage.setItem('refresh_token', refreshToken);
                }
                const tokenRefresh = meta.tokenRefresh === true;
                set((state) => ({
                    user,
                    accessToken,
                    refreshToken: refreshToken ?? state.refreshToken,
                    isAuthenticated: !!accessToken,
                    loading: false,
                    sessionStartedAt: tokenRefresh
                        ? state.sessionStartedAt
                        : Date.now(),
                }));
            },

            /** Legacy sessions without sessionStartedAt: anchor once for absolute timeout */
            ensureSessionAnchor: () =>
                set((state) => {
                    if (!state.isAuthenticated || !state.accessToken) return state;
                    if (state.sessionStartedAt != null) return state;
                    return { sessionStartedAt: Date.now() };
                }),

            updateUser: (userData) => {
                const currentUser = get().user;
                set({ user: { ...currentUser, ...userData } });
            },

            logout: () => {
                set({
                    user: null,
                    accessToken: null,
                    isAuthenticated: false,
                    sessionStartedAt: null,
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
                isAuthenticated: state.isAuthenticated,
                sessionStartedAt: state.sessionStartedAt,
            }),
        }
    )
);

export default useAuthStore;
