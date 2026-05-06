import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { Button, Modal, Space, Table } from "antd";
import dayjs from "dayjs";
const typeMap = {
    single: "单选题",
    multiple: "多选题",
    coding: "编程题",
};
const columns = [
    { title: "标题", dataIndex: "title", key: "title", width: 300 },
    {
        title: "题型",
        dataIndex: "type",
        key: "type",
        width: 100,
        render: (value) => typeMap[value] || value
    },
    {
        title: "创建人",
        dataIndex: "creatorName",
        key: "creatorName",
        width: 100,
        render: (value) => value || "-",
    },
    { title: "语言", dataIndex: "language", key: "language", width: 120 },
    {
        title: "创建时间",
        dataIndex: "createdAt",
        key: "createdAt",
        width: 180,
        render: (value) => dayjs(value).format("YYYY-MM-DD HH:mm"),
    },
];
export function QuestionTable({ data, loading, pagination, onEdit, onDelete, onSelectionChange }) {
    return (_jsx(Table, { rowKey: "id", dataSource: data, loading: loading, rowSelection: onSelectionChange ? {
            onChange: (selectedRowKeys) => onSelectionChange(selectedRowKeys),
        } : undefined, columns: [
            ...columns,
            {
                title: "操作",
                key: "action",
                width: 160,
                render: (_, record) => (_jsxs(Space, { children: [_jsx(Button, { type: "link", onClick: () => onEdit(record), children: "\u7F16\u8F91" }), _jsx(Button, { danger: true, type: "link", onClick: () => {
                                Modal.confirm({
                                    title: "确认删除",
                                    content: "确定要删除该题目吗？此操作无法撤销。",
                                    okText: "确认",
                                    cancelText: "取消",
                                    onOk: () => onDelete(record.id),
                                    centered: true,
                                });
                            }, children: "\u5220\u9664" })] })),
            },
        ], pagination: {
            ...pagination,
            showTotal: (total) => `共 ${total} 条`,
            showSizeChanger: true,
            showQuickJumper: true,
        } }));
}
