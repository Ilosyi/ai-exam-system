import { useEffect, useState } from "react";
import { Card, List, Button, Tag, Typography, Empty, Spin } from "antd";
import { EnterOutlined, ClockCircleOutlined } from "@ant-design/icons";
import { fetchPublishedPapers, startAttempt } from "../api/exam";
import type { PublishedPaper } from "../api/exam";
import { useNavigate } from "react-router-dom";
import dayjs from "dayjs";
import { useAuth } from "../hooks/useAuth";

const { Title } = Typography;

function getExamStatus(paper: PublishedPaper) {
  const now = dayjs();
  const start = dayjs(paper.startTime);
  const end = dayjs(paper.endTime);
  const isActive = (now.isAfter(start) || now.isSame(start)) && (now.isBefore(end) || now.isSame(end));

  if ((paper.attemptStatus === "submitted" || paper.attemptStatus === "timeout") && paper.attemptId) {
    return {
      tagColor: paper.attemptStatus === "timeout" ? "orange" : "green",
      tagText: paper.attemptStatus === "timeout" ? "超时提交" : "已交卷",
      actionText: "查看详情",
      disabled: false,
      target: `/exam/${paper.attemptId}/result`,
    };
  }

  if (paper.attemptStatus === "in_progress" && paper.attemptId) {
    return {
      tagColor: isActive ? "blue" : "orange",
      tagText: isActive ? "答题中" : "已开始未交卷",
      actionText: isActive ? "继续答题" : "已结束",
      disabled: !isActive,
      target: `/exam/${paper.attemptId}/take`,
    };
  }

  if (isActive) {
    return {
      tagColor: "blue",
      tagText: "进行中",
      actionText: "进入答题",
      disabled: false,
      target: "",
    };
  }

  return {
    tagColor: now.isBefore(start) ? "default" : "red",
    tagText: now.isBefore(start) ? "未开始" : "未参加",
    actionText: now.isBefore(start) ? "未到考试时间" : "已结束",
    disabled: true,
    target: "",
  };
}

export function ExamListPage() {
  const [papers, setPapers] = useState<PublishedPaper[]>([]);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { user, logout } = useAuth();

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const res = await fetchPublishedPapers();
        setPapers(res.data || []);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  const handleStart = async (paperId: number) => {
    try {
      const res = await startAttempt(paperId);
      navigate(`/exam/${res.data.id}/take`);
    } catch (err: unknown) {
      // Error will be shown by the API response
      alert(err instanceof Error ? err.message : "开始答题失败");
    }
  };

  return (
    <div className="exam-shell">
      <div className="exam-shell__inner">
        <div className="exam-topbar app-fade-up">
          <div>
            <div className="page-hero__eyebrow" style={{ background: "rgba(39,84,197,0.08)", color: "var(--app-brand-strong)" }}>
              Student Entrance
            </div>
            <Title level={2} style={{ margin: "12px 0 0", color: "var(--app-brand-strong)", fontFamily: "\"Baskerville\", \"Songti SC\", serif" }}>
              在线考试
            </Title>
            <Typography.Paragraph type="secondary" style={{ margin: "8px 0 0" }}>
              按开放时间窗进入考试，系统会自动处理保存、倒计时和交卷流程。
            </Typography.Paragraph>
          </div>
          <div className="panel-surface" style={{ padding: "12px 14px", borderRadius: 20 }}>
            <Tag color="blue" style={{ marginRight: 8 }}>{user?.username}</Tag>
            <Button
              onClick={() => {
                logout();
                navigate("/login", { replace: true });
              }}
            >
              退出登录
            </Button>
          </div>
        </div>

        {loading ? (
          <div className="panel-surface app-fade-up" style={{ textAlign: "center", padding: 80, borderRadius: 28 }}><Spin size="large" /></div>
        ) : papers.length === 0 ? (
          <Card className="section-card app-fade-up" style={{ borderRadius: 28 }}>
            <Empty description="暂无已发布考试" />
          </Card>
        ) : (
          <List
            grid={{ gutter: 20, column: 1 }}
            dataSource={papers}
            renderItem={(paper) => {
              const start = dayjs(paper.startTime);
              const end = dayjs(paper.endTime);
              const status = getExamStatus(paper);

              return (
                <List.Item className="app-fade-up">
                  <Card className="glass-card" hoverable style={{ borderRadius: 24 }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 20, flexWrap: "wrap" }}>
                      <div>
                        <Title level={4} style={{ margin: 0, color: "var(--app-brand-strong)" }}>{paper.title}</Title>
                        <div style={{ marginTop: 10, color: "#566579" }}>
                          <Tag color="blue">{paper.language}</Tag>
                          <Tag color="gold">总分 {paper.totalScore}</Tag>
                          <Tag color={status.tagColor}>{status.tagText}</Tag>
                          {paper.attemptScore != null && <Tag color="green">得分 {paper.attemptScore}</Tag>}
                        </div>
                        <div style={{ marginTop: 12, color: "#738196", fontSize: 13 }}>
                          <ClockCircleOutlined style={{ marginRight: 4 }} />
                          {start.format("YYYY-MM-DD HH:mm")} ~ {end.format("YYYY-MM-DD HH:mm")}
                          {paper.duration > 0 && <span style={{ marginLeft: 8 }}>答题限时: {paper.duration} 分钟</span>}
                        </div>
                      </div>
                      <Button
                        type="primary"
                        icon={<EnterOutlined />}
                        disabled={status.disabled}
                        onClick={() => {
                          if (status.target) {
                            navigate(status.target);
                            return;
                          }
                          void handleStart(paper.paperId);
                        }}
                      >
                        {status.actionText}
                      </Button>
                    </div>
                  </Card>
                </List.Item>
              );
            }}
          />
        )}
      </div>
    </div>
  );
}
