import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useContext, useEffect, useState } from "react";
import { fetchCurrentUser, login as loginApi, register as registerApi } from "../api/auth";
import { AUTH_STORAGE_KEY, AUTH_UNAUTHORIZED_EVENT } from "../api/client";
const AuthContext = createContext(undefined);
function readStoredSession() {
    const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
    if (!raw) {
        return null;
    }
    try {
        return JSON.parse(raw);
    }
    catch {
        window.localStorage.removeItem(AUTH_STORAGE_KEY);
        return null;
    }
}
function writeStoredSession(session) {
    if (!session) {
        window.localStorage.removeItem(AUTH_STORAGE_KEY);
        return;
    }
    window.localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(session));
}
export function getDefaultRouteByRole(role) {
    if (role === "student") {
        return "/exam";
    }
    return "/questions";
}
export function AuthProvider({ children }) {
    const [session, setSession] = useState(() => readStoredSession());
    const [loading, setLoading] = useState(true);
    const applySession = (nextSession) => {
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
        }
        catch {
            applySession(null);
            return null;
        }
        finally {
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
    const login = async (input) => {
        const res = await loginApi(input);
        applySession(res.data);
        setLoading(false);
        return res.data;
    };
    const register = async (input) => {
        const res = await registerApi(input);
        applySession(res.data);
        setLoading(false);
        return res.data;
    };
    return (_jsx(AuthContext.Provider, { value: {
            user: session?.user ?? null,
            token: session?.token ?? null,
            loading,
            login,
            register,
            logout,
            refreshUser,
        }, children: children }));
}
export function useAuth() {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error("useAuth must be used within AuthProvider");
    }
    return context;
}
