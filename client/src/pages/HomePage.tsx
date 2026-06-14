import { useEffect, useState } from "react";
import { Button, Empty, Spin, Tag, Typography, message } from "antd";
import { CalendarOutlined, CheckCircleOutlined, ClockCircleOutlined, LogoutOutlined, ReadOutlined, UserOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import dayjs from "dayjs";
import { fetchPublishedPapers, startAttempt, type PublishedPaper } from "../api/exam";
import { useAuth } from "../hooks/useAuth";
import { useDocumentCourses } from "../hooks/useDocuments";

const { Title, Text, Paragraph } = Typography;

type ExamCardState = {
  tagColor: string;
  tagText: string;
  actionText: string;
  disabled: boolean;
  onAction: () => void;
};

function roleLabel(role?: string) {
  if (role === "admin") {
    return "管理员";
  }
  if (role === "student") {
    return "学生";
  }
  return "教师";
}

export function HomePage() {
  const [papers, setPapers] = useState<PublishedPaper[]>([]);
  const [examLoading, setExamLoading] = useState(false);
  const { courses, loading: coursesLoading } = useDocumentCourses();
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const loadPapers = async () => {
      setExamLoading(true);
      try {
        const res = await fetchPublishedPapers();
        setPapers(res.data ?? []);
      } catch (error) {
        message.error(error instanceof Error ? error.message : "考试列表加载失败");
      } finally {
        setExamLoading(false);
      }
    };

    void loadPapers();
  }, []);

  const handleStart = async (paperId: number) => {
    try {
      const res = await startAttempt(paperId);
      navigate(`/exam/${res.data.id}/take`);
    } catch (error) {
      message.error(error instanceof Error ? error.message : "开始答题失败");
    }
  };

  const handleLogout = () => {
    logout();
    navigate("/login", { replace: true });
  };

  const getExamCardState = (paper: PublishedPaper): ExamCardState => {
    const now = dayjs();
    const start = dayjs(paper.startTime);
    const end = dayjs(paper.endTime);
    const isActive = (now.isAfter(start) || now.isSame(start)) && (now.isBefore(end) || now.isSame(end));
    const isFuture = now.isBefore(start);

    if ((paper.attemptStatus === "submitted" || paper.attemptStatus === "timeout") && paper.attemptId) {
      return {
        tagColor: paper.attemptStatus === "timeout" ? "orange" : "green",
        tagText: paper.attemptStatus === "timeout" ? "超时提交" : "已交卷",
        actionText: "查看详情",
        disabled: false,
        onAction: () => navigate(`/exam/${paper.attemptId}/result`),
      };
    }

    if (paper.attemptStatus === "in_progress" && paper.attemptId) {
      return {
        tagColor: isActive ? "blue" : "orange",
        tagText: isActive ? "答题中" : "答题结束",
        actionText: isActive ? "继续答题" : "已结束",
        disabled: !isActive,
        onAction: () => navigate(`/exam/${paper.attemptId}/take`),
      };
    }

    if (isActive) {
      return {
        tagColor: "blue",
        tagText: "进行中",
        actionText: "开始考试",
        disabled: false,
        onAction: () => handleStart(paper.paperId),
      };
    }

    return {
      tagColor: isFuture ? "default" : "red",
      tagText: isFuture ? "未开始" : "未参加",
      actionText: isFuture ? "未到考试时间" : "已结束",
      disabled: true,
      onAction: () => undefined,
    };
  };

  return (
    <main className="student-home">
      <header className="student-home__top">
        <Title level={3} className="student-home__title">
          个人中心
        </Title>
        <Button icon={<LogoutOutlined />} onClick={handleLogout}>
          退出登录
        </Button>
      </header>

      <section className="student-profile-card">
        <div className="student-profile-card__avatar" aria-hidden="true">
          <UserOutlined />
        </div>
        <div className="student-profile-card__main">
          <Text className="student-profile-card__label">Welcome back</Text>
          <Title level={2} className="student-profile-card__name">
            {user?.username ?? "同学"}
          </Title>
          <div className="student-profile-card__meta">
            <span>ID: {user?.id ?? "-"}</span>
            <span>角色: {roleLabel(user?.role)}</span>
            <span>状态: {user?.status ?? "-"}</span>
            <span>班级: {user?.classId ?? "未分配"}</span>
          </div>
        </div>
        <div className="student-profile-card__plane" aria-hidden="true" />
      </section>

      <section className="student-section">
        <div className="student-section__head">
          <Title level={4}>课程列表</Title>
          <Text type="secondary">选择课程文档，继续阅读本周内容</Text>
        </div>

        {coursesLoading ? (
          <div className="student-home__loading">
            <Spin />
          </div>
        ) : courses.length === 0 ? (
          <div className="student-home__empty">
            <Empty description="暂无课程文档" />
          </div>
        ) : (
          <div className="course-card-grid">
            {courses.map((course) => (
              <article className="course-card" key={course.id}>
                <div className="course-card__icon">
                  <ReadOutlined />
                </div>
                <Title level={5}>{course.title}</Title>
                <Paragraph ellipsis={{ rows: 2 }} className="course-card__desc">
                  {course.description || "暂无课程说明"}
                </Paragraph>
                <div className="course-card__docs">
                  {course.documents.length === 0 ? (
                    <Text type="secondary">暂无课件</Text>
                  ) : (
                    course.documents.map((doc, index) => (
                      <button
                        className="course-doc-link"
                        key={doc.id}
                        type="button"
                        onClick={() => navigate(`/home/courses/${course.id}/docs/${doc.id}`)}
                      >
                        <span className="course-doc-link__badge">{index + 1}</span>
                        <span>{doc.title}</span>
                      </button>
                    ))
                  )}
                </div>
              </article>
            ))}
          </div>
        )}
      </section>

      <section className="student-section">
        <div className="student-section__head">
          <Title level={4}>我的考试</Title>
          <Text type="secondary">查看已发布考试、历史答题记录和成绩详情</Text>
        </div>

        {examLoading ? (
          <div className="student-home__loading">
            <Spin />
          </div>
        ) : papers.length === 0 ? (
          <div className="student-home__empty">
            <Empty description="暂无已发布考试" />
          </div>
        ) : (
          <div className="student-exam-grid">
            {papers.map((paper) => {
              const start = dayjs(paper.startTime);
              const end = dayjs(paper.endTime);
              const cardState = getExamCardState(paper);

              return (
                <article className="student-exam-card" key={paper.paperId}>
                  <div className="student-exam-card__top">
                    <CalendarOutlined />
                    <Tag color={cardState.tagColor}>{cardState.tagText}</Tag>
                  </div>
                  <Title level={5}>{paper.title}</Title>
                  <div className="student-exam-card__meta">
                    <span>总分 {paper.totalScore}</span>
                    <span>{paper.duration > 0 ? `${paper.duration} 分钟` : "不限时"}</span>
                  </div>
                  <div className="student-exam-card__score">
                    {paper.attemptStatus === "submitted" || paper.attemptStatus === "timeout" ? (
                      <>
                        <CheckCircleOutlined />
                        <span>得分 {paper.attemptScore ?? "-"}</span>
                        <Text type="secondary">/ {paper.totalScore}</Text>
                      </>
                    ) : paper.attemptStatus === "in_progress" ? (
                      <>
                        <ClockCircleOutlined />
                        <span>已开始</span>
                      </>
                    ) : (
                      <Text type="secondary">暂无答题记录</Text>
                    )}
                  </div>
                  <Text className="student-exam-card__time">
                    {start.format("YYYY-MM-DD HH:mm")} - {end.format("YYYY-MM-DD HH:mm")}
                  </Text>
                  <Button type="primary" disabled={cardState.disabled} block onClick={cardState.onAction}>
                    {cardState.actionText}
                  </Button>
                </article>
              );
            })}
          </div>
        )}
      </section>
    </main>
  );
}
