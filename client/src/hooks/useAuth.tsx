import { createContext, useContext, useEffect, useState, type PropsWithChildren } from "react";
import { fetchCurrentUser, login as loginApi, register as registerApi, type AuthPayload, type AuthUser, type LoginInput, type RegisterInput } from "../api/auth";
import { AUTH_STORAGE_KEY, AUTH_UNAUTHORIZED_EVENT } from "../api/client";

interface AuthContextValue {
  user: AuthUser | null;
  token: string | null;
  loading: boolean;
  login: (input: LoginInput) => Promise<AuthPayload>;
  register: (input: RegisterInput) => Promise<AuthPayload>;
  logout: () => void;
  refreshUser: () => Promise<AuthUser | null>;
}

interface StoredSession {
  token: string;
  user: AuthUser;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

function readStoredSession(): StoredSession | null {
  const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as StoredSession;
  } catch {
    window.localStorage.removeItem(AUTH_STORAGE_KEY);
    return null;
  }
}

function writeStoredSession(session: StoredSession | null) {
  if (!session) {
    window.localStorage.removeItem(AUTH_STORAGE_KEY);
    return;
  }
  window.localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(session));
}

export function getDefaultRouteByRole(role: AuthUser["role"]): string {
  if (role === "student") {
    return "/home";
  }
  return "/questions";
}

export function AuthProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<StoredSession | null>(() => readStoredSession());
  const [loading, setLoading] = useState(true);

  const applySession = (nextSession: StoredSession | null) => {
    setSession(nextSession);
    writeStoredSession(nextSession);
  };

  const logout = () => {
    applySession(null);
  };

  const refreshUser = async () => {
    if (!session?.token) {
      setLoading(false);
      return null;
    }
    try {
      const res = await fetchCurrentUser();
      const nextSession = { token: session.token, user: res.data };
      applySession(nextSession);
      return res.data;
    } catch {
      applySession(null);
      return null;
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void refreshUser();
  }, []);

  useEffect(() => {
    const handleUnauthorized = () => {
      applySession(null);
      setLoading(false);
    };
    window.addEventListener(AUTH_UNAUTHORIZED_EVENT, handleUnauthorized);
    return () => window.removeEventListener(AUTH_UNAUTHORIZED_EVENT, handleUnauthorized);
  }, []);

  const login = async (input: LoginInput) => {
    const res = await loginApi(input);
    applySession(res.data);
    setLoading(false);
    return res.data;
  };

  const register = async (input: RegisterInput) => {
    const res = await registerApi(input);
    applySession(res.data);
    setLoading(false);
    return res.data;
  };

  return (
    <AuthContext.Provider
      value={{
        user: session?.user ?? null,
        token: session?.token ?? null,
        loading,
        login,
        register,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}
