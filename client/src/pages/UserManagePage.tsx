import { useCallback, useEffect, useMemo, useState } from "react";
import { Button, Card, Form, Input, Modal, Select, Space, Table, Tag, message } from "antd";
import { DeleteOutlined, EditOutlined } from "@ant-design/icons";
import { deleteUser, fetchUsers, updateUser, type AuthUser, type UpdateUserInput, type UserListFilters } from "../api/auth";
import { fetchClasses, type ClassItem } from "../api/class";

const roleLabel: Record<AuthUser["role"], string> = {
    admin: "管理员",
    teacher: "教师",
    student: "学生",
};

interface UserEditValues {
    role: AuthUser["role"];
    status: "active" | "disabled";
    classId?: number;
}

export function UserManagePage() {
    const [filters, setFilters] = useState<UserListFilters>({ page: 1, pageSize: 10, keyword: "", role: undefined, status: undefined, classId: undefined });
    const [data, setData] = useState<AuthUser[]>([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(false);

    const [classes, setClasses] = useState<ClassItem[]>([]);
    const [classLoading, setClassLoading] = useState(false);

    const [editing, setEditing] = useState<AuthUser | null>(null);
    const [modalOpen, setModalOpen] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const [form] = Form.useForm<UserEditValues>();

    const loadUsers = useCallback(async () => {
        setLoading(true);
        try {
            const res = await fetchUsers(filters);
            setData(res.data);
            setTotal(res.total);
        } catch (err: unknown) {
            message.error(err instanceof Error ? err.message : "用户列表加载失败");
        } finally {
            setLoading(false);
        }
    }, [filters]);

    const loadClasses = useCallback(async () => {
        setClassLoading(true);
        try {
            const res = await fetchClasses({ page: 1, pageSize: 200 });
            setClasses(res.data);
        } catch (err: unknown) {
            message.error(err instanceof Error ? err.message : "班级列表加载失败");
        } finally {
            setClassLoading(false);
        }
    }, []);

    useEffect(() => {
        void loadUsers();
    }, [loadUsers]);

    useEffect(() => {
        void loadClasses();
    }, [loadClasses]);

    const openEdit = (item: AuthUser) => {
        setEditing(item);
        form.setFieldsValue({
            role: item.role,
            status: item.status === "disabled" ? "disabled" : "active",
            classId: item.classId ?? undefined,
        });
        setModalOpen(true);
    };

    const onSubmit = async (values: UserEditValues) => {
        if (!editing) return;
        setSubmitting(true);
        try {
            const payload: UpdateUserInput = {
                role: values.role,
                status: values.status,
                classId: values.classId,
            };
            await updateUser(editing.id, payload);
            message.success("更新成功");
            setModalOpen(false);
            void loadUsers();
        } catch (err: unknown) {
            message.error(err instanceof Error ? err.message : "更新失败");
        } finally {
            setSubmitting(false);
        }
    };

    const handleDelete = (item: AuthUser) => {
        Modal.confirm({
            title: "确认删除",
            content: `确定删除用户「${item.username}」吗？`,
            centered: true,
            onOk: async () => {
                try {
                    await deleteUser(item.id);
                    message.success("删除成功");
                    void loadUsers();
                } catch (err: unknown) {
                    message.error(err instanceof Error ? err.message : "删除失败");
                }
            },
        });
    };

    const classNameMap = useMemo(() => {
        const map = new Map<number, string>();
        classes.forEach((item) => {
            map.set(item.id, item.name);
        });
        return map;
    }, [classes]);

    const pagination = useMemo(
        () => ({
            current: filters.page ?? 1,
            pageSize: filters.pageSize ?? 10,
            total,
            onChange: (page: number, pageSize?: number) => {
                setFilters((prev) => ({ ...prev, page, pageSize: pageSize ?? prev.pageSize }));
            },
        }),
        [filters.page, filters.pageSize, total],
    );

    return (
        <>
            <Card title="用户管理">
                <Space style={{ marginBottom: 16 }} wrap>
                    <Input
                        placeholder="搜索用户名"
                        value={String(filters.keyword ?? "")}
                        onChange={(e) => setFilters((prev) => ({ ...prev, keyword: e.target.value, page: 1 }))}
                        allowClear
                        style={{ width: 220 }}
                    />
                    <Select
                        placeholder="角色"
                        value={filters.role}
                        allowClear
                        style={{ width: 140 }}
                        options={[
                            { label: "管理员", value: "admin" },
                            { label: "教师", value: "teacher" },
                            { label: "学生", value: "student" },
                        ]}
                        onChange={(value) => setFilters((prev) => ({ ...prev, role: value, page: 1 }))}
                    />
                    <Select
                        placeholder="状态"
                        value={filters.status}
                        allowClear
                        style={{ width: 140 }}
                        options={[
                            { label: "启用", value: "active" },
                            { label: "禁用", value: "disabled" },
                        ]}
                        onChange={(value) => setFilters((prev) => ({ ...prev, status: value, page: 1 }))}
                    />
                    <Select
                        placeholder="班级"
                        value={filters.classId}
                        allowClear
                        loading={classLoading}
                        style={{ width: 220 }}
                        options={classes.map((item) => ({ label: item.name, value: item.id }))}
                        onChange={(value) => setFilters((prev) => ({ ...prev, classId: value, page: 1 }))}
                    />
                </Space>

                <Table
                    rowKey="id"
                    dataSource={data}
                    loading={loading}
                    pagination={pagination}
                    columns={[
                        { title: "ID", dataIndex: "id", key: "id", width: 80 },
                        { title: "用户名", dataIndex: "username", key: "username" },
                        {
                            title: "角色",
                            dataIndex: "role",
                            key: "role",
                            width: 120,
                            render: (role: AuthUser["role"]) => <Tag color={role === "admin" ? "red" : role === "teacher" ? "blue" : "green"}>{roleLabel[role]}</Tag>,
                        },
                        {
                            title: "班级",
                            dataIndex: "classId",
                            key: "classId",
                            width: 200,
                            render: (classId: number | null) => {
                                if (!classId) return <Tag>未分配</Tag>;
                                return <Tag color="gold">{classNameMap.get(classId) || `#${classId}`}</Tag>;
                            },
                        },
                        {
                            title: "状态",
                            dataIndex: "status",
                            key: "status",
                            width: 100,
                            render: (status: string) => <Tag color={status === "active" ? "green" : "default"}>{status === "active" ? "启用" : "禁用"}</Tag>,
                        },
                        {
                            title: "操作",
                            key: "action",
                            width: 190,
                            render: (_: unknown, record: AuthUser) => (
                                <Space>
                                    <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEdit(record)}>
                                        编辑
                                    </Button>
                                    <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record)}>
                                        删除
                                    </Button>
                                </Space>
                            ),
                        },
                    ]}
                />
            </Card>

            <Modal
                title={editing ? `编辑用户：${editing.username}` : "编辑用户"}
                open={modalOpen}
                onCancel={() => setModalOpen(false)}
                footer={null}
                destroyOnClose
            >
                <Form form={form} layout="vertical" onFinish={onSubmit}>
                    <Form.Item name="role" label="角色" rules={[{ required: true, message: "请选择角色" }]}>
                        <Select
                            options={[
                                { label: "管理员", value: "admin" },
                                { label: "教师", value: "teacher" },
                                { label: "学生", value: "student" },
                            ]}
                        />
                    </Form.Item>
                    <Form.Item name="status" label="状态" rules={[{ required: true, message: "请选择状态" }]}>
                        <Select
                            options={[
                                { label: "启用", value: "active" },
                                { label: "禁用", value: "disabled" },
                            ]}
                        />
                    </Form.Item>
                    <Form.Item name="classId" label="班级" extra="仅学生需要分配班级，其他角色可留空">
                        <Select
                            allowClear
                            placeholder="请选择班级"
                            loading={classLoading}
                            options={classes.map((item) => ({ value: item.id, label: item.name }))}
                        />
                    </Form.Item>
                    <Form.Item>
                        <Space>
                            <Button onClick={() => setModalOpen(false)}>取消</Button>
                            <Button type="primary" htmlType="submit" loading={submitting}>
                                保存
                            </Button>
                        </Space>
                    </Form.Item>
                </Form>
            </Modal>
        </>
    );
}
