import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { BrowserRouter, Navigate, Route, Routes, useLocation } from "react-router-dom";
import { Spin } from "antd";
import { AppLayout } from "./components/AppLayout";
import { QuestionPage } from "./pages/QuestionPage";
import { NotesPage } from "./pages/NotesPage";
import { PaperGeneratePage } from "./pages/PaperGeneratePage";
import { PaperManagePage } from "./pages/PaperManagePage";
import { PaperEditPage } from "./pages/PaperEditPage";
import { ClassManagePage } from "./pages/ClassManagePage";
import { UserManagePage } from "./pages/UserManagePage";
import { ExamListPage } from "./pages/ExamListPage";
import { ExamTakePage } from "./pages/ExamTakePage";
import { ExamResultPage } from "./pages/ExamResultPage";
import { LoginPage } from "./pages/LoginPage";
import { AuthProvider, getDefaultRouteByRole, useAuth } from "./hooks/useAuth";
function FullPageSpin() {
    return (_jsx("div", { style: { minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center" }, children: _jsx(Spin, { size: "large" }) }));
}
function RequireAuth({ roles, children }) {
    const { user, loading } = useAuth();
    const location = useLocation();
    if (loading) {
        return _jsx(FullPageSpin, {});
    }
    if (!user) {
        return _jsx(Navigate, { to: "/login", replace: true, state: { from: location } });
    }
    if (!roles.includes(user.role)) {
        return _jsx(Navigate, { to: getDefaultRouteByRole(user.role), replace: true });
    }
    return children;
}
function PublicOnly({ children }) {
    const { user, loading } = useAuth();
    if (loading) {
        return _jsx(FullPageSpin, {});
    }
    if (user) {
        return _jsx(Navigate, { to: getDefaultRouteByRole(user.role), replace: true });
    }
    return children;
}
export default function App() {
    return (_jsx(BrowserRouter, { children: _jsx(AuthProvider, { children: _jsxs(Routes, { children: [_jsx(Route, { path: "/login", element: _jsx(PublicOnly, { children: _jsx(LoginPage, {}) }) }), _jsx(Route, { path: "/exam", element: _jsx(RequireAuth, { roles: ["admin", "student"], children: _jsx(ExamListPage, {}) }) }), _jsx(Route, { path: "/exam/:id/take", element: _jsx(RequireAuth, { roles: ["admin", "student"], children: _jsx(ExamTakePage, {}) }) }), _jsx(Route, { path: "/exam/:id/result", element: _jsx(RequireAuth, { roles: ["admin", "student"], children: _jsx(ExamResultPage, {}) }) }), _jsxs(Route, { element: _jsx(RequireAuth, { roles: ["admin", "teacher"], children: _jsx(AppLayout, {}) }), children: [_jsx(Route, { path: "/questions", element: _jsx(QuestionPage, {}) }), _jsx(Route, { path: "/notes", element: _jsx(NotesPage, {}) }), _jsx(Route, { path: "/papers", element: _jsx(PaperManagePage, {}) }), _jsx(Route, { path: "/papers/generate", element: _jsx(PaperGeneratePage, {}) }), _jsx(Route, { path: "/papers/:id/edit", element: _jsx(PaperEditPage, {}) }), _jsx(Route, { path: "/classes", element: _jsx(ClassManagePage, {}) }), _jsx(Route, { path: "/users", element: _jsx(RequireAuth, { roles: ["admin"], children: _jsx(UserManagePage, {}) }) })] }), _jsx(Route, { path: "/", element: _jsx(Navigate, { to: "/login", replace: true }) }), _jsx(Route, { path: "*", element: _jsx(Navigate, { to: "/login", replace: true }) })] }) }) }));
}
