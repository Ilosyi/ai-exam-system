import { useCallback, useEffect, useMemo, useState } from "react";
import type { Paper, PaperFilters } from "../types/paper";
import { fetchPapers, deletePaper } from "../api/paper";

const DEFAULT_FILTERS: PaperFilters = {
  keyword: "",
  status: "",
  page: 1,
  pageSize: 10,
};

export function usePapers() {
  const [filters, setFilters] = useState<PaperFilters>(DEFAULT_FILTERS);
  const [data, setData] = useState<Paper[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetchPapers(filters);
      setData(res.data);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    void load();
  }, [load]);

  const pagination = useMemo(
    () => ({
      current: filters.page ?? 1,
      pageSize: filters.pageSize ?? DEFAULT_FILTERS.pageSize!,
      total,
      onChange: (page: number, pageSize?: number) =>
        setFilters((prev) => ({ ...prev, page, pageSize })),
    }),
    [filters.page, filters.pageSize, total],
  );

  const onDelete = useCallback(async (id: number) => {
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
