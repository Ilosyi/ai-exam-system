export interface Question {
  id: number;
  type: "single" | "multiple" | "coding";
  language: "go" | "cpp" | "java" | "javascript" | "python";
  title: string;
  content: string;
  options: string[];
  answers: number[];
  creatorName: string;
  createdAt: string;
}

export interface QuestionFilters {
  keyword?: string;
  type?: Question["type"] | "";
  language?: Question["language"] | "";
  page?: number;
  pageSize?: number;
  [key: string]: string | number | undefined;
}
