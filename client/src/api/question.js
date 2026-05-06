import { apiDelete, apiDeleteJson, apiGet, apiPost, apiPut } from "./client";
export function fetchQuestions(filters) {
    return apiGet("/questions", filters);
}
export function createQuestion(payload) {
    return apiPost("/questions", payload);
}
export function updateQuestion(id, payload) {
    return apiPut(`/questions/${id}`, payload);
}
export function deleteQuestion(id) {
    return apiDelete(`/questions/${id}`);
}
export function deleteQuestionsBulk(ids) {
    return apiDeleteJson(`/questions`, { ids });
}
