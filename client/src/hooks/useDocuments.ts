import { useCallback, useEffect, useRef, useState } from "react";
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
  const mountedRef = useRef(false);
  const requestIdRef = useRef(0);

  const reload = useCallback(async () => {
    const requestId = requestIdRef.current + 1;
    requestIdRef.current = requestId;
    if (mountedRef.current) {
      setLoading(true);
    }
    try {
      const res = await fetchDocumentCourses();
      if (mountedRef.current && requestIdRef.current === requestId) {
        setCourses(res.data);
      }
    } catch (error) {
      if (mountedRef.current && requestIdRef.current === requestId) {
        message.error(error instanceof Error ? error.message : "文档课程加载失败");
      }
    } finally {
      if (mountedRef.current && requestIdRef.current === requestId) {
        setLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    void reload();

    return () => {
      mountedRef.current = false;
      requestIdRef.current += 1;
    };
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
  const mountedRef = useRef(false);
  const requestIdRef = useRef(0);

  const reload = useCallback(async () => {
    const requestId = requestIdRef.current + 1;
    requestIdRef.current = requestId;

    if (!courseId || !docId) {
      if (mountedRef.current) {
        setDetail(null);
        setLoading(false);
      }
      return;
    }

    if (mountedRef.current) {
      setLoading(true);
    }
    try {
      const res = await fetchDocumentDetail(courseId, docId);
      if (mountedRef.current && requestIdRef.current === requestId) {
        setDetail(res.data);
      }
    } catch (error) {
      if (mountedRef.current && requestIdRef.current === requestId) {
        setDetail(null);
        message.error(error instanceof Error ? error.message : "文档详情加载失败");
      }
    } finally {
      if (mountedRef.current && requestIdRef.current === requestId) {
        setLoading(false);
      }
    }
  }, [courseId, docId]);

  useEffect(() => {
    mountedRef.current = true;
    void reload();

    return () => {
      mountedRef.current = false;
      requestIdRef.current += 1;
    };
  }, [reload]);

  return {
    detail,
    loading,
    reload,
  };
}
