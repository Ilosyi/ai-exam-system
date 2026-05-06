import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useEffect, useMemo, useState } from "react";
import { Card, Table, Tag, Button, Space, Select, Input, Modal, Form, DatePicker, InputNumber, Statistic, Typography, Empty, message } from "antd";
import { PlusOutlined, DeleteOutlined, EditOutlined, SendOutlined, StopOutlined, BarChartOutlined } from "@ant-design/icons";
import { usePapers } from "../hooks/usePapers";
import { fetchPaperSubmissions, publishPaper, unpublishPaper } from "../api/paper";
import dayjs from "dayjs";
import { useNavigate } from "react-router-dom";
import { fetchClasses } from "../api/class";
const { RangePicker } = DatePicker;
const statusMap = {
    draft: { color: "default", text: "草稿" },
    published: { color: "green", text: "已发布" },
    closed: { color: "red", text: "已关闭" },
};
export function PaperManagePage() {
    const { filters, setFilters, data, total, pagination, loading, onDelete, refresh } = usePapers();
    const navigate = useNavigate();
    const [publishModalOpen, setPublishModalOpen] = useState(false);
    const [publishingPaper, setPublishingPaper] = useState(null);
    const [publishLoading, setPublishLoading] = useState(false);
    const [classOptions, setClassOptions] = useState([]);
    const [classLoading, setClassLoading] = useState(false);
    const [submissionModalOpen, setSubmissionModalOpen] = useState(false);
    const [submissionPaper, setSubmissionPaper] = useState(null);
    const [submissionLoading, setSubmissionLoading] = useState(false);
    const [submissionStats, setSubmissionStats] = useState(null);
    const [submissionClassId, setSubmissionClassId] = useState(undefined);
    useEffect(() => {
        if (!publishModalOpen && !submissionModalOpen)
            return;
        let cancelled = false;
        const loadClasses = async () => {
            setClassLoading(true);
            try {
                const res = await fetchClasses({ page: 1, pageSize: 200 });
                if (!cancelled) {
                    setClassOptions(res.data);
                }
            }
            catch (err) {
                message.error(err instanceof Error ? err.message : "班级列表加载失败");
            }
            finally {
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
    const loadSubmissionStats = async (paperId, classId) => {
        setSubmissionLoading(true);
        try {
            const res = await fetchPaperSubmissions(paperId, classId);
            setSubmissionStats(res.data);
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "提交情况加载失败");
        }
        finally {
            setSubmissionLoading(false);
        }
    };
    const handlePublish = async (values) => {
        if (!publishingPaper)
            return;
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
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "发布失败");
        }
        finally {
            setPublishLoading(false);
        }
    };
    const handleDelete = (record) => {
        Modal.confirm({
            title: "确认删除",
            content: `确定要删除试卷「${record.title}」吗？`,
            centered: true,
            onOk: () => onDelete(record.id),
        });
    };
    const openSubmissionModal = (record) => {
        setSubmissionPaper(record);
        setSubmissionClassId(undefined);
        setSubmissionModalOpen(true);
        void loadSubmissionStats(record.id);
    };
    const submissionColumns = useMemo(() => [
        {
            title: "学生",
            dataIndex: "username",
            key: "username",
            render: (value) => _jsx(Typography.Text, { strong: true, children: value }),
        },
        {
            title: "状态",
            dataIndex: "status",
            key: "status",
            width: 120,
            render: (status) => {
                const text = status === "submitted" ? "已交卷" : status === "timeout" ? "超时交卷" : status;
                const color = status === "submitted" ? "green" : status === "timeout" ? "orange" : "blue";
                return _jsx(Tag, { color: color, children: text });
            },
        },
        {
            title: "得分",
            dataIndex: "totalScore",
            key: "totalScore",
            width: 100,
            render: (score) => (score ?? "-"),
        },
        {
            title: "交卷时间",
            dataIndex: "submittedAt",
            key: "submittedAt",
            width: 180,
            render: (value) => (value ? dayjs(value).format("YYYY-MM-DD HH:mm") : "-"),
        },
    ], []);
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
            render: (status) => {
                const info = statusMap[status] || { color: "default", text: status };
                return _jsx(Tag, { color: info.color, children: info.text });
            },
        },
        {
            title: "创建时间",
            dataIndex: "createdAt",
            key: "createdAt",
            width: 180,
            render: (val) => (val ? dayjs(val).format("YYYY-MM-DD HH:mm") : "-"),
        },
        {
            title: "操作",
            key: "action",
            width: 340,
            render: (_, record) => (_jsxs(Space, { children: [_jsx(Button, { type: "link", size: "small", icon: _jsx(EditOutlined, {}), onClick: () => navigate(`/papers/${record.id}/edit`), children: "\u7F16\u8F91" }), _jsx(Button, { type: "link", size: "small", icon: _jsx(BarChartOutlined, {}), onClick: () => openSubmissionModal(record), children: "\u63D0\u4EA4\u60C5\u51B5" }), record.status === "draft" && (_jsx(Button, { type: "link", size: "small", icon: _jsx(SendOutlined, {}), onClick: () => { setPublishingPaper(record); setPublishModalOpen(true); }, children: "\u53D1\u5E03" })), record.status === "published" && (_jsx(Button, { type: "link", size: "small", danger: true, icon: _jsx(StopOutlined, {}), onClick: async () => {
                            try {
                                await unpublishPaper(record.id);
                                message.success("取消发布成功");
                                refresh();
                            }
                            catch (err) {
                                message.error(err instanceof Error ? err.message : "操作失败");
                            }
                        }, children: "\u53D6\u6D88\u53D1\u5E03" })), _jsx(Button, { type: "link", size: "small", danger: true, icon: _jsx(DeleteOutlined, {}), onClick: () => handleDelete(record), children: "\u5220\u9664" })] })),
        },
    ];
    return (_jsxs("div", { className: "page-shell", children: [_jsxs("section", { className: "page-hero app-fade-up", children: [_jsx("div", { className: "page-hero__eyebrow", children: "Exam Publishing Desk" }), _jsx("h1", { className: "page-hero__title", children: "\u8BD5\u5377\u7BA1\u7406" }), _jsx("p", { className: "page-hero__description", children: "\u8FD9\u91CC\u4E32\u8054\u53D1\u5E03\u3001\u53D6\u6D88\u53D1\u5E03\u3001\u73ED\u7EA7\u8FC7\u6EE4\u4E0E\u4EA4\u5377\u7EDF\u8BA1\uFF0C\u5E2E\u52A9\u6559\u5E08\u5FEB\u901F\u5224\u65AD\u8BD5\u5377\u72B6\u6001\u4E0E\u5B66\u751F\u5B8C\u6210\u5EA6\u3002" })] }), _jsxs(Card, { className: "section-card app-fade-up", title: _jsxs(Space, { direction: "vertical", size: 0, children: [_jsx(Typography.Title, { level: 4, style: { margin: 0 }, children: "\u8BD5\u5377\u7BA1\u7406" }), _jsx(Typography.Text, { type: "secondary", children: "\u4ECE\u53D1\u5E03\u3001\u7B5B\u9009\u5230\u4EA4\u5377\u7EDF\u8BA1\uFF0C\u90FD\u5728\u540C\u4E00\u5F20\u6559\u52A1\u603B\u89C8\u8868\u91CC\u5B8C\u6210" })] }), extra: _jsx(Button, { type: "primary", icon: _jsx(PlusOutlined, {}), onClick: () => navigate("/papers/generate"), children: "\u667A\u80FD\u7EC4\u5377" }), children: [_jsxs("div", { className: "dashboard-toolbar", style: { marginBottom: 16 }, children: [_jsxs("div", { className: "dashboard-toolbar__left", children: [_jsx(Input, { placeholder: "\u641C\u7D22\u6807\u9898", value: filters.keyword, onChange: (e) => setFilters((prev) => ({ ...prev, keyword: e.target.value, page: 1 })), style: { width: 220 }, allowClear: true }), _jsx(Select, { placeholder: "\u72B6\u6001", value: filters.status || undefined, onChange: (val) => setFilters((prev) => ({ ...prev, status: val, page: 1 })), style: { width: 140 }, allowClear: true, options: [
                                            { value: "draft", label: "草稿" },
                                            { value: "published", label: "已发布" },
                                            { value: "closed", label: "已关闭" },
                                        ] })] }), _jsx(Typography.Text, { className: "surface-note", children: "\u5EFA\u8BAE\u6559\u5E08\u5728\u53D1\u5E03\u524D\u5148\u786E\u8BA4\u73ED\u7EA7\u3001\u65F6\u95F4\u7A97\u548C\u7EDF\u8BA1\u8303\u56F4\u3002" })] }), _jsx(Table, { rowKey: "id", columns: columns, dataSource: data, loading: loading, pagination: pagination, size: "middle" })] }), _jsx(Modal, { title: "\u53D1\u5E03\u8BD5\u5377", open: publishModalOpen, onCancel: () => setPublishModalOpen(false), footer: null, destroyOnClose: true, children: _jsxs(Form, { layout: "vertical", initialValues: {
                        timeRange: [dayjs().add(8, "hour").startOf("hour"), dayjs().add(1, "day").add(8, "hour").startOf("hour")],
                        duration: 120,
                        classId: undefined,
                    }, onFinish: (values) => handlePublish(values), children: [_jsx(Form.Item, { name: "timeRange", label: "\u8003\u8BD5\u65F6\u95F4", rules: [{ required: true, message: "请选择考试时间" }], children: _jsx(RangePicker, { showTime: true, style: { width: "100%" } }) }), _jsx(Form.Item, { name: "duration", label: "\u7B54\u9898\u65F6\u957F\uFF08\u5206\u949F\uFF09", extra: "0 \u8868\u793A\u4E0D\u9650\u65F6\uFF0C\u5230\u8003\u8BD5\u7ED3\u675F\u81EA\u52A8\u4EA4\u5377", children: _jsx(InputNumber, { min: 0, max: 600, style: { width: "100%" }, placeholder: "0 = \u4E0D\u9650\u65F6" }) }), _jsx(Form.Item, { name: "classId", label: "\u53D1\u5E03\u73ED\u7EA7", extra: "\u4E0D\u9009\u62E9\u5219\u53D1\u5E03\u4E3A\u516C\u5171\u8BD5\u5377\uFF08\u6240\u6709\u5B66\u751F\u53EF\u89C1\uFF09", children: _jsx(Select, { allowClear: true, placeholder: "\u8BF7\u9009\u62E9\u73ED\u7EA7\uFF08\u53EF\u9009\uFF09", loading: classLoading, options: classOptions.map((item) => ({ value: item.id, label: item.name })) }) }), _jsx(Form.Item, { children: _jsxs(Space, { children: [_jsx(Button, { onClick: () => setPublishModalOpen(false), children: "\u53D6\u6D88" }), _jsx(Button, { type: "primary", htmlType: "submit", loading: publishLoading, children: "\u786E\u8BA4\u53D1\u5E03" })] }) })] }) }), _jsx(Modal, { title: submissionPaper ? `提交情况 · ${submissionPaper.title}` : "提交情况", open: submissionModalOpen, onCancel: () => {
                    setSubmissionModalOpen(false);
                    setSubmissionPaper(null);
                    setSubmissionStats(null);
                    setSubmissionClassId(undefined);
                }, footer: null, width: 920, destroyOnClose: true, children: _jsxs(Space, { direction: "vertical", size: "large", style: { width: "100%" }, children: [_jsx("div", { style: {
                                padding: 20,
                                borderRadius: 20,
                                background: "linear-gradient(135deg, #fff8e8 0%, #f3efe3 50%, #e5dcc7 100%)",
                                border: "1px solid rgba(111, 87, 51, 0.15)",
                            }, children: _jsxs(Space, { size: "large", wrap: true, style: { width: "100%", justifyContent: "space-between" }, children: [_jsxs(Space, { direction: "vertical", size: 2, children: [_jsx(Typography.Text, { style: { letterSpacing: 1.2, color: "#866126" }, children: "SUBMISSION SNAPSHOT" }), _jsx(Typography.Title, { level: 3, style: { margin: 0 }, children: "\u4EA4\u5377\u6001\u52BF" }), _jsx(Typography.Text, { type: "secondary", children: "\u652F\u6301\u6309\u73ED\u7EA7\u7B5B\u9009\uFF0C\u5FEB\u901F\u786E\u8BA4\u8C01\u5DF2\u4EA4\u5377\u3001\u8C01\u8FD8\u672A\u5B8C\u6210\u3002" })] }), _jsxs(Space, { wrap: true, children: [_jsx(Statistic, { title: "\u5E94\u4EA4\u4EBA\u6570", value: submissionStats?.expectedCount ?? 0 }), _jsx(Statistic, { title: "\u5DF2\u4EA4\u4EBA\u6570", value: submissionStats?.submittedCount ?? 0, valueStyle: { color: "#1d6b57" } }), _jsx(Statistic, { title: "\u672A\u4EA4\u4EBA\u6570", value: submissionStats?.unsubmittedCount ?? 0, valueStyle: { color: "#c65a28" } })] })] }) }), _jsxs(Space, { style: { width: "100%", justifyContent: "space-between" }, wrap: true, children: [_jsx(Typography.Text, { type: "secondary", children: "\u5DF2\u4EA4\u5B66\u751F\u660E\u7EC6\u6309\u6700\u8FD1\u4EA4\u5377\u65F6\u95F4\u6392\u5E8F\u5C55\u793A\u3002" }), _jsx(Select, { allowClear: true, placeholder: "\u6309\u73ED\u7EA7\u7B5B\u9009", style: { width: 220 }, loading: classLoading, value: submissionClassId, options: classOptions.map((item) => ({ value: item.id, label: item.name })), onChange: (value) => {
                                        setSubmissionClassId(value);
                                        if (submissionPaper) {
                                            void loadSubmissionStats(submissionPaper.id, value);
                                        }
                                    } })] }), _jsx(Table, { rowKey: "studentId", loading: submissionLoading, dataSource: submissionStats?.submittedStudents ?? [], columns: submissionColumns, pagination: false, locale: { emptyText: _jsx(Empty, { description: "\u5F53\u524D\u7B5B\u9009\u6761\u4EF6\u4E0B\u6682\u65E0\u5DF2\u4EA4\u5B66\u751F" }) } })] }) })] }));
}
