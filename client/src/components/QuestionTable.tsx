import { Button, Modal, Space, Table } from "antd";
import type { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";
import type { Question } from "../types/question";

/**
 * QuestionTable
 * -------------
 * 封装题目表格的展示逻辑：列定义、序列化（如题型中文映射）、以及操作列。
 * - 将 type/语言等字段以用户友好的形式展示
 * - 提供行选择回调，供上层执行批量操作
 */

interface Props {
  data: Question[];
  loading: boolean;
  pagination: {
    current: number;
    pageSize: number;
    total: number;
    onChange: (page: number, pageSize?: number) => void;
  };
  onEdit: (record: Question) => void;
  onDelete: (id: number) => void;
  onSelectionChange?: (ids: number[]) => void;
}

const typeMap: Record<string, string> = {
  single: "单选题",
  multiple: "多选题",
  coding: "编程题",
};

const columns: ColumnsType<Question> = [
  { title: "标题", dataIndex: "title", key: "title", width: 300 },
  { 
    title: "题型", 
    dataIndex: "type", 
    key: "type", 
    width: 100,
    render: (value: string) => typeMap[value] || value
  },
  {
    title: "创建人",
    dataIndex: "creatorName",
    key: "creatorName",
    width: 100,
    render: (value: string) => value || "-",
  },
  { title: "语言", dataIndex: "language", key: "language", width: 120 },
  {
    title: "创建时间",
    dataIndex: "createdAt",
    key: "createdAt",
    width: 180,
    render: (value: string) => dayjs(value).format("YYYY-MM-DD HH:mm"),
  },
];

export function QuestionTable({ data, loading, pagination, onEdit, onDelete, onSelectionChange }: Props) {
  return (
    <Table<Question>
      rowKey="id"
      dataSource={data}
      loading={loading}
      rowSelection={onSelectionChange ? {
        onChange: (selectedRowKeys) => onSelectionChange(selectedRowKeys as number[]),
      } : undefined}
      columns={[
        ...columns,
        {
          title: "操作",
          key: "action",
          width: 160,
          render: (_: unknown, record: Question) => (
            <Space>
              {/* 编辑：由上层传入 onEdit 以打开编辑弹窗并注入当前记录 */}
              <Button type="link" onClick={() => onEdit(record)}>
                编辑
              </Button>
              {/* 删除：在页面中心弹窗确认（Modal.confirm），以便用户更明显地感知不可逆操作 */}
              <Button 
                danger 
                type="link" 
                onClick={() => {
                  Modal.confirm({
                    title: "确认删除",
                    content: "确定要删除该题目吗？此操作无法撤销。",
                    okText: "确认",
                    cancelText: "取消",
                    onOk: () => onDelete(record.id),
                    centered: true,
                  });
                }}
              >
                删除
              </Button>
            </Space>
          ),
        },
      ]}
      pagination={{
        ...pagination,
        showTotal: (total) => `共 ${total} 条`,
        showSizeChanger: true,
        showQuickJumper: true,
      }}
    />
  );
}
