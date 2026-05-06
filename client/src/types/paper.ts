export interface PaperItem {
  id: number;
  paperId: number;
  questionId: number;
  type: "single" | "multiple" | "coding";
  score: number;
  sortNo: number;
  question?: QuestionInline;
}

export interface QuestionInline {
  id: number;
  type: "single" | "multiple" | "coding";
  language: "go" | "cpp" | "java" | "javascript" | "python";
  title: string;
  content: string;
  options: string[];
  answers: number[];
  createdAt: string;
}

export interface Paper {
  id: number;
  title: string;
  language: string;
  totalScore: number;
  status: "draft" | "published" | "closed";
  createdBy: number;
  createdAt: string;
  updatedAt: string;
  items?: PaperItem[];
  publication?: Publication;
}

export interface Publication {
  id: number;
  paperId: number;
  startTime: string;
  endTime: string;
  duration: number; // 答题时长(分钟), 0=不限时
  isPublished: boolean;
}

export interface PaperFilters {
  keyword?: string;
  status?: string;
  page?: number;
  pageSize?: number;
  [key: string]: string | number | undefined;
}

export interface GenerateRequest {
  language: string;
  singleCount: number;
  multipleCount: number;
  codingCount: number;
  totalScore: number;
}

export interface GenerateResult {
  items: PaperItem[];
  language: string;
  totalScore: number;
}

export interface SavePaperRequest {
  title: string;
  language: string;
  totalScore: number;
  items: {
    questionId: number;
    type: string;
    score: number;
    sortNo: number;
  }[];
}

export interface PublishRequest {
  startTime: string;
  endTime: string;
  duration: number; // 答题时长(分钟), 0=不限时
  classId?: number;
}
