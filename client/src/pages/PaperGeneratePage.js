import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState } from "react";
import { Card, Form, Select, InputNumber, Button, Space, Tag, Empty } from "antd";
import { ThunderboltOutlined, DeleteOutlined, SwapOutlined, SaveOutlined } from "@ant-design/icons";
import { usePaperGenerate } from "../hooks/usePaperGenerate";
import { generatePaper } from "../api/paper";
import { useNavigate } from "react-router-dom";
import { Input, Modal } from "antd";
const typeMap = { single: "单选", multiple: "多选", coding: "编程" };
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
    const handleGenerate = async (values) => {
        await generate(values);
    };
    const handleReplace = async (item, index) => {
        try {
            // Re-generate a single question of the same type/language
            const req = {
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
        }
        catch {
            // Error already handled in api layer
        }
    };
    const handleSave = async (title) => {
        if (items.length === 0)
            return;
        setSaving(true);
        const result = await save(title, language, totalScore, items);
        setSaving(false);
        setSaveModalOpen(false);
        if (result) {
            navigate(`/papers/${result.id}/edit`);
        }
    };
    return (_jsxs("div", { style: { display: "flex", gap: 16, height: "calc(100vh - 144px)" }, children: [_jsx("div", { style: { flex: 3, overflow: "auto" }, children: _jsx(Card, { title: `题目预览 (${items.length} 题)`, style: { height: "100%" }, extra: _jsxs("span", { children: ["\u603B\u5206: ", totalScore] }), children: items.length === 0 ? (_jsx(Empty, { description: "\u8BF7\u5148\u914D\u7F6E\u53C2\u6570\u5E76\u7EC4\u5377" })) : (_jsx(Space, { direction: "vertical", style: { width: "100%" }, size: "middle", children: items.map((item, index) => (_jsx(Card, { size: "small", title: _jsxs(Space, { children: [_jsxs("span", { children: ["\u7B2C ", index + 1, " \u9898"] }), _jsx(Tag, { color: item.type === "single" ? "blue" : item.type === "multiple" ? "orange" : "green", children: typeMap[item.type] || item.type }), _jsxs("span", { style: { color: "#999" }, children: [item.score, " \u5206"] })] }), extra: _jsxs(Space, { children: [_jsx(Button, { size: "small", icon: _jsx(SwapOutlined, {}), onClick: () => handleReplace(item, index), children: "\u66FF\u6362" }), _jsx(Button, { size: "small", danger: true, icon: _jsx(DeleteOutlined, {}), onClick: () => removeItem(item.id), children: "\u5220\u9664" })] }), children: item.question ? (_jsxs("div", { children: [_jsx("div", { style: { fontWeight: 500, marginBottom: 8 }, children: item.question.title }), _jsx("div", { style: { color: "#666", whiteSpace: "pre-wrap", fontSize: 13 }, children: item.question.content }), item.question.options && item.question.options.length > 0 && (_jsx("div", { style: { marginTop: 8 }, children: item.question.options.map((opt, i) => (_jsxs("div", { style: { padding: "2px 0" }, children: [String.fromCharCode(65 + i), ". ", opt] }, i))) }))] })) : (_jsxs("div", { style: { color: "#999" }, children: ["\u9898\u76EE ID: ", item.questionId] })) }, item.questionId + "-" + index))) })) }) }), _jsx("div", { style: { flex: 1, minWidth: 320, maxWidth: 360 }, children: _jsxs(Card, { title: "\u7EC4\u5377\u914D\u7F6E", style: { height: "100%", overflow: "auto" }, children: [_jsxs(Form, { form: form, layout: "vertical", initialValues: { language: "go", singleCount: 5, multipleCount: 3, codingCount: 2, totalScore: 100 }, onFinish: handleGenerate, children: [_jsx(Form.Item, { name: "language", label: "\u7F16\u7A0B\u8BED\u8A00", rules: [{ required: true }], children: _jsx(Select, { options: languageOptions }) }), _jsx(Form.Item, { name: "singleCount", label: "\u5355\u9009\u9898\u6570\u91CF", children: _jsx(InputNumber, { min: 0, max: 50, style: { width: "100%" } }) }), _jsx(Form.Item, { name: "multipleCount", label: "\u591A\u9009\u9898\u6570\u91CF", children: _jsx(InputNumber, { min: 0, max: 50, style: { width: "100%" } }) }), _jsx(Form.Item, { name: "codingCount", label: "\u7F16\u7A0B\u9898\u6570\u91CF", children: _jsx(InputNumber, { min: 0, max: 20, style: { width: "100%" } }) }), _jsx(Form.Item, { name: "totalScore", label: "\u603B\u5206", rules: [{ required: true }], children: _jsx(InputNumber, { min: 1, max: 1000, style: { width: "100%" } }) }), _jsx(Form.Item, { children: _jsxs(Space, { style: { width: "100%", justifyContent: "center" }, children: [_jsx(Button, { type: "primary", htmlType: "submit", icon: _jsx(ThunderboltOutlined, {}), loading: loading, children: "\u968F\u673A\u7EC4\u5377" }), _jsx(Button, { icon: _jsx(SaveOutlined, {}), disabled: items.length === 0, onClick: () => setSaveModalOpen(true), children: "\u4FDD\u5B58\u8BD5\u5377" })] }) })] }), items.length > 0 && (_jsxs("div", { style: { marginTop: 16, padding: 12, background: "#f5f5f5", borderRadius: 8 }, children: [_jsx("div", { style: { fontWeight: 500, marginBottom: 8 }, children: "\u9898\u578B\u7EDF\u8BA1" }), _jsxs("div", { children: ["\u5355\u9009: ", items.filter((i) => i.type === "single").length, " \u9898"] }), _jsxs("div", { children: ["\u591A\u9009: ", items.filter((i) => i.type === "multiple").length, " \u9898"] }), _jsxs("div", { children: ["\u7F16\u7A0B: ", items.filter((i) => i.type === "coding").length, " \u9898"] })] }))] }) }), _jsx(Modal, { title: "\u4FDD\u5B58\u8BD5\u5377", open: saveModalOpen, onCancel: () => setSaveModalOpen(false), onOk: () => {
                    const title = document.getElementById("save-paper-title")?.value;
                    if (title)
                        handleSave(title);
                }, confirmLoading: saving, children: _jsx(Input, { id: "save-paper-title", placeholder: "\u8BF7\u8F93\u5165\u8BD5\u5377\u6807\u9898" }) })] }));
}
