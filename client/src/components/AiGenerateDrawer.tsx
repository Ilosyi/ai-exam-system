import { useEffect, useMemo, useState } from "react";
import { Alert, Button, Card, Divider, Drawer, Form, Input, Select, Space, Statistic, Typography, message } from "antd";
import type { Question } from "../types/question";
import { generateQuestions, testAiConnection } from "../api/ai";
import aiLogo from "../assets/AIlogo.png";

/**
 * AiGenerateDrawer
 * -----------------
 * 负责展示 AI 出题的侧滑面板：
 * - 左侧为构建 AI 请求的表单（题型、数量、语言、关键词）
 * - 右侧为 AI 返回的试题预览区域
 * - 提供单条插入与一键插入全部的能力
 *
 * 说明：该组件不直接处理数据持久化，而是通过 onInsert 回调将生成的题目交给上层进行插入。
 */

interface Props {
  onInsert: (question: Partial<Question>) => Promise<void>;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function AiGenerateDrawer({ onInsert, open: controlledOpen, onOpenChange }: Props) {
  const [uncontrolledOpen, setUncontrolledOpen] = useState(false);
  const open = controlledOpen ?? uncontrolledOpen;
  const setOpen = useMemo(() => onOpenChange ?? setUncontrolledOpen, [onOpenChange]);
  const [loading, setLoading] = useState(false);
  const [inserting, setInserting] = useState(false);
  const [result, setResult] = useState<Partial<Question>[]>([]);
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const [testing, setTesting] = useState(false);
  const [diagnostic, setDiagnostic] = useState<{
    model: string;
    baseURL: string;
    elapsedMs: number;
    reply: string;
    replyDigest: string;
  } | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    if (!loading) {
      setElapsedSeconds(0);
      return;
    }
    const startedAt = Date.now();
    setElapsedSeconds(0);
    const timer = window.setInterval(() => {
      setElapsedSeconds(Math.floor((Date.now() - startedAt) / 1000));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [loading]);

  const handleGenerate = async () => {
    // 验证表单并构造请求参数，向后端 AI 接口请求生成试题
    const values = await form.validateFields(); // { type, count, language, keyword }
    setLoading(true);
    try {
      const res = await generateQuestions(values);
      message.success(`共生成 ${res.questions.length} 道题目`);
      setResult(res.questions);
    } finally {
      setLoading(false);
    }
  };

  const handleTestConnection = async () => {
    setTesting(true);
    try {
      const res = await testAiConnection();
      setDiagnostic(res.data);
      message.success("AI 连接测试成功");
    } catch (error) {
      setDiagnostic(null);
      message.error(error instanceof Error ? error.message : "AI 连接测试失败");
    } finally {
      setTesting(false);
    }
  };

  const handleApply = async (item: Partial<Question>) => {
    // 单条插入：将单个题目交给外层处理并提示用户
    await onInsert(item);
    message.success(`已插入 ${item.title}`);
  };

  const handleBatchInsert = async () => {
    if (result.length === 0) return;
    setInserting(true);
    try {
      for (const item of result) {
        await onInsert(item);
      }
      message.success(`成功插入 ${result.length} 道题目`);
      setResult([]); // 可选：插入后清空结果
    } catch (error) {
      message.error("批量插入过程中出现错误");
    } finally {
      setInserting(false);
    }
  };

  const waitingTone = elapsedSeconds >= 60 ? "warning" : "info";
  const waitingMessage =
    elapsedSeconds >= 60
      ? "本次生成等待较久，可能是上游模型仍在出结果。若最终超时，服务商侧仍可能已消耗 token。"
      : elapsedSeconds >= 15
        ? "模型正在组织题目，请保持当前窗口。较复杂题目或高峰时段可能会等待更久。"
        : "正在向模型请求题目，请稍候。";

  return (
    <>
      <Drawer
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <img src={aiLogo} alt="AI Logo" style={{ width: 24, height: 24 }} />
            <span>AI 生成试题</span>
          </div>
        }
        open={open}
        width={900}
        onClose={() => setOpen(false)}
      >
        <div style={{ display: 'flex', gap: 16 }}>
          {/* 左侧表单 */}
          <div style={{ width: 320 }}>
            <Form form={form} layout="vertical" initialValues={{ type: "single", count: 3, language: "go", keyword: "" }}>
              <div style={{ display: 'flex', gap: 12 }}>
                <Form.Item name="type" label="题型" rules={[{ required: true }]} style={{ flex: 1 }}> 
                  <Select options={[
                    { label: "单选题", value: "single" },
                    { label: "多选题", value: "multiple" },
                    { label: "编程题", value: "coding" },
                  ]} />
                </Form.Item>
                <Form.Item name="count" label="题目数量" rules={[{ required: true }]} style={{ width: 120 }}> 
                  <Select options={[
                    { label: "1", value: 1 },
                    { label: "3", value: 3 },
                    { label: "5", value: 5 },
                    { label: "10", value: 10 },
                  ]} />
                </Form.Item>
              </div>

              <Form.Item name="language" label="语言" rules={[{ required: true }]}> 
                <Select options={[
                  { label: "Go", value: "go" },
                  { label: "C++", value: "cpp" },
                  { label: "Java", value: "java" },
                  { label: "JavaScript", value: "javascript" },
                  { label: "Python", value: "python" },
                ]} />
              </Form.Item>

              <Form.Item name="keyword" label="关键词" rules={[{ required: true, message: "请填写关键词" }]}> 
                <Input placeholder="如：二叉树 遍历" />
              </Form.Item>

              <Button type="primary" loading={loading} onClick={handleGenerate} block style={{ marginBottom: 16, backgroundColor: '#1677ff' }}>
                生成并预览题库
              </Button>

              <Button loading={testing} onClick={handleTestConnection} block style={{ marginBottom: 16 }}>
                测试连接
              </Button>

              {loading && (
                <Space direction="vertical" size={12} style={{ width: "100%", marginBottom: 16 }}>
                  <div
                    style={{
                      padding: 16,
                      borderRadius: 12,
                      background: "linear-gradient(135deg, rgba(22,119,255,0.12) 0%, rgba(22,119,255,0.04) 100%)",
                      border: "1px solid rgba(22,119,255,0.15)",
                    }}
                  >
                    <Space align="center" size="large">
                      <Statistic title="已等待" value={elapsedSeconds} suffix="秒" />
                      <div>
                        <Typography.Text strong style={{ display: "block", marginBottom: 4 }}>
                          AI 正在生成题目
                        </Typography.Text>
                        <Typography.Text type="secondary">
                          当前请求已自动追加 `/no_think`，以减少不必要的思考耗时。
                        </Typography.Text>
                      </div>
                    </Space>
                  </div>
                  <Alert type={waitingTone} showIcon message={waitingMessage} />
                </Space>
              )}

              {!loading && diagnostic && (
                <Card
                  size="small"
                  style={{
                    marginBottom: 16,
                    borderRadius: 14,
                    background: "linear-gradient(135deg, #f6fbff 0%, #edf5ff 52%, #f7f5ee 100%)",
                    border: "1px solid rgba(17, 86, 161, 0.14)",
                    boxShadow: "0 14px 30px rgba(12, 69, 128, 0.08)",
                  }}
                  title={
                    <Space direction="vertical" size={0}>
                      <Typography.Text strong>AI 连接诊断</Typography.Text>
                      <Typography.Text type="secondary">快速确认当前兼容接口是否健康可用</Typography.Text>
                    </Space>
                  }
                >
                  <Space size="large" wrap style={{ width: "100%", justifyContent: "space-between" }}>
                    <Statistic title="模型" value={diagnostic.model} valueStyle={{ fontSize: 18 }} />
                    <Statistic title="耗时" value={diagnostic.elapsedMs} suffix="ms" valueStyle={{ fontSize: 18, color: diagnostic.elapsedMs > 30000 ? "#c65a28" : "#1d6b57" }} />
                  </Space>
                  <Divider style={{ margin: "16px 0" }} />
                  <Space direction="vertical" size={10} style={{ width: "100%" }}>
                    <div>
                      <Typography.Text type="secondary">Base URL</Typography.Text>
                      <Typography.Paragraph copyable style={{ marginBottom: 0 }}>
                        {diagnostic.baseURL}
                      </Typography.Paragraph>
                    </div>
                    <div>
                      <Typography.Text type="secondary">回复摘要</Typography.Text>
                      <Typography.Paragraph style={{ marginBottom: 0 }}>
                        {diagnostic.replyDigest}
                      </Typography.Paragraph>
                    </div>
                    <div>
                      <Typography.Text type="secondary">原始回复</Typography.Text>
                      <div
                        style={{
                          marginTop: 6,
                          padding: 12,
                          borderRadius: 10,
                          background: "rgba(255,255,255,0.72)",
                          border: "1px solid rgba(0,0,0,0.06)",
                          maxHeight: 120,
                          overflow: "auto",
                          whiteSpace: "pre-wrap",
                          color: "#44556c",
                          fontSize: 13,
                          lineHeight: 1.6,
                        }}
                      >
                        {diagnostic.reply}
                      </div>
                    </div>
                  </Space>
                </Card>
              )}
              
              {result.length > 0 && (
                <Button loading={inserting} onClick={handleBatchInsert} block style={{ marginBottom: 16 }}>
                  一键插入全部 ({result.length})
                </Button>
              )}
            </Form>
          </div>

          {/* 右侧预览 */}
          {/* 右侧预览区域：展示 AI 生成的题目列表，支持单条插入与一键插入 */}
          <div style={{ flex: 1, border: '2px dashed #e8e8e8', borderRadius: 8, padding: 24, minHeight: 420, display: 'flex', flexDirection: 'column' }}>
            {loading ? (
              <div style={{ margin: "auto", textAlign: "center" }}>
                <Typography.Title level={4} style={{ marginBottom: 8 }}>
                  正在生成题目
                </Typography.Title>
                <Typography.Text type="secondary">
                  已等待 {elapsedSeconds} 秒，生成完成后会在这里展示预览结果。
                </Typography.Text>
              </div>
            ) : result.length === 0 ? (
              // 空状态提示
              <div style={{ color: '#d9d9d9', fontSize: 24, fontWeight: 500, margin: 'auto' }}>AI 生成区域</div>
            ) : (
              <div className="space-y-4 overflow-auto flex-1">
                {result.map((item: Partial<Question>, idx: number) => (
                  <div key={`${item.title}-${idx}`} className="border rounded p-4 bg-white shadow-sm">
                    <div className="flex justify-between items-start mb-2">
                      {/* 标题 + 操作 */}
                      <div className="font-bold text-base">{idx + 1}. {item.title}</div>
                      <Button size="small" type="primary" ghost onClick={() => handleApply(item)}>插入题库</Button>
                    </div>
                    {/* meta 信息（题型 / 语言） */}
                    <div className="text-xs text-gray-500 mb-2">
                      <span className="bg-blue-50 text-blue-600 px-2 py-0.5 rounded mr-2">{item.type}</span>
                      <span className="bg-green-50 text-green-600 px-2 py-0.5 rounded">{item.language}</span>
                    </div>
                    {/* 题目正文与选项 */}
                    <div className="text-sm text-gray-700 whitespace-pre-wrap mb-3 bg-gray-50 p-2 rounded">{item.content}</div>
                    {Array.isArray(item.options) && item.options.length > 0 && (
                      <div className="bg-gray-50 p-2 rounded">
                        <ul className="text-sm space-y-1">
                          {item.options.map((opt, i) => (
                            <li key={i} className="flex">
                              <span className="font-medium mr-2">{String.fromCharCode(65 + i)}.</span>
                              <span>{opt}</span>
                            </li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </Drawer>
    </>
  );
}
