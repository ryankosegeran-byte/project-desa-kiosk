import { useEffect, type ReactNode } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { getUser } from "../lib/api";

interface NavItem {
  to: string;
  icon: string;
  label: string;
  admin?: boolean;
}

const NAV_ITEMS: NavItem[] = [
  { to: "/dashboard", icon: "🏠", label: "Dashboard" },
  { to: "/warga", icon: "👥", label: "Data Warga" },
  { to: "/warga/register", icon: "➕", label: "Registrasi Warga" },
  { to: "/surat", icon: "📄", label: "Arsip Surat" },
  { to: "/templates", icon: "📋", label: "Template Cetak" },
  { to: "/kiosk-status", icon: "📡", label: "Status Kiosk" },
  { to: "/admin/desa", icon: "🏘️", label: "Kelola Desa", admin: true },
  { to: "/admin/users", icon: "👤", label: "Kelola User", admin: true },
  { to: "/admin/jenis-surat", icon: "📝", label: "Jenis Surat", admin: true },
  { to: "/admin/activity", icon: "📊", label: "Monitoring PIC", admin: true },
  { to: "/settings/ocr", icon: "⚙️", label: "Pengaturan OCR", admin: true },
];

function isActive(pathname: string, to: string): boolean {
  if (to === "/dashboard") return pathname === to;
  return pathname === to || pathname.startsWith(to + "/");
}

interface Props {
  title: string;
  children: ReactNode;
}

export default function DashboardLayout({ title, children }: Props) {
  const location = useLocation();
  const navigate = useNavigate();
  const user = getUser();

  useEffect(() => {
    document.title = `${title} - Dashboard Kiosk Desa`;
  }, [title]);

  const isSuperadmin = user?.role === "superadmin";
  const villageLabel = user?.desa_id ? "Wilayah Kiosk" : "Superadmin Hub";

  const handleLogout = () => {
    localStorage.clear();
    navigate("/", { replace: true });
  };

  const items = NAV_ITEMS.filter((it) => !it.admin || isSuperadmin);
  const generalItems = items.filter((it) => !it.admin);
  const adminItems = items.filter((it) => it.admin);

  return (
    <div className="dashboard-wrapper">
      <aside className="sidebar">
        <div className="brand-section">
          <div className="brand-logo">KD</div>
          <div className="brand-title">
            <h1>Kiosk Desa</h1>
            <p id="village-name">{villageLabel}</p>
          </div>
        </div>

        <nav className="nav-menu">
          {generalItems.map((it) => (
            <Link
              key={it.to}
              to={it.to}
              className={`nav-item${isActive(location.pathname, it.to) ? " active" : ""}`}
            >
              <span>{it.icon}</span> {it.label}
            </Link>
          ))}

          {isSuperadmin && adminItems.length > 0 && (
            <div>
              <div
                style={{
                  padding: "16px 16px 4px 16px",
                  fontSize: "11px",
                  fontWeight: "bold",
                  textTransform: "uppercase",
                  color: "var(--text-muted)",
                  borderTop: "1px solid var(--border-color)",
                  marginTop: "8px",
                }}
              >
                Admin Panel
              </div>
              {adminItems.map((it) => (
                <Link
                  key={it.to}
                  to={it.to}
                  className={`nav-item${isActive(location.pathname, it.to) ? " active" : ""}`}
                >
                  <span>{it.icon}</span> {it.label}
                </Link>
              ))}
            </div>
          )}
        </nav>

        <div className="user-profile-section">
          <div className="user-info">
            <span className="user-name" id="user-display-name">
              {user?.nama ?? "Perangkat Desa"}
            </span>
            <span className="user-role" id="user-display-role">
              {user?.jabatan ?? user?.role ?? "Operator"}
            </span>
          </div>
          <button
            className="btn btn-secondary"
            style={{ marginTop: "10px" }}
            onClick={handleLogout}
          >
            Keluar
          </button>
        </div>
      </aside>

      <main className="main-content">{children}</main>
    </div>
  );
}
