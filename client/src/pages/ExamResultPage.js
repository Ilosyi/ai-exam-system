import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useEffect, useState } from "react";
import { Card, Tag, Space, Typography, Spin, Button, Descriptions } from "antd";
import { ArrowLeftOutlined, CheckCircleOutlined, CloseCircleOutlined } from "@ant-design/icons";
import { useParams, useNavigate } from "react-router-dom";
import { getAttemptResult } from "../api/exam";
import dayjs from "dayjs";
const { Title, Text } = Typography;
const typeMap = {
    single: { color: "blue", text: "单选" },
    multiple: { color: "orange", text: "多选" },
    coding: { color: "green", text: "编程" },
};
export function ExamResultPage() {
    const { id } = useParams();
    const attemptId = Number(id);
    const navigate = useNavigate();
    const [attempt, setAttempt] = useState(null);
    const [loading, setLoading] = useState(true);
    useEffect(() => {
        const load = async () => {
            try {
                const res = await getAttemptResult(attemptId);
                setAttempt(res.data);
            }
            catch {
                navigate("/exam");
            }
            finally {
                setLoading(false);
            }
        };
        load();
    }, [attemptId, navigate]);
    if (loading) {
        return _jsx("div", { style: { textAlign: "center", padding: 100 }, children: _jsx(Spin, { size: "large" }) });
    }
    if (!attempt?.paper) {
        return _jsx("div", { style: { textAlign: "center", padding: 100 }, children: "\u65E0\u6CD5\u52A0\u8F7D\u7B54\u5377\u4FE1\u606F" });
    }
    const items = attempt.paper.items || [];
    const answerMap = new Map((attempt.answers || []).map((a) => [a.questionId, a]));
    return (_jsx("div", { style: { minHeight: "100vh", background: "#f0f2f5", padding: "40px 24px" }, children: _jsxs("div", { style: { maxWidth: 900, margin: "0 auto" }, children: [_jsx("div", { style: { marginBottom: 24 }, children: _jsx(Button, { icon: _jsx(ArrowLeftOutlined, {}), onClick: () => navigate("/exam"), children: "\u8FD4\u56DE\u8003\u8BD5\u5217\u8868" }) }), _jsxs(Card, { style: { marginBottom: 24 }, children: [_jsx(Title, { level: 3, style: { marginTop: 0 }, children: attempt.paper.title }), _jsxs(Descriptions, { column: 3, children: [_jsx(Descriptions.Item, { label: "\u5F97\u5206", children: attempt.totalScore ?? "-" }), _jsx(Descriptions.Item, { label: "\u603B\u5206", children: attempt.paper.totalScore }), _jsx(Descriptions.Item, { label: "\u72B6\u6001", children: _jsx(Tag, { color: attempt.status === "submitted" ? "green" : "orange", children: attempt.status === "submitted" ? "已交卷" : "超时自动提交" }) }), _jsx(Descriptions.Item, { label: "\u5F00\u59CB\u65F6\u95F4", children: dayjs(attempt.startedAt).format("YYYY-MM-DD HH:mm:ss") }), _jsx(Descriptions.Item, { label: "\u63D0\u4EA4\u65F6\u95F4", children: attempt.submittedAt ? dayjs(attempt.submittedAt).format("YYYY-MM-DD HH:mm:ss") : "-" })] })] }), _jsx(Card, { title: "\u7B54\u9898\u8BE6\u60C5", children: _jsx(Space, { direction: "vertical", style: { width: "100%" }, size: "middle", children: items.map((item, index) => {
                            const answer = answerMap.get(item.questionId);
                            const typeInfo = typeMap[item.type] || { color: "default", text: item.type };
                            const question = item.question;
                            let studentAnswerText = "未作答";
                            let correctAnswerText = "-";
                            if (question?.options && question.options.length > 0) {
                                const correctIdxs = question.answers || [];
                                correctAnswerText = correctIdxs.map((i) => String.fromCharCode(65 + i)).join(", ");
                                if (answer?.answerJson) {
                                    try {
                                        const studentIdxs = JSON.parse(answer.answerJson);
                                        studentAnswerText = studentIdxs.map((i) => String.fromCharCode(65 + i)).join(", ");
                                    }
                                    catch {
                                        studentAnswerText = answer.answerJson;
                                    }
                                }
                            }
                            return (_jsx(Card, { size: "small", title: _jsxs(Space, { children: [_jsxs("span", { children: ["\u7B2C ", index + 1, " \u9898"] }), _jsx(Tag, { color: typeInfo.color, children: typeInfo.text }), _jsxs("span", { style: { color: "#999" }, children: [item.score, " \u5206"] }), answer?.isCorrect === true && _jsx(CheckCircleOutlined, { style: { color: "#52c41a" } }), answer?.isCorrect === false && _jsx(CloseCircleOutlined, { style: { color: "#ff4d4f" } }), answer?.score != null && _jsxs(Text, { type: "secondary", children: ["\u5F97\u5206: ", answer.score] })] }), children: question ? (_jsxs("div", { children: [_jsx("div", { style: { fontWeight: 500, marginBottom: 8 }, children: question.title }), question.options && question.options.length > 0 && (_jsx("div", { style: { marginBottom: 12 }, children: question.options.map((opt, i) => (_jsxs("div", { style: { padding: "2px 0" }, children: [String.fromCharCode(65 + i), ". ", opt] }, i))) })), _jsxs("div", { style: { display: "flex", gap: 24, marginTop: 12 }, children: [_jsxs("div", { children: [_jsx(Text, { type: "secondary", children: "\u4F60\u7684\u7B54\u6848: " }), _jsx(Text, { style: { color: answer?.isCorrect ? "#52c41a" : "#ff4d4f" }, children: studentAnswerText })] }), _jsxs("div", { children: [_jsx(Text, { type: "secondary", children: "\u6B63\u786E\u7B54\u6848: " }), _jsx(Text, { style: { color: "#52c41a" }, children: correctAnswerText })] })] })] })) : (_jsxs("div", { style: { color: "#999" }, children: ["\u9898\u76EE ID: ", item.questionId] })) }, item.questionId));
                        }) }) })] }) }));
}
