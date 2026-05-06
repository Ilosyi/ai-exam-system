import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { useNotes } from "../hooks/useNotes";

export function NotesPage() {
  const content = useNotes();

  return (
    <div className="markdown-body" style={{ minHeight: 400, background: "#fff", padding: 24, borderRadius: 8 }}>
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
    </div>
  );
}
