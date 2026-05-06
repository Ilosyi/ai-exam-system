import { useState, type ChangeEvent } from "react";
import { Button, Card, Flex, Form, Input, Select, Space, Dropdown, Modal, Typography } from "antd";
import { useQuestions } from "../hooks/useQuestions";

/**
 * QuestionPage
 * ------------
 * 题库管理页面：负责渲染筛选栏、出题按钮、题目列表（QuestionTable）以及与 hook 的交互。
 * 主要职责：
 * - 控制筛选条件并传入 useQuestions
 * - 打开/关闭新增或编辑弹窗
 * - 处理批量操作（批量删除）并给出确认提示
 */
import { QuestionFormModal } from "../components/QuestionFormModal";
import { QuestionTable } from "../components/QuestionTable";
import { AiGenerateDrawer } from "../components/AiGenerateDrawer";
import type { Question } from "../types/question";

export function QuestionPage() {
  const { data, filters, setFilters, pagination, loading, onCreate, onUpdate, onDelete, onDeleteMany } = useQuestions();
  const [editing, setEditing] = useState<Question | undefined>();
  const [modalOpen, setModalOpen] = useState(false);
  const [aiOpen, setAiOpen] = useState(false);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);

  const onFilterChange = (changed: Partial<typeof filters>) => {
    setFilters((prev: typeof filters) => ({ ...prev, ...changed, page: 1 }));
  };

  return (
    <div className="page-shell">
      <section className="page-hero app-fade-up">
        <div className="page-hero__eyebrow">Question Studio</div>
        <h1 className="page-hero__title">题库管理</h1>
        <p className="page-hero__description">
          在同一页面内维护结构化题库、调用 AI 辅助出题，并快速完成批量清理与人工编辑。
        </p>
      </section>

      <Card className="section-card app-fade-up">
        <Form layout="inline">
          <Flex gap="small" wrap>
            <Form.Item label="题型">
              <Select
                options={[
                  { label: "全部", value: "" },
                  { label: "单选", value: "single" },
                  { label: "多选", value: "multiple" },
                  { label: "编程", value: "coding" },
                ]}
                value={filters.type}
                onChange={(value) => onFilterChange({ type: value as any })}
                style={{ width: 160 }}
              />
            </Form.Item>
            <Form.Item label="语言">
              <Select
                options={[
                  { label: "全部", value: "" },
                  { label: "Go", value: "go" },
                  { label: "C++", value: "cpp" },
                  { label: "Java", value: "java" },
                  { label: "JavaScript", value: "javascript" },
                  { label: "Python", value: "python" },
                ]}
                value={filters.language}
                onChange={(value) => onFilterChange({ language: value as any })}
                style={{ width: 180 }}
              />
            </Form.Item>
            <Form.Item label="搜索">
              <Input
                placeholder="请输入标题或内容"
                value={filters.keyword}
                onChange={(e: ChangeEvent<HTMLInputElement>) => onFilterChange({ keyword: e.target.value })}
                allowClear
                style={{ width: 220 }}
              />
            </Form.Item>
          </Flex>
        </Form>
      </Card>

      <Card
        className="section-card app-fade-up"
        title="题目列表"
        extra={
          <Space>
            <Dropdown.Button
              type="primary"
              onClick={() => setAiOpen(true)}
              menu={{
                items: [
                  { key: "ai", label: "AI 出题" },
                  { key: "manual", label: "自主出题" },
                ],
                onClick: ({ key }) => {
                  if (key === "ai") setAiOpen(true);
                  if (key === "manual") {
                    setEditing(undefined);
                    setModalOpen(true);
                  }
                },
              }}
            >
              出题
            </Dropdown.Button>
            <Button 
              danger 
              disabled={!selectedIds.length} 
              onClick={() => {
                Modal.confirm({
                  title: "确认批量删除",
                  content: `确定要删除选中的 ${selectedIds.length} 个题目吗？此操作无法撤销。`,
                  okText: "确认",
                  cancelText: "取消",
                  onOk: () => onDeleteMany(selectedIds),
                  centered: true,
                });
              }}
            >
              批量删除
            </Button>
          </Space>
        }
      >
        <div className="dashboard-toolbar" style={{ marginBottom: 16 }}>
          <Typography.Text className="surface-note">
            当前列表支持按题型、语言、关键词联动筛选，右上角主按钮默认直达 AI 出题抽屉。
          </Typography.Text>
          <Space>
            <Typography.Text type="secondary">已选 {selectedIds.length} 条</Typography.Text>
          </Space>
        </div>
        <QuestionTable
          data={data}
          loading={loading}
          pagination={pagination}
          onSelectionChange={setSelectedIds}
          onEdit={(record) => {
            setEditing(record);
            setModalOpen(true);
          }}
          onDelete={(id) => onDelete(id)}
        />
      </Card>
      <AiGenerateDrawer onInsert={onCreate} open={aiOpen} onOpenChange={setAiOpen} />
      <QuestionFormModal
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        initialValues={editing}
        onSubmit={(values) => (editing ? onUpdate(editing.id, values) : onCreate(values))}
      />
    </div>
  );
}
