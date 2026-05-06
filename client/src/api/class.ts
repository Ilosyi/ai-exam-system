import { apiDelete, apiGet, apiPost, apiPut } from "./client";

export interface ClassItem {
    id: number;
    name: string;
    teacherId: number;
}

export interface ClassListFilters {
    keyword?: string;
    teacherId?: number;
    page?: number;
    pageSize?: number;
    [key: string]: string | number | undefined;
}

export interface ClassListResponse {
    data: ClassItem[];
    total: number;
}

export interface ClassPayload {
    name: string;
    teacherId?: number;
}

export interface ClassStudent {
    id: number;
    username: string;
    status: string;
    classId?: number;
    inClass: boolean;
}

export interface ClassStudentListResponse {
    data: ClassStudent[];
    total: number;
}

export interface StudentExamRecord {
    attemptId: number;
    paperId: number;
    paperTitle: string;
    status: string;
    totalScore?: number;
    startedAt: string;
    submittedAt?: string;
}

export interface StudentExamListResponse {
    data: StudentExamRecord[];
    total: number;
}

export async function fetchClasses(filters: ClassListFilters = {}): Promise<ClassListResponse> {
    return apiGet<ClassListResponse>("/classes", filters);
}

export async function createClass(payload: ClassPayload): Promise<{ data: ClassItem; message: string }> {
    return apiPost<{ data: ClassItem; message: string }>("/classes", payload);
}

export async function updateClass(id: number, payload: Partial<ClassPayload>): Promise<{ data: ClassItem; message: string }> {
    return apiPut<{ data: ClassItem; message: string }>(`/classes/${id}`, payload);
}

export async function deleteClass(id: number): Promise<void> {
    return apiDelete<void>(`/classes/${id}`);
}

export async function fetchClassStudents(
    classId: number,
    filters: { keyword?: string; status?: string; scope?: "class" | "all"; page?: number; pageSize?: number } = {},
): Promise<ClassStudentListResponse> {
    return apiGet<ClassStudentListResponse>(`/classes/${classId}/students`, filters as Record<string, string | number | undefined>);
}

export async function batchEditClassStudents(
    classId: number,
    action: "add" | "remove",
    studentIds: number[],
): Promise<{ message: string }> {
    return apiPost<{ message: string }>(`/classes/${classId}/students/batch-edit`, { action, studentIds });
}

export async function fetchStudentExams(
    classId: number,
    studentId: number,
    filters: { page?: number; pageSize?: number } = {},
): Promise<StudentExamListResponse> {
    return apiGet<StudentExamListResponse>(`/classes/${classId}/students/${studentId}/exams`, filters as Record<string, string | number | undefined>);
}
