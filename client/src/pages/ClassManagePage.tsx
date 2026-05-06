import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Button, Card, Drawer, Empty, Form, Input, InputNumber, type InputRef, Modal, Segmented, Select, Space, Statistic, Table, Tag, Typography, message } from "antd";
import { DeleteOutlined, EditOutlined, PlusOutlined, TeamOutlined, BookOutlined, UserAddOutlined, UserDeleteOutlined } from "@ant-design/icons";
import {
    batchEditClassStudents,
    createClass,
    deleteClass,
    fetchClasses,
    fetchClassStudents,
    fetchStudentExams,
    updateClass,
    type ClassItem,
    type ClassStudent,
    type StudentExamRecord,
} from "../api/class";
import { useAuth } from "../hooks/useAuth";
import dayjs from "dayjs";

interface ClassFormValues {
    name: string;
    teacherId?: number;
}

interface MemberFilters {
    keyword: string;
    status?: string;
    scope: "class" | "all";
    page: number;
    pageSize: number;
}

export function ClassManagePage() {
    const { user } = useAuth();
    const [filters, setFilters] = useState({ keyword: "", page: 1, pageSize: 10 });
    const [data, setData] = useState<ClassItem[]>([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(false);
    const [modalOpen, setModalOpen] = useState(false);
    const [editing, setEditing] = useState<ClassItem | null>(null);
    const [submitting, setSubmitting] = useState(false);
    const [memberDrawerOpen, setMemberDrawerOpen] = useState(false);
    const [activeClass, setActiveClass] = useState<ClassItem | null>(null);
    const [memberFilters, setMemberFilters] = useState<MemberFilters>({ keyword: "", status: undefined, scope: "all", page: 1, pageSize: 8 });
    const [students, setStudents] = useState<ClassStudent[]>([]);
    const [studentsTotal, setStudentsTotal] = useState(0);
    const [studentsLoading, setStudentsLoading] = useState(false);
    const [selectedStudentIds, setSelectedStudentIds] = useState<number[]>([]);
    const [selectedStudents, setSelectedStudents] = useState<ClassStudent[]>([]);
    const [memberSubmitting, setMemberSubmitting] = useState(false);
    const [studentExamModalOpen, setStudentExamModalOpen] = useState(false);
    const [activeStudent, setActiveStudent] = useState<ClassStudent | null>(null);
    const [studentExams, setStudentExams] = useState<StudentExamRecord[]>([]);
    const [studentExamsTotal, setStudentExamsTotal] = useState(0);
    const [studentExamLoading, setStudentExamLoading] = useState(false);
    const [studentExamPage, setStudentExamPage] = useState(1);
    const [form] = Form.useForm<ClassFormValues>();
    const searchInputRef = useRef<InputRef>(null);

    const isAdmin = user?.role === "admin";

    const load = useCallback(async () => {
        setLoading(true);
        try {
            const res = await fetchClasses(filters);
            setData(res.data);
            setTotal(res.total);
        } catch (err: unknown) {
            message.error(err instanceof Error ? err.message : "班级列表加载失败");
        } finally {
            setLoading(false);
        }
    }, [filters]);

    useEffect(() => {
        void load();
    }, [load]);

    const loadStudents = useCallback(async () => {
        if (!memberDrawerOpen || !activeClass) return;
        setStudentsLoading(true);
        try {
            const res = await fetchClassStudents(activeClass.id, memberFilters);
            setStudents(res.data);
            setStudentsTotal(res.total);
        } catch (err: unknown) {
            message.error(err instanceof Error ? err.message : "学生列表加载失败");
        } finally {
            setStudentsLoading(false);
        }
    }, [activeClass, memberDrawerOpen, memberFilters, message]);

    useEffect(() => {
        void loadStudents();
    }, [loadStudents]);

    const loadStudentExams = useCallback(async () => {
        if (!studentExamModalOpen || !activeClass || !activeStudent) return;
        setStudentExamLoading(true);
        try {
            const res = await fetchStudentExams(activeClass.id, activeStudent.id, { page: studentExamPage, pageSize: 6 });
            setStudentExams(res.data);
            setStudentExamsTotal(res.total);
        } catch (err: unknown) {
            message.error(err instanceof Error ? err.message : "考试记录加载失败");
        } finally {
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

    const openEdit = (item: ClassItem) => {
        setEditing(item);
        form.setFieldsValue({ name: item.name, teacherId: item.teacherId });
        setModalOpen(true);
    };

    const openMemberDrawer = (item: ClassItem) => {
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

    const openStudentExams = (student: ClassStudent) => {
        setActiveStudent(student);
        setStudentExamPage(1);
        setStudentExamModalOpen(true);
    };

    const onSubmit = async (values: ClassFormValues) => {
        setSubmitting(true);
        try {
            if (editing) {
                await updateClass(editing.id, {
                    name: values.name,
                    teacherId: isAdmin ? values.teacherId : undefined,
                });
                message.success("更新成功");
            } else {
                await createClass({
                    name: values.name,
                    teacherId: isAdmin ? values.teacherId : undefined,
                });
                message.success("创建成功");
            }
            setModalOpen(false);
            void load();
        } catch (err: unknown) {
            message.error(err instanceof Error ? err.message : "保存失败");
        } finally {
            setSubmitting(false);
        }
    };

    const handleDelete = (item: ClassItem) => {
        Modal.confirm({
            title: "确认删除",
            content: `确定删除班级「${item.name}」吗？`,
            centered: true,
            onOk: async () => {
                try {
                    await deleteClass(item.id);
                    message.success("删除成功");
                    void load();
                } catch (err: unknown) {
                    message.error(err instanceof Error ? err.message : "删除失败");
                }
            },
        });
    };

    const selectedInClassIds = selectedStudents.filter((item) => item.inClass).map((item) => item.id);
    const selectedOutClassIds = selectedStudents.filter((item) => !item.inClass).map((item) => item.id);

    const handleBatchEdit = async (action: "add" | "remove") => {
        if (!activeClass) return;
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
        } catch (err: unknown) {
            message.error(err instanceof Error ? err.message : "成员操作失败");
        } finally {
            setMemberSubmitting(false);
        }
    };

    const pagination = useMemo(
        () => ({
            current: filters.page,
            pageSize: filters.pageSize,
            total,
            onChange: (page: number, pageSize?: number) => {
                setFilters((prev) => ({ ...prev, page, pageSize: pageSize ?? prev.pageSize }));
            },
        }),
        [filters.page, filters.pageSize, total],
    );

    const studentPagination = useMemo(
        () => ({
            current: memberFilters.page,
            pageSize: memberFilters.pageSize,
            total: studentsTotal,
            onChange: (page: number, pageSize?: number) => {
                setMemberFilters((prev) => ({ ...prev, page, pageSize: pageSize ?? prev.pageSize }));
            },
        }),
        [memberFilters.page, memberFilters.pageSize, studentsTotal],
    );

    const examPagination = useMemo(
        () => ({
            current: studentExamPage,
            pageSize: 6,
            total: studentExamsTotal,
            onChange: (page: number) => {
                setStudentExamPage(page);
            },
        }),
        [studentExamPage, studentExamsTotal],
    );

    const inClassCount = students.filter((item) => item.inClass).length;

    return (
        <div className="page-shell">
            <section className="page-hero app-fade-up">
                <div className="page-hero__eyebrow">Class Governance</div>
                <h1 className="page-hero__title">班级管理</h1>
                <p className="page-hero__description">
                    在教师视角下统一维护班级、成员与学生考试记录，让教务协作从单点操作变成连续工作流。
                </p>
            </section>

            <Card
                className="section-card app-fade-up"
                title={
                    <Space direction="vertical" size={0}>
                        <Typography.Title level={4} style={{ margin: 0 }}>
                            班级管理
                        </Typography.Title>
                        <Typography.Text type="secondary">
                            教师视角下的班级治理台，支持成员维护与学生考试追踪
                        </Typography.Text>
                    </Space>
                }
                extra={
                    <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
                        新建班级
                    </Button>
                }
            >
                <div className="dashboard-toolbar" style={{ marginBottom: 16 }}>
                    <div className="dashboard-toolbar__left">
                        <Input
                            placeholder="搜索班级名称"
                            value={filters.keyword}
                            onChange={(e) => setFilters((prev) => ({ ...prev, keyword: e.target.value, page: 1 }))}
                            allowClear
                            style={{ width: 240 }}
                        />
                    </div>
                    <Typography.Text className="surface-note">
                        可从这里继续进入成员管理抽屉，完成学生分配与考试记录查看。
                    </Typography.Text>
                </div>

                <Table
                    rowKey="id"
                    dataSource={data}
                    loading={loading}
                    pagination={pagination}
                    columns={[
                        { title: "班级ID", dataIndex: "id", key: "id", width: 100 },
                        { title: "班级名称", dataIndex: "name", key: "name" },
                        {
                            title: "教师ID",
                            dataIndex: "teacherId",
                            key: "teacherId",
                            width: 120,
                            render: (teacherId: number) => <Tag color="blue">{teacherId}</Tag>,
                        },
                        {
                            title: "操作",
                            key: "action",
                            width: 280,
                            render: (_: unknown, record: ClassItem) => (
                                <Space>
                                    <Button type="link" size="small" icon={<TeamOutlined />} onClick={() => openMemberDrawer(record)}>
                                        成员管理
                                    </Button>
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
                title={editing ? "编辑班级" : "新建班级"}
                open={modalOpen}
                onCancel={() => setModalOpen(false)}
                footer={null}
                destroyOnClose
            >
                <Form form={form} layout="vertical" onFinish={onSubmit}>
                    <Form.Item name="name" label="班级名称" rules={[{ required: true, message: "请输入班级名称" }]}>
                        <Input placeholder="请输入班级名称" maxLength={64} />
                    </Form.Item>
                    {isAdmin ? (
                        <Form.Item name="teacherId" label="教师ID" extra="管理员可指定班级所属教师">
                            <InputNumber min={1} style={{ width: "100%" }} placeholder="可选" />
                        </Form.Item>
                    ) : null}
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

            <Drawer
                title={
                    <Space direction="vertical" size={0}>
                        <Typography.Text style={{ fontSize: 18, fontWeight: 700 }}>
                            {activeClass?.name || "班级成员管理"}
                        </Typography.Text>
                        <Typography.Text type="secondary">
                            在一个工作台里完成成员维护和学生考试追踪
                        </Typography.Text>
                    </Space>
                }
                placement="right"
                width={860}
                open={memberDrawerOpen}
                onClose={() => {
                    setMemberDrawerOpen(false);
                    setActiveClass(null);
                    setSelectedStudentIds([]);
                    setSelectedStudents([]);
                }}
                extra={
                    <Space>
                        <Button
                            icon={<UserAddOutlined />}
                            type="primary"
                            ghost
                            disabled={!selectedStudentIds.length}
                            loading={memberSubmitting}
                            onClick={() => void handleBatchEdit("add")}
                        >
                            加入班级
                        </Button>
                        <Button
                            icon={<UserDeleteOutlined />}
                            danger
                            disabled={!selectedStudentIds.length}
                            loading={memberSubmitting}
                            onClick={() => void handleBatchEdit("remove")}
                        >
                            移出班级
                        </Button>
                    </Space>
                }
            >
                <Space direction="vertical" size="large" style={{ width: "100%" }}>
                    <div
                        style={{
                            padding: 20,
                            borderRadius: 20,
                            background: "linear-gradient(135deg, #122033 0%, #1f3a5f 58%, #3f6aa1 100%)",
                            color: "#f8fbff",
                            boxShadow: "0 22px 48px rgba(16, 36, 62, 0.18)",
                        }}
                    >
                        <Space size="large" style={{ width: "100%", justifyContent: "space-between" }} wrap>
                            <Space direction="vertical" size={2}>
                                <Typography.Text style={{ color: "rgba(255,255,255,0.72)", letterSpacing: 1.2 }}>
                                    CLASS OPERATIONS
                                </Typography.Text>
                                <Typography.Title level={3} style={{ color: "#fff", margin: 0 }}>
                                    成员治理抽屉
                                </Typography.Title>
                                <Typography.Text style={{ color: "rgba(255,255,255,0.78)" }}>
                                    选中任意学生后，可以直接批量加入或移出当前班级。
                                </Typography.Text>
                            </Space>
                            <Space size="middle">
                                <Statistic title={<span style={{ color: "rgba(255,255,255,0.72)" }}>当前页人数</span>} value={students.length} valueStyle={{ color: "#fff" }} />
                                <Statistic title={<span style={{ color: "rgba(255,255,255,0.72)" }}>当前页班内</span>} value={inClassCount} valueStyle={{ color: "#fff" }} />
                                <Statistic title={<span style={{ color: "rgba(255,255,255,0.72)" }}>已选学生</span>} value={selectedStudentIds.length} valueStyle={{ color: "#fff" }} />
                            </Space>
                        </Space>
                    </div>

                    <Card bodyStyle={{ paddingBottom: 8 }}>
                        <Space wrap style={{ width: "100%", justifyContent: "space-between" }}>
                            <Space wrap>
                                <Input
                                    ref={searchInputRef}
                                    placeholder="搜索学生用户名"
                                    value={memberFilters.keyword}
                                    onChange={(e) => setMemberFilters((prev) => ({ ...prev, keyword: e.target.value, page: 1 }))}
                                    allowClear
                                    style={{ width: 240 }}
                                />
                                <Select
                                    allowClear
                                    placeholder="状态"
                                    style={{ width: 140 }}
                                    value={memberFilters.status}
                                    options={[
                                        { label: "启用", value: "active" },
                                        { label: "禁用", value: "disabled" },
                                    ]}
                                    onChange={(value) => setMemberFilters((prev) => ({ ...prev, status: value, page: 1 }))}
                                />
                                <Segmented
                                    value={memberFilters.scope}
                                    onChange={(value) => setMemberFilters((prev) => ({ ...prev, scope: value as "class" | "all", page: 1 }))}
                                    options={[
                                        { label: "候选学生池", value: "all" },
                                        { label: "班内成员", value: "class" },
                                    ]}
                                />
                            </Space>
                            <Space>
                                <Tag color="blue">可加入 {selectedOutClassIds.length}</Tag>
                                <Tag color="volcano">可移出 {selectedInClassIds.length}</Tag>
                            </Space>
                        </Space>
                    </Card>

                    <Table
                        rowKey="id"
                        dataSource={students}
                        loading={studentsLoading}
                        pagination={studentPagination}
                        rowSelection={{
                            selectedRowKeys: selectedStudentIds,
                            onChange: (selectedRowKeys, selectedRows) => {
                                setSelectedStudentIds(selectedRowKeys as number[]);
                                setSelectedStudents(selectedRows as ClassStudent[]);
                            },
                        }}
                        locale={{
                            emptyText: <Empty description={memberFilters.scope === "class" ? "班级下还没有学生" : "没有找到匹配学生"} />,
                        }}
                        columns={[
                            {
                                title: "用户名",
                                dataIndex: "username",
                                key: "username",
                                render: (value: string, record: ClassStudent) => (
                                    <Space direction="vertical" size={2}>
                                        <Typography.Text strong>{value}</Typography.Text>
                                        <Space size={6}>
                                            <Tag color={record.inClass ? "green" : "default"}>{record.inClass ? "已在班级" : "待加入"}</Tag>
                                            <Tag color={record.status === "active" ? "blue" : "default"}>{record.status === "active" ? "启用" : "禁用"}</Tag>
                                        </Space>
                                    </Space>
                                ),
                            },
                            {
                                title: "主班级",
                                dataIndex: "classId",
                                key: "classId",
                                width: 120,
                                render: (classId?: number) => (classId ? <Tag color="gold">#{classId}</Tag> : <Tag>未分配</Tag>),
                            },
                            {
                                title: "操作",
                                key: "action",
                                width: 140,
                                render: (_: unknown, record: ClassStudent) => (
                                    <Button
                                        type="link"
                                        size="small"
                                        icon={<BookOutlined />}
                                        disabled={!record.inClass}
                                        onClick={() => openStudentExams(record)}
                                    >
                                        考试记录
                                    </Button>
                                ),
                            },
                        ]}
                    />
                </Space>
            </Drawer>

            <Modal
                title={activeStudent ? `学生考试记录 · ${activeStudent.username}` : "学生考试记录"}
                open={studentExamModalOpen}
                onCancel={() => {
                    setStudentExamModalOpen(false);
                    setActiveStudent(null);
                }}
                footer={null}
                width={820}
                destroyOnClose
            >
                <Table
                    rowKey="attemptId"
                    dataSource={studentExams}
                    loading={studentExamLoading}
                    pagination={examPagination}
                    locale={{ emptyText: <Empty description="暂无考试记录" /> }}
                    columns={[
                        { title: "试卷", dataIndex: "paperTitle", key: "paperTitle" },
                        {
                            title: "状态",
                            dataIndex: "status",
                            key: "status",
                            width: 120,
                            render: (status: string) => {
                                const color = status === "submitted" ? "green" : status === "timeout" ? "orange" : "blue";
                                const text = status === "submitted" ? "已交卷" : status === "timeout" ? "超时交卷" : "进行中";
                                return <Tag color={color}>{text}</Tag>;
                            },
                        },
                        {
                            title: "得分",
                            dataIndex: "totalScore",
                            key: "totalScore",
                            width: 100,
                            render: (score?: number) => (score ?? "-"),
                        },
                        {
                            title: "开始时间",
                            dataIndex: "startedAt",
                            key: "startedAt",
                            width: 170,
                            render: (value: string) => dayjs(value).format("YYYY-MM-DD HH:mm"),
                        },
                        {
                            title: "交卷时间",
                            dataIndex: "submittedAt",
                            key: "submittedAt",
                            width: 170,
                            render: (value?: string) => (value ? dayjs(value).format("YYYY-MM-DD HH:mm") : "-"),
                        },
                    ]}
                />
            </Modal>
        </div>
    );
}
