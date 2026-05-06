import { useCallback, useEffect, useRef, useState } from "react";
import { Button, Card, Radio, Checkbox, Tag, Space, Modal, Spin, Typography } from "antd";
import { ArrowLeftOutlined, ArrowRightOutlined, SendOutlined, ClockCircleOutlined } from "@ant-design/icons";
import { useParams, useNavigate } from "react-router-dom";
import { getAttempt, saveAnswers, submitAttempt, recordProctorEvent } from "../api/exam";
import type { ExamAttempt } from "../api/exam";
import type { PaperItem } from "../types/paper";
import dayjs from "dayjs";

const { Title, Text } = Typography;

const typeMap: Record<string, { color: string; text: string }> = {
  single: { color: "blue", text: "单选" },
  multiple: { color: "orange", text: "多选" },
  coding: { color: "green", text: "编程" },
};

function formatCountdown(ms: number): string {
  if (ms <= 0) return "00:00:00";
  const totalSec = Math.floor(ms / 1000);
  const h = Math.floor(totalSec / 3600);
  const m = Math.floor((totalSec % 3600) / 60);
  const s = totalSec % 60;
  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

export function ExamTakePage() {
  const { id } = useParams<{ id: string }>();
  const attemptId = Number(id);
  const navigate = useNavigate();
  const [attempt, setAttempt] = useState<ExamAttempt | null>(null);
  const [loading, setLoading] = useState(true);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [answers, setAnswers] = useState<Record<number, number[]>>({});
  const [submitting, setSubmitting] = useState(false);
  const [remainingMs, setRemainingMs] = useState<number | null>(null);
  const autoSaveTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  const countdownTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  const submittedRef = useRef(false);

  const loadAttempt = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getAttempt(attemptId);
      setAttempt(res.data);
      // Restore saved answers
      if (res.data.answers) {
        const saved: Record<number, number[]> = {};
        for (const a of res.data.answers) {
          try {
            saved[a.questionId] = JSON.parse(a.answerJson || "[]");
          } catch {
            saved[a.questionId] = [];
          }
        }
        setAnswers(saved);
      }
    } catch {
      navigate("/exam");
    } finally {
      setLoading(false);
    }
  }, [attemptId, navigate]);

  useEffect(() => {
    loadAttempt();
  }, [loadAttempt]);

  // Countdown timer
  useEffect(() => {
    if (!attempt?.deadline || attempt.status !== "in_progress") return;

    const deadlineMs = dayjs(attempt.deadline).valueOf();

    const tick = () => {
      const diff = deadlineMs - Date.now();
      setRemainingMs(diff);
      if (diff <= 0 && !submittedRef.current) {
        submittedRef.current = true;
        // Auto-submit
        Modal.warning({
          title: "答题时间已到",
          content: "系统将自动为您交卷",
          centered: true,
          onOk: async () => {
            const entries = Object.entries(answers).map(([qId, ans]) => ({
              questionId: Number(qId),
              answerJson: JSON.stringify(ans),
            }));
            if (entries.length > 0) {
              await saveAnswers(attemptId, entries).catch(() => {});
            }
            await submitAttempt(attemptId).catch(() => {});
            navigate(`/exam/${attemptId}/result`);
          },
        });
        if (countdownTimer.current) clearInterval(countdownTimer.current);
      }
    };

    tick(); // initial tick
    countdownTimer.current = setInterval(tick, 1000);
    return () => {
      if (countdownTimer.current) clearInterval(countdownTimer.current);
    };
  }, [attempt?.deadline, attempt?.status, attemptId, answers, navigate]);

  // Auto-save every 30 seconds
  useEffect(() => {
    autoSaveTimer.current = setInterval(() => {
      const entries = Object.entries(answers).map(([qId, ans]) => ({
        questionId: Number(qId),
        answerJson: JSON.stringify(ans),
      }));
      if (entries.length > 0) {
        saveAnswers(attemptId, entries).catch(() => {});
      }
    }, 30000);
    return () => {
      if (autoSaveTimer.current) clearInterval(autoSaveTimer.current);
    };
  }, [attemptId, answers]);

  // Proctor events
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden) {
        recordProctorEvent(attemptId, "visibilitychange", JSON.stringify({ hidden: true }));
      }
    };
    const handleBlur = () => {
      recordProctorEvent(attemptId, "blur", JSON.stringify({ timestamp: Date.now() }));
    };
    document.addEventListener("visibilitychange", handleVisibilityChange);
    window.addEventListener("blur", handleBlur);
    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      window.removeEventListener("blur", handleBlur);
    };
  }, [attemptId]);

  if (loading || !attempt?.paper?.items) {
    return <div style={{ textAlign: "center", padding: 100 }}><Spin size="large" /></div>;
  }

  const items = attempt.paper.items;
  const currentItem = items[currentIndex];
  const question = currentItem?.question;

  // Group items by type for navigator
  const grouped = items.reduce<Record<string, PaperItem[]>>((acc, item) => {
    if (!acc[item.type]) acc[item.type] = [];
    acc[item.type].push(item);
    return acc;
  }, {});

  const handleAnswer = (questionId: number, answer: number[]) => {
    setAnswers((prev) => ({ ...prev, [questionId]: answer }));
  };

  const handleSubmit = async () => {
    const answeredCount = Object.keys(answers).filter((k) => answers[Number(k)].length > 0).length;
    const unansweredCount = items.length - answeredCount;

    Modal.confirm({
      title: "确认交卷",
      content: (
        <div>
          <div style={{ marginBottom: 12 }}>
            {items.map((item, index) => {
              const ans = answers[item.questionId];
              const isAnswered = ans && ans.length > 0;
              return (
                <span
                  key={item.questionId}
                  style={{
                    display: "inline-block",
                    width: 32,
                    height: 32,
                    lineHeight: "32px",
                    textAlign: "center",
                    margin: 2,
                    borderRadius: 4,
                    background: isAnswered ? "#1890ff" : "#d9d9d9",
                    color: isAnswered ? "#fff" : "#666",
                    fontSize: 13,
                  }}
                >
                  {index + 1}
                </span>
              );
            })}
          </div>
          {unansweredCount > 0 && (
            <div style={{ color: "#faad14" }}>还有 {unansweredCount} 题未作答，确定要交卷吗？</div>
          )}
        </div>
      ),
      centered: true,
      okText: "确认交卷",
      cancelText: "继续答题",
      onOk: async () => {
        submittedRef.current = true;
        setSubmitting(true);
        const entries = Object.entries(answers).map(([qId, ans]) => ({
          questionId: Number(qId),
          answerJson: JSON.stringify(ans),
        }));
        if (entries.length > 0) {
          await saveAnswers(attemptId, entries);
        }
        await submitAttempt(attemptId);
        navigate(`/exam/${attemptId}/result`);
      },
    });
  };

  // Countdown display style
  const isUrgent = remainingMs !== null && remainingMs > 0 && remainingMs < 5 * 60 * 1000; // < 5 min
  const isExpired = remainingMs !== null && remainingMs <= 0;

  return (
    <div style={{ minHeight: "100vh", background: "#f0f2f5" }}>
      {/* Header */}
      <div style={{ background: "#fff", padding: "12px 24px", display: "flex", justifyContent: "space-between", alignItems: "center", boxShadow: "0 2px 8px rgba(0,0,0,0.06)" }}>
        <Title level={4} style={{ margin: 0 }}>{attempt.paper.title}</Title>
        <Space size="large">
          {remainingMs !== null && !isExpired && (
            <span style={{
              fontSize: 18,
              fontWeight: 700,
              fontFamily: "monospace",
              color: isUrgent ? "#ff4d4f" : "#52c41a",
              background: isUrgent ? "#fff1f0" : "#f6ffed",
              padding: "4px 16px",
              borderRadius: 6,
              border: `1px solid ${isUrgent ? "#ffa39e" : "#b7eb8f"}`,
            }}>
              <ClockCircleOutlined style={{ marginRight: 6 }} />
              {formatCountdown(remainingMs)}
            </span>
          )}
          <Button type="primary" danger icon={<SendOutlined />} onClick={handleSubmit} loading={submitting}>
            交卷
          </Button>
        </Space>
      </div>

      <div style={{ display: "flex", height: "calc(100vh - 60px)" }}>
        {/* Left: Question navigator */}
        <div style={{ width: 200, background: "#fff", padding: 16, overflow: "auto", borderRight: "1px solid #f0f0f0" }}>
          {Object.entries(grouped).map(([type, typeItems]) => {
            const typeInfo = typeMap[type] || { color: "default", text: type };
            return (
              <div key={type} style={{ marginBottom: 16 }}>
                <Tag color={typeInfo.color} style={{ marginBottom: 8 }}>{typeInfo.text}</Tag>
                <div style={{ display: "flex", flexWrap: "wrap", gap: 4 }}>
                  {typeItems.map((item) => {
                    const globalIndex = items.findIndex((i) => i.questionId === item.questionId);
                    const ans = answers[item.questionId];
                    const isAnswered = ans && ans.length > 0;
                    const isCurrent = globalIndex === currentIndex;
                    return (
                      <span
                        key={item.questionId}
                        onClick={() => setCurrentIndex(globalIndex)}
                        style={{
                          display: "inline-block",
                          width: 32,
                          height: 32,
                          lineHeight: "32px",
                          textAlign: "center",
                          borderRadius: 4,
                          cursor: "pointer",
                          background: isCurrent ? "#1890ff" : isAnswered ? "#e6f7ff" : "#f5f5f5",
                          color: isCurrent ? "#fff" : isAnswered ? "#1890ff" : "#999",
                          border: isCurrent ? "2px solid #1890ff" : "1px solid #d9d9d9",
                          fontWeight: isCurrent ? 700 : 400,
                        }}
                      >
                        {globalIndex + 1}
                      </span>
                    );
                  })}
                </div>
              </div>
            );
          })}
        </div>

        {/* Right: Question area */}
        <div style={{ flex: 1, padding: 24, overflow: "auto" }}>
          {question ? (
            <Card>
              <div style={{ marginBottom: 16 }}>
                <Space>
                  <span style={{ fontSize: 16, fontWeight: 600 }}>第 {currentIndex + 1} 题</span>
                  <Tag color={typeMap[currentItem.type]?.color}>{typeMap[currentItem.type]?.text}</Tag>
                  <Text type="secondary">{currentItem.score} 分</Text>
                </Space>
              </div>
              <div style={{ fontWeight: 500, marginBottom: 12, fontSize: 15 }}>{question.title}</div>
              <div style={{ color: "#666", whiteSpace: "pre-wrap", marginBottom: 20, fontSize: 14 }}>{question.content}</div>

              {currentItem.type === "single" && question.options && (
                <Radio.Group
                  value={answers[currentItem.questionId]?.[0]}
                  onChange={(e) => handleAnswer(currentItem.questionId, [e.target.value])}
                >
                  <Space direction="vertical">
                    {question.options.map((opt, i) => (
                      <Radio key={i} value={i}>{String.fromCharCode(65 + i)}. {opt}</Radio>
                    ))}
                  </Space>
                </Radio.Group>
              )}

              {currentItem.type === "multiple" && question.options && (
                <Checkbox.Group
                  value={answers[currentItem.questionId] || []}
                  onChange={(vals) => handleAnswer(currentItem.questionId, vals as number[])}
                >
                  <Space direction="vertical">
                    {question.options.map((opt, i) => (
                      <Checkbox key={i} value={i}>{String.fromCharCode(65 + i)}. {opt}</Checkbox>
                    ))}
                  </Space>
                </Checkbox.Group>
              )}

              {currentItem.type === "coding" && (
                <div style={{ color: "#999" }}>编程题请在下方编写代码（功能开发中）</div>
              )}
            </Card>
          ) : (
            <Card><div style={{ textAlign: "center", color: "#999" }}>题目加载中...</div></Card>
          )}

          {/* Navigation buttons */}
          <div style={{ marginTop: 16, display: "flex", justifyContent: "space-between" }}>
            <Button
              icon={<ArrowLeftOutlined />}
              disabled={currentIndex === 0}
              onClick={() => setCurrentIndex((prev) => prev - 1)}
            >
              上一题
            </Button>
            <Button
              icon={<ArrowRightOutlined />}
              disabled={currentIndex === items.length - 1}
              onClick={() => setCurrentIndex((prev) => prev + 1)}
            >
              下一题
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
