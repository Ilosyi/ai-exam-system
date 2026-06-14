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
import { ExamTakePage } from "./pages/ExamTakePage";
import { ExamResultPage } from "./pages/ExamResultPage";
import { HomePage } from "./pages/HomePage";
import { DocumentReaderPage } from "./pages/DocumentReaderPage";
import { LoginPage } from "./pages/LoginPage";
import { AuthProvider, getDefaultRouteByRole, useAuth } from "./hooks/useAuth";

type AppRole = "admin" | "teacher" | "student";

export const routeRoleAccess: Record<string, AppRole[]> = {
  studentHome: ["admin", "student"],
  documentReader: ["admin", "teacher", "student"],
  studentExam: ["admin", "student"],
  teacherWorkspace: ["admin", "teacher"],
  adminOnly: ["admin"],
};

function FullPageSpin() {
  return (
    <div style={{ minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center" }}>
      <Spin size="large" />
    </div>
  );
}

function RequireAuth({ roles, children }: { roles: AppRole[]; children: JSX.Element }) {
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
            path="/home"
            element={
              <RequireAuth roles={routeRoleAccess.studentHome}>
                <HomePage />
              </RequireAuth>
            }
          />
          <Route
            path="/home/courses/:courseId/docs/:docId"
            element={
              <RequireAuth roles={routeRoleAccess.documentReader}>
                <DocumentReaderPage />
              </RequireAuth>
            }
          />
          <Route
            path="/exam"
            element={
              <RequireAuth roles={routeRoleAccess.studentExam}>
                <Navigate to="/home" replace />
              </RequireAuth>
            }
          />
          <Route
            path="/exam/:id/take"
            element={
              <RequireAuth roles={routeRoleAccess.studentExam}>
                <ExamTakePage />
              </RequireAuth>
            }
          />
          <Route
            path="/exam/:id/result"
            element={
              <RequireAuth roles={routeRoleAccess.studentExam}>
                <ExamResultPage />
              </RequireAuth>
            }
          />
          <Route
            path="/exam/papers/:paperId/detail"
            element={
              <RequireAuth roles={routeRoleAccess.studentExam}>
                <ExamResultPage />
              </RequireAuth>
            }
          />

          <Route
            element={
              <RequireAuth roles={routeRoleAccess.teacherWorkspace}>
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
                <RequireAuth roles={routeRoleAccess.adminOnly}>
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
