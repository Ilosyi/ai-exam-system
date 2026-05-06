export const AUTH_STORAGE_KEY = "question-bank-auth";
export const AUTH_UNAUTHORIZED_EVENT = "auth:unauthorized";
const API_BASE = import.meta.env.VITE_API_BASE ?? "/api";
function getAuthToken() {
    const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
    if (!raw) {
        return null;
    }
    try {
        const parsed = JSON.parse(raw);
        return parsed.token ?? null;
    }
    catch {
        window.localStorage.removeItem(AUTH_STORAGE_KEY);
        return null;
    }
}
async function handleResponse(res) {
    const raw = await res.text();
    if (!res.ok) {
        if (res.status === 401) {
            window.dispatchEvent(new Event(AUTH_UNAUTHORIZED_EVENT));
        }
        let message = raw || "请求失败";
        try {
            const parsed = JSON.parse(raw);
            message = parsed.message || parsed.error || message;
        }
        catch {
            // Keep the raw message when the response body is not JSON.
        }
        throw new Error(message);
    }
    if (!raw) {
        return undefined;
    }
    return JSON.parse(raw);
}
function buildHeaders() {
    const token = getAuthToken();
    return token ? { Authorization: `Bearer ${token}` } : {};
}
export async function apiGet(path, params) {
    const url = new URL(path, window.location.origin);
    url.pathname = `${API_BASE}${path}`;
    if (params) {
        Object.entries(params).forEach(([key, value]) => {
            if (value !== undefined && value !== "") {
                url.searchParams.set(key, String(value));
            }
        });
    }
    const res = await fetch(url.toString(), {
        headers: buildHeaders(),
    });
    return handleResponse(res);
}
export async function apiPost(path, body) {
    const res = await fetch(`${API_BASE}${path}`, {
        method: "POST",
        headers: { "Content-Type": "application/json", ...buildHeaders() },
        body: JSON.stringify(body),
    });
    return handleResponse(res);
}
export async function apiPut(path, body) {
    const res = await fetch(`${API_BASE}${path}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", ...buildHeaders() },
        body: JSON.stringify(body),
    });
    return handleResponse(res);
}
export async function apiDelete(path) {
    const res = await fetch(`${API_BASE}${path}`, {
        method: "DELETE",
        headers: buildHeaders(),
    });
    return handleResponse(res);
}
export async function apiDeleteJson(path, body) {
    const res = await fetch(`${API_BASE}${path}`, {
        method: "DELETE",
        headers: { "Content-Type": "application/json", ...buildHeaders() },
        body: JSON.stringify(body),
    });
    return handleResponse(res);
}
