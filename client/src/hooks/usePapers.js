import { useCallback, useEffect, useMemo, useState } from "react";
import { fetchPapers, deletePaper } from "../api/paper";
const DEFAULT_FILTERS = {
    keyword: "",
    status: "",
    page: 1,
    pageSize: 10,
};
export function usePapers() {
    const [filters, setFilters] = useState(DEFAULT_FILTERS);
    const [data, setData] = useState([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(false);
    const load = useCallback(async () => {
        setLoading(true);
        try {
            const res = await fetchPapers(filters);
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
    const onDelete = useCallback(async (id) => {
        await deletePaper(id);
        await load();
    }, [load]);
    return {
        filters,
        setFilters,
        data,
        total,
        pagination,
        loading,
        onDelete,
        refresh: load,
    };
}
