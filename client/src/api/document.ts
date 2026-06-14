import { apiDelete, apiGet, apiPost, apiPut } from "./client";

export interface DocumentMeta {
  id: string;
  title: string;
  order: number;
}

export interface CourseDocument {
  id: string;
  title: string;
  description: string;
  order: number;
  documents: DocumentMeta[];
}

export interface DocumentDetail extends DocumentMeta {
  markdown: string;
}

export interface CourseInput {
  id: string;
  title: string;
  description: string;
  order: number;
}

export interface DocumentInput {
  id: string;
  title: string;
  order: number;
  markdown: string;
}

export function fetchDocumentCourses(): Promise<{ data: CourseDocument[]; total: number }> {
  return apiGet<{ data: CourseDocument[]; total: number }>("/documents/courses");
}

export function fetchDocumentCourse(courseId: string): Promise<{ data: CourseDocument }> {
  return apiGet<{ data: CourseDocument }>(`/documents/courses/${courseId}`);
}

export function createDocumentCourse(input: CourseInput): Promise<{ data: CourseDocument }> {
  return apiPost<{ data: CourseDocument }>("/documents/courses", input);
}

export function updateDocumentCourse(courseId: string, input: CourseInput): Promise<{ data: CourseDocument }> {
  return apiPut<{ data: CourseDocument }>(`/documents/courses/${courseId}`, input);
}

export function deleteDocumentCourse(courseId: string): Promise<void> {
  return apiDelete<void>(`/documents/courses/${courseId}`);
}

export function fetchDocumentDetail(courseId: string, docId: string): Promise<{ data: DocumentDetail }> {
  return apiGet<{ data: DocumentDetail }>(`/documents/courses/${courseId}/docs/${docId}`);
}

export function createDocument(courseId: string, input: DocumentInput): Promise<{ data: DocumentDetail }> {
  return apiPost<{ data: DocumentDetail }>(`/documents/courses/${courseId}/docs`, input);
}

export function updateDocument(courseId: string, docId: string, input: DocumentInput): Promise<{ data: DocumentDetail }> {
  return apiPut<{ data: DocumentDetail }>(`/documents/courses/${courseId}/docs/${docId}`, input);
}

export function deleteDocument(courseId: string, docId: string): Promise<void> {
  return apiDelete<void>(`/documents/courses/${courseId}/docs/${docId}`);
}
