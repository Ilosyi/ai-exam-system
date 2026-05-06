import { useEffect } from "react";
import { Form, Input, Modal, Select, Radio, Checkbox, Space } from "antd";
import type { Question } from "../types/question";

/**
 * QuestionFormModal
 * -----------------
 * 新增 / 编辑题目的模态窗口。
 * 目标：尽量仿照设计图的布局（类型/语言在一行，标题与内容全宽，选项垂直排列，答案选项横向）。
 *
 * 注意事项：
 * - 对于编程题（coding），隐藏选项与答案字段
 * - 提交时根据题型组装正确的 payload（single => singleAnswer 转为 answers 数组）
 */

interface Props {
  open: boolean;
  onCancel: () => void;
  onSubmit: (values: Partial<Question>) => Promise<void>;
  initialValues?: Partial<Question>;
}

type QuestionFormValues = {
  type: Question["type"]; 
  language: Question["language"]; 
  title: string; 
  content: string; 
  options?: string[]; 
  answers?: number[];
  singleAnswer?: number;
};

export function QuestionFormModal({ open, onCancel, onSubmit, initialValues }: Props) {
  const [form] = Form.useForm<QuestionFormValues>();

  useEffect(() => {
    if (open) {
      form.resetFields();
      if (initialValues) {
        const vals: any = {
          ...initialValues,
          singleAnswer: (initialValues.type === 'single' && initialValues.answers?.[0] !== undefined)
            ? initialValues.answers[0]
            : undefined
        };
        form.setFieldsValue(vals);
      } else {
        form.setFieldsValue({ type: "single", language: "go" });
      }
    }
  }, [open, initialValues, form]);

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      const base = {
        type: values.type,
        language: values.language,
        title: values.title,
        content: values.content,
      } as Partial<Question>;

      if (values.type === "coding") {
        await onSubmit({ ...base, options: [], answers: [] });
      } else {
        const opts = (values.options ?? []).slice(0, 4);
        let ans: number[] = [];
        if (values.type === "single") {
          ans = typeof values.singleAnswer === "number" ? [values.singleAnswer] : [];
        } else {
          ans = Array.isArray(values.answers) ? values.answers : [];
        }
        await onSubmit({ ...base, options: opts, answers: ans });
      }
      onCancel();
    } catch (e) {
      // validation failed
    }
  };

  return (
    <Modal
      title={initialValues ? "编辑题目" : "创建题目"}
      open={open}
      onCancel={onCancel}
      onOk={handleOk}
      width={800}
      destroyOnClose
      maskClosable={false}
      okText="确认"
      cancelText="取消"
    >
      {/* 表单布局：使用水平布局以便自定义 label 宽度，模拟设计图的对齐 */}
      <Form
        form={form}
        layout="horizontal"
        labelCol={{ style: { width: 60, textAlign: 'left', fontWeight: 500 } }}
        wrapperCol={{ flex: 1 }}
        colon={false}
      >
        {/* Row 1: Type and Language（同一行） */}
        <div style={{ display: 'flex', alignItems: 'flex-start', marginBottom: 24 }}>
          <Form.Item
            name="type"
            label="类型"
            rules={[{ required: true }]}
            style={{ marginBottom: 0, width: 300, marginRight: 20 }}
            labelCol={{ style: { width: 50, textAlign: 'left' } }}
          >
            <Select options={[
              { label: "单选题", value: "single" },
              { label: "多选题", value: "multiple" },
              { label: "编程题", value: "coding" },
            ]} />
          </Form.Item>
          <Form.Item
            name="language"
            label="语言"
            rules={[{ required: true }]}
            style={{ marginBottom: 0, width: 300 }}
            labelCol={{ style: { width: 50, textAlign: 'left' } }}
          >
            <Select options={[
              { label: "Go", value: "go" },
              { label: "C++", value: "cpp" },
              { label: "Java", value: "java" },
              { label: "JavaScript", value: "javascript" },
              { label: "Python", value: "python" },
            ]} allowClear />
          </Form.Item>
        </div>

        <Form.Item name="title" label="标题" rules={[{ required: true, message: "请输入标题" }]}>
          <Input placeholder="请输入标题" />
        </Form.Item>

        <Form.Item name="content" label="内容" rules={[{ required: true, message: "请输入内容" }]}>
          <Input.TextArea
            rows={6}
            placeholder=""
            showCount
            maxLength={2000}
          />
        </Form.Item>

        <Form.Item noStyle shouldUpdate={(prev, curr) => prev.type !== curr.type}>
          {({ getFieldValue }) => {
            const type = getFieldValue("type");
            if (type === "coding") return null;
            return (
              <>
                {['A', 'B', 'C', 'D'].map((label, index) => (
                  <Form.Item
                    key={label}
                    label={label}
                    name={["options", index]}
                    rules={[{ required: true, message: "请输入选项内容" }]}
                  >
                    <Input placeholder="请输入选项内容" />
                  </Form.Item>
                ))}

                <Form.Item
                  label="答案"
                  name={type === "single" ? "singleAnswer" : "answers"}
                  rules={[{ required: true, message: "请选择答案" }]}
                >
                  {type === "single" ? (
                    <Radio.Group>
                      <Space size="large">
                        <Radio value={0}>A</Radio>
                        <Radio value={1}>B</Radio>
                        <Radio value={2}>C</Radio>
                        <Radio value={3}>D</Radio>
                      </Space>
                    </Radio.Group>
                  ) : (
                    <Checkbox.Group>
                      <Space size="large">
                        <Checkbox value={0}>A</Checkbox>
                        <Checkbox value={1}>B</Checkbox>
                        <Checkbox value={2}>C</Checkbox>
                        <Checkbox value={3}>D</Checkbox>
                      </Space>
                    </Checkbox.Group>
                  )}
                </Form.Item>
              </>
            );
          }}
        </Form.Item>
      </Form>
    </Modal>
  );
}
