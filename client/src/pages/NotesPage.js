import { jsx as _jsx } from "react/jsx-runtime";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { useNotes } from "../hooks/useNotes";
export function NotesPage() {
    const content = useNotes();
    return (_jsx("div", { className: "markdown-body", style: { minHeight: 400, background: "#fff", padding: 24, borderRadius: 8 }, children: _jsx(ReactMarkdown, { remarkPlugins: [remarkGfm], children: content }) }));
}
