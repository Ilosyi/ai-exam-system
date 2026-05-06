import { Layout, Menu, theme } from "antd";
import { ReadOutlined, DatabaseOutlined, FileTextOutlined, FormOutlined, TeamOutlined, UserOutlined } from "@ant-design/icons";
import { Link, Outlet, useLocation, useNavigate } from "react-router-dom";
import { useState } from "react";
import { Button, Form, Input, Modal, Space, Tag, Typography, message } from "antd";
import { useAuth } from "../hooks/useAuth";
import { changePassword } from "../api/auth";

function buildMenuItems(role?: "admin" | "teacher" | "student") {
  const base = [
    { key: "/notes", label: <Link to="/notes">学习心得</Link>, icon: <ReadOutlined /> },
    { key: "/questions", label: <Link to="/questions">题库管理</Link>, icon: <DatabaseOutlined /> },
    { key: "/papers/generate", label: <Link to="/papers/generate">智能组卷</Link>, icon: <FormOutlined /> },
    { key: "/papers", label: <Link to="/papers">试卷管理</Link>, icon: <FileTextOutlined /> },
    { key: "/classes", label: <Link to="/classes">班级管理</Link>, icon: <TeamOutlined /> },
  ];
  if (role === "admin") {
    base.push({ key: "/users", label: <Link to="/users">用户管理</Link>, icon: <UserOutlined /> });
  }
  return base;
}

function getSelectedKey(pathname: string): string {
  if (pathname.startsWith("/notes")) return "/notes";
  if (pathname.startsWith("/questions")) return "/questions";
  if (pathname.startsWith("/papers/generate")) return "/papers/generate";
  if (pathname.startsWith("/classes")) return "/classes";
  if (pathname.startsWith("/users")) return "/users";
  if (pathname.startsWith("/papers")) return "/papers";
  return "/questions";
}

export function AppLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const [collapsed, setCollapsed] = useState(false);
  const [passwordModalOpen, setPasswordModalOpen] = useState(false);
  const [savingPassword, setSavingPassword] = useState(false);
  const [passwordForm] = Form.useForm<{ oldPassword: string; newPassword: string; confirmPassword: string }>();
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  const handleChangePassword = async (values: { oldPassword: string; newPassword: string; confirmPassword: string }) => {
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
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "修改密码失败");
    } finally {
      setSavingPassword(false);
    }
  };

  return (
    <Layout className="app-shell" style={{ minHeight: "100vh", background: "transparent", alignItems: "stretch" }}>
      <Layout.Sider
        width={220}
        theme="dark"
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        collapsedWidth={68}
        trigger={null}
        style={{
          background: "#031b3c",
          alignSelf: "stretch",
          position: "sticky",
          top: 0,
          height: "100vh",
          overflow: "hidden",
          zIndex: 20,
        }}
      >
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            minHeight: "100%",
            height: "100%",
            padding: collapsed ? "14px 10px 12px" : "14px 0 12px",
            background: "#031b3c",
            color: "#eef4ff",
          }}
        >
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              minHeight: collapsed ? 60 : 72,
              padding: 0,
              marginBottom: 10,
            }}
          >
            <div
              style={{
                width: collapsed ? 46 : 118,
                height: collapsed ? 46 : 54,
              }}
            >
              <img
                src="/kingsoft-logo.png"
                alt="Kingsoft Office"
                style={{
                  width: "100%",
                  height: "100%",
                  objectFit: "contain",
                }}
              />
            </div>
          </div>

          <Menu
            theme="dark"
            items={buildMenuItems(user?.role)}
            selectedKeys={[getSelectedKey(location.pathname)]}
            style={{
              background: "#031b3c",
              borderInlineEnd: "none",
              flex: 1,
              paddingInline: collapsed ? 0 : 8,
            }}
          />

          <div style={{ padding: collapsed ? "0" : "0 8px", marginTop: 12 }}>
            <Button
              type="text"
              block
              style={{
                height: 44,
                color: "#f2f6ff",
                borderRadius: collapsed ? 12 : 0,
                background: "rgba(255,255,255,0.02)",
              }}
              onClick={() => setCollapsed((prev) => !prev)}
            >
              {collapsed ? ">" : "<"}
            </Button>
          </div>
        </div>
      </Layout.Sider>
      <Layout style={{ background: "transparent", position: "relative", zIndex: 1 }}>
        <Layout.Header
          style={{
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
          }}
        >
          <div style={{ flex: 1 }}>
            <Typography.Title level={4} style={{ margin: 0, color: "#fff", fontFamily: "\"Baskerville\", \"Songti SC\", serif" }}>
              华中科技大学 losyi 大作业
            </Typography.Title>
          </div>

          <Space>
            <Tag color="blue" style={{ margin: 0 }}>{user?.role ?? "guest"}</Tag>
            <Typography.Text strong style={{ color: "#f2f6ff" }}>{user?.username ?? "未登录"}</Typography.Text>
            <Button size="small" onClick={() => setPasswordModalOpen(true)}>
              修改密码
            </Button>
            <Button
              size="small"
              onClick={() => {
                logout();
                navigate("/login", { replace: true });
              }}
            >
              退出登录
            </Button>
          </Space>
        </Layout.Header>
        <Layout.Content style={{ margin: "16px 22px 22px" }}>
          <div
            className="panel-surface app-fade-up"
            style={{ padding: 24, background: colorBgContainer, borderRadius: borderRadiusLG, maxWidth: "var(--page-width)", margin: "0 auto" }}
          >
            <Outlet />
          </div>
        </Layout.Content>
      </Layout>

      <Modal
        title="修改密码"
        open={passwordModalOpen}
        onCancel={() => setPasswordModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form form={passwordForm} layout="vertical" onFinish={handleChangePassword}>
          <Form.Item name="oldPassword" label="旧密码" rules={[{ required: true, message: "请输入旧密码" }]}>
            <Input.Password placeholder="请输入旧密码" />
          </Form.Item>
          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[
              { required: true, message: "请输入新密码" },
              { min: 6, message: "密码至少 6 位" },
            ]}
          >
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label="确认新密码"
            rules={[{ required: true, message: "请再次输入新密码" }]}
          >
            <Input.Password placeholder="请再次输入新密码" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button onClick={() => setPasswordModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit" loading={savingPassword}>
                确认修改
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Layout>
  );
}
