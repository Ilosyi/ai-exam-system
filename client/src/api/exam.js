import { apiGet, apiPost, apiPut } from "./client";
export async function fetchPublishedPapers() {
    return apiGet("/exam/published");
}
export async function startAttempt(paperId) {
    return apiPost(`/exam/papers/${paperId}/start`, {});
}
export async function getAttempt(attemptId) {
    return apiGet(`/exam/attempts/${attemptId}`);
}
export async function saveAnswers(attemptId, answers) {
    return apiPut(`/exam/attempts/${attemptId}/answers`, { answers });
}
export async function submitAttempt(attemptId) {
    return apiPost(`/exam/attempts/${attemptId}/submit`, {});
}
export async function getAttemptResult(attemptId) {
    return apiGet(`/exam/attempts/${attemptId}/result`);
}
export async function recordProctorEvent(attemptId, eventType, payloadJson) {
    return apiPost(`/exam/attempts/${attemptId}/events`, { eventType, payloadJson: payloadJson ?? "{}" });
}
