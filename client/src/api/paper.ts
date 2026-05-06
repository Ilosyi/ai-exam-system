import { apiGet, apiPost, apiPut, apiDelete } from "./client";
import type { Paper, PaperFilters, GenerateRequest, GenerateResult, SavePaperRequest, PublishRequest } from "../types/paper";

interface PaperListResponse {
  data: Paper[];
  total: number;
}

interface PaperDetailResponse {
  data: Paper;
}

interface GenerateResponse {
  data: GenerateResult;
}

export async function fetchPapers(filters: PaperFilters): Promise<PaperListResponse> {
  return apiGet<PaperListResponse>("/papers", filters);
}

export async function fetchPaper(id: number): Promise<PaperDetailResponse> {
  return apiGet<PaperDetailResponse>(`/papers/${id}`);
}

export async function generatePaper(payload: GenerateRequest): Promise<GenerateResponse> {
  return apiPost<GenerateResponse>("/papers/generate", payload);
}

export async function createPaper(payload: SavePaperRequest): Promise<PaperDetailResponse> {
  return apiPost<PaperDetailResponse>("/papers", payload);
}

export async function updatePaper(id: number, payload: Partial<Paper>): Promise<PaperDetailResponse> {
  return apiPut<PaperDetailResponse>(`/papers/${id}`, payload);
}

export async function deletePaper(id: number): Promise<void> {
  return apiDelete<void>(`/papers/${id}`);
}

export async function replaceQuestion(paperId: number, itemId: number, questionId?: number): Promise<PaperDetailResponse> {
  return apiPost<PaperDetailResponse>(`/papers/${paperId}/replace-question`, { itemId, questionId });
}

export async function deletePaperItem(paperId: number, itemId: number): Promise<PaperDetailResponse> {
  return apiDelete<PaperDetailResponse>(`/papers/${paperId}/items/${itemId}`);
}

export async function publishPaper(paperId: number, payload: PublishRequest): Promise<{ message: string }> {
  return apiPost<{ message: string }>(`/papers/${paperId}/publish`, payload);
}

export async function unpublishPaper(paperId: number): Promise<{ message: string }> {
  return apiPost<{ message: string }>(`/papers/${paperId}/unpublish`, {});
}

export interface PaperSubmittedStudent {
  studentId: number;
  username: string;
  status: string;
  totalScore?: number;
  submittedAt?: string;
}

export interface PaperSubmissionStats {
  paperId: number;
  classId?: number;
  expectedCount: number;
  submittedCount: number;
  unsubmittedCount: number;
  submittedStudents: PaperSubmittedStudent[];
}

export async function fetchPaperSubmissions(paperId: number, classId?: number): Promise<{ data: PaperSubmissionStats }> {
  const params: Record<string, string | number | undefined> = {};
  if (classId !== undefined) params.classId = classId;
  return apiGet<{ data: PaperSubmissionStats }>(`/papers/${paperId}/submissions`, params);
}
