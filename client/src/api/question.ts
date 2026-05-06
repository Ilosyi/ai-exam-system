import { apiDelete, apiDeleteJson, apiGet, apiPost, apiPut } from "./client";
import type { Question, QuestionFilters } from "../types/question";

export interface QuestionListResponse {
  data: Question[];
  total: number;
}

export function fetchQuestions(filters: QuestionFilters) {
  return apiGet<QuestionListResponse>("/questions", filters);
}

export function createQuestion(payload: Partial<Question>) {
  return apiPost<Question>("/questions", payload);
}

export function updateQuestion(id: number, payload: Partial<Question>) {
  return apiPut<Question>(`/questions/${id}`, payload);
}

export function deleteQuestion(id: number) {
  return apiDelete<void>(`/questions/${id}`);
}

export function deleteQuestionsBulk(ids: number[]) {
  return apiDeleteJson<void>(`/questions`, { ids });
}
