import { apiPost } from "./client";
export function generateQuestions(payload) {
    return apiPost("/ai/generate", payload);
}
export function testAiConnection() {
    return apiPost("/ai/test", {});
}
