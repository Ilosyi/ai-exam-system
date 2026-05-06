import { useCallback, useState } from "react";
import { fetchPaper, replaceQuestion, deletePaperItem, publishPaper, unpublishPaper, updatePaper } from "../api/paper";
import { message } from "antd";
export function usePaperEdit(paperId) {
    const [paper, setPaper] = useState(null);
    const [loading, setLoading] = useState(false);
    const load = useCallback(async () => {
        setLoading(true);
        try {
            const res = await fetchPaper(paperId);
            setPaper(res.data);
        }
        finally {
            setLoading(false);
        }
    }, [paperId]);
    const onDeleteItem = useCallback(async (itemId) => {
        try {
            const res = await deletePaperItem(paperId, itemId);
            setPaper(res.data);
            message.success("删除成功");
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "删除失败");
        }
    }, [paperId]);
    const onReplaceItem = useCallback(async (itemId, questionId) => {
        try {
            const res = await replaceQuestion(paperId, itemId, questionId);
            setPaper(res.data);
            message.success("替换成功");
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "替换失败");
        }
    }, [paperId]);
    const onPublish = useCallback(async (startTime, endTime, duration, classId) => {
        try {
            await publishPaper(paperId, { startTime, endTime, duration, classId });
            message.success("发布成功");
            await load();
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "发布失败");
        }
    }, [paperId, load]);
    const onUnpublish = useCallback(async () => {
        try {
            await unpublishPaper(paperId);
            message.success("取消发布成功");
            await load();
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "取消发布失败");
        }
    }, [paperId, load]);
    const onUpdateTitle = useCallback(async (title) => {
        try {
            await updatePaper(paperId, { title });
            message.success("更新成功");
            await load();
        }
        catch (err) {
            message.error(err instanceof Error ? err.message : "更新失败");
        }
    }, [paperId, load]);
    const getItemStats = useCallback(() => {
        if (!paper?.items)
            return { single: 0, multiple: 0, coding: 0, total: 0 };
        const stats = { single: 0, multiple: 0, coding: 0, total: paper.items.length };
        for (const item of paper.items) {
            if (item.type in stats) {
                stats[item.type]++;
            }
        }
        return stats;
    }, [paper]);
    return {
        paper,
        loading,
        load,
        onDeleteItem,
        onReplaceItem,
        onPublish,
        onUnpublish,
        onUpdateTitle,
        getItemStats,
    };
}
