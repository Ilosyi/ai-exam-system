import { useCallback, useEffect, useState } from "react";
import { message } from "antd";
import {
  fetchDocumentCourses,
  fetchDocumentDetail,
  type CourseDocument,
  type DocumentDetail,
} from "../api/document";

export function useDocumentCourses() {
  const [courses, setCourses] = useState<CourseDocument[]>([]);
  const [loading, setLoading] = useState(false);

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetchDocumentCourses();
      setCourses(res.data);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "文档课程加载失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void reload();
  }, [reload]);

  return {
    courses,
    loading,
    reload,
    setCourses,
  };
}

export function useDocumentDetail(courseId?: string, docId?: string) {
  const [detail, setDetail] = useState<DocumentDetail | null>(null);
  const [loading, setLoading] = useState(false);

  const reload = useCallback(async () => {
    if (!courseId || !docId) {
      setDetail(null);
      return;
    }

    setLoading(true);
    try {
      const res = await fetchDocumentDetail(courseId, docId);
      setDetail(res.data);
    } catch (error) {
      setDetail(null);
      message.error(error instanceof Error ? error.message : "文档详情加载失败");
    } finally {
      setLoading(false);
    }
  }, [courseId, docId]);

  useEffect(() => {
    void reload();
  }, [reload]);

  return {
    detail,
    loading,
    reload,
  };
}
