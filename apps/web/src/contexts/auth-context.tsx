'use client';

import { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { apiClient } from '@/lib/api-client';
import { parseAccessTokenClaims } from '@/lib/access-control';

interface AuthUser {
  id: string;
  username: string;
  email: string;
  role: string;
  permissions: string[];
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

  const normalizeUser = useCallback((data?: Partial<AuthUser> | null, permissions: string[] = []): AuthUser | null => {
    if (!data?.id || !data.username || !data.email || !data.role) {
      return null;
    }
    return {
      id: data.id,
      username: data.username,
      email: data.email,
      role: data.role,
      permissions,
      force_password_reset: data.force_password_reset ?? false,
      email_verified_at: data.email_verified_at ?? null,
    };
  }, []);

  const fetchMe = useCallback(async (token: string) => {
    apiClient.setToken(token);
    const claims = parseAccessTokenClaims(token);
    try {
      const data = await apiClient.getMe();
      setUser(normalizeUser(data, claims?.permissions ?? []));
    } catch (err: any) {
      if (err?.status === 401 || err?.code === 40101 || err?.code === 40102) {
        apiClient.setToken(null);
        setUser(null);
      }
      // Keep the current session for transient errors such as rate limiting or timeouts.
      // This avoids logging users out just because one bootstrap request failed.
      return;
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
    const claims = parseAccessTokenClaims(token);
    const normalizedUser = normalizeUser(nextUser, claims?.permissions ?? []);
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
