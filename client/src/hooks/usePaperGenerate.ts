import { useCallback, useState } from "react";
import type { PaperItem, GenerateRequest } from "../types/paper";
import { generatePaper, createPaper } from "../api/paper";
import { message } from "antd";

export function usePaperGenerate() {
  const [items, setItems] = useState<PaperItem[]>([]);
  const [language, setLanguage] = useState<string>("go");
  const [totalScore, setTotalScore] = useState<number>(100);
  const [loading, setLoading] = useState(false);

  const generate = useCallback(async (req: GenerateRequest) => {
    setLoading(true);
    try {
      const res = await generatePaper(req);
      setItems(res.data.items);
      setLanguage(res.data.language);
      setTotalScore(res.data.totalScore);
      message.success("组卷成功");
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "组卷失败");
    } finally {
      setLoading(false);
    }
  }, []);

  const removeItem = useCallback((itemId: number) => {
    setItems((prev) => prev.filter((item) => item.id !== itemId));
  }, []);

  const save = useCallback(async (title: string, language: string, totalScore: number, items: PaperItem[]) => {
    try {
      const res = await createPaper({
        title,
        language,
        totalScore,
        items: items.map((item, index) => ({
          questionId: item.questionId,
          type: item.type,
          score: item.score,
          sortNo: index + 1,
        })),
      });
      message.success("保存成功");
      return res.data;
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : "保存失败");
      return null;
    }
  }, []);

  return {
    items,
    setItems,
    language,
    totalScore,
    loading,
    generate,
    removeItem,
    save,
  };
}
