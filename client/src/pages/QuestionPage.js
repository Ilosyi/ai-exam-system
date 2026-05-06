import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState } from "react";
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
export function QuestionPage() {
    const { data, filters, setFilters, pagination, loading, onCreate, onUpdate, onDelete, onDeleteMany } = useQuestions();
    const [editing, setEditing] = useState();
    const [modalOpen, setModalOpen] = useState(false);
    const [aiOpen, setAiOpen] = useState(false);
    const [selectedIds, setSelectedIds] = useState([]);
    const onFilterChange = (changed) => {
        setFilters((prev) => ({ ...prev, ...changed, page: 1 }));
    };
    return (_jsxs("div", { className: "page-shell", children: [_jsxs("section", { className: "page-hero app-fade-up", children: [_jsx("div", { className: "page-hero__eyebrow", children: "Question Studio" }), _jsx("h1", { className: "page-hero__title", children: "\u9898\u5E93\u7BA1\u7406" }), _jsx("p", { className: "page-hero__description", children: "\u5728\u540C\u4E00\u9875\u9762\u5185\u7EF4\u62A4\u7ED3\u6784\u5316\u9898\u5E93\u3001\u8C03\u7528 AI \u8F85\u52A9\u51FA\u9898\uFF0C\u5E76\u5FEB\u901F\u5B8C\u6210\u6279\u91CF\u6E05\u7406\u4E0E\u4EBA\u5DE5\u7F16\u8F91\u3002" })] }), _jsx(Card, { className: "section-card app-fade-up", children: _jsx(Form, { layout: "inline", children: _jsxs(Flex, { gap: "small", wrap: true, children: [_jsx(Form.Item, { label: "\u9898\u578B", children: _jsx(Select, { options: [
                                        { label: "全部", value: "" },
                                        { label: "单选", value: "single" },
                                        { label: "多选", value: "multiple" },
                                        { label: "编程", value: "coding" },
                                    ], value: filters.type, onChange: (value) => onFilterChange({ type: value }), style: { width: 160 } }) }), _jsx(Form.Item, { label: "\u8BED\u8A00", children: _jsx(Select, { options: [
                                        { label: "全部", value: "" },
                                        { label: "Go", value: "go" },
                                        { label: "C++", value: "cpp" },
                                        { label: "Java", value: "java" },
                                        { label: "JavaScript", value: "javascript" },
                                        { label: "Python", value: "python" },
                                    ], value: filters.language, onChange: (value) => onFilterChange({ language: value }), style: { width: 180 } }) }), _jsx(Form.Item, { label: "\u641C\u7D22", children: _jsx(Input, { placeholder: "\u8BF7\u8F93\u5165\u6807\u9898\u6216\u5185\u5BB9", value: filters.keyword, onChange: (e) => onFilterChange({ keyword: e.target.value }), allowClear: true, style: { width: 220 } }) })] }) }) }), _jsxs(Card, { className: "section-card app-fade-up", title: "\u9898\u76EE\u5217\u8868", extra: _jsxs(Space, { children: [_jsx(Dropdown.Button, { type: "primary", onClick: () => setAiOpen(true), menu: {
                                items: [
                                    { key: "ai", label: "AI 出题" },
                                    { key: "manual", label: "自主出题" },
                                ],
                                onClick: ({ key }) => {
                                    if (key === "ai")
                                        setAiOpen(true);
                                    if (key === "manual") {
                                        setEditing(undefined);
                                        setModalOpen(true);
                                    }
                                },
                            }, children: "\u51FA\u9898" }), _jsx(Button, { danger: true, disabled: !selectedIds.length, onClick: () => {
                                Modal.confirm({
                                    title: "确认批量删除",
                                    content: `确定要删除选中的 ${selectedIds.length} 个题目吗？此操作无法撤销。`,
                                    okText: "确认",
                                    cancelText: "取消",
                                    onOk: () => onDeleteMany(selectedIds),
                                    centered: true,
                                });
                            }, children: "\u6279\u91CF\u5220\u9664" })] }), children: [_jsxs("div", { className: "dashboard-toolbar", style: { marginBottom: 16 }, children: [_jsx(Typography.Text, { className: "surface-note", children: "\u5F53\u524D\u5217\u8868\u652F\u6301\u6309\u9898\u578B\u3001\u8BED\u8A00\u3001\u5173\u952E\u8BCD\u8054\u52A8\u7B5B\u9009\uFF0C\u53F3\u4E0A\u89D2\u4E3B\u6309\u94AE\u9ED8\u8BA4\u76F4\u8FBE AI \u51FA\u9898\u62BD\u5C49\u3002" }), _jsx(Space, { children: _jsxs(Typography.Text, { type: "secondary", children: ["\u5DF2\u9009 ", selectedIds.length, " \u6761"] }) })] }), _jsx(QuestionTable, { data: data, loading: loading, pagination: pagination, onSelectionChange: setSelectedIds, onEdit: (record) => {
                            setEditing(record);
                            setModalOpen(true);
                        }, onDelete: (id) => onDelete(id) })] }), _jsx(AiGenerateDrawer, { onInsert: onCreate, open: aiOpen, onOpenChange: setAiOpen }), _jsx(QuestionFormModal, { open: modalOpen, onCancel: () => setModalOpen(false), initialValues: editing, onSubmit: (values) => (editing ? onUpdate(editing.id, values) : onCreate(values)) })] }));
}
