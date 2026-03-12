'use client';

import { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { apiClient } from '@/lib/api-client';

interface AuthUser {
  id: string;
  username: string;
  email: string;
  role: string;
  email_verified_at?: string | null;
}

interface AuthContextValue {
  user: AuthUser | null;
  isLoggedIn: boolean;
  loading: boolean;
  login: (token: string, refreshToken?: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue>({
  user: null,
  isLoggedIn: false,
  loading: true,
  login: async (_token: string, _refreshToken?: string) => {},
  logout: () => {},
});

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchMe = useCallback(async (token: string) => {
    apiClient.setToken(token);
    try {
      const data = await apiClient.getMe();
      setUser({
        id: data.id,
        username: data.username,
        email: data.email,
        role: data.role,
        email_verified_at: data.email_verified_at ?? null,
      });
    } catch {
      apiClient.setToken(null);
      setUser(null);
    }
  }, []);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) {
      fetchMe(token).finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, [fetchMe]);

  const login = useCallback(async (token: string, refreshToken?: string) => {
    if (refreshToken) apiClient.setRefreshToken(refreshToken);
    await fetchMe(token);
  }, [fetchMe]);

  const logout = useCallback(() => {
    apiClient.setToken(null);
    document.cookie = '_auth=; path=/; max-age=0';
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, isLoggedIn: !!user, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
