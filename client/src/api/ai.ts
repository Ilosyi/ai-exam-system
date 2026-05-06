import { apiPost } from "./client";
import type { Question } from "../types/question";

export interface AiGeneratePayload {
  type: "single" | "multiple" | "coding";
  count: number;
  language: "go" | "cpp" | "java" | "javascript" | "python";
  keyword: string;
}

export interface AiGenerateResponse {
  prompt: string;
  raw: string;
  questions: Array<Pick<Question, "type" | "language" | "title" | "content" | "options" | "answers">>;
}

export function generateQuestions(payload: AiGeneratePayload) {
  return apiPost<AiGenerateResponse>("/ai/generate", payload);
}

export interface AiTestResponse {
  data: {
    model: string;
    baseURL: string;
    elapsedMs: number;
    reply: string;
    replyDigest: string;
  };
  message: string;
}

export function testAiConnection() {
  return apiPost<AiTestResponse>("/ai/test", {});
}
