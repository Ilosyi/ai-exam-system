import { useEffect, useState } from "react";
import { Card, Tag, Space, Typography, Spin, Button, Descriptions } from "antd";
import { ArrowLeftOutlined, CheckCircleOutlined, CloseCircleOutlined } from "@ant-design/icons";
import { useParams, useNavigate } from "react-router-dom";
import { getAttemptResult } from "../api/exam";
import type { ExamAttempt } from "../api/exam";
import type { PaperItem } from "../types/paper";
import dayjs from "dayjs";

const { Title, Text } = Typography;

const typeMap: Record<string, { color: string; text: string }> = {
  single: { color: "blue", text: "单选" },
  multiple: { color: "orange", text: "多选" },
  coding: { color: "green", text: "编程" },
};

export function ExamResultPage() {
  const { id } = useParams<{ id: string }>();
  const attemptId = Number(id);
  const navigate = useNavigate();
  const [attempt, setAttempt] = useState<ExamAttempt | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        const res = await getAttemptResult(attemptId);
        setAttempt(res.data);
      } catch {
        navigate("/exam");
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [attemptId, navigate]);

  if (loading) {
    return <div style={{ textAlign: "center", padding: 100 }}><Spin size="large" /></div>;
  }

  if (!attempt?.paper) {
    return <div style={{ textAlign: "center", padding: 100 }}>无法加载答卷信息</div>;
  }

  const items = attempt.paper.items || [];
  const answerMap = new Map((attempt.answers || []).map((a) => [a.questionId, a]));

  return (
    <div style={{ minHeight: "100vh", background: "#f0f2f5", padding: "40px 24px" }}>
      <div style={{ maxWidth: 900, margin: "0 auto" }}>
        <div style={{ marginBottom: 24 }}>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/exam")}>返回考试列表</Button>
        </div>

        {/* Summary card */}
        <Card style={{ marginBottom: 24 }}>
          <Title level={3} style={{ marginTop: 0 }}>{attempt.paper.title}</Title>
          <Descriptions column={3}>
            <Descriptions.Item label="得分">{attempt.totalScore ?? "-"}</Descriptions.Item>
            <Descriptions.Item label="总分">{attempt.paper.totalScore}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={attempt.status === "submitted" ? "green" : "orange"}>
                {attempt.status === "submitted" ? "已交卷" : "超时自动提交"}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="开始时间">{dayjs(attempt.startedAt).format("YYYY-MM-DD HH:mm:ss")}</Descriptions.Item>
            <Descriptions.Item label="提交时间">{attempt.submittedAt ? dayjs(attempt.submittedAt).format("YYYY-MM-DD HH:mm:ss") : "-"}</Descriptions.Item>
          </Descriptions>
        </Card>

        {/* Answer details */}
        <Card title="答题详情">
          <Space direction="vertical" style={{ width: "100%" }} size="middle">
            {items.map((item: PaperItem, index: number) => {
              const answer = answerMap.get(item.questionId);
              const typeInfo = typeMap[item.type] || { color: "default", text: item.type };
              const question = item.question;

              let studentAnswerText = "未作答";
              let correctAnswerText = "-";

              if (question?.options && question.options.length > 0) {
                const correctIdxs = question.answers || [];
                correctAnswerText = correctIdxs.map((i) => String.fromCharCode(65 + i)).join(", ");

                if (answer?.answerJson) {
                  try {
                    const studentIdxs: number[] = JSON.parse(answer.answerJson);
                    studentAnswerText = studentIdxs.map((i) => String.fromCharCode(65 + i)).join(", ");
                  } catch {
                    studentAnswerText = answer.answerJson;
                  }
                }
              }

              return (
                <Card
                  key={item.questionId}
                  size="small"
                  title={
                    <Space>
                      <span>第 {index + 1} 题</span>
                      <Tag color={typeInfo.color}>{typeInfo.text}</Tag>
                      <span style={{ color: "#999" }}>{item.score} 分</span>
                      {answer?.isCorrect === true && <CheckCircleOutlined style={{ color: "#52c41a" }} />}
                      {answer?.isCorrect === false && <CloseCircleOutlined style={{ color: "#ff4d4f" }} />}
                      {answer?.score != null && <Text type="secondary">得分: {answer.score}</Text>}
                    </Space>
                  }
                >
                  {question ? (
                    <div>
                      <div style={{ fontWeight: 500, marginBottom: 8 }}>{question.title}</div>
                      {question.options && question.options.length > 0 && (
                        <div style={{ marginBottom: 12 }}>
                          {question.options.map((opt, i) => (
                            <div key={i} style={{ padding: "2px 0" }}>
                              {String.fromCharCode(65 + i)}. {opt}
                            </div>
                          ))}
                        </div>
                      )}
                      <div style={{ display: "flex", gap: 24, marginTop: 12 }}>
                        <div>
                          <Text type="secondary">你的答案: </Text>
                          <Text style={{ color: answer?.isCorrect ? "#52c41a" : "#ff4d4f" }}>{studentAnswerText}</Text>
                        </div>
                        <div>
                          <Text type="secondary">正确答案: </Text>
                          <Text style={{ color: "#52c41a" }}>{correctAnswerText}</Text>
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div style={{ color: "#999" }}>题目 ID: {item.questionId}</div>
                  )}
                </Card>
              );
            })}
          </Space>
        </Card>
      </div>
    </div>
  );
}
