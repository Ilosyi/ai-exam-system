import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Button, Card, Drawer, Empty, Form, Input, InputNumber, Modal, Segmented, Select, Space, Statistic, Table, Tag, Typography, message } from "antd";
import { DeleteOutlined, EditOutlined, PlusOutlined, TeamOutlined, BookOutlined, UserAddOutlined, UserDeleteOutlined } from "@ant-design/icons";
import { batchEditClassStudents, createClass, deleteClass, fetchClasses, fetchClassStudents, fetchStudentExams, updateClass, } from "../api/class";
import { useAuth } from "../hooks/useAuth";
import dayjs from "dayjs";
export function ClassManagePage() {
    const { user } = useAuth();
    const [filters, setFilters] = useState({ keyword: "", page: 1, pageSize: 10 });
    const [data, setData] = useState([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(false);
    const [modalOpen, setModalOpen] = useState(false);
    const [editing, setEditing] = useState(null);
    const [submitting, setSubmitting] = useState(false);
    const [memberDrawerOpen, setMemberDrawerOpen] = useState(false);
    const [activeClass, setActiveClass] = useState(null);
    const [memberFilters, setMemberFilters] = useState({ keyword: "", status: undefined, scope: "all", page: 1, pageSize: 8 });
    const [students, setStudents] = useState([]);
    const [studentsTotal, setStudentsTotal] = useState(0);
    const [studentsLoading, setStudentsLoading] = useState(false);
    const [selectedStudentIds, setSelectedStudentIds] = useState([]);
    const [selectedStudents, setSelectedStudents] = useState([]);
    const [memberSubmitting, setMemberSubmitting] = useState(false);
    const [studentExamModalOpen, setStudentExamModalOpen] = useState(false);
    const [activeStudent, setActiveStudent] = useState(null);
    const [studentExams, setStudentExams] = useState([]);
    const [studentExamsTotal, setStudentExamsTotal] = useState(0);
    const [studentExamLoading, setStudentExamLoading] = useState(false);
    const [studentExamPage, setStudentExamPage] = useState(1);
    const [form] = Form.useForm();
    const searchInputRef = useRef(null);
    const isAdmin = user?.role === "admin";
    const load = useCallback(async () => {
        setLoading(true);
        try {
            const res = await fetchClasses(filters);
            setData(res.data);
            setTotal(res.total);
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "班级列表加载失败");
        }
        finally {
            setLoading(false);
        }
    }, [filters]);
    useEffect(() => {
        void load();
    }, [load]);
    const loadStudents = useCallback(async () => {
        if (!memberDrawerOpen || !activeClass)
            return;
        setStudentsLoading(true);
        try {
            const res = await fetchClassStudents(activeClass.id, memberFilters);
            setStudents(res.data);
            setStudentsTotal(res.total);
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "学生列表加载失败");
        }
        finally {
            setStudentsLoading(false);
        }
    }, [activeClass, memberDrawerOpen, memberFilters, message]);
    useEffect(() => {
        void loadStudents();
    }, [loadStudents]);
    const loadStudentExams = useCallback(async () => {
        if (!studentExamModalOpen || !activeClass || !activeStudent)
            return;
        setStudentExamLoading(true);
        try {
            const res = await fetchStudentExams(activeClass.id, activeStudent.id, { page: studentExamPage, pageSize: 6 });
            setStudentExams(res.data);
            setStudentExamsTotal(res.total);
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "考试记录加载失败");
        }
        finally {
            setStudentExamLoading(false);
        }
    }, [activeClass, activeStudent, studentExamModalOpen, studentExamPage, message]);
    useEffect(() => {
        void loadStudentExams();
    }, [loadStudentExams]);
    const openCreate = () => {
        setEditing(null);
        form.setFieldsValue({ name: "", teacherId: undefined });
        setModalOpen(true);
    };
    const openEdit = (item) => {
        setEditing(item);
        form.setFieldsValue({ name: item.name, teacherId: item.teacherId });
        setModalOpen(true);
    };
    const openMemberDrawer = (item) => {
        setActiveClass(item);
        setSelectedStudentIds([]);
        setSelectedStudents([]);
        setMemberFilters({ keyword: "", status: undefined, scope: "all", page: 1, pageSize: 8 });
        setMemberDrawerOpen(true);
    };
    useEffect(() => {
        if (memberDrawerOpen) {
            searchInputRef.current?.focus();
        }
    }, [memberDrawerOpen]);
    const openStudentExams = (student) => {
        setActiveStudent(student);
        setStudentExamPage(1);
        setStudentExamModalOpen(true);
    };
    const onSubmit = async (values) => {
        setSubmitting(true);
        try {
            if (editing) {
                await updateClass(editing.id, {
                    name: values.name,
                    teacherId: isAdmin ? values.teacherId : undefined,
                });
                message.success("更新成功");
            }
            else {
                await createClass({
                    name: values.name,
                    teacherId: isAdmin ? values.teacherId : undefined,
                });
                message.success("创建成功");
            }
            setModalOpen(false);
            void load();
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "保存失败");
        }
        finally {
            setSubmitting(false);
        }
    };
    const handleDelete = (item) => {
        Modal.confirm({
            title: "确认删除",
            content: `确定删除班级「${item.name}」吗？`,
            centered: true,
            onOk: async () => {
                try {
                    await deleteClass(item.id);
                    message.success("删除成功");
                    void load();
                }
                catch (err) {
                    message.error(err instanceof Error ? err.message : "删除失败");
                }
            },
        });
    };
    const selectedInClassIds = selectedStudents.filter((item) => item.inClass).map((item) => item.id);
    const selectedOutClassIds = selectedStudents.filter((item) => !item.inClass).map((item) => item.id);
    const handleBatchEdit = async (action) => {
        if (!activeClass)
            return;
        const targetIds = action === "add" ? selectedOutClassIds : selectedInClassIds;
        if (targetIds.length === 0) {
            message.warning(action === "add" ? "当前选中学生都已在班级中" : "当前选中学生都不在班级中");
            return;
        }
        setMemberSubmitting(true);
        try {
            await batchEditClassStudents(activeClass.id, action, targetIds);
            message.success(action === "add" ? `已加入 ${targetIds.length} 名学生` : `已移出 ${targetIds.length} 名学生`);
            setSelectedStudentIds([]);
            setSelectedStudents([]);
            await loadStudents();
            await load();
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "成员操作失败");
        }
        finally {
            setMemberSubmitting(false);
        }
    };
    const pagination = useMemo(() => ({
        current: filters.page,
        pageSize: filters.pageSize,
        total,
        onChange: (page, pageSize) => {
            setFilters((prev) => ({ ...prev, page, pageSize: pageSize ?? prev.pageSize }));
        },
    }), [filters.page, filters.pageSize, total]);
    const studentPagination = useMemo(() => ({
        current: memberFilters.page,
        pageSize: memberFilters.pageSize,
        total: studentsTotal,
        onChange: (page, pageSize) => {
            setMemberFilters((prev) => ({ ...prev, page, pageSize: pageSize ?? prev.pageSize }));
        },
    }), [memberFilters.page, memberFilters.pageSize, studentsTotal]);
    const examPagination = useMemo(() => ({
        current: studentExamPage,
        pageSize: 6,
        total: studentExamsTotal,
        onChange: (page) => {
            setStudentExamPage(page);
        },
    }), [studentExamPage, studentExamsTotal]);
    const inClassCount = students.filter((item) => item.inClass).length;
    return (_jsxs("div", { className: "page-shell", children: [_jsxs("section", { className: "page-hero app-fade-up", children: [_jsx("div", { className: "page-hero__eyebrow", children: "Class Governance" }), _jsx("h1", { className: "page-hero__title", children: "\u73ED\u7EA7\u7BA1\u7406" }), _jsx("p", { className: "page-hero__description", children: "\u5728\u6559\u5E08\u89C6\u89D2\u4E0B\u7EDF\u4E00\u7EF4\u62A4\u73ED\u7EA7\u3001\u6210\u5458\u4E0E\u5B66\u751F\u8003\u8BD5\u8BB0\u5F55\uFF0C\u8BA9\u6559\u52A1\u534F\u4F5C\u4ECE\u5355\u70B9\u64CD\u4F5C\u53D8\u6210\u8FDE\u7EED\u5DE5\u4F5C\u6D41\u3002" })] }), _jsxs(Card, { className: "section-card app-fade-up", title: _jsxs(Space, { direction: "vertical", size: 0, children: [_jsx(Typography.Title, { level: 4, style: { margin: 0 }, children: "\u73ED\u7EA7\u7BA1\u7406" }), _jsx(Typography.Text, { type: "secondary", children: "\u6559\u5E08\u89C6\u89D2\u4E0B\u7684\u73ED\u7EA7\u6CBB\u7406\u53F0\uFF0C\u652F\u6301\u6210\u5458\u7EF4\u62A4\u4E0E\u5B66\u751F\u8003\u8BD5\u8FFD\u8E2A" })] }), extra: _jsx(Button, { type: "primary", icon: _jsx(PlusOutlined, {}), onClick: openCreate, children: "\u65B0\u5EFA\u73ED\u7EA7" }), children: [_jsxs("div", { className: "dashboard-toolbar", style: { marginBottom: 16 }, children: [_jsx("div", { className: "dashboard-toolbar__left", children: _jsx(Input, { placeholder: "\u641C\u7D22\u73ED\u7EA7\u540D\u79F0", value: filters.keyword, onChange: (e) => setFilters((prev) => ({ ...prev, keyword: e.target.value, page: 1 })), allowClear: true, style: { width: 240 } }) }), _jsx(Typography.Text, { className: "surface-note", children: "\u53EF\u4ECE\u8FD9\u91CC\u7EE7\u7EED\u8FDB\u5165\u6210\u5458\u7BA1\u7406\u62BD\u5C49\uFF0C\u5B8C\u6210\u5B66\u751F\u5206\u914D\u4E0E\u8003\u8BD5\u8BB0\u5F55\u67E5\u770B\u3002" })] }), _jsx(Table, { rowKey: "id", dataSource: data, loading: loading, pagination: pagination, columns: [
                            { title: "班级ID", dataIndex: "id", key: "id", width: 100 },
                            { title: "班级名称", dataIndex: "name", key: "name" },
                            {
                                title: "教师ID",
                                dataIndex: "teacherId",
                                key: "teacherId",
                                width: 120,
                                render: (teacherId) => _jsx(Tag, { color: "blue", children: teacherId }),
                            },
                            {
                                title: "操作",
                                key: "action",
                                width: 280,
                                render: (_, record) => (_jsxs(Space, { children: [_jsx(Button, { type: "link", size: "small", icon: _jsx(TeamOutlined, {}), onClick: () => openMemberDrawer(record), children: "\u6210\u5458\u7BA1\u7406" }), _jsx(Button, { type: "link", size: "small", icon: _jsx(EditOutlined, {}), onClick: () => openEdit(record), children: "\u7F16\u8F91" }), _jsx(Button, { type: "link", size: "small", danger: true, icon: _jsx(DeleteOutlined, {}), onClick: () => handleDelete(record), children: "\u5220\u9664" })] })),
                            },
                        ] })] }), _jsx(Modal, { title: editing ? "编辑班级" : "新建班级", open: modalOpen, onCancel: () => setModalOpen(false), footer: null, destroyOnClose: true, children: _jsxs(Form, { form: form, layout: "vertical", onFinish: onSubmit, children: [_jsx(Form.Item, { name: "name", label: "\u73ED\u7EA7\u540D\u79F0", rules: [{ required: true, message: "请输入班级名称" }], children: _jsx(Input, { placeholder: "\u8BF7\u8F93\u5165\u73ED\u7EA7\u540D\u79F0", maxLength: 64 }) }), isAdmin ? (_jsx(Form.Item, { name: "teacherId", label: "\u6559\u5E08ID", extra: "\u7BA1\u7406\u5458\u53EF\u6307\u5B9A\u73ED\u7EA7\u6240\u5C5E\u6559\u5E08", children: _jsx(InputNumber, { min: 1, style: { width: "100%" }, placeholder: "\u53EF\u9009" }) })) : null, _jsx(Form.Item, { children: _jsxs(Space, { children: [_jsx(Button, { onClick: () => setModalOpen(false), children: "\u53D6\u6D88" }), _jsx(Button, { type: "primary", htmlType: "submit", loading: submitting, children: "\u4FDD\u5B58" })] }) })] }) }), _jsx(Drawer, { title: _jsxs(Space, { direction: "vertical", size: 0, children: [_jsx(Typography.Text, { style: { fontSize: 18, fontWeight: 700 }, children: activeClass?.name || "班级成员管理" }), _jsx(Typography.Text, { type: "secondary", children: "\u5728\u4E00\u4E2A\u5DE5\u4F5C\u53F0\u91CC\u5B8C\u6210\u6210\u5458\u7EF4\u62A4\u548C\u5B66\u751F\u8003\u8BD5\u8FFD\u8E2A" })] }), placement: "right", width: 860, open: memberDrawerOpen, onClose: () => {
                    setMemberDrawerOpen(false);
                    setActiveClass(null);
                    setSelectedStudentIds([]);
                    setSelectedStudents([]);
                }, extra: _jsxs(Space, { children: [_jsx(Button, { icon: _jsx(UserAddOutlined, {}), type: "primary", ghost: true, disabled: !selectedStudentIds.length, loading: memberSubmitting, onClick: () => void handleBatchEdit("add"), children: "\u52A0\u5165\u73ED\u7EA7" }), _jsx(Button, { icon: _jsx(UserDeleteOutlined, {}), danger: true, disabled: !selectedStudentIds.length, loading: memberSubmitting, onClick: () => void handleBatchEdit("remove"), children: "\u79FB\u51FA\u73ED\u7EA7" })] }), children: _jsxs(Space, { direction: "vertical", size: "large", style: { width: "100%" }, children: [_jsx("div", { style: {
                                padding: 20,
                                borderRadius: 20,
                                background: "linear-gradient(135deg, #122033 0%, #1f3a5f 58%, #3f6aa1 100%)",
                                color: "#f8fbff",
                                boxShadow: "0 22px 48px rgba(16, 36, 62, 0.18)",
                            }, children: _jsxs(Space, { size: "large", style: { width: "100%", justifyContent: "space-between" }, wrap: true, children: [_jsxs(Space, { direction: "vertical", size: 2, children: [_jsx(Typography.Text, { style: { color: "rgba(255,255,255,0.72)", letterSpacing: 1.2 }, children: "CLASS OPERATIONS" }), _jsx(Typography.Title, { level: 3, style: { color: "#fff", margin: 0 }, children: "\u6210\u5458\u6CBB\u7406\u62BD\u5C49" }), _jsx(Typography.Text, { style: { color: "rgba(255,255,255,0.78)" }, children: "\u9009\u4E2D\u4EFB\u610F\u5B66\u751F\u540E\uFF0C\u53EF\u4EE5\u76F4\u63A5\u6279\u91CF\u52A0\u5165\u6216\u79FB\u51FA\u5F53\u524D\u73ED\u7EA7\u3002" })] }), _jsxs(Space, { size: "middle", children: [_jsx(Statistic, { title: _jsx("span", { style: { color: "rgba(255,255,255,0.72)" }, children: "\u5F53\u524D\u9875\u4EBA\u6570" }), value: students.length, valueStyle: { color: "#fff" } }), _jsx(Statistic, { title: _jsx("span", { style: { color: "rgba(255,255,255,0.72)" }, children: "\u5F53\u524D\u9875\u73ED\u5185" }), value: inClassCount, valueStyle: { color: "#fff" } }), _jsx(Statistic, { title: _jsx("span", { style: { color: "rgba(255,255,255,0.72)" }, children: "\u5DF2\u9009\u5B66\u751F" }), value: selectedStudentIds.length, valueStyle: { color: "#fff" } })] })] }) }), _jsx(Card, { bodyStyle: { paddingBottom: 8 }, children: _jsxs(Space, { wrap: true, style: { width: "100%", justifyContent: "space-between" }, children: [_jsxs(Space, { wrap: true, children: [_jsx(Input, { ref: searchInputRef, placeholder: "\u641C\u7D22\u5B66\u751F\u7528\u6237\u540D", value: memberFilters.keyword, onChange: (e) => setMemberFilters((prev) => ({ ...prev, keyword: e.target.value, page: 1 })), allowClear: true, style: { width: 240 } }), _jsx(Select, { allowClear: true, placeholder: "\u72B6\u6001", style: { width: 140 }, value: memberFilters.status, options: [
                                                    { label: "启用", value: "active" },
                                                    { label: "禁用", value: "disabled" },
                                                ], onChange: (value) => setMemberFilters((prev) => ({ ...prev, status: value, page: 1 })) }), _jsx(Segmented, { value: memberFilters.scope, onChange: (value) => setMemberFilters((prev) => ({ ...prev, scope: value, page: 1 })), options: [
                                                    { label: "候选学生池", value: "all" },
                                                    { label: "班内成员", value: "class" },
                                                ] })] }), _jsxs(Space, { children: [_jsxs(Tag, { color: "blue", children: ["\u53EF\u52A0\u5165 ", selectedOutClassIds.length] }), _jsxs(Tag, { color: "volcano", children: ["\u53EF\u79FB\u51FA ", selectedInClassIds.length] })] })] }) }), _jsx(Table, { rowKey: "id", dataSource: students, loading: studentsLoading, pagination: studentPagination, rowSelection: {
                                selectedRowKeys: selectedStudentIds,
                                onChange: (selectedRowKeys, selectedRows) => {
                                    setSelectedStudentIds(selectedRowKeys);
                                    setSelectedStudents(selectedRows);
                                },
                            }, locale: {
                                emptyText: _jsx(Empty, { description: memberFilters.scope === "class" ? "班级下还没有学生" : "没有找到匹配学生" }),
                            }, columns: [
                                {
                                    title: "用户名",
                                    dataIndex: "username",
                                    key: "username",
                                    render: (value, record) => (_jsxs(Space, { direction: "vertical", size: 2, children: [_jsx(Typography.Text, { strong: true, children: value }), _jsxs(Space, { size: 6, children: [_jsx(Tag, { color: record.inClass ? "green" : "default", children: record.inClass ? "已在班级" : "待加入" }), _jsx(Tag, { color: record.status === "active" ? "blue" : "default", children: record.status === "active" ? "启用" : "禁用" })] })] })),
                                },
                                {
                                    title: "主班级",
                                    dataIndex: "classId",
                                    key: "classId",
                                    width: 120,
                                    render: (classId) => (classId ? _jsxs(Tag, { color: "gold", children: ["#", classId] }) : _jsx(Tag, { children: "\u672A\u5206\u914D" })),
                                },
                                {
                                    title: "操作",
                                    key: "action",
                                    width: 140,
                                    render: (_, record) => (_jsx(Button, { type: "link", size: "small", icon: _jsx(BookOutlined, {}), disabled: !record.inClass, onClick: () => openStudentExams(record), children: "\u8003\u8BD5\u8BB0\u5F55" })),
                                },
                            ] })] }) }), _jsx(Modal, { title: activeStudent ? `学生考试记录 · ${activeStudent.username}` : "学生考试记录", open: studentExamModalOpen, onCancel: () => {
                    setStudentExamModalOpen(false);
                    setActiveStudent(null);
                }, footer: null, width: 820, destroyOnClose: true, children: _jsx(Table, { rowKey: "attemptId", dataSource: studentExams, loading: studentExamLoading, pagination: examPagination, locale: { emptyText: _jsx(Empty, { description: "\u6682\u65E0\u8003\u8BD5\u8BB0\u5F55" }) }, columns: [
                        { title: "试卷", dataIndex: "paperTitle", key: "paperTitle" },
                        {
                            title: "状态",
                            dataIndex: "status",
                            key: "status",
                            width: 120,
                            render: (status) => {
                                const color = status === "submitted" ? "green" : status === "timeout" ? "orange" : "blue";
                                const text = status === "submitted" ? "已交卷" : status === "timeout" ? "超时交卷" : "进行中";
                                return _jsx(Tag, { color: color, children: text });
                            },
                        },
                        {
                            title: "得分",
                            dataIndex: "totalScore",
                            key: "totalScore",
                            width: 100,
                            render: (score) => (score ?? "-"),
                        },
                        {
                            title: "开始时间",
                            dataIndex: "startedAt",
                            key: "startedAt",
                            width: 170,
                            render: (value) => dayjs(value).format("YYYY-MM-DD HH:mm"),
                        },
                        {
                            title: "交卷时间",
                            dataIndex: "submittedAt",
                            key: "submittedAt",
                            width: 170,
                            render: (value) => (value ? dayjs(value).format("YYYY-MM-DD HH:mm") : "-"),
                        },
                    ] }) })] }));
}
