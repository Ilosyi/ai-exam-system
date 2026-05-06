import { useCallback, useEffect, useMemo, useState } from "react";
import { fetchQuestions, createQuestion, updateQuestion, deleteQuestion, deleteQuestionsBulk } from "../api/question";
/**
 * useQuestions Hook
 * ------------------
 * 封装题库列表相关的状态与操作，包含：
 * - 当前筛选条件（keyword/type/language/page/pageSize）
 * - 列表数据和总数
 * - loading 状态
 * - 分页对象（兼容 Ant Design Table 的 pagination props）
 * - CRUD 操作：create/update/delete/bulk delete
 *
 * 设计要点：
 * - 将所有与题目列表相关的逻辑集中到一个 hook 中，便于页面复用与单测
 * - 每次变更筛选或调用 CRUD 后均刷新列表（load）以确保数据一致性
 */
const DEFAULT_FILTERS = {
    keyword: "",
    type: "",
    language: "",
    page: 1,
    pageSize: 8,
};
export function useQuestions() {
    const [filters, setFilters] = useState(DEFAULT_FILTERS);
    const [data, setData] = useState([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(false);
    const load = useCallback(async () => {
        setLoading(true);
        try {
            // 向后端请求符合当前 filters 的题目列表
            const res = await fetchQuestions(filters);
            setData(res.data);
            setTotal(res.total);
        }
        finally {
            setLoading(false);
        }
    }, [filters]);
    useEffect(() => {
        void load();
    }, [load]);
    const pagination = useMemo(() => ({
        current: filters.page ?? 1,
        pageSize: filters.pageSize ?? DEFAULT_FILTERS.pageSize,
        total,
        onChange: (page, pageSize) => setFilters((prev) => ({ ...prev, page, pageSize })),
    }), [filters.page, filters.pageSize, total]);
    const onCreate = useCallback(async (payload) => {
        // 创建题目后重新加载列表以显示最新记录
        await createQuestion(payload);
        await load();
    }, [load]);
    const onUpdate = useCallback(async (id, payload) => {
        await updateQuestion(id, payload);
        await load();
    }, [load]);
    const onDelete = useCallback(async (id) => {
        // 删除单个题目并刷新
        await deleteQuestion(id);
        await load();
    }, [load]);
    const onDeleteMany = useCallback(async (ids) => {
        // 批量删除并刷新
        await deleteQuestionsBulk(ids);
        await load();
    }, [load]);
    return {
        filters,
        setFilters,
        data,
        total,
        pagination,
        loading,
        onCreate,
        onUpdate,
        onDelete,
        onDeleteMany,
    };
}
