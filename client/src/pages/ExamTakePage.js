import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useCallback, useEffect, useRef, useState } from "react";
import { Button, Card, Radio, Checkbox, Tag, Space, Modal, Spin, Typography } from "antd";
import { ArrowLeftOutlined, ArrowRightOutlined, SendOutlined, ClockCircleOutlined } from "@ant-design/icons";
import { useParams, useNavigate } from "react-router-dom";
import { getAttempt, saveAnswers, submitAttempt, recordProctorEvent } from "../api/exam";
import dayjs from "dayjs";
const { Title, Text } = Typography;
const typeMap = {
    single: { color: "blue", text: "单选" },
    multiple: { color: "orange", text: "多选" },
    coding: { color: "green", text: "编程" },
};
function formatCountdown(ms) {
    if (ms <= 0)
        return "00:00:00";
    const totalSec = Math.floor(ms / 1000);
    const h = Math.floor(totalSec / 3600);
    const m = Math.floor((totalSec % 3600) / 60);
    const s = totalSec % 60;
    return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}
export function ExamTakePage() {
    const { id } = useParams();
    const attemptId = Number(id);
    const navigate = useNavigate();
    const [attempt, setAttempt] = useState(null);
    const [loading, setLoading] = useState(true);
    const [currentIndex, setCurrentIndex] = useState(0);
    const [answers, setAnswers] = useState({});
    const [submitting, setSubmitting] = useState(false);
    const [remainingMs, setRemainingMs] = useState(null);
    const autoSaveTimer = useRef(null);
    const countdownTimer = useRef(null);
    const submittedRef = useRef(false);
    const loadAttempt = useCallback(async () => {
        setLoading(true);
        try {
            const res = await getAttempt(attemptId);
            setAttempt(res.data);
            // Restore saved answers
            if (res.data.answers) {
                const saved = {};
                for (const a of res.data.answers) {
                    try {
                        saved[a.questionId] = JSON.parse(a.answerJson || "[]");
                    }
                    catch {
                        saved[a.questionId] = [];
                    }
                }
                setAnswers(saved);
            }
        }
        catch {
            navigate("/exam");
        }
        finally {
            setLoading(false);
        }
    }, [attemptId, navigate]);
    useEffect(() => {
        loadAttempt();
    }, [loadAttempt]);
    // Countdown timer
    useEffect(() => {
        if (!attempt?.deadline || attempt.status !== "in_progress")
            return;
        const deadlineMs = dayjs(attempt.deadline).valueOf();
        const tick = () => {
            const diff = deadlineMs - Date.now();
            setRemainingMs(diff);
            if (diff <= 0 && !submittedRef.current) {
                submittedRef.current = true;
                // Auto-submit
                Modal.warning({
                    title: "答题时间已到",
                    content: "系统将自动为您交卷",
                    centered: true,
                    onOk: async () => {
                        const entries = Object.entries(answers).map(([qId, ans]) => ({
                            questionId: Number(qId),
                            answerJson: JSON.stringify(ans),
                        }));
                        if (entries.length > 0) {
                            await saveAnswers(attemptId, entries).catch(() => { });
                        }
                        await submitAttempt(attemptId).catch(() => { });
                        navigate(`/exam/${attemptId}/result`);
                    },
                });
                if (countdownTimer.current)
                    clearInterval(countdownTimer.current);
            }
        };
        tick(); // initial tick
        countdownTimer.current = setInterval(tick, 1000);
        return () => {
            if (countdownTimer.current)
                clearInterval(countdownTimer.current);
        };
    }, [attempt?.deadline, attempt?.status, attemptId, answers, navigate]);
    // Auto-save every 30 seconds
    useEffect(() => {
        autoSaveTimer.current = setInterval(() => {
            const entries = Object.entries(answers).map(([qId, ans]) => ({
                questionId: Number(qId),
                answerJson: JSON.stringify(ans),
            }));
            if (entries.length > 0) {
                saveAnswers(attemptId, entries).catch(() => { });
            }
        }, 30000);
        return () => {
            if (autoSaveTimer.current)
                clearInterval(autoSaveTimer.current);
        };
    }, [attemptId, answers]);
    // Proctor events
    useEffect(() => {
        const handleVisibilityChange = () => {
            if (document.hidden) {
                recordProctorEvent(attemptId, "visibilitychange", JSON.stringify({ hidden: true }));
            }
        };
        const handleBlur = () => {
            recordProctorEvent(attemptId, "blur", JSON.stringify({ timestamp: Date.now() }));
        };
        document.addEventListener("visibilitychange", handleVisibilityChange);
        window.addEventListener("blur", handleBlur);
        return () => {
            document.removeEventListener("visibilitychange", handleVisibilityChange);
            window.removeEventListener("blur", handleBlur);
        };
    }, [attemptId]);
    if (loading || !attempt?.paper?.items) {
        return _jsx("div", { style: { textAlign: "center", padding: 100 }, children: _jsx(Spin, { size: "large" }) });
    }
    const items = attempt.paper.items;
    const currentItem = items[currentIndex];
    const question = currentItem?.question;
    // Group items by type for navigator
    const grouped = items.reduce((acc, item) => {
        if (!acc[item.type])
            acc[item.type] = [];
        acc[item.type].push(item);
        return acc;
    }, {});
    const handleAnswer = (questionId, answer) => {
        setAnswers((prev) => ({ ...prev, [questionId]: answer }));
    };
    const handleSubmit = async () => {
        const answeredCount = Object.keys(answers).filter((k) => answers[Number(k)].length > 0).length;
        const unansweredCount = items.length - answeredCount;
        Modal.confirm({
            title: "确认交卷",
            content: (_jsxs("div", { children: [_jsx("div", { style: { marginBottom: 12 }, children: items.map((item, index) => {
                            const ans = answers[item.questionId];
                            const isAnswered = ans && ans.length > 0;
                            return (_jsx("span", { style: {
                                    display: "inline-block",
                                    width: 32,
                                    height: 32,
                                    lineHeight: "32px",
                                    textAlign: "center",
                                    margin: 2,
                                    borderRadius: 4,
                                    background: isAnswered ? "#1890ff" : "#d9d9d9",
                                    color: isAnswered ? "#fff" : "#666",
                                    fontSize: 13,
                                }, children: index + 1 }, item.questionId));
                        }) }), unansweredCount > 0 && (_jsxs("div", { style: { color: "#faad14" }, children: ["\u8FD8\u6709 ", unansweredCount, " \u9898\u672A\u4F5C\u7B54\uFF0C\u786E\u5B9A\u8981\u4EA4\u5377\u5417\uFF1F"] }))] })),
            centered: true,
            okText: "确认交卷",
            cancelText: "继续答题",
            onOk: async () => {
                submittedRef.current = true;
                setSubmitting(true);
                const entries = Object.entries(answers).map(([qId, ans]) => ({
                    questionId: Number(qId),
                    answerJson: JSON.stringify(ans),
                }));
                if (entries.length > 0) {
                    await saveAnswers(attemptId, entries);
                }
                await submitAttempt(attemptId);
                navigate(`/exam/${attemptId}/result`);
            },
        });
    };
    // Countdown display style
    const isUrgent = remainingMs !== null && remainingMs > 0 && remainingMs < 5 * 60 * 1000; // < 5 min
    const isExpired = remainingMs !== null && remainingMs <= 0;
    return (_jsxs("div", { style: { minHeight: "100vh", background: "#f0f2f5" }, children: [_jsxs("div", { style: { background: "#fff", padding: "12px 24px", display: "flex", justifyContent: "space-between", alignItems: "center", boxShadow: "0 2px 8px rgba(0,0,0,0.06)" }, children: [_jsx(Title, { level: 4, style: { margin: 0 }, children: attempt.paper.title }), _jsxs(Space, { size: "large", children: [remainingMs !== null && !isExpired && (_jsxs("span", { style: {
                                    fontSize: 18,
                                    fontWeight: 700,
                                    fontFamily: "monospace",
                                    color: isUrgent ? "#ff4d4f" : "#52c41a",
                                    background: isUrgent ? "#fff1f0" : "#f6ffed",
                                    padding: "4px 16px",
                                    borderRadius: 6,
                                    border: `1px solid ${isUrgent ? "#ffa39e" : "#b7eb8f"}`,
                                }, children: [_jsx(ClockCircleOutlined, { style: { marginRight: 6 } }), formatCountdown(remainingMs)] })), _jsx(Button, { type: "primary", danger: true, icon: _jsx(SendOutlined, {}), onClick: handleSubmit, loading: submitting, children: "\u4EA4\u5377" })] })] }), _jsxs("div", { style: { display: "flex", height: "calc(100vh - 60px)" }, children: [_jsx("div", { style: { width: 200, background: "#fff", padding: 16, overflow: "auto", borderRight: "1px solid #f0f0f0" }, children: Object.entries(grouped).map(([type, typeItems]) => {
                            const typeInfo = typeMap[type] || { color: "default", text: type };
                            return (_jsxs("div", { style: { marginBottom: 16 }, children: [_jsx(Tag, { color: typeInfo.color, style: { marginBottom: 8 }, children: typeInfo.text }), _jsx("div", { style: { display: "flex", flexWrap: "wrap", gap: 4 }, children: typeItems.map((item) => {
                                            const globalIndex = items.findIndex((i) => i.questionId === item.questionId);
                                            const ans = answers[item.questionId];
                                            const isAnswered = ans && ans.length > 0;
                                            const isCurrent = globalIndex === currentIndex;
                                            return (_jsx("span", { onClick: () => setCurrentIndex(globalIndex), style: {
                                                    display: "inline-block",
                                                    width: 32,
                                                    height: 32,
                                                    lineHeight: "32px",
                                                    textAlign: "center",
                                                    borderRadius: 4,
                                                    cursor: "pointer",
                                                    background: isCurrent ? "#1890ff" : isAnswered ? "#e6f7ff" : "#f5f5f5",
                                                    color: isCurrent ? "#fff" : isAnswered ? "#1890ff" : "#999",
                                                    border: isCurrent ? "2px solid #1890ff" : "1px solid #d9d9d9",
                                                    fontWeight: isCurrent ? 700 : 400,
                                                }, children: globalIndex + 1 }, item.questionId));
                                        }) })] }, type));
                        }) }), _jsxs("div", { style: { flex: 1, padding: 24, overflow: "auto" }, children: [question ? (_jsxs(Card, { children: [_jsx("div", { style: { marginBottom: 16 }, children: _jsxs(Space, { children: [_jsxs("span", { style: { fontSize: 16, fontWeight: 600 }, children: ["\u7B2C ", currentIndex + 1, " \u9898"] }), _jsx(Tag, { color: typeMap[currentItem.type]?.color, children: typeMap[currentItem.type]?.text }), _jsxs(Text, { type: "secondary", children: [currentItem.score, " \u5206"] })] }) }), _jsx("div", { style: { fontWeight: 500, marginBottom: 12, fontSize: 15 }, children: question.title }), _jsx("div", { style: { color: "#666", whiteSpace: "pre-wrap", marginBottom: 20, fontSize: 14 }, children: question.content }), currentItem.type === "single" && question.options && (_jsx(Radio.Group, { value: answers[currentItem.questionId]?.[0], onChange: (e) => handleAnswer(currentItem.questionId, [e.target.value]), children: _jsx(Space, { direction: "vertical", children: question.options.map((opt, i) => (_jsxs(Radio, { value: i, children: [String.fromCharCode(65 + i), ". ", opt] }, i))) }) })), currentItem.type === "multiple" && question.options && (_jsx(Checkbox.Group, { value: answers[currentItem.questionId] || [], onChange: (vals) => handleAnswer(currentItem.questionId, vals), children: _jsx(Space, { direction: "vertical", children: question.options.map((opt, i) => (_jsxs(Checkbox, { value: i, children: [String.fromCharCode(65 + i), ". ", opt] }, i))) }) })), currentItem.type === "coding" && (_jsx("div", { style: { color: "#999" }, children: "\u7F16\u7A0B\u9898\u8BF7\u5728\u4E0B\u65B9\u7F16\u5199\u4EE3\u7801\uFF08\u529F\u80FD\u5F00\u53D1\u4E2D\uFF09" }))] })) : (_jsx(Card, { children: _jsx("div", { style: { textAlign: "center", color: "#999" }, children: "\u9898\u76EE\u52A0\u8F7D\u4E2D..." }) })), _jsxs("div", { style: { marginTop: 16, display: "flex", justifyContent: "space-between" }, children: [_jsx(Button, { icon: _jsx(ArrowLeftOutlined, {}), disabled: currentIndex === 0, onClick: () => setCurrentIndex((prev) => prev - 1), children: "\u4E0A\u4E00\u9898" }), _jsx(Button, { icon: _jsx(ArrowRightOutlined, {}), disabled: currentIndex === items.length - 1, onClick: () => setCurrentIndex((prev) => prev + 1), children: "\u4E0B\u4E00\u9898" })] })] })] })] }));
}
