import { useMemo } from "react";
import type { ReactNode } from "react";
import { Button, Empty, Spin, Typography } from "antd";
import { ArrowLeftOutlined, FileTextOutlined } from "@ant-design/icons";
import { useNavigate, useParams } from "react-router-dom";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeHighlight from "rehype-highlight";
import "github-markdown-css/github-markdown-light.css";
import "highlight.js/styles/github.css";
import { useDocumentCourses, useDocumentDetail } from "../hooks/useDocuments";

const { Title, Text } = Typography;

interface TocItem {
  id: string;
  text: string;
  level: number;
}

function slugifyHeading(text: string) {
  const slug = text
    .trim()
    .toLowerCase()
    .replace(/[^\p{L}\p{N}\s-]/gu, "")
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");

  return slug || "section";
}

function childrenToText(children: ReactNode): string {
  if (typeof children === "string" || typeof children === "number") {
    return String(children);
  }
  if (Array.isArray(children)) {
    return children.map(childrenToText).join("");
  }
  if (children && typeof children === "object" && "props" in children) {
    return childrenToText((children as { props?: { children?: ReactNode } }).props?.children);
  }
  return "";
}

function makeUniqueSlug(baseSlug: string, usedSlugs: Map<string, number>) {
  const usedCount = usedSlugs.get(baseSlug) ?? 0;
  usedSlugs.set(baseSlug, usedCount + 1);

  if (usedCount === 0) {
    return baseSlug;
  }
  return `${baseSlug}-${usedCount + 1}`;
}

function getFenceMarker(line: string) {
  const leadingSpaces = line.match(/^ */)?.[0].length ?? 0;
  if (leadingSpaces > 3) {
    return null;
  }

  const match = /^(`{3,}|~{3,})/.exec(line.slice(leadingSpaces));
  if (!match) {
    return null;
  }

  return {
    char: match[1][0],
    length: match[1].length,
  };
}

function isIndentedCodeLine(line: string) {
  return line.startsWith("\t") || /^ {4,}/.test(line);
}

function buildHeadingList(markdown: string): TocItem[] {
  const usedSlugs = new Map<string, number>();
  let fence: { char: string; length: number } | null = null;

  return markdown
    .split("\n")
    .map((line) => {
      const marker = getFenceMarker(line);
      if (marker) {
        if (!fence) {
          fence = marker;
          return null;
        }
        if (marker.char === fence.char && marker.length >= fence.length) {
          fence = null;
        }
        return null;
      }
      if (fence) {
        return null;
      }
      if (isIndentedCodeLine(line)) {
        return null;
      }

      const match = /^(#{1,3})\s+(.+)$/.exec(line.trim());
      if (!match) {
        return null;
      }
      const text = match[2].replace(/#+$/, "").trim();
      const baseSlug = slugifyHeading(text);
      return {
        id: makeUniqueSlug(baseSlug, usedSlugs),
        text,
        level: match[1].length,
      };
    })
    .filter((item): item is TocItem => Boolean(item));
}

export function DocumentReaderPage() {
  const { courseId, docId } = useParams();
  const navigate = useNavigate();
  const { courses, loading: coursesLoading } = useDocumentCourses();
  const { detail, loading: detailLoading } = useDocumentDetail(courseId, docId);

  const currentCourse = courses.find((course) => course.id === courseId);
  const headingList = useMemo(() => buildHeadingList(detail?.markdown ?? ""), [detail?.markdown]);
  const headingIdsByText = useMemo(() => {
    const idsByText = new Map<string, string[]>();
    headingList.forEach((heading) => {
      const normalizedText = heading.text.trim();
      const ids = idsByText.get(normalizedText) ?? [];
      ids.push(heading.id);
      idsByText.set(normalizedText, ids);
    });
    return idsByText;
  }, [headingList]);
  const toc = headingList;
  const loading = coursesLoading || detailLoading;
  const renderedHeadingCount = new Map<string, number>();

  const renderHeading = (level: 1 | 2 | 3, children: ReactNode) => {
    const text = childrenToText(children);
    const normalizedText = text.trim();
    const usedCount = renderedHeadingCount.get(normalizedText) ?? 0;
    renderedHeadingCount.set(normalizedText, usedCount + 1);
    const id = headingIdsByText.get(normalizedText)?.[usedCount] ?? slugifyHeading(normalizedText);
    const Tag = `h${level}` as "h1" | "h2" | "h3";

    return <Tag id={id}>{children}</Tag>;
  };

  return (
    <main className="document-reader">
      <header className="document-reader__bar">
        <div>
          <Text type="secondary">{currentCourse?.title ?? "课程文档"}</Text>
          <Title level={3}>{detail?.title ?? "文档阅读"}</Title>
        </div>
        <div className="document-reader__actions">
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/home")}>
            返回首页
          </Button>
          <Button
            type="primary"
            icon={<FileTextOutlined />}
            disabled={!currentCourse?.documents.length}
            onClick={() => {
              const firstDoc = currentCourse?.documents[0];
              if (currentCourse && firstDoc) {
                navigate(`/home/courses/${currentCourse.id}/docs/${firstDoc.id}`);
              }
            }}
          >
            切换课件
          </Button>
        </div>
      </header>

      <div className="document-reader__layout">
        <aside className="document-reader__sidebar">
          <section>
            <Title level={5}>课程文档</Title>
            <div className="reader-doc-list">
              {currentCourse?.documents.map((doc) => (
                <button
                  className={doc.id === docId ? "reader-doc-link reader-doc-link--active" : "reader-doc-link"}
                  key={doc.id}
                  type="button"
                  onClick={() => navigate(`/home/courses/${currentCourse.id}/docs/${doc.id}`)}
                >
                  {doc.title}
                </button>
              )) ?? <Text type="secondary">暂无课程文档</Text>}
            </div>
          </section>

          <section>
            <Title level={5}>标题目录</Title>
            {toc.length === 0 ? (
              <Text type="secondary">暂无目录</Text>
            ) : (
              <nav className="reader-toc">
                {toc.map((item) => (
                  <a className={`reader-toc__item reader-toc__item--h${item.level}`} href={`#${item.id}`} key={item.id}>
                    {item.text}
                  </a>
                ))}
              </nav>
            )}
          </section>
        </aside>

        <article className="document-reader__content">
          {loading ? (
            <div className="document-reader__state">
              <Spin />
            </div>
          ) : !detail ? (
            <div className="document-reader__state">
              <Empty description="文档不存在或暂不可访问" />
            </div>
          ) : (
            <ReactMarkdown
              className="markdown-body"
              remarkPlugins={[remarkGfm]}
              rehypePlugins={[rehypeHighlight]}
              components={{
                h1: ({ children }) => renderHeading(1, children),
                h2: ({ children }) => renderHeading(2, children),
                h3: ({ children }) => renderHeading(3, children),
              }}
            >
              {detail.markdown}
            </ReactMarkdown>
          )}
        </article>
      </div>
    </main>
  );
}
