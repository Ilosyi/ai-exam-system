import { BrowserRouter, Navigate, Route, Routes, useLocation } from "react-router-dom";
import { Spin } from "antd";
import { AppLayout } from "./components/AppLayout";
import { QuestionPage } from "./pages/QuestionPage";
import { NotesPage } from "./pages/NotesPage";
import { DocumentManagePage } from "./pages/DocumentManagePage";
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
  return (
    <div style={{ minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center" }}>
      <Spin size="large" />
    </div>
  );
}

function RequireAuth({ roles, children }: { roles: Array<"admin" | "teacher" | "student">; children: JSX.Element }) {
  const { user, loading } = useAuth();
  const location = useLocation();

  if (loading) {
    return <FullPageSpin />;
  }
  if (!user) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }
  if (!roles.includes(user.role)) {
    return <Navigate to={getDefaultRouteByRole(user.role)} replace />;
  }
  return children;
}

function PublicOnly({ children }: { children: JSX.Element }) {
  const { user, loading } = useAuth();

  if (loading) {
    return <FullPageSpin />;
  }
  if (user) {
    return <Navigate to={getDefaultRouteByRole(user.role)} replace />;
  }
  return children;
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route
            path="/login"
            element={
              <PublicOnly>
                <LoginPage />
              </PublicOnly>
            }
          />

          <Route
            path="/exam"
            element={
              <RequireAuth roles={["admin", "student"]}>
                <ExamListPage />
              </RequireAuth>
            }
          />
          <Route
            path="/exam/:id/take"
            element={
              <RequireAuth roles={["admin", "student"]}>
                <ExamTakePage />
              </RequireAuth>
            }
          />
          <Route
            path="/exam/:id/result"
            element={
              <RequireAuth roles={["admin", "student"]}>
                <ExamResultPage />
              </RequireAuth>
            }
          />

          <Route
            element={
              <RequireAuth roles={["admin", "teacher"]}>
                <AppLayout />
              </RequireAuth>
            }
          >
            <Route path="/questions" element={<QuestionPage />} />
            <Route path="/notes" element={<NotesPage />} />
            <Route path="/documents" element={<DocumentManagePage />} />
            <Route path="/papers" element={<PaperManagePage />} />
            <Route path="/papers/generate" element={<PaperGeneratePage />} />
            <Route path="/papers/:id/edit" element={<PaperEditPage />} />
            <Route path="/classes" element={<ClassManagePage />} />
            <Route
              path="/users"
              element={
                <RequireAuth roles={["admin"]}>
                  <UserManagePage />
                </RequireAuth>
              }
            />
          </Route>

          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
