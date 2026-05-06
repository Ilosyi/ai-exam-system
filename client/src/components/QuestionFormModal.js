import { jsx as _jsx, jsxs as _jsxs, Fragment as _Fragment } from "react/jsx-runtime";
import { useEffect } from "react";
import { Form, Input, Modal, Select, Radio, Checkbox, Space } from "antd";
export function QuestionFormModal({ open, onCancel, onSubmit, initialValues }) {
    const [form] = Form.useForm();
    useEffect(() => {
        if (open) {
            form.resetFields();
            if (initialValues) {
                const vals = {
                    ...initialValues,
                    singleAnswer: (initialValues.type === 'single' && initialValues.answers?.[0] !== undefined)
                        ? initialValues.answers[0]
                        : undefined
                };
                form.setFieldsValue(vals);
            }
            else {
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
            };
            if (values.type === "coding") {
                await onSubmit({ ...base, options: [], answers: [] });
            }
            else {
                const opts = (values.options ?? []).slice(0, 4);
                let ans = [];
                if (values.type === "single") {
                    ans = typeof values.singleAnswer === "number" ? [values.singleAnswer] : [];
                }
                else {
                    ans = Array.isArray(values.answers) ? values.answers : [];
                }
                await onSubmit({ ...base, options: opts, answers: ans });
            }
            onCancel();
        }
        catch (e) {
            // validation failed
        }
    };
    return (_jsx(Modal, { title: initialValues ? "编辑题目" : "创建题目", open: open, onCancel: onCancel, onOk: handleOk, width: 800, destroyOnClose: true, maskClosable: false, okText: "\u786E\u8BA4", cancelText: "\u53D6\u6D88", children: _jsxs(Form, { form: form, layout: "horizontal", labelCol: { style: { width: 60, textAlign: 'left', fontWeight: 500 } }, wrapperCol: { flex: 1 }, colon: false, children: [_jsxs("div", { style: { display: 'flex', alignItems: 'flex-start', marginBottom: 24 }, children: [_jsx(Form.Item, { name: "type", label: "\u7C7B\u578B", rules: [{ required: true }], style: { marginBottom: 0, width: 300, marginRight: 20 }, labelCol: { style: { width: 50, textAlign: 'left' } }, children: _jsx(Select, { options: [
                                    { label: "单选题", value: "single" },
                                    { label: "多选题", value: "multiple" },
                                    { label: "编程题", value: "coding" },
                                ] }) }), _jsx(Form.Item, { name: "language", label: "\u8BED\u8A00", rules: [{ required: true }], style: { marginBottom: 0, width: 300 }, labelCol: { style: { width: 50, textAlign: 'left' } }, children: _jsx(Select, { options: [
                                    { label: "Go", value: "go" },
                                    { label: "C++", value: "cpp" },
                                    { label: "Java", value: "java" },
                                    { label: "JavaScript", value: "javascript" },
                                    { label: "Python", value: "python" },
                                ], allowClear: true }) })] }), _jsx(Form.Item, { name: "title", label: "\u6807\u9898", rules: [{ required: true, message: "请输入标题" }], children: _jsx(Input, { placeholder: "\u8BF7\u8F93\u5165\u6807\u9898" }) }), _jsx(Form.Item, { name: "content", label: "\u5185\u5BB9", rules: [{ required: true, message: "请输入内容" }], children: _jsx(Input.TextArea, { rows: 6, placeholder: "", showCount: true, maxLength: 2000 }) }), _jsx(Form.Item, { noStyle: true, shouldUpdate: (prev, curr) => prev.type !== curr.type, children: ({ getFieldValue }) => {
                        const type = getFieldValue("type");
                        if (type === "coding")
                            return null;
                        return (_jsxs(_Fragment, { children: [['A', 'B', 'C', 'D'].map((label, index) => (_jsx(Form.Item, { label: label, name: ["options", index], rules: [{ required: true, message: "请输入选项内容" }], children: _jsx(Input, { placeholder: "\u8BF7\u8F93\u5165\u9009\u9879\u5185\u5BB9" }) }, label))), _jsx(Form.Item, { label: "\u7B54\u6848", name: type === "single" ? "singleAnswer" : "answers", rules: [{ required: true, message: "请选择答案" }], children: type === "single" ? (_jsx(Radio.Group, { children: _jsxs(Space, { size: "large", children: [_jsx(Radio, { value: 0, children: "A" }), _jsx(Radio, { value: 1, children: "B" }), _jsx(Radio, { value: 2, children: "C" }), _jsx(Radio, { value: 3, children: "D" })] }) })) : (_jsx(Checkbox.Group, { children: _jsxs(Space, { size: "large", children: [_jsx(Checkbox, { value: 0, children: "A" }), _jsx(Checkbox, { value: 1, children: "B" }), _jsx(Checkbox, { value: 2, children: "C" }), _jsx(Checkbox, { value: 3, children: "D" })] }) })) })] }));
                    } })] }) }));
}
