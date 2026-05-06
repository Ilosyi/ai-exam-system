import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useEffect, useState } from "react";
import { Card, Button, Space, Tag, Statistic, Row, Col, Modal, DatePicker, Form, InputNumber, Select, message } from "antd";
import { SwapOutlined, DeleteOutlined, SendOutlined, StopOutlined, ArrowLeftOutlined } from "@ant-design/icons";
import { usePaperEdit } from "../hooks/usePaperEdit";
import { useNavigate, useParams } from "react-router-dom";
import dayjs from "dayjs";
import { fetchClasses } from "../api/class";
const { RangePicker } = DatePicker;
const typeMap = {
    single: { color: "blue", text: "单选" },
    multiple: { color: "orange", text: "多选" },
    coding: { color: "green", text: "编程" },
};
const statusMap = {
    draft: { color: "default", text: "草稿" },
    published: { color: "green", text: "已发布" },
    closed: { color: "red", text: "已关闭" },
};
export function PaperEditPage() {
    const { id } = useParams();
    const paperId = Number(id);
    const { paper, loading, load, onDeleteItem, onReplaceItem, onPublish, onUnpublish, getItemStats } = usePaperEdit(paperId);
    const navigate = useNavigate();
    const [publishModalOpen, setPublishModalOpen] = useState(false);
    const [classOptions, setClassOptions] = useState([]);
    const [classLoading, setClassLoading] = useState(false);
    useEffect(() => {
        if (paperId)
            load();
    }, [paperId, load]);
    useEffect(() => {
        if (!publishModalOpen)
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
    }, [publishModalOpen]);
    if (loading || !paper) {
        return _jsx(Card, { loading: true });
    }
    const stats = getItemStats();
    const statusInfo = statusMap[paper.status] || { color: "default", text: paper.status };
    return (_jsxs("div", { children: [_jsxs("div", { style: { marginBottom: 16, display: "flex", justifyContent: "space-between", alignItems: "center" }, children: [_jsxs(Space, { children: [_jsx(Button, { icon: _jsx(ArrowLeftOutlined, {}), onClick: () => navigate("/papers"), children: "\u8FD4\u56DE" }), _jsx("h2", { style: { margin: 0 }, children: paper.title }), _jsx(Tag, { color: statusInfo.color, children: statusInfo.text })] }), _jsxs(Space, { children: [paper.status === "draft" && (_jsx(Button, { type: "primary", icon: _jsx(SendOutlined, {}), onClick: () => setPublishModalOpen(true), children: "\u53D1\u5E03" })), paper.status === "published" && (_jsx(Button, { danger: true, icon: _jsx(StopOutlined, {}), onClick: onUnpublish, children: "\u53D6\u6D88\u53D1\u5E03" }))] })] }), _jsxs(Row, { gutter: 16, style: { marginBottom: 16 }, children: [_jsx(Col, { span: 6, children: _jsx(Card, { size: "small", children: _jsx(Statistic, { title: "\u603B\u5206", value: paper.totalScore }) }) }), _jsx(Col, { span: 6, children: _jsx(Card, { size: "small", children: _jsx(Statistic, { title: "\u603B\u9898\u6570", value: stats.total }) }) }), _jsx(Col, { span: 6, children: _jsx(Card, { size: "small", children: _jsx(Statistic, { title: "\u5355\u9009\u9898", value: stats.single }) }) }), _jsx(Col, { span: 6, children: _jsx(Card, { size: "small", children: _jsx(Statistic, { title: "\u591A\u9009\u9898", value: stats.multiple }) }) })] }), _jsx(Card, { title: `题目列表 (${paper.items?.length || 0} 题)`, children: _jsx(Space, { direction: "vertical", style: { width: "100%" }, size: "middle", children: (paper.items || []).map((item, index) => {
                        const typeInfo = typeMap[item.type] || { color: "default", text: item.type };
                        return (_jsx(Card, { size: "small", title: _jsxs(Space, { children: [_jsxs("span", { children: ["\u7B2C ", index + 1, " \u9898"] }), _jsx(Tag, { color: typeInfo.color, children: typeInfo.text }), _jsxs("span", { style: { color: "#999" }, children: [item.score, " \u5206"] })] }), extra: paper.status === "draft" ? (_jsxs(Space, { children: [_jsx(Button, { size: "small", icon: _jsx(SwapOutlined, {}), onClick: () => onReplaceItem(item.id), children: "\u66FF\u6362" }), _jsx(Button, { size: "small", danger: true, icon: _jsx(DeleteOutlined, {}), onClick: () => {
                                            Modal.confirm({
                                                title: "确认删除",
                                                content: "确定要删除这道题吗？删除后总分将自动调整。",
                                                centered: true,
                                                onOk: () => onDeleteItem(item.id),
                                            });
                                        }, children: "\u5220\u9664" })] })) : null, children: item.question ? (_jsxs("div", { children: [_jsx("div", { style: { fontWeight: 500, marginBottom: 8 }, children: item.question.title }), _jsx("div", { style: { color: "#666", whiteSpace: "pre-wrap", fontSize: 13 }, children: item.question.content }), item.question.options && item.question.options.length > 0 && (_jsx("div", { style: { marginTop: 8 }, children: item.question.options.map((opt, i) => (_jsxs("div", { style: { padding: "2px 0" }, children: [String.fromCharCode(65 + i), ". ", opt] }, i))) }))] })) : (_jsxs("div", { style: { color: "#999" }, children: ["\u9898\u76EE ID: ", item.questionId] })) }, item.id));
                    }) }) }), _jsx(Modal, { title: "\u53D1\u5E03\u8BD5\u5377", open: publishModalOpen, onCancel: () => setPublishModalOpen(false), footer: null, destroyOnClose: true, children: _jsxs(Form, { layout: "vertical", initialValues: {
                        timeRange: [dayjs().add(8, "hour").startOf("hour"), dayjs().add(1, "day").add(8, "hour").startOf("hour")],
                        duration: 120,
                        classId: undefined,
                    }, onFinish: (values) => {
                        onPublish(values.timeRange[0].toISOString(), values.timeRange[1].toISOString(), values.duration || 0, values.classId);
                        setPublishModalOpen(false);
                    }, children: [_jsx(Form.Item, { name: "timeRange", label: "\u8003\u8BD5\u65F6\u95F4", rules: [{ required: true, message: "请选择考试时间" }], children: _jsx(RangePicker, { showTime: true, style: { width: "100%" } }) }), _jsx(Form.Item, { name: "duration", label: "\u7B54\u9898\u65F6\u957F\uFF08\u5206\u949F\uFF09", extra: "0 \u8868\u793A\u4E0D\u9650\u65F6\uFF0C\u5230\u8003\u8BD5\u7ED3\u675F\u81EA\u52A8\u4EA4\u5377", children: _jsx(InputNumber, { min: 0, max: 600, style: { width: "100%" }, placeholder: "0 = \u4E0D\u9650\u65F6" }) }), _jsx(Form.Item, { name: "classId", label: "\u53D1\u5E03\u73ED\u7EA7", extra: "\u4E0D\u9009\u62E9\u5219\u53D1\u5E03\u4E3A\u516C\u5171\u8BD5\u5377\uFF08\u6240\u6709\u5B66\u751F\u53EF\u89C1\uFF09", children: _jsx(Select, { allowClear: true, placeholder: "\u8BF7\u9009\u62E9\u73ED\u7EA7\uFF08\u53EF\u9009\uFF09", loading: classLoading, options: classOptions.map((item) => ({ value: item.id, label: item.name })) }) }), _jsx(Form.Item, { children: _jsxs(Space, { children: [_jsx(Button, { onClick: () => setPublishModalOpen(false), children: "\u53D6\u6D88" }), _jsx(Button, { type: "primary", htmlType: "submit", children: "\u786E\u8BA4\u53D1\u5E03" })] }) })] }) })] }));
}
