import { useMemo, useRef, useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeHighlight from "rehype-highlight";
import { Button, Card, Col, Drawer, Empty, Form, Input, InputNumber, Modal, Row, Space, Spin, Tree, Typography, message } from "antd";
import { DeleteOutlined, EyeOutlined, FileMarkdownOutlined, FolderAddOutlined, PlusOutlined, SaveOutlined } from "@ant-design/icons";
import type { DataNode } from "antd/es/tree";
import {
  createDocument,
  createDocumentCourse,
  deleteDocument,
  deleteDocumentCourse,
  fetchDocumentDetail,
  updateDocument,
  updateDocumentCourse,
  type CourseInput,
  type DocumentInput,
} from "../api/document";
import { useDocumentCourses } from "../hooks/useDocuments";

type Selection =
  | { type: "course"; courseId?: string; isNew: boolean }
  | { type: "doc"; courseId: string; docId?: string; isNew: boolean }
  | null;

const courseKey = (courseId: string) => `course:${encodeURIComponent(courseId)}`;
const docKey = (courseId: string, docId: string) => `doc:${encodeURIComponent(courseId)}:${encodeURIComponent(docId)}`;

export function DocumentManagePage() {
  const { courses, loading, reload } = useDocumentCourses();
  const [selection, setSelection] = useState<Selection>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [previewOpen, setPreviewOpen] = useState(false);
  const detailRequestIdRef = useRef(0);
  const [courseForm] = Form.useForm<CourseInput>();
  const [docForm] = Form.useForm<DocumentInput>();
  const previewMarkdown = Form.useWatch("markdown", docForm) ?? "";

  const treeData = useMemo<DataNode[]>(
    () =>
      courses.map((course) => ({
        key: courseKey(course.id),
        title: course.title,
        icon: <FolderAddOutlined />,
        children: course.documents.map((doc) => ({
          key: docKey(course.id, doc.id),
          title: doc.title,
          icon: <FileMarkdownOutlined />,
        })),
      })),
    [courses],
  );

  const selectedTreeKeys = useMemo(() => {
    if (!selection || selection.isNew) return [];
    if (selection.type === "course" && selection.courseId) return [courseKey(selection.courseId)];
    if (selection.type === "doc" && selection.docId) return [docKey(selection.courseId, selection.docId)];
    return [];
  }, [selection]);

  const selectedCourse = useMemo(() => {
    if (!selection?.courseId) return null;
    return courses.find((course) => course.id === selection.courseId) ?? null;
  }, [courses, selection]);

  const currentCourseId = selection?.type === "doc" ? selection.courseId : selection?.courseId;

  const cancelDetailLoad = () => {
    detailRequestIdRef.current += 1;
    setDetailLoading(false);
  };

  const handleCreateCourse = () => {
    cancelDetailLoad();
    setSelection({ type: "course", isNew: true });
    courseForm.setFieldsValue({ id: "", title: "", description: "", order: courses.length + 1 });
  };

  const handleCreateDocument = () => {
    const targetCourseId = currentCourseId ?? courses[0]?.id;
    if (!targetCourseId) {
      message.warning("请先创建课程，再新增文档");
      return;
    }
    cancelDetailLoad();
    setSelection({ type: "doc", courseId: targetCourseId, isNew: true });
    docForm.setFieldsValue({ id: "", title: "", order: (courses.find((course) => course.id === targetCourseId)?.documents.length ?? 0) + 1, markdown: "" });
  };

  const handleSelect = async (_keys: React.Key[], info: { node: DataNode }) => {
    const key = String(info.node.key);
    if (key.startsWith("course:")) {
      const courseId = decodeURIComponent(key.replace(/^course:/, ""));
      const course = courses.find((item) => item.id === courseId);
      if (!course) return;
      cancelDetailLoad();
      setSelection({ type: "course", courseId, isNew: false });
      courseForm.setFieldsValue({
        id: course.id,
        title: course.title,
        description: course.description,
        order: course.order,
      });
      return;
    }

    if (key.startsWith("doc:")) {
      const [, encodedCourseId, encodedDocId] = key.split(":");
      const courseId = decodeURIComponent(encodedCourseId);
      const docId = decodeURIComponent(encodedDocId);
      const course = courses.find((item) => item.id === courseId);
      const doc = course?.documents.find((item) => item.id === docId);
      if (!course || !doc) return;

      const requestId = detailRequestIdRef.current + 1;
      detailRequestIdRef.current = requestId;
      setSelection({ type: "doc", courseId, docId, isNew: false });
      docForm.setFieldsValue({ id: doc.id, title: doc.title, order: doc.order, markdown: "" });
      setDetailLoading(true);
      try {
        const res = await fetchDocumentDetail(courseId, docId);
        if (detailRequestIdRef.current === requestId) {
          docForm.setFieldsValue(res.data);
        }
      } catch (err: unknown) {
        if (detailRequestIdRef.current === requestId) {
          message.error(err instanceof Error ? err.message : "文档详情加载失败");
        }
      } finally {
        if (detailRequestIdRef.current === requestId) {
          setDetailLoading(false);
        }
      }
    }
  };

  const handleSaveCourse = async (values: CourseInput) => {
    setSaving(true);
    try {
      if (selection?.type === "course" && !selection.isNew && selection.courseId) {
        await updateDocumentCourse(selection.courseId, values);
        message.success("课程保存成功");
        await reload();
        setSelection({ type: "course", courseId: selection.courseId, isNew: false });
      } else {
        await createDocumentCourse(values);
        message.success("课程创建成功");
        await reload();
        setSelection({ type: "course", courseId: values.id, isNew: false });
      }
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "课程保存失败");
    } finally {
      setSaving(false);
    }
  };

  const handleSaveDocument = async (values: DocumentInput) => {
    if (selection?.type !== "doc") return;
    setSaving(true);
    try {
      if (!selection.isNew && selection.docId) {
        await updateDocument(selection.courseId, selection.docId, values);
        message.success("文档保存成功");
        await reload();
        setSelection({ type: "doc", courseId: selection.courseId, docId: selection.docId, isNew: false });
      } else {
        await createDocument(selection.courseId, values);
        message.success("文档创建成功");
        await reload();
        setSelection({ type: "doc", courseId: selection.courseId, docId: values.id, isNew: false });
      }
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "文档保存失败");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = () => {
    if (!selection || selection.isNew) return;
    if (selection.type === "course" && selection.courseId) {
      const courseTitle = selectedCourse?.title ?? selection.courseId;
      Modal.confirm({
        title: "确认删除课程",
        content: `删除课程「${courseTitle}」会同时删除该课程下所有 Markdown 文档，确定继续吗？`,
        okText: "确认删除",
        okButtonProps: { danger: true },
        centered: true,
        onOk: async () => {
          try {
            await deleteDocumentCourse(selection.courseId!);
            message.success("课程删除成功");
            setSelection(null);
            await reload();
          } catch (err: unknown) {
            message.error(err instanceof Error ? err.message : "课程删除失败");
          }
        },
      });
      return;
    }

    if (selection.type === "doc" && selection.docId) {
      const docTitle = docForm.getFieldValue("title") || selection.docId;
      Modal.confirm({
        title: "确认删除文档",
        content: `删除文档「${docTitle}」后不可恢复，确定继续吗？`,
        okText: "确认删除",
        okButtonProps: { danger: true },
        centered: true,
        onOk: async () => {
          try {
            await deleteDocument(selection.courseId, selection.docId!);
            message.success("文档删除成功");
            setSelection({ type: "course", courseId: selection.courseId, isNew: false });
            await reload();
          } catch (err: unknown) {
            message.error(err instanceof Error ? err.message : "文档删除失败");
          }
        },
      });
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <div style={{ display: "flex", justifyContent: "space-between", gap: 16, alignItems: "flex-start" }}>
        <div>
          <Typography.Title level={3} style={{ marginTop: 0, marginBottom: 8 }}>
            文档管理
          </Typography.Title>
          <Typography.Text type="secondary">按课程维护 Markdown 课件，学生端会在首页同步展示。</Typography.Text>
        </div>
        <Space wrap>
          <Button icon={<FolderAddOutlined />} onClick={handleCreateCourse}>
            新增课程
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreateDocument}>
            新增文档
          </Button>
        </Space>
      </div>

      <Row gutter={16}>
        <Col xs={24} lg={8}>
          <Card title="课程 / 文档" bodyStyle={{ minHeight: 520 }}>
            <Spin spinning={loading}>
              {treeData.length > 0 ? (
                <Tree
                  showIcon
                  blockNode
                  defaultExpandAll
                  treeData={treeData}
                  selectedKeys={selectedTreeKeys}
                  onSelect={handleSelect}
                />
              ) : (
                <Empty description="暂无课程文档" />
              )}
            </Spin>
          </Card>
        </Col>
        <Col xs={24} lg={16}>
          <Card
            title={selection?.type === "doc" ? "编辑文档" : "编辑课程"}
            extra={
              <Space>
                {selection?.type === "doc" && (
                  <Button icon={<EyeOutlined />} onClick={() => setPreviewOpen(true)}>
                    文档预览
                  </Button>
                )}
                {selection && !selection.isNew && (
                  <Button danger icon={<DeleteOutlined />} onClick={handleDelete}>
                    删除
                  </Button>
                )}
              </Space>
            }
            bodyStyle={{ minHeight: 520 }}
          >
            {!selection ? (
              <Empty description="请选择左侧课程或文档，或点击新增按钮开始维护" />
            ) : selection.type === "course" ? (
              <Form form={courseForm} layout="vertical" onFinish={handleSaveCourse}>
                <Row gutter={16}>
                  <Col xs={24} md={12}>
                    <Form.Item name="id" label="课程 ID" rules={[{ required: true, message: "请输入课程 ID" }]}>
                      <Input disabled={!selection.isNew} placeholder="例如 go-backend" />
                    </Form.Item>
                  </Col>
                  <Col xs={24} md={12}>
                    <Form.Item name="order" label="排序" rules={[{ required: true, message: "请输入排序" }]}>
                      <InputNumber min={0} style={{ width: "100%" }} />
                    </Form.Item>
                  </Col>
                </Row>
                <Form.Item name="title" label="课程标题" rules={[{ required: true, message: "请输入课程标题" }]}>
                  <Input placeholder="例如 Go 后端开发" />
                </Form.Item>
                <Form.Item name="description" label="课程说明">
                  <Input.TextArea rows={5} placeholder="简要说明课程内容和适用对象" />
                </Form.Item>
                <Form.Item>
                  <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saving}>
                    保存课程
                  </Button>
                </Form.Item>
              </Form>
            ) : (
              <Spin spinning={detailLoading}>
                <Typography.Text type="secondary" style={{ display: "block", marginBottom: 16 }}>
                  所属课程：{selectedCourse?.title ?? selection.courseId}
                </Typography.Text>
                <Form form={docForm} layout="vertical" onFinish={handleSaveDocument}>
                  <Row gutter={16}>
                    <Col xs={24} md={12}>
                      <Form.Item name="id" label="文档 ID" rules={[{ required: true, message: "请输入文档 ID" }]}>
                        <Input disabled={!selection.isNew} placeholder="例如 week01-intro" />
                      </Form.Item>
                    </Col>
                    <Col xs={24} md={12}>
                      <Form.Item name="order" label="排序" rules={[{ required: true, message: "请输入排序" }]}>
                        <InputNumber min={0} style={{ width: "100%" }} />
                      </Form.Item>
                    </Col>
                  </Row>
                  <Form.Item name="title" label="文档标题" rules={[{ required: true, message: "请输入文档标题" }]}>
                    <Input placeholder="例如 第一周：环境准备" />
                  </Form.Item>
                  <Form.Item name="markdown" label="Markdown 内容" rules={[{ required: true, message: "请输入 Markdown 内容" }]}>
                    <Input.TextArea rows={18} placeholder="# 标题&#10;&#10;在这里编写课件内容..." />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saving}>
                      保存文档
                    </Button>
                  </Form.Item>
                </Form>
              </Spin>
            )}
          </Card>
        </Col>
      </Row>

      <Drawer
        title="文档预览"
        open={previewOpen}
        width={760}
        onClose={() => setPreviewOpen(false)}
        destroyOnClose
      >
        <div className="markdown-body" style={{ minHeight: 360 }}>
          {previewMarkdown ? (
            <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeHighlight]}>
              {previewMarkdown}
            </ReactMarkdown>
          ) : (
            <Empty description="暂无 Markdown 内容" />
          )}
        </div>
      </Drawer>
    </Space>
  );
}
