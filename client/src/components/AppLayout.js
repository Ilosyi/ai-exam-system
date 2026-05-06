import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { Layout, Menu, theme } from "antd";
import { ReadOutlined, DatabaseOutlined, FileTextOutlined, FormOutlined, TeamOutlined, UserOutlined } from "@ant-design/icons";
import { Link, Outlet, useLocation, useNavigate } from "react-router-dom";
import { useState } from "react";
import { Button, Form, Input, Modal, Space, Tag, Typography, message } from "antd";
import { useAuth } from "../hooks/useAuth";
import { changePassword } from "../api/auth";
function buildMenuItems(role) {
    const base = [
        { key: "/notes", label: _jsx(Link, { to: "/notes", children: "\u5B66\u4E60\u5FC3\u5F97" }), icon: _jsx(ReadOutlined, {}) },
        { key: "/questions", label: _jsx(Link, { to: "/questions", children: "\u9898\u5E93\u7BA1\u7406" }), icon: _jsx(DatabaseOutlined, {}) },
        { key: "/papers/generate", label: _jsx(Link, { to: "/papers/generate", children: "\u667A\u80FD\u7EC4\u5377" }), icon: _jsx(FormOutlined, {}) },
        { key: "/papers", label: _jsx(Link, { to: "/papers", children: "\u8BD5\u5377\u7BA1\u7406" }), icon: _jsx(FileTextOutlined, {}) },
        { key: "/classes", label: _jsx(Link, { to: "/classes", children: "\u73ED\u7EA7\u7BA1\u7406" }), icon: _jsx(TeamOutlined, {}) },
    ];
    if (role === "admin") {
        base.push({ key: "/users", label: _jsx(Link, { to: "/users", children: "\u7528\u6237\u7BA1\u7406" }), icon: _jsx(UserOutlined, {}) });
    }
    return base;
}
function getSelectedKey(pathname) {
    if (pathname.startsWith("/notes"))
        return "/notes";
    if (pathname.startsWith("/questions"))
        return "/questions";
    if (pathname.startsWith("/papers/generate"))
        return "/papers/generate";
    if (pathname.startsWith("/classes"))
        return "/classes";
    if (pathname.startsWith("/users"))
        return "/users";
    if (pathname.startsWith("/papers"))
        return "/papers";
    return "/questions";
}
export function AppLayout() {
    const location = useLocation();
    const navigate = useNavigate();
    const { user, logout } = useAuth();
    const [collapsed, setCollapsed] = useState(false);
    const [passwordModalOpen, setPasswordModalOpen] = useState(false);
    const [savingPassword, setSavingPassword] = useState(false);
    const [passwordForm] = Form.useForm();
    const { token: { colorBgContainer, borderRadiusLG }, } = theme.useToken();
    const handleChangePassword = async (values) => {
        if (values.newPassword !== values.confirmPassword) {
            message.error("两次输入的新密码不一致");
            return;
        }
        setSavingPassword(true);
        try {
            await changePassword({ oldPassword: values.oldPassword, newPassword: values.newPassword });
            message.success("密码修改成功");
            setPasswordModalOpen(false);
            passwordForm.resetFields();
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "修改密码失败");
        }
        finally {
            setSavingPassword(false);
        }
    };
    return (_jsxs(Layout, { className: "app-shell", style: { minHeight: "100vh", background: "transparent", alignItems: "stretch" }, children: [_jsx(Layout.Sider, { width: 220, theme: "dark", collapsible: true, collapsed: collapsed, onCollapse: setCollapsed, collapsedWidth: 68, trigger: null, style: {
                    background: "#031b3c",
                    alignSelf: "stretch",
                    position: "sticky",
                    top: 0,
                    height: "100vh",
                    overflow: "hidden",
                    zIndex: 20,
                }, children: _jsxs("div", { style: {
                        display: "flex",
                        flexDirection: "column",
                        minHeight: "100%",
                        height: "100%",
                        padding: collapsed ? "14px 10px 12px" : "14px 0 12px",
                        background: "#031b3c",
                        color: "#eef4ff",
                    }, children: [_jsx("div", { style: {
                                display: "flex",
                                alignItems: "center",
                                justifyContent: "center",
                                minHeight: collapsed ? 60 : 72,
                                padding: 0,
                                marginBottom: 10,
                            }, children: _jsx("div", { style: {
                                    width: collapsed ? 46 : 118,
                                    height: collapsed ? 46 : 54,
                                }, children: _jsx("img", { src: "/kingsoft-logo.png", alt: "Kingsoft Office", style: {
                                        width: "100%",
                                        height: "100%",
                                        objectFit: "contain",
                                    } }) }) }), _jsx(Menu, { theme: "dark", items: buildMenuItems(user?.role), selectedKeys: [getSelectedKey(location.pathname)], style: {
                                background: "#031b3c",
                                borderInlineEnd: "none",
                                flex: 1,
                                paddingInline: collapsed ? 0 : 8,
                            } }), _jsx("div", { style: { padding: collapsed ? "0" : "0 8px", marginTop: 12 }, children: _jsx(Button, { type: "text", block: true, style: {
                                    height: 44,
                                    color: "#f2f6ff",
                                    borderRadius: collapsed ? 12 : 0,
                                    background: "rgba(255,255,255,0.02)",
                                }, onClick: () => setCollapsed((prev) => !prev), children: collapsed ? ">" : "<" }) })] }) }), _jsxs(Layout, { style: { background: "transparent", position: "relative", zIndex: 1 }, children: [_jsxs(Layout.Header, { style: {
                            background: "#031b3c",
                            color: "#fff",
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "space-between",
                            gap: 16,
                            padding: "0 22px",
                            height: 68,
                            position: "sticky",
                            top: 0,
                            zIndex: 15,
                        }, children: [_jsx("div", { style: { flex: 1 }, children: _jsx(Typography.Title, { level: 4, style: { margin: 0, color: "#fff", fontFamily: "\"Baskerville\", \"Songti SC\", serif" }, children: "\u534E\u4E2D\u79D1\u6280\u5927\u5B66 \u67EF\u4FCA\u7FD4 \u5927\u4F5C\u4E1A" }) }), _jsxs(Space, { children: [_jsx(Tag, { color: "blue", style: { margin: 0 }, children: user?.role ?? "guest" }), _jsx(Typography.Text, { strong: true, style: { color: "#f2f6ff" }, children: user?.username ?? "未登录" }), _jsx(Button, { size: "small", onClick: () => setPasswordModalOpen(true), children: "\u4FEE\u6539\u5BC6\u7801" }), _jsx(Button, { size: "small", onClick: () => {
                                            logout();
                                            navigate("/login", { replace: true });
                                        }, children: "\u9000\u51FA\u767B\u5F55" })] })] }), _jsx(Layout.Content, { style: { margin: "16px 22px 22px" }, children: _jsx("div", { className: "panel-surface app-fade-up", style: { padding: 24, background: colorBgContainer, borderRadius: borderRadiusLG, maxWidth: "var(--page-width)", margin: "0 auto" }, children: _jsx(Outlet, {}) }) })] }), _jsx(Modal, { title: "\u4FEE\u6539\u5BC6\u7801", open: passwordModalOpen, onCancel: () => setPasswordModalOpen(false), footer: null, destroyOnClose: true, children: _jsxs(Form, { form: passwordForm, layout: "vertical", onFinish: handleChangePassword, children: [_jsx(Form.Item, { name: "oldPassword", label: "\u65E7\u5BC6\u7801", rules: [{ required: true, message: "请输入旧密码" }], children: _jsx(Input.Password, { placeholder: "\u8BF7\u8F93\u5165\u65E7\u5BC6\u7801" }) }), _jsx(Form.Item, { name: "newPassword", label: "\u65B0\u5BC6\u7801", rules: [
                                { required: true, message: "请输入新密码" },
                                { min: 6, message: "密码至少 6 位" },
                            ], children: _jsx(Input.Password, { placeholder: "\u8BF7\u8F93\u5165\u65B0\u5BC6\u7801" }) }), _jsx(Form.Item, { name: "confirmPassword", label: "\u786E\u8BA4\u65B0\u5BC6\u7801", rules: [{ required: true, message: "请再次输入新密码" }], children: _jsx(Input.Password, { placeholder: "\u8BF7\u518D\u6B21\u8F93\u5165\u65B0\u5BC6\u7801" }) }), _jsx(Form.Item, { children: _jsxs(Space, { children: [_jsx(Button, { onClick: () => setPasswordModalOpen(false), children: "\u53D6\u6D88" }), _jsx(Button, { type: "primary", htmlType: "submit", loading: savingPassword, children: "\u786E\u8BA4\u4FEE\u6539" })] }) })] }) })] }));
}
