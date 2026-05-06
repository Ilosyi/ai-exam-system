import { apiDelete, apiGet, apiPost, apiPut } from "./client";
export async function fetchClasses(filters = {}) {
    return apiGet("/classes", filters);
}
export async function createClass(payload) {
    return apiPost("/classes", payload);
}
export async function updateClass(id, payload) {
    return apiPut(`/classes/${id}`, payload);
}
export async function deleteClass(id) {
    return apiDelete(`/classes/${id}`);
}
export async function fetchClassStudents(classId, filters = {}) {
    return apiGet(`/classes/${classId}/students`, filters);
}
export async function batchEditClassStudents(classId, action, studentIds) {
    return apiPost(`/classes/${classId}/students/batch-edit`, { action, studentIds });
}
export async function fetchStudentExams(classId, studentId, filters = {}) {
    return apiGet(`/classes/${classId}/students/${studentId}/exams`, filters);
}
