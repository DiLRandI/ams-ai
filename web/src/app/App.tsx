import { Navigate, Route, Routes } from "react-router-dom";
import { AppLayout } from "./AppLayout";
import { LoginPage } from "../features/auth/LoginPage";
import { ProfilePage } from "../features/auth/ProfilePage";
import { DashboardPage } from "../features/dashboard/DashboardPage";
import { AssetDetailPage } from "../features/assets/AssetDetailPage";
import { AssetFormPage } from "../features/assets/AssetFormPage";
import { AssetsPage } from "../features/assets/AssetsPage";
import { CategoriesPage } from "../features/assets/CategoriesPage";
import { RemindersPage } from "../features/reminders/RemindersPage";
import { ReportsPage } from "../features/reports/ReportsPage";
import { ProtectedRoute } from "../features/auth/ProtectedRoute";

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppLayout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/assets" element={<AssetsPage />} />
          <Route path="/assets/categories" element={<CategoriesPage />} />
          <Route path="/assets/new" element={<AssetFormPage />} />
          <Route path="/assets/:id" element={<AssetDetailPage />} />
          <Route path="/assets/:id/edit" element={<AssetFormPage />} />
          <Route path="/reminders" element={<RemindersPage />} />
          <Route path="/reports" element={<ReportsPage />} />
          <Route path="/profile" element={<ProfilePage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}
