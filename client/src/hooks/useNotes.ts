import { useEffect, useState } from "react";
import { apiGet } from "../api/client";

export function useNotes() {
  const [content, setContent] = useState("加载中...");

  useEffect(() => {
    apiGet<{ markdown: string }>("/notes")
      .then((data) => setContent(data.markdown ?? ""))
      .catch(() => setContent("无法加载 README 内容"));
  }, []);

  return content;
}
