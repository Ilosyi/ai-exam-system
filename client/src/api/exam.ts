import { apiGet, apiPost, apiPut } from "./client";

export interface PublishedPaper {
  paperId: number;
  title: string;
  language: string;
  totalScore: number;
  startTime: string;
  endTime: string;
  duration: number; // 答题时长(分钟), 0=不限时
  attemptId?: number;
  attemptStatus?: "in_progress" | "submitted" | "timeout";
  attemptScore?: number | null;
  attemptStartedAt?: string;
  attemptSubmittedAt?: string;
}

export interface ExamAnswer {
  id: number;
  attemptId: number;
  questionId: number;
  answerJson: string;
  isCorrect: boolean | null;
  score: number | null;
}

export interface ExamAttempt {
  id: number;
  paperId: number;
  studentId: number;
  startedAt: string;
  submittedAt: string | null;
  status: "in_progress" | "submitted" | "timeout";
  totalScore: number | null;
  deadline: string | null; // 答题截止时间
  paper?: import("../types/paper").Paper;
  answers?: ExamAnswer[];
}

export async function fetchPublishedPapers(): Promise<{ data: PublishedPaper[] }> {
  return apiGet<{ data: PublishedPaper[] }>("/exam/published");
}

export async function startAttempt(paperId: number): Promise<{ message: string; data: ExamAttempt }> {
  return apiPost<{ message: string; data: ExamAttempt }>(`/exam/papers/${paperId}/start`, {});
}

export async function getAttempt(attemptId: number): Promise<{ data: ExamAttempt }> {
  return apiGet<{ data: ExamAttempt }>(`/exam/attempts/${attemptId}`);
}

export async function saveAnswers(attemptId: number, answers: { questionId: number; answerJson: string }[]): Promise<{ message: string }> {
  return apiPut<{ message: string }>(`/exam/attempts/${attemptId}/answers`, { answers });
}

export async function submitAttempt(attemptId: number): Promise<{ message: string }> {
  return apiPost<{ message: string }>(`/exam/attempts/${attemptId}/submit`, {});
}

export async function getAttemptResult(attemptId: number): Promise<{ data: ExamAttempt }> {
  return apiGet<{ data: ExamAttempt }>(`/exam/attempts/${attemptId}/result`);
}

export async function recordProctorEvent(attemptId: number, eventType: string, payloadJson?: string): Promise<{ message: string }> {
  return apiPost<{ message: string }>(`/exam/attempts/${attemptId}/events`, { eventType, payloadJson: payloadJson ?? "{}" });
}
