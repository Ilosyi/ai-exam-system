import { useState } from "react";
import { Card, Form, Select, InputNumber, Button, Space, Tag, Empty, Spin } from "antd";
import { ThunderboltOutlined, DeleteOutlined, SwapOutlined, SaveOutlined } from "@ant-design/icons";
import { usePaperGenerate } from "../hooks/usePaperGenerate";
import { generatePaper } from "../api/paper";
import type { PaperItem, GenerateRequest } from "../types/paper";
import { useNavigate } from "react-router-dom";
import { Input, Modal } from "antd";

const typeMap: Record<string, string> = { single: "单选", multiple: "多选", coding: "编程" };
const languageOptions = [
  { value: "go", label: "Go" },
  { value: "cpp", label: "C++" },
  { value: "java", label: "Java" },
  { value: "javascript", label: "JavaScript" },
  { value: "python", label: "Python" },
];

export function PaperGeneratePage() {
  const { items, setItems, language, totalScore, loading, generate, removeItem, save } = usePaperGenerate();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [saveModalOpen, setSaveModalOpen] = useState(false);
  const [saving, setSaving] = useState(false);

  const handleGenerate = async (values: GenerateRequest) => {
    await generate(values);
  };

  const handleReplace = async (item: PaperItem, index: number) => {
    try {
      // Re-generate a single question of the same type/language
      const req: GenerateRequest = {
        language: language,
        singleCount: item.type === "single" ? 1 : 0,
        multipleCount: item.type === "multiple" ? 1 : 0,
        codingCount: item.type === "coding" ? 1 : 0,
        totalScore: item.score,
      };
      const res = await generatePaper(req);
      if (res.data.items.length > 0) {
        const newItems = [...items];
        newItems[index] = res.data.items[0];
        newItems[index].score = item.score;
        newItems[index].sortNo = item.sortNo;
        setItems(newItems);
      }
    } catch {
      // Error already handled in api layer
    }
  };

  const handleSave = async (title: string) => {
    if (items.length === 0) return;
    setSaving(true);
    const result = await save(title, language, totalScore, items);
    setSaving(false);
    setSaveModalOpen(false);
    if (result) {
      navigate(`/papers/${result.id}/edit`);
    }
  };

  return (
    <div style={{ display: "flex", gap: 16, height: "calc(100vh - 144px)" }}>
      {/* Left: Preview area (3/4) */}
      <div style={{ flex: 3, overflow: "auto" }}>
        <Card title={`题目预览 (${items.length} 题)`} style={{ height: "100%" }} extra={<span>总分: {totalScore}</span>}>
          {items.length === 0 ? (
            <Empty description="请先配置参数并组卷" />
          ) : (
            <Space direction="vertical" style={{ width: "100%" }} size="middle">
              {items.map((item, index) => (
                <Card
                  key={item.questionId + "-" + index}
                  size="small"
                  title={
                    <Space>
                      <span>第 {index + 1} 题</span>
                      <Tag color={item.type === "single" ? "blue" : item.type === "multiple" ? "orange" : "green"}>
                        {typeMap[item.type] || item.type}
                      </Tag>
                      <span style={{ color: "#999" }}>{item.score} 分</span>
                    </Space>
                  }
                  extra={
                    <Space>
                      <Button size="small" icon={<SwapOutlined />} onClick={() => handleReplace(item, index)}>
                        替换
                      </Button>
                      <Button size="small" danger icon={<DeleteOutlined />} onClick={() => removeItem(item.id)}>
                        删除
                      </Button>
                    </Space>
                  }
                >
                  {item.question ? (
                    <div>
                      <div style={{ fontWeight: 500, marginBottom: 8 }}>{item.question.title}</div>
                      <div style={{ color: "#666", whiteSpace: "pre-wrap", fontSize: 13 }}>{item.question.content}</div>
                      {item.question.options && item.question.options.length > 0 && (
                        <div style={{ marginTop: 8 }}>
                          {item.question.options.map((opt, i) => (
                            <div key={i} style={{ padding: "2px 0" }}>
                              {String.fromCharCode(65 + i)}. {opt}
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  ) : (
                    <div style={{ color: "#999" }}>题目 ID: {item.questionId}</div>
                  )}
                </Card>
              ))}
            </Space>
          )}
        </Card>
      </div>

      {/* Right: Config panel (1/4) */}
      <div style={{ flex: 1, minWidth: 320, maxWidth: 360 }}>
        <Card title="组卷配置" style={{ height: "100%", overflow: "auto" }}>
          <Form form={form} layout="vertical" initialValues={{ language: "go", singleCount: 5, multipleCount: 3, codingCount: 2, totalScore: 100 }} onFinish={handleGenerate}>
            <Form.Item name="language" label="编程语言" rules={[{ required: true }]}>
              <Select options={languageOptions} />
            </Form.Item>
            <Form.Item name="singleCount" label="单选题数量">
              <InputNumber min={0} max={50} style={{ width: "100%" }} />
            </Form.Item>
            <Form.Item name="multipleCount" label="多选题数量">
              <InputNumber min={0} max={50} style={{ width: "100%" }} />
            </Form.Item>
            <Form.Item name="codingCount" label="编程题数量">
              <InputNumber min={0} max={20} style={{ width: "100%" }} />
            </Form.Item>
            <Form.Item name="totalScore" label="总分" rules={[{ required: true }]}>
              <InputNumber min={1} max={1000} style={{ width: "100%" }} />
            </Form.Item>
            <Form.Item>
              <Space style={{ width: "100%", justifyContent: "center" }}>
                <Button type="primary" htmlType="submit" icon={<ThunderboltOutlined />} loading={loading}>
                  随机组卷
                </Button>
                <Button icon={<SaveOutlined />} disabled={items.length === 0} onClick={() => setSaveModalOpen(true)}>
                  保存试卷
                </Button>
              </Space>
            </Form.Item>
          </Form>

          {items.length > 0 && (
            <div style={{ marginTop: 16, padding: 12, background: "#f5f5f5", borderRadius: 8 }}>
              <div style={{ fontWeight: 500, marginBottom: 8 }}>题型统计</div>
              <div>单选: {items.filter((i) => i.type === "single").length} 题</div>
              <div>多选: {items.filter((i) => i.type === "multiple").length} 题</div>
              <div>编程: {items.filter((i) => i.type === "coding").length} 题</div>
            </div>
          )}
        </Card>
      </div>

      <Modal title="保存试卷" open={saveModalOpen} onCancel={() => setSaveModalOpen(false)} onOk={() => {
        const title = (document.getElementById("save-paper-title") as HTMLInputElement)?.value;
        if (title) handleSave(title);
      }} confirmLoading={saving}>
        <Input id="save-paper-title" placeholder="请输入试卷标题" />
      </Modal>
    </div>
  );
}
