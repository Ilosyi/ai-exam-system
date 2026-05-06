import { apiDelete, apiGet, apiPost, apiPut } from "./client";

export interface AuthUser {
  id: number;
  username: string;
  role: "admin" | "teacher" | "student";
  classId: number | null;
  status: string;
}

export interface AuthPayload {
  token: string;
  user: AuthUser;
}

export interface LoginInput {
  username: string;
  password: string;
}

export interface RegisterInput {
  username: string;
  password: string;
  role: "teacher" | "student";
  classId?: number;
}

export interface ChangePasswordInput {
  oldPassword: string;
  newPassword: string;
}

export interface UserListFilters {
  keyword?: string;
  role?: "admin" | "teacher" | "student";
  status?: "active" | "disabled";
  classId?: number;
  page?: number;
  pageSize?: number;
  [key: string]: string | number | undefined;
}

export interface UserListResponse {
  data: AuthUser[];
  total: number;
}

export interface UpdateUserInput {
  role?: "admin" | "teacher" | "student";
  classId?: number;
  status?: "active" | "disabled";
}

export async function login(input: LoginInput): Promise<{ data: AuthPayload }> {
  return apiPost<{ data: AuthPayload }>("/auth/login", input);
}

export async function register(input: RegisterInput): Promise<{ data: AuthPayload }> {
  return apiPost<{ data: AuthPayload }>("/auth/register", input);
}

export async function fetchCurrentUser(): Promise<{ data: AuthUser }> {
  return apiGet<{ data: AuthUser }>("/auth/me");
}

export async function refreshToken(token: string): Promise<{ data: AuthPayload }> {
  return apiPost<{ data: AuthPayload }>("/auth/refresh", { token });
}

export async function logoutApi(): Promise<{ message: string }> {
  return apiPost<{ message: string }>("/auth/logout", {});
}

export async function changePassword(input: ChangePasswordInput): Promise<{ message: string }> {
  return apiPost<{ message: string }>("/auth/change-password", input);
}

export async function fetchUsers(filters: UserListFilters): Promise<UserListResponse> {
  return apiGet<UserListResponse>("/users", filters);
}

export async function updateUser(id: number, input: UpdateUserInput): Promise<{ data: AuthUser; message: string }> {
  return apiPut<{ data: AuthUser; message: string }>(`/users/${id}`, input);
}

export async function deleteUser(id: number): Promise<void> {
  return apiDelete<void>(`/users/${id}`);
}
