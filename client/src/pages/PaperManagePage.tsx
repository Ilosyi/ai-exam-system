import { useEffect, useMemo, useState } from "react";
import { Card, Table, Tag, Button, Space, Select, Input, Modal, Form, DatePicker, InputNumber, Statistic, Typography, Empty, message } from "antd";
import { PlusOutlined, DeleteOutlined, EditOutlined, SendOutlined, StopOutlined, BarChartOutlined } from "@ant-design/icons";
import { usePapers } from "../hooks/usePapers";
import { fetchPaperSubmissions, publishPaper, unpublishPaper, type PaperSubmissionStats } from "../api/paper";
import type { Paper } from "../types/paper";
import dayjs from "dayjs";
import { useNavigate } from "react-router-dom";
import { fetchClasses, type ClassItem } from "../api/class";

const { RangePicker } = DatePicker;

const statusMap: Record<string, { color: string; text: string }> = {
  draft: { color: "default", text: "草稿" },
  published: { color: "green", text: "已发布" },
  closed: { color: "red", text: "已关闭" },
};

export function PaperManagePage() {
  const { filters, setFilters, data, total, pagination, loading, onDelete, refresh } = usePapers();
  const navigate = useNavigate();
  const [publishModalOpen, setPublishModalOpen] = useState(false);
  const [publishingPaper, setPublishingPaper] = useState<Paper | null>(null);
  const [publishLoading, setPublishLoading] = useState(false);
  const [classOptions, setClassOptions] = useState<ClassItem[]>([]);
  const [classLoading, setClassLoading] = useState(false);
  const [submissionModalOpen, setSubmissionModalOpen] = useState(false);
  const [submissionPaper, setSubmissionPaper] = useState<Paper | null>(null);
  const [submissionLoading, setSubmissionLoading] = useState(false);
  const [submissionStats, setSubmissionStats] = useState<PaperSubmissionStats | null>(null);
  const [submissionClassId, setSubmissionClassId] = useState<number | undefined>(undefined);

  useEffect(() => {
    if (!publishModalOpen && !submissionModalOpen) return;
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
  }, [publishModalOpen, submissionModalOpen, message]);

  const loadSubmissionStats = async (paperId: number, classId?: number) => {
    setSubmissionLoading(true);
    try {
      const res = await fetchPaperSubmissions(paperId, classId);
      setSubmissionStats(res.data);
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "提交情况加载失败");
    } finally {
      setSubmissionLoading(false);
    }
  };

  const handlePublish = async (values: { timeRange: [dayjs.Dayjs, dayjs.Dayjs]; duration: number; classId?: number }) => {
    if (!publishingPaper) return;
    setPublishLoading(true);
    try {
      await publishPaper(publishingPaper.id, {
        startTime: values.timeRange[0].toISOString(),
        endTime: values.timeRange[1].toISOString(),
        duration: values.duration || 0,
        classId: values.classId,
      });
      message.success("发布成功");
      setPublishModalOpen(false);
      refresh();
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "发布失败");
    } finally {
      setPublishLoading(false);
    }
  };

  const handleDelete = (record: Paper) => {
    Modal.confirm({
      title: "确认删除",
      content: `确定要删除试卷「${record.title}」吗？`,
      centered: true,
      onOk: () => onDelete(record.id),
    });
  };

  const openSubmissionModal = (record: Paper) => {
    setSubmissionPaper(record);
    setSubmissionClassId(undefined);
    setSubmissionModalOpen(true);
    void loadSubmissionStats(record.id);
  };

  const submissionColumns = useMemo(
    () => [
      {
        title: "学生",
        dataIndex: "username",
        key: "username",
        render: (value: string) => <Typography.Text strong>{value}</Typography.Text>,
      },
      {
        title: "状态",
        dataIndex: "status",
        key: "status",
        width: 120,
        render: (status: string) => {
          const text = status === "submitted" ? "已交卷" : status === "timeout" ? "超时交卷" : status;
          const color = status === "submitted" ? "green" : status === "timeout" ? "orange" : "blue";
          return <Tag color={color}>{text}</Tag>;
        },
      },
      {
        title: "得分",
        dataIndex: "totalScore",
        key: "totalScore",
        width: 100,
        render: (score?: number) => (score ?? "-"),
      },
      {
        title: "交卷时间",
        dataIndex: "submittedAt",
        key: "submittedAt",
        width: 180,
        render: (value?: string) => (value ? dayjs(value).format("YYYY-MM-DD HH:mm") : "-"),
      },
    ],
    [],
  );

  const columns = [
    { title: "标题", dataIndex: "title", key: "title", ellipsis: true },
    { title: "语言", dataIndex: "language", key: "language", width: 100 },
    {
      title: "总分",
      dataIndex: "totalScore",
      key: "totalScore",
      width: 80,
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
      width: 100,
      render: (status: string) => {
        const info = statusMap[status] || { color: "default", text: status };
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      key: "createdAt",
      width: 180,
      render: (val: string) => (val ? dayjs(val).format("YYYY-MM-DD HH:mm") : "-"),
    },
    {
      title: "操作",
      key: "action",
      width: 340,
      render: (_: unknown, record: Paper) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => navigate(`/papers/${record.id}/edit`)}>
            编辑
          </Button>
          <Button type="link" size="small" icon={<BarChartOutlined />} onClick={() => openSubmissionModal(record)}>
            提交情况
          </Button>
          {record.status === "draft" && (
            <Button type="link" size="small" icon={<SendOutlined />} onClick={() => { setPublishingPaper(record); setPublishModalOpen(true); }}>
              发布
            </Button>
          )}
          {record.status === "published" && (
            <Button type="link" size="small" danger icon={<StopOutlined />} onClick={async () => {
              try {
                await unpublishPaper(record.id);
                message.success("取消发布成功");
                refresh();
              } catch (err: unknown) {
                message.error(err instanceof Error ? err.message : "操作失败");
              }
            }}>
              取消发布
            </Button>
          )}
          <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-shell">
      <section className="page-hero app-fade-up">
        <div className="page-hero__eyebrow">Exam Publishing Desk</div>
        <h1 className="page-hero__title">试卷管理</h1>
        <p className="page-hero__description">
          这里串联发布、取消发布、班级过滤与交卷统计，帮助教师快速判断试卷状态与学生完成度。
        </p>
      </section>

      <Card
        className="section-card app-fade-up"
        title={
          <Space direction="vertical" size={0}>
            <Typography.Title level={4} style={{ margin: 0 }}>
              试卷管理
            </Typography.Title>
            <Typography.Text type="secondary">
              从发布、筛选到交卷统计，都在同一张教务总览表里完成
            </Typography.Text>
          </Space>
        }
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate("/papers/generate")}>
            智能组卷
          </Button>
        }
      >
        <div className="dashboard-toolbar" style={{ marginBottom: 16 }}>
          <div className="dashboard-toolbar__left">
            <Input
              placeholder="搜索标题"
              value={filters.keyword as string}
              onChange={(e) => setFilters((prev) => ({ ...prev, keyword: e.target.value, page: 1 }))}
              style={{ width: 220 }}
              allowClear
            />
            <Select
              placeholder="状态"
              value={filters.status || undefined}
              onChange={(val) => setFilters((prev) => ({ ...prev, status: val, page: 1 }))}
              style={{ width: 140 }}
              allowClear
              options={[
                { value: "draft", label: "草稿" },
                { value: "published", label: "已发布" },
                { value: "closed", label: "已关闭" },
              ]}
            />
          </div>
          <Typography.Text className="surface-note">
            建议教师在发布前先确认班级、时间窗和统计范围。
          </Typography.Text>
        </div>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          pagination={pagination}
          size="middle"
        />
      </Card>

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
          onFinish={(values: { timeRange: [dayjs.Dayjs, dayjs.Dayjs]; duration: number; classId?: number }) => handlePublish(values)}
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
              <Button type="primary" htmlType="submit" loading={publishLoading}>确认发布</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={submissionPaper ? `提交情况 · ${submissionPaper.title}` : "提交情况"}
        open={submissionModalOpen}
        onCancel={() => {
          setSubmissionModalOpen(false);
          setSubmissionPaper(null);
          setSubmissionStats(null);
          setSubmissionClassId(undefined);
        }}
        footer={null}
        width={920}
        destroyOnClose
      >
        <Space direction="vertical" size="large" style={{ width: "100%" }}>
          <div
            style={{
              padding: 20,
              borderRadius: 20,
              background: "linear-gradient(135deg, #fff8e8 0%, #f3efe3 50%, #e5dcc7 100%)",
              border: "1px solid rgba(111, 87, 51, 0.15)",
            }}
          >
            <Space size="large" wrap style={{ width: "100%", justifyContent: "space-between" }}>
              <Space direction="vertical" size={2}>
                <Typography.Text style={{ letterSpacing: 1.2, color: "#866126" }}>SUBMISSION SNAPSHOT</Typography.Text>
                <Typography.Title level={3} style={{ margin: 0 }}>
                  交卷态势
                </Typography.Title>
                <Typography.Text type="secondary">
                  支持按班级筛选，快速确认谁已交卷、谁还未完成。
                </Typography.Text>
              </Space>
              <Space wrap>
                <Statistic title="应交人数" value={submissionStats?.expectedCount ?? 0} />
                <Statistic title="已交人数" value={submissionStats?.submittedCount ?? 0} valueStyle={{ color: "#1d6b57" }} />
                <Statistic title="未交人数" value={submissionStats?.unsubmittedCount ?? 0} valueStyle={{ color: "#c65a28" }} />
              </Space>
            </Space>
          </div>

          <Space style={{ width: "100%", justifyContent: "space-between" }} wrap>
            <Typography.Text type="secondary">
              已交学生明细按最近交卷时间排序展示。
            </Typography.Text>
            <Select
              allowClear
              placeholder="按班级筛选"
              style={{ width: 220 }}
              loading={classLoading}
              value={submissionClassId}
              options={classOptions.map((item) => ({ value: item.id, label: item.name }))}
              onChange={(value) => {
                setSubmissionClassId(value);
                if (submissionPaper) {
                  void loadSubmissionStats(submissionPaper.id, value);
                }
              }}
            />
          </Space>

          <Table
            rowKey="studentId"
            loading={submissionLoading}
            dataSource={submissionStats?.submittedStudents ?? []}
            columns={submissionColumns}
            pagination={false}
            locale={{ emptyText: <Empty description="当前筛选条件下暂无已交学生" /> }}
          />
        </Space>
      </Modal>
    </div>
  );
}
