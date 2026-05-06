import { jsx as _jsx, jsxs as _jsxs, Fragment as _Fragment } from "react/jsx-runtime";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Button, Card, Form, Input, Modal, Select, Space, Table, Tag, message } from "antd";
import { DeleteOutlined, EditOutlined } from "@ant-design/icons";
import { deleteUser, fetchUsers, updateUser } from "../api/auth";
import { fetchClasses } from "../api/class";
const roleLabel = {
    admin: "管理员",
    teacher: "教师",
    student: "学生",
};
export function UserManagePage() {
    const [filters, setFilters] = useState({ page: 1, pageSize: 10, keyword: "", role: undefined, status: undefined, classId: undefined });
    const [data, setData] = useState([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(false);
    const [classes, setClasses] = useState([]);
    const [classLoading, setClassLoading] = useState(false);
    const [editing, setEditing] = useState(null);
    const [modalOpen, setModalOpen] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const [form] = Form.useForm();
    const loadUsers = useCallback(async () => {
        setLoading(true);
        try {
            const res = await fetchUsers(filters);
            setData(res.data);
            setTotal(res.total);
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "用户列表加载失败");
        }
        finally {
            setLoading(false);
        }
    }, [filters]);
    const loadClasses = useCallback(async () => {
        setClassLoading(true);
        try {
            const res = await fetchClasses({ page: 1, pageSize: 200 });
            setClasses(res.data);
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "班级列表加载失败");
        }
        finally {
            setClassLoading(false);
        }
    }, []);
    useEffect(() => {
        void loadUsers();
    }, [loadUsers]);
    useEffect(() => {
        void loadClasses();
    }, [loadClasses]);
    const openEdit = (item) => {
        setEditing(item);
        form.setFieldsValue({
            role: item.role,
            status: item.status === "disabled" ? "disabled" : "active",
            classId: item.classId ?? undefined,
        });
        setModalOpen(true);
    };
    const onSubmit = async (values) => {
        if (!editing)
            return;
        setSubmitting(true);
        try {
            const payload = {
                role: values.role,
                status: values.status,
                classId: values.classId,
            };
            await updateUser(editing.id, payload);
            message.success("更新成功");
            setModalOpen(false);
            void loadUsers();
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "更新失败");
        }
        finally {
            setSubmitting(false);
        }
    };
    const handleDelete = (item) => {
        Modal.confirm({
            title: "确认删除",
            content: `确定删除用户「${item.username}」吗？`,
            centered: true,
            onOk: async () => {
                try {
                    await deleteUser(item.id);
                    message.success("删除成功");
                    void loadUsers();
                }
                catch (err) {
                    message.error(err instanceof Error ? err.message : "删除失败");
                }
            },
        });
    };
    const classNameMap = useMemo(() => {
        const map = new Map();
        classes.forEach((item) => {
            map.set(item.id, item.name);
        });
        return map;
    }, [classes]);
    const pagination = useMemo(() => ({
        current: filters.page ?? 1,
        pageSize: filters.pageSize ?? 10,
        total,
        onChange: (page, pageSize) => {
            setFilters((prev) => ({ ...prev, page, pageSize: pageSize ?? prev.pageSize }));
        },
    }), [filters.page, filters.pageSize, total]);
    return (_jsxs(_Fragment, { children: [_jsxs(Card, { title: "\u7528\u6237\u7BA1\u7406", children: [_jsxs(Space, { style: { marginBottom: 16 }, wrap: true, children: [_jsx(Input, { placeholder: "\u641C\u7D22\u7528\u6237\u540D", value: String(filters.keyword ?? ""), onChange: (e) => setFilters((prev) => ({ ...prev, keyword: e.target.value, page: 1 })), allowClear: true, style: { width: 220 } }), _jsx(Select, { placeholder: "\u89D2\u8272", value: filters.role, allowClear: true, style: { width: 140 }, options: [
                                    { label: "管理员", value: "admin" },
                                    { label: "教师", value: "teacher" },
                                    { label: "学生", value: "student" },
                                ], onChange: (value) => setFilters((prev) => ({ ...prev, role: value, page: 1 })) }), _jsx(Select, { placeholder: "\u72B6\u6001", value: filters.status, allowClear: true, style: { width: 140 }, options: [
                                    { label: "启用", value: "active" },
                                    { label: "禁用", value: "disabled" },
                                ], onChange: (value) => setFilters((prev) => ({ ...prev, status: value, page: 1 })) }), _jsx(Select, { placeholder: "\u73ED\u7EA7", value: filters.classId, allowClear: true, loading: classLoading, style: { width: 220 }, options: classes.map((item) => ({ label: item.name, value: item.id })), onChange: (value) => setFilters((prev) => ({ ...prev, classId: value, page: 1 })) })] }), _jsx(Table, { rowKey: "id", dataSource: data, loading: loading, pagination: pagination, columns: [
                            { title: "ID", dataIndex: "id", key: "id", width: 80 },
                            { title: "用户名", dataIndex: "username", key: "username" },
                            {
                                title: "角色",
                                dataIndex: "role",
                                key: "role",
                                width: 120,
                                render: (role) => _jsx(Tag, { color: role === "admin" ? "red" : role === "teacher" ? "blue" : "green", children: roleLabel[role] }),
                            },
                            {
                                title: "班级",
                                dataIndex: "classId",
                                key: "classId",
                                width: 200,
                                render: (classId) => {
                                    if (!classId)
                                        return _jsx(Tag, { children: "\u672A\u5206\u914D" });
                                    return _jsx(Tag, { color: "gold", children: classNameMap.get(classId) || `#${classId}` });
                                },
                            },
                            {
                                title: "状态",
                                dataIndex: "status",
                                key: "status",
                                width: 100,
                                render: (status) => _jsx(Tag, { color: status === "active" ? "green" : "default", children: status === "active" ? "启用" : "禁用" }),
                            },
                            {
                                title: "操作",
                                key: "action",
                                width: 190,
                                render: (_, record) => (_jsxs(Space, { children: [_jsx(Button, { type: "link", size: "small", icon: _jsx(EditOutlined, {}), onClick: () => openEdit(record), children: "\u7F16\u8F91" }), _jsx(Button, { type: "link", size: "small", danger: true, icon: _jsx(DeleteOutlined, {}), onClick: () => handleDelete(record), children: "\u5220\u9664" })] })),
                            },
                        ] })] }), _jsx(Modal, { title: editing ? `编辑用户：${editing.username}` : "编辑用户", open: modalOpen, onCancel: () => setModalOpen(false), footer: null, destroyOnClose: true, children: _jsxs(Form, { form: form, layout: "vertical", onFinish: onSubmit, children: [_jsx(Form.Item, { name: "role", label: "\u89D2\u8272", rules: [{ required: true, message: "请选择角色" }], children: _jsx(Select, { options: [
                                    { label: "管理员", value: "admin" },
                                    { label: "教师", value: "teacher" },
                                    { label: "学生", value: "student" },
                                ] }) }), _jsx(Form.Item, { name: "status", label: "\u72B6\u6001", rules: [{ required: true, message: "请选择状态" }], children: _jsx(Select, { options: [
                                    { label: "启用", value: "active" },
                                    { label: "禁用", value: "disabled" },
                                ] }) }), _jsx(Form.Item, { name: "classId", label: "\u73ED\u7EA7", extra: "\u4EC5\u5B66\u751F\u9700\u8981\u5206\u914D\u73ED\u7EA7\uFF0C\u5176\u4ED6\u89D2\u8272\u53EF\u7559\u7A7A", children: _jsx(Select, { allowClear: true, placeholder: "\u8BF7\u9009\u62E9\u73ED\u7EA7", loading: classLoading, options: classes.map((item) => ({ value: item.id, label: item.name })) }) }), _jsx(Form.Item, { children: _jsxs(Space, { children: [_jsx(Button, { onClick: () => setModalOpen(false), children: "\u53D6\u6D88" }), _jsx(Button, { type: "primary", htmlType: "submit", loading: submitting, children: "\u4FDD\u5B58" })] }) })] }) })] }));
}
