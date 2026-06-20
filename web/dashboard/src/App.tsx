import { type ReactNode } from "react";
import { Routes, Route, Navigate } from "react-router-dom";

import RequireAuth from "./components/RequireAuth";
import DashboardLayout from "./layouts/DashboardLayout";

import LoginForm from "./components/LoginForm";
import DashboardOverview from "./components/DashboardOverview";
import WargaList from "./components/WargaList";
import WargaRegister from "./components/WargaRegister";
import WargaDraftComplete from "./components/WargaDraftComplete";
import SuratTable from "./components/SuratTable";
import TemplatesList from "./components/TemplatesList";
import KioskStatus from "./components/KioskStatus";
import DesaManager from "./components/DesaManager";
import UserManager from "./components/UserManager";
import JenisSuratManager from "./components/JenisSuratManager";
import ActivityLogList from "./components/ActivityLogList";
import OCRProviderConfig from "./components/OCRProviderConfig";

// Wraps a page in the auth guard + dashboard chrome.
function Page({ title, children }: { title: string; children: ReactNode }) {
  return (
    <RequireAuth>
      <DashboardLayout title={title}>{children}</DashboardLayout>
    </RequireAuth>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LoginForm />} />

      <Route path="/dashboard" element={<Page title="Dashboard Overview"><DashboardOverview /></Page>} />

      <Route path="/warga" element={<Page title="Data Warga Desa"><WargaList /></Page>} />
      <Route path="/warga/register" element={<Page title="Registrasi Warga Baru"><WargaRegister /></Page>} />
      <Route path="/warga/draft" element={<Page title="Lengkapi Data Warga"><WargaDraftComplete /></Page>} />

      <Route path="/surat" element={<Page title="Arsip Cetak Surat"><SuratTable /></Page>} />
      <Route path="/templates" element={<Page title="Kelola Template Cetak"><TemplatesList /></Page>} />
      <Route path="/kiosk-status" element={<Page title="Status Terminal Kiosk"><KioskStatus /></Page>} />

      <Route path="/admin/desa" element={<Page title="Kelola Desa"><DesaManager /></Page>} />
      <Route path="/admin/users" element={<Page title="Kelola User PIC"><UserManager /></Page>} />
      <Route path="/admin/jenis-surat" element={<Page title="Kelola Jenis Surat"><JenisSuratManager /></Page>} />
      <Route path="/admin/activity" element={<Page title="Log Aktivitas Audit"><ActivityLogList /></Page>} />

      <Route path="/settings/ocr" element={<Page title="Pengaturan AI OCR"><OCRProviderConfig /></Page>} />

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
