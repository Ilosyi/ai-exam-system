import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useEffect, useState } from "react";
import { Card, List, Button, Tag, Typography, Empty, Spin } from "antd";
import { EnterOutlined, ClockCircleOutlined } from "@ant-design/icons";
import { fetchPublishedPapers, startAttempt } from "../api/exam";
import { useNavigate } from "react-router-dom";
import dayjs from "dayjs";
import { useAuth } from "../hooks/useAuth";
const { Title } = Typography;
export function ExamListPage() {
    const [papers, setPapers] = useState([]);
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();
    const { user, logout } = useAuth();
    useEffect(() => {
        const load = async () => {
            setLoading(true);
            try {
                const res = await fetchPublishedPapers();
                setPapers(res.data || []);
            }
            finally {
                setLoading(false);
            }
        };
        load();
    }, []);
    const handleStart = async (paperId) => {
        try {
            const res = await startAttempt(paperId);
            navigate(`/exam/${res.data.id}/take`);
        }
        catch (err) {
            // Error will be shown by the API response
            alert(err instanceof Error ? err.message : "开始答题失败");
        }
    };
    return (_jsx("div", { className: "exam-shell", children: _jsxs("div", { className: "exam-shell__inner", children: [_jsxs("div", { className: "exam-topbar app-fade-up", children: [_jsxs("div", { children: [_jsx("div", { className: "page-hero__eyebrow", style: { background: "rgba(39,84,197,0.08)", color: "var(--app-brand-strong)" }, children: "Student Entrance" }), _jsx(Title, { level: 2, style: { margin: "12px 0 0", color: "var(--app-brand-strong)", fontFamily: "\"Baskerville\", \"Songti SC\", serif" }, children: "\u5728\u7EBF\u8003\u8BD5" }), _jsx(Typography.Paragraph, { type: "secondary", style: { margin: "8px 0 0" }, children: "\u6309\u5F00\u653E\u65F6\u95F4\u7A97\u8FDB\u5165\u8003\u8BD5\uFF0C\u7CFB\u7EDF\u4F1A\u81EA\u52A8\u5904\u7406\u4FDD\u5B58\u3001\u5012\u8BA1\u65F6\u548C\u4EA4\u5377\u6D41\u7A0B\u3002" })] }), _jsxs("div", { className: "panel-surface", style: { padding: "12px 14px", borderRadius: 20 }, children: [_jsx(Tag, { color: "blue", style: { marginRight: 8 }, children: user?.username }), _jsx(Button, { onClick: () => {
                                        logout();
                                        navigate("/login", { replace: true });
                                    }, children: "\u9000\u51FA\u767B\u5F55" })] })] }), loading ? (_jsx("div", { className: "panel-surface app-fade-up", style: { textAlign: "center", padding: 80, borderRadius: 28 }, children: _jsx(Spin, { size: "large" }) })) : papers.length === 0 ? (_jsx(Card, { className: "section-card app-fade-up", style: { borderRadius: 28 }, children: _jsx(Empty, { description: "\u6682\u65E0\u53EF\u53C2\u52A0\u7684\u8003\u8BD5" }) })) : (_jsx(List, { grid: { gutter: 20, column: 1 }, dataSource: papers, renderItem: (paper) => {
                        const now = dayjs();
                        const start = dayjs(paper.startTime);
                        const end = dayjs(paper.endTime);
                        const isActive = now.isAfter(start) && now.isBefore(end);
                        return (_jsx(List.Item, { className: "app-fade-up", children: _jsx(Card, { className: "glass-card", hoverable: true, style: { borderRadius: 24 }, children: _jsxs("div", { style: { display: "flex", justifyContent: "space-between", alignItems: "center", gap: 20, flexWrap: "wrap" }, children: [_jsxs("div", { children: [_jsx(Title, { level: 4, style: { margin: 0, color: "var(--app-brand-strong)" }, children: paper.title }), _jsxs("div", { style: { marginTop: 10, color: "#566579" }, children: [_jsx(Tag, { color: "blue", children: paper.language }), _jsxs(Tag, { color: "gold", children: ["\u603B\u5206 ", paper.totalScore] })] }), _jsxs("div", { style: { marginTop: 12, color: "#738196", fontSize: 13 }, children: [_jsx(ClockCircleOutlined, { style: { marginRight: 4 } }), start.format("YYYY-MM-DD HH:mm"), " ~ ", end.format("YYYY-MM-DD HH:mm"), paper.duration > 0 && _jsxs("span", { style: { marginLeft: 8 }, children: ["\u7B54\u9898\u9650\u65F6: ", paper.duration, " \u5206\u949F"] })] })] }), _jsx(Button, { type: "primary", icon: _jsx(EnterOutlined, {}), disabled: !isActive, onClick: () => handleStart(paper.paperId), children: isActive ? "进入答题" : "不在考试时间" })] }) }) }));
                    } }))] }) }));
}
