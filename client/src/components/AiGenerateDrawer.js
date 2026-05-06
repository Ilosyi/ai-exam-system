import { jsx as _jsx, jsxs as _jsxs, Fragment as _Fragment } from "react/jsx-runtime";
import { useEffect, useMemo, useState } from "react";
import { Alert, Button, Card, Divider, Drawer, Form, Input, Select, Space, Statistic, Typography, message } from "antd";
import { generateQuestions, testAiConnection } from "../api/ai";
import aiLogo from "../assets/AIlogo.png";
export function AiGenerateDrawer({ onInsert, open: controlledOpen, onOpenChange }) {
    const [uncontrolledOpen, setUncontrolledOpen] = useState(false);
    const open = controlledOpen ?? uncontrolledOpen;
    const setOpen = useMemo(() => onOpenChange ?? setUncontrolledOpen, [onOpenChange]);
    const [loading, setLoading] = useState(false);
    const [inserting, setInserting] = useState(false);
    const [result, setResult] = useState([]);
    const [elapsedSeconds, setElapsedSeconds] = useState(0);
    const [testing, setTesting] = useState(false);
    const [diagnostic, setDiagnostic] = useState(null);
    const [form] = Form.useForm();
    useEffect(() => {
        if (!loading) {
            setElapsedSeconds(0);
            return;
        }
        const startedAt = Date.now();
        setElapsedSeconds(0);
        const timer = window.setInterval(() => {
            setElapsedSeconds(Math.floor((Date.now() - startedAt) / 1000));
        }, 1000);
        return () => window.clearInterval(timer);
    }, [loading]);
    const handleGenerate = async () => {
        // 验证表单并构造请求参数，向后端 AI 接口请求生成试题
        const values = await form.validateFields(); // { type, count, language, keyword }
        setLoading(true);
        try {
            const res = await generateQuestions(values);
            message.success(`共生成 ${res.questions.length} 道题目`);
            setResult(res.questions);
        }
        finally {
            setLoading(false);
        }
    };
    const handleTestConnection = async () => {
        setTesting(true);
        try {
            const res = await testAiConnection();
            setDiagnostic(res.data);
            message.success("AI 连接测试成功");
        }
        catch (error) {
            setDiagnostic(null);
            message.error(error instanceof Error ? error.message : "AI 连接测试失败");
        }
        finally {
            setTesting(false);
        }
    };
    const handleApply = async (item) => {
        // 单条插入：将单个题目交给外层处理并提示用户
        await onInsert(item);
        message.success(`已插入 ${item.title}`);
    };
    const handleBatchInsert = async () => {
        if (result.length === 0)
            return;
        setInserting(true);
        try {
            for (const item of result) {
                await onInsert(item);
            }
            message.success(`成功插入 ${result.length} 道题目`);
            setResult([]); // 可选：插入后清空结果
        }
        catch (error) {
            message.error("批量插入过程中出现错误");
        }
        finally {
            setInserting(false);
        }
    };
    const waitingTone = elapsedSeconds >= 60 ? "warning" : "info";
    const waitingMessage = elapsedSeconds >= 60
        ? "本次生成等待较久，可能是上游模型仍在出结果。若最终超时，服务商侧仍可能已消耗 token。"
        : elapsedSeconds >= 15
            ? "模型正在组织题目，请保持当前窗口。较复杂题目或高峰时段可能会等待更久。"
            : "正在向模型请求题目，请稍候。";
    return (_jsx(_Fragment, { children: _jsx(Drawer, { title: _jsxs("div", { style: { display: 'flex', alignItems: 'center', gap: 8 }, children: [_jsx("img", { src: aiLogo, alt: "AI Logo", style: { width: 24, height: 24 } }), _jsx("span", { children: "AI \u751F\u6210\u8BD5\u9898" })] }), open: open, width: 900, onClose: () => setOpen(false), children: _jsxs("div", { style: { display: 'flex', gap: 16 }, children: [_jsx("div", { style: { width: 320 }, children: _jsxs(Form, { form: form, layout: "vertical", initialValues: { type: "single", count: 3, language: "go", keyword: "" }, children: [_jsxs("div", { style: { display: 'flex', gap: 12 }, children: [_jsx(Form.Item, { name: "type", label: "\u9898\u578B", rules: [{ required: true }], style: { flex: 1 }, children: _jsx(Select, { options: [
                                                    { label: "单选题", value: "single" },
                                                    { label: "多选题", value: "multiple" },
                                                    { label: "编程题", value: "coding" },
                                                ] }) }), _jsx(Form.Item, { name: "count", label: "\u9898\u76EE\u6570\u91CF", rules: [{ required: true }], style: { width: 120 }, children: _jsx(Select, { options: [
                                                    { label: "1", value: 1 },
                                                    { label: "3", value: 3 },
                                                    { label: "5", value: 5 },
                                                    { label: "10", value: 10 },
                                                ] }) })] }), _jsx(Form.Item, { name: "language", label: "\u8BED\u8A00", rules: [{ required: true }], children: _jsx(Select, { options: [
                                            { label: "Go", value: "go" },
                                            { label: "C++", value: "cpp" },
                                            { label: "Java", value: "java" },
                                            { label: "JavaScript", value: "javascript" },
                                            { label: "Python", value: "python" },
                                        ] }) }), _jsx(Form.Item, { name: "keyword", label: "\u5173\u952E\u8BCD", rules: [{ required: true, message: "请填写关键词" }], children: _jsx(Input, { placeholder: "\u5982\uFF1A\u4E8C\u53C9\u6811 \u904D\u5386" }) }), _jsx(Button, { type: "primary", loading: loading, onClick: handleGenerate, block: true, style: { marginBottom: 16, backgroundColor: '#1677ff' }, children: "\u751F\u6210\u5E76\u9884\u89C8\u9898\u5E93" }), _jsx(Button, { loading: testing, onClick: handleTestConnection, block: true, style: { marginBottom: 16 }, children: "\u6D4B\u8BD5\u8FDE\u63A5" }), loading && (_jsxs(Space, { direction: "vertical", size: 12, style: { width: "100%", marginBottom: 16 }, children: [_jsx("div", { style: {
                                                padding: 16,
                                                borderRadius: 12,
                                                background: "linear-gradient(135deg, rgba(22,119,255,0.12) 0%, rgba(22,119,255,0.04) 100%)",
                                                border: "1px solid rgba(22,119,255,0.15)",
                                            }, children: _jsxs(Space, { align: "center", size: "large", children: [_jsx(Statistic, { title: "\u5DF2\u7B49\u5F85", value: elapsedSeconds, suffix: "\u79D2" }), _jsxs("div", { children: [_jsx(Typography.Text, { strong: true, style: { display: "block", marginBottom: 4 }, children: "AI \u6B63\u5728\u751F\u6210\u9898\u76EE" }), _jsx(Typography.Text, { type: "secondary", children: "\u5F53\u524D\u8BF7\u6C42\u5DF2\u81EA\u52A8\u8FFD\u52A0 `/no_think`\uFF0C\u4EE5\u51CF\u5C11\u4E0D\u5FC5\u8981\u7684\u601D\u8003\u8017\u65F6\u3002" })] })] }) }), _jsx(Alert, { type: waitingTone, showIcon: true, message: waitingMessage })] })), !loading && diagnostic && (_jsxs(Card, { size: "small", style: {
                                        marginBottom: 16,
                                        borderRadius: 14,
                                        background: "linear-gradient(135deg, #f6fbff 0%, #edf5ff 52%, #f7f5ee 100%)",
                                        border: "1px solid rgba(17, 86, 161, 0.14)",
                                        boxShadow: "0 14px 30px rgba(12, 69, 128, 0.08)",
                                    }, title: _jsxs(Space, { direction: "vertical", size: 0, children: [_jsx(Typography.Text, { strong: true, children: "AI \u8FDE\u63A5\u8BCA\u65AD" }), _jsx(Typography.Text, { type: "secondary", children: "\u5FEB\u901F\u786E\u8BA4\u5F53\u524D\u517C\u5BB9\u63A5\u53E3\u662F\u5426\u5065\u5EB7\u53EF\u7528" })] }), children: [_jsxs(Space, { size: "large", wrap: true, style: { width: "100%", justifyContent: "space-between" }, children: [_jsx(Statistic, { title: "\u6A21\u578B", value: diagnostic.model, valueStyle: { fontSize: 18 } }), _jsx(Statistic, { title: "\u8017\u65F6", value: diagnostic.elapsedMs, suffix: "ms", valueStyle: { fontSize: 18, color: diagnostic.elapsedMs > 30000 ? "#c65a28" : "#1d6b57" } })] }), _jsx(Divider, { style: { margin: "16px 0" } }), _jsxs(Space, { direction: "vertical", size: 10, style: { width: "100%" }, children: [_jsxs("div", { children: [_jsx(Typography.Text, { type: "secondary", children: "Base URL" }), _jsx(Typography.Paragraph, { copyable: true, style: { marginBottom: 0 }, children: diagnostic.baseURL })] }), _jsxs("div", { children: [_jsx(Typography.Text, { type: "secondary", children: "\u56DE\u590D\u6458\u8981" }), _jsx(Typography.Paragraph, { style: { marginBottom: 0 }, children: diagnostic.replyDigest })] }), _jsxs("div", { children: [_jsx(Typography.Text, { type: "secondary", children: "\u539F\u59CB\u56DE\u590D" }), _jsx("div", { style: {
                                                                marginTop: 6,
                                                                padding: 12,
                                                                borderRadius: 10,
                                                                background: "rgba(255,255,255,0.72)",
                                                                border: "1px solid rgba(0,0,0,0.06)",
                                                                maxHeight: 120,
                                                                overflow: "auto",
                                                                whiteSpace: "pre-wrap",
                                                                color: "#44556c",
                                                                fontSize: 13,
                                                                lineHeight: 1.6,
                                                            }, children: diagnostic.reply })] })] })] })), result.length > 0 && (_jsxs(Button, { loading: inserting, onClick: handleBatchInsert, block: true, style: { marginBottom: 16 }, children: ["\u4E00\u952E\u63D2\u5165\u5168\u90E8 (", result.length, ")"] }))] }) }), _jsx("div", { style: { flex: 1, border: '2px dashed #e8e8e8', borderRadius: 8, padding: 24, minHeight: 420, display: 'flex', flexDirection: 'column' }, children: loading ? (_jsxs("div", { style: { margin: "auto", textAlign: "center" }, children: [_jsx(Typography.Title, { level: 4, style: { marginBottom: 8 }, children: "\u6B63\u5728\u751F\u6210\u9898\u76EE" }), _jsxs(Typography.Text, { type: "secondary", children: ["\u5DF2\u7B49\u5F85 ", elapsedSeconds, " \u79D2\uFF0C\u751F\u6210\u5B8C\u6210\u540E\u4F1A\u5728\u8FD9\u91CC\u5C55\u793A\u9884\u89C8\u7ED3\u679C\u3002"] })] })) : result.length === 0 ? (
                        // 空状态提示
                        _jsx("div", { style: { color: '#d9d9d9', fontSize: 24, fontWeight: 500, margin: 'auto' }, children: "AI \u751F\u6210\u533A\u57DF" })) : (_jsx("div", { className: "space-y-4 overflow-auto flex-1", children: result.map((item, idx) => (_jsxs("div", { className: "border rounded p-4 bg-white shadow-sm", children: [_jsxs("div", { className: "flex justify-between items-start mb-2", children: [_jsxs("div", { className: "font-bold text-base", children: [idx + 1, ". ", item.title] }), _jsx(Button, { size: "small", type: "primary", ghost: true, onClick: () => handleApply(item), children: "\u63D2\u5165\u9898\u5E93" })] }), _jsxs("div", { className: "text-xs text-gray-500 mb-2", children: [_jsx("span", { className: "bg-blue-50 text-blue-600 px-2 py-0.5 rounded mr-2", children: item.type }), _jsx("span", { className: "bg-green-50 text-green-600 px-2 py-0.5 rounded", children: item.language })] }), _jsx("div", { className: "text-sm text-gray-700 whitespace-pre-wrap mb-3 bg-gray-50 p-2 rounded", children: item.content }), Array.isArray(item.options) && item.options.length > 0 && (_jsx("div", { className: "bg-gray-50 p-2 rounded", children: _jsx("ul", { className: "text-sm space-y-1", children: item.options.map((opt, i) => (_jsxs("li", { className: "flex", children: [_jsxs("span", { className: "font-medium mr-2", children: [String.fromCharCode(65 + i), "."] }), _jsx("span", { children: opt })] }, i))) }) }))] }, `${item.title}-${idx}`))) })) })] }) }) }));
}
