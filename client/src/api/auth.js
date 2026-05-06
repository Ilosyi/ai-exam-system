import { apiDelete, apiGet, apiPost, apiPut } from "./client";
export async function login(input) {
    return apiPost("/auth/login", input);
}
export async function register(input) {
    return apiPost("/auth/register", input);
}
export async function fetchCurrentUser() {
    return apiGet("/auth/me");
}
export async function refreshToken(token) {
    return apiPost("/auth/refresh", { token });
}
export async function logoutApi() {
    return apiPost("/auth/logout", {});
}
export async function changePassword(input) {
    return apiPost("/auth/change-password", input);
}
export async function fetchUsers(filters) {
    return apiGet("/users", filters);
}
export async function updateUser(id, input) {
    return apiPut(`/users/${id}`, input);
}
export async function deleteUser(id) {
    return apiDelete(`/users/${id}`);
}
