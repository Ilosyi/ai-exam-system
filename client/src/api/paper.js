import { apiGet, apiPost, apiPut, apiDelete } from "./client";
export async function fetchPapers(filters) {
    return apiGet("/papers", filters);
}
export async function fetchPaper(id) {
    return apiGet(`/papers/${id}`);
}
export async function generatePaper(payload) {
    return apiPost("/papers/generate", payload);
}
export async function createPaper(payload) {
    return apiPost("/papers", payload);
}
export async function updatePaper(id, payload) {
    return apiPut(`/papers/${id}`, payload);
}
export async function deletePaper(id) {
    return apiDelete(`/papers/${id}`);
}
export async function replaceQuestion(paperId, itemId, questionId) {
    return apiPost(`/papers/${paperId}/replace-question`, { itemId, questionId });
}
export async function deletePaperItem(paperId, itemId) {
    return apiDelete(`/papers/${paperId}/items/${itemId}`);
}
export async function publishPaper(paperId, payload) {
    return apiPost(`/papers/${paperId}/publish`, payload);
}
export async function unpublishPaper(paperId) {
    return apiPost(`/papers/${paperId}/unpublish`, {});
}
export async function fetchPaperSubmissions(paperId, classId) {
    const params = {};
    if (classId !== undefined)
        params.classId = classId;
    return apiGet(`/papers/${paperId}/submissions`, params);
}
