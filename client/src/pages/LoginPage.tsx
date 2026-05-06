import { useMemo, useState } from "react";
import { Alert, Button, Card, Form, Input, Select, Space, Tabs, Typography } from "antd";
import { LockOutlined, UserOutlined } from "@ant-design/icons";
import { useLocation, useNavigate } from "react-router-dom";
import { getDefaultRouteByRole, useAuth } from "../hooks/useAuth";

const { Paragraph, Text, Title } = Typography;

interface LoginFormValues {
  username: string;
  password: string;
}

interface RegisterFormValues extends LoginFormValues {
  role: "teacher" | "student";
  confirmPassword: string;
}

export function LoginPage() {
  const { login, register } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [submitting, setSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  const redirectTo = useMemo(() => {
    const state = location.state as { from?: { pathname?: string } } | null;
    return state?.from?.pathname;
  }, [location.state]);

  const resolveNextPath = (role: "admin" | "teacher" | "student") => {
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

  const handleLogin = async (values: LoginFormValues) => {
    setSubmitting(true);
    setErrorMessage("");
    try {
      const payload = await login(values);
      navigate(resolveNextPath(payload.user.role), { replace: true });
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "登录失败");
    } finally {
      setSubmitting(false);
    }
  };

  const handleRegister = async (values: RegisterFormValues) => {
    setSubmitting(true);
    setErrorMessage("");
    try {
      const payload = await register({
        username: values.username,
        password: values.password,
        role: values.role,
      });
      navigate(resolveNextPath(payload.user.role), { replace: true });
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "注册失败");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: 24,
        background:
          "radial-gradient(circle at left top, rgba(39,84,197,0.12), transparent 22%), radial-gradient(circle at right top, rgba(39,84,197,0.1), transparent 18%), linear-gradient(135deg, #f7f9fc 0%, #e9eff7 52%, #f8fbff 100%)",
      }}
    >
      <Card
        className="glass-card app-fade-up"
        style={{ width: 500, borderRadius: 28, overflow: "hidden" }}
        bodyStyle={{ padding: 28 }}
      >
        <Space direction="vertical" size={10} style={{ width: "100%", marginBottom: 18 }}>
          <div className="page-hero__eyebrow" style={{ background: "rgba(39,84,197,0.08)", color: "var(--app-brand-strong)" }}>
            Campus Assessment System
          </div>
          <Title level={2} style={{ marginBottom: 0, color: "var(--app-brand-strong)", fontFamily: "\"Baskerville\", \"Songti SC\", serif" }}>
            AI 题库系统
          </Title>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            面向教学场景的一体化命题与考试平台。教师进入工作台，学生进入在线考试端。
          </Paragraph>
          <div
            style={{
              padding: "12px 14px",
              borderRadius: 18,
              background: "linear-gradient(135deg, rgba(39,84,197,0.08) 0%, rgba(39,84,197,0.03) 100%)",
              border: "1px solid rgba(34,52,79,0.08)",
            }}
          >
            <Text type="secondary">默认账号：admin / admin123，teacher01 / teacher123，student01 / student123</Text>
          </div>
        </Space>

        {errorMessage ? <Alert type="error" message={errorMessage} showIcon style={{ marginBottom: 16 }} /> : null}

        <Tabs
          defaultActiveKey="login"
          style={{ marginTop: 8 }}
          items={[
            {
              key: "login",
              label: "登录",
              children: (
                <Form layout="vertical" onFinish={handleLogin} autoComplete="off">
                  <Form.Item name="username" label="用户名" rules={[{ required: true, message: "请输入用户名" }]}>
                    <Input prefix={<UserOutlined />} placeholder="请输入用户名" />
                  </Form.Item>
                  <Form.Item name="password" label="密码" rules={[{ required: true, message: "请输入密码" }]}>
                    <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" />
                  </Form.Item>
                  <Button type="primary" htmlType="submit" block loading={submitting}>
                    登录
                  </Button>
                </Form>
              ),
            },
            {
              key: "register",
              label: "注册",
              children: (
                <Form layout="vertical" onFinish={handleRegister} autoComplete="off">
                  <Form.Item name="username" label="用户名" rules={[{ required: true, message: "请输入用户名" }]}>
                    <Input prefix={<UserOutlined />} placeholder="请输入用户名" />
                  </Form.Item>
                  <Form.Item
                    name="role"
                    label="角色"
                    initialValue="student"
                    rules={[{ required: true, message: "请选择角色" }]}
                  >
                    <Select
                      options={[
                        { label: "学生", value: "student" },
                        { label: "教师", value: "teacher" },
                      ]}
                    />
                  </Form.Item>
                  <Form.Item
                    name="password"
                    label="密码"
                    rules={[
                      { required: true, message: "请输入密码" },
                      { min: 6, message: "密码至少 6 位" },
                    ]}
                  >
                    <Input.Password prefix={<LockOutlined />} placeholder="请设置密码" />
                  </Form.Item>
                  <Form.Item
                    name="confirmPassword"
                    label="确认密码"
                    dependencies={["password"]}
                    rules={[
                      { required: true, message: "请再次输入密码" },
                      ({ getFieldValue }) => ({
                        validator(_, value) {
                          if (!value || getFieldValue("password") === value) {
                            return Promise.resolve();
                          }
                          return Promise.reject(new Error("两次输入的密码不一致"));
                        },
                      }),
                    ]}
                  >
                    <Input.Password prefix={<LockOutlined />} placeholder="请再次输入密码" />
                  </Form.Item>
                  <Button type="primary" htmlType="submit" block loading={submitting}>
                    注册并登录
                  </Button>
                </Form>
              ),
            },
          ]}
        />
      </Card>
    </div>
  );
}
