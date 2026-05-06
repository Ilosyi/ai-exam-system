import { useEffect, useState } from "react";
import { Card, Button, Space, Tag, Statistic, Row, Col, Modal, DatePicker, Form, InputNumber, Select, message } from "antd";
import { SwapOutlined, DeleteOutlined, SendOutlined, StopOutlined, ArrowLeftOutlined } from "@ant-design/icons";
import { usePaperEdit } from "../hooks/usePaperEdit";
import { useNavigate, useParams } from "react-router-dom";
import type { PaperItem } from "../types/paper";
import dayjs from "dayjs";
import { fetchClasses, type ClassItem } from "../api/class";

const { RangePicker } = DatePicker;

const typeMap: Record<string, { color: string; text: string }> = {
  single: { color: "blue", text: "单选" },
  multiple: { color: "orange", text: "多选" },
  coding: { color: "green", text: "编程" },
};

const statusMap: Record<string, { color: string; text: string }> = {
  draft: { color: "default", text: "草稿" },
  published: { color: "green", text: "已发布" },
  closed: { color: "red", text: "已关闭" },
};

export function PaperEditPage() {
  const { id } = useParams<{ id: string }>();
  const paperId = Number(id);
  const { paper, loading, load, onDeleteItem, onReplaceItem, onPublish, onUnpublish, getItemStats } = usePaperEdit(paperId);
  const navigate = useNavigate();
  const [publishModalOpen, setPublishModalOpen] = useState(false);
  const [classOptions, setClassOptions] = useState<ClassItem[]>([]);
  const [classLoading, setClassLoading] = useState(false);

  useEffect(() => {
    if (paperId) load();
  }, [paperId, load]);

  useEffect(() => {
    if (!publishModalOpen) return;
    let cancelled = false;
    const loadClasses = async () => {
      setClassLoading(true);
      try {
        const res = await fetchClasses({ page: 1, pageSize: 200 });
        if (!cancelled) {
          setClassOptions(res.data);
        }
      } catch (err: unknown) {
        message.error(err instanceof Error ? err.message : "班级列表加载失败");
      } finally {
        if (!cancelled) {
          setClassLoading(false);
        }
      }
    };
    void loadClasses();
    return () => {
      cancelled = true;
    };
  }, [publishModalOpen]);

  if (loading || !paper) {
    return <Card loading />;
  }

  const stats = getItemStats();
  const statusInfo = statusMap[paper.status] || { color: "default", text: paper.status };

  return (
    <div>
      <div style={{ marginBottom: 16, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/papers")}>返回</Button>
          <h2 style={{ margin: 0 }}>{paper.title}</h2>
          <Tag color={statusInfo.color}>{statusInfo.text}</Tag>
        </Space>
        <Space>
          {paper.status === "draft" && (
            <Button type="primary" icon={<SendOutlined />} onClick={() => setPublishModalOpen(true)}>发布</Button>
          )}
          {paper.status === "published" && (
            <Button danger icon={<StopOutlined />} onClick={onUnpublish}>取消发布</Button>
          )}
        </Space>
      </div>

      {/* Stats bar */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card size="small"><Statistic title="总分" value={paper.totalScore} /></Card>
        </Col>
        <Col span={6}>
          <Card size="small"><Statistic title="总题数" value={stats.total} /></Card>
        </Col>
        <Col span={6}>
          <Card size="small"><Statistic title="单选题" value={stats.single} /></Card>
        </Col>
        <Col span={6}>
          <Card size="small"><Statistic title="多选题" value={stats.multiple} /></Card>
        </Col>
      </Row>

      {/* Item list */}
      <Card title={`题目列表 (${paper.items?.length || 0} 题)`}>
        <Space direction="vertical" style={{ width: "100%" }} size="middle">
          {(paper.items || []).map((item: PaperItem, index: number) => {
            const typeInfo = typeMap[item.type] || { color: "default", text: item.type };
            return (
              <Card
                key={item.id}
                size="small"
                title={
                  <Space>
                    <span>第 {index + 1} 题</span>
                    <Tag color={typeInfo.color}>{typeInfo.text}</Tag>
                    <span style={{ color: "#999" }}>{item.score} 分</span>
                  </Space>
                }
                extra={
                  paper.status === "draft" ? (
                    <Space>
                      <Button size="small" icon={<SwapOutlined />} onClick={() => onReplaceItem(item.id)}>替换</Button>
                      <Button size="small" danger icon={<DeleteOutlined />} onClick={() => {
                        Modal.confirm({
                          title: "确认删除",
                          content: "确定要删除这道题吗？删除后总分将自动调整。",
                          centered: true,
                          onOk: () => onDeleteItem(item.id),
                        });
                      }}>删除</Button>
                    </Space>
                  ) : null
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
            );
          })}
        </Space>
      </Card>

      {/* Publish Modal */}
      <Modal
        title="发布试卷"
        open={publishModalOpen}
        onCancel={() => setPublishModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form
          layout="vertical"
          initialValues={{
            timeRange: [dayjs().add(8, "hour").startOf("hour"), dayjs().add(1, "day").add(8, "hour").startOf("hour")],
            duration: 120,
            classId: undefined,
          }}
          onFinish={(values: { timeRange: [dayjs.Dayjs, dayjs.Dayjs]; duration: number; classId?: number }) => {
            onPublish(values.timeRange[0].toISOString(), values.timeRange[1].toISOString(), values.duration || 0, values.classId);
            setPublishModalOpen(false);
          }}
        >
          <Form.Item name="timeRange" label="考试时间" rules={[{ required: true, message: "请选择考试时间" }]}>
            <RangePicker showTime style={{ width: "100%" }} />
          </Form.Item>
          <Form.Item name="duration" label="答题时长（分钟）" extra="0 表示不限时，到考试结束自动交卷">
            <InputNumber min={0} max={600} style={{ width: "100%" }} placeholder="0 = 不限时" />
          </Form.Item>
          <Form.Item name="classId" label="发布班级" extra="不选择则发布为公共试卷（所有学生可见）">
            <Select
              allowClear
              placeholder="请选择班级（可选）"
              loading={classLoading}
              options={classOptions.map((item) => ({ value: item.id, label: item.name }))}
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button onClick={() => setPublishModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit">确认发布</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
