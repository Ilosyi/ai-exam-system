import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useMemo, useState } from "react";
import { Alert, Button, Card, Form, Input, Select, Space, Tabs, Typography } from "antd";
import { LockOutlined, UserOutlined } from "@ant-design/icons";
import { useLocation, useNavigate } from "react-router-dom";
import { getDefaultRouteByRole, useAuth } from "../hooks/useAuth";
const { Paragraph, Text, Title } = Typography;
export function LoginPage() {
    const { login, register } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const [submitting, setSubmitting] = useState(false);
    const [errorMessage, setErrorMessage] = useState("");
    const redirectTo = useMemo(() => {
        const state = location.state;
        return state?.from?.pathname;
    }, [location.state]);
    const resolveNextPath = (role) => {
        if (redirectTo) {
            if (role === "student" && redirectTo.startsWith("/exam")) {
                return redirectTo;
            }
            if ((role === "admin" || role === "teacher") && !redirectTo.startsWith("/exam")) {
                return redirectTo;
            }
        }
        return getDefaultRouteByRole(role);
    };
    const handleLogin = async (values) => {
        setSubmitting(true);
        setErrorMessage("");
        try {
            const payload = await login(values);
            navigate(resolveNextPath(payload.user.role), { replace: true });
        }
        catch (error) {
            setErrorMessage(error instanceof Error ? error.message : "登录失败");
        }
        finally {
            setSubmitting(false);
        }
    };
    const handleRegister = async (values) => {
        setSubmitting(true);
        setErrorMessage("");
        try {
            const payload = await register({
                username: values.username,
                password: values.password,
                role: values.role,
            });
            navigate(resolveNextPath(payload.user.role), { replace: true });
        }
        catch (error) {
            setErrorMessage(error instanceof Error ? error.message : "注册失败");
        }
        finally {
            setSubmitting(false);
        }
    };
    return (_jsx("div", { style: {
            minHeight: "100vh",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            padding: 24,
            background: "radial-gradient(circle at left top, rgba(39,84,197,0.12), transparent 22%), radial-gradient(circle at right top, rgba(39,84,197,0.1), transparent 18%), linear-gradient(135deg, #f7f9fc 0%, #e9eff7 52%, #f8fbff 100%)",
        }, children: _jsxs(Card, { className: "glass-card app-fade-up", style: { width: 500, borderRadius: 28, overflow: "hidden" }, bodyStyle: { padding: 28 }, children: [_jsxs(Space, { direction: "vertical", size: 10, style: { width: "100%", marginBottom: 18 }, children: [_jsx("div", { className: "page-hero__eyebrow", style: { background: "rgba(39,84,197,0.08)", color: "var(--app-brand-strong)" }, children: "Campus Assessment System" }), _jsx(Title, { level: 2, style: { marginBottom: 0, color: "var(--app-brand-strong)", fontFamily: "\"Baskerville\", \"Songti SC\", serif" }, children: "AI \u9898\u5E93\u7CFB\u7EDF" }), _jsx(Paragraph, { type: "secondary", style: { marginBottom: 0 }, children: "\u9762\u5411\u6559\u5B66\u573A\u666F\u7684\u4E00\u4F53\u5316\u547D\u9898\u4E0E\u8003\u8BD5\u5E73\u53F0\u3002\u6559\u5E08\u8FDB\u5165\u5DE5\u4F5C\u53F0\uFF0C\u5B66\u751F\u8FDB\u5165\u5728\u7EBF\u8003\u8BD5\u7AEF\u3002" }), _jsx("div", { style: {
                                padding: "12px 14px",
                                borderRadius: 18,
                                background: "linear-gradient(135deg, rgba(39,84,197,0.08) 0%, rgba(39,84,197,0.03) 100%)",
                                border: "1px solid rgba(34,52,79,0.08)",
                            }, children: _jsx(Text, { type: "secondary", children: "\u9ED8\u8BA4\u8D26\u53F7\uFF1Aadmin / admin123\uFF0Cteacher01 / teacher123\uFF0Cstudent01 / student123" }) })] }), errorMessage ? _jsx(Alert, { type: "error", message: errorMessage, showIcon: true, style: { marginBottom: 16 } }) : null, _jsx(Tabs, { defaultActiveKey: "login", style: { marginTop: 8 }, items: [
                        {
                            key: "login",
                            label: "登录",
                            children: (_jsxs(Form, { layout: "vertical", onFinish: handleLogin, autoComplete: "off", children: [_jsx(Form.Item, { name: "username", label: "\u7528\u6237\u540D", rules: [{ required: true, message: "请输入用户名" }], children: _jsx(Input, { prefix: _jsx(UserOutlined, {}), placeholder: "\u8BF7\u8F93\u5165\u7528\u6237\u540D" }) }), _jsx(Form.Item, { name: "password", label: "\u5BC6\u7801", rules: [{ required: true, message: "请输入密码" }], children: _jsx(Input.Password, { prefix: _jsx(LockOutlined, {}), placeholder: "\u8BF7\u8F93\u5165\u5BC6\u7801" }) }), _jsx(Button, { type: "primary", htmlType: "submit", block: true, loading: submitting, children: "\u767B\u5F55" })] })),
                        },
                        {
                            key: "register",
                            label: "注册",
                            children: (_jsxs(Form, { layout: "vertical", onFinish: handleRegister, autoComplete: "off", children: [_jsx(Form.Item, { name: "username", label: "\u7528\u6237\u540D", rules: [{ required: true, message: "请输入用户名" }], children: _jsx(Input, { prefix: _jsx(UserOutlined, {}), placeholder: "\u8BF7\u8F93\u5165\u7528\u6237\u540D" }) }), _jsx(Form.Item, { name: "role", label: "\u89D2\u8272", initialValue: "student", rules: [{ required: true, message: "请选择角色" }], children: _jsx(Select, { options: [
                                                { label: "学生", value: "student" },
                                                { label: "教师", value: "teacher" },
                                            ] }) }), _jsx(Form.Item, { name: "password", label: "\u5BC6\u7801", rules: [
                                            { required: true, message: "请输入密码" },
                                            { min: 6, message: "密码至少 6 位" },
                                        ], children: _jsx(Input.Password, { prefix: _jsx(LockOutlined, {}), placeholder: "\u8BF7\u8BBE\u7F6E\u5BC6\u7801" }) }), _jsx(Form.Item, { name: "confirmPassword", label: "\u786E\u8BA4\u5BC6\u7801", dependencies: ["password"], rules: [
                                            { required: true, message: "请再次输入密码" },
                                            ({ getFieldValue }) => ({
                                                validator(_, value) {
                                                    if (!value || getFieldValue("password") === value) {
                                                        return Promise.resolve();
                                                    }
                                                    return Promise.reject(new Error("两次输入的密码不一致"));
                                                },
                                            }),
                                        ], children: _jsx(Input.Password, { prefix: _jsx(LockOutlined, {}), placeholder: "\u8BF7\u518D\u6B21\u8F93\u5165\u5BC6\u7801" }) }), _jsx(Button, { type: "primary", htmlType: "submit", block: true, loading: submitting, children: "\u6CE8\u518C\u5E76\u767B\u5F55" })] })),
                        },
                    ] })] }) }));
}
