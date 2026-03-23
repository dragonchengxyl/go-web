'use client';

import { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { apiClient } from '@/lib/api-client';

interface AuthUser {
  id: string;
  username: string;
  email: string;
  role: string;
  force_password_reset?: boolean;
  email_verified_at?: string | null;
}

interface AuthContextValue {
  user: AuthUser | null;
  isLoggedIn: boolean;
  loading: boolean;
  login: (token: string, refreshToken?: string, user?: Partial<AuthUser> | null) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue>({
  user: null,
  isLoggedIn: false,
  loading: true,
  login: async (_token: string, _refreshToken?: string, _user?: Partial<AuthUser> | null) => {},
  logout: () => {},
});

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);

  const normalizeUser = useCallback((data?: Partial<AuthUser> | null): AuthUser | null => {
    if (!data?.id || !data.username || !data.email || !data.role) {
      return null;
    }
    return {
      id: data.id,
      username: data.username,
      email: data.email,
      role: data.role,
      force_password_reset: data.force_password_reset ?? false,
      email_verified_at: data.email_verified_at ?? null,
    };
  }, []);

  const fetchMe = useCallback(async (token: string) => {
    apiClient.setToken(token);
    try {
      const data = await apiClient.getMe();
      setUser(normalizeUser(data));
    } catch {
      apiClient.setToken(null);
      setUser(null);
    }
  }, [normalizeUser]);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) {
      fetchMe(token).finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, [fetchMe]);

  const login = useCallback(async (token: string, refreshToken?: string, nextUser?: Partial<AuthUser> | null) => {
    apiClient.setToken(token);
    if (refreshToken) apiClient.setRefreshToken(refreshToken);
    const normalizedUser = normalizeUser(nextUser);
    if (normalizedUser) {
      setUser(normalizedUser);
      return;
    }
    await fetchMe(token);
  }, [fetchMe, normalizeUser]);

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
