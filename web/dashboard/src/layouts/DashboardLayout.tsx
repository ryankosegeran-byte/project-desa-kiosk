import { useEffect, type ReactNode } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { Sun, Moon } from "lucide-react";
import { getUser } from "../lib/api";
import { useTheme } from "../lib/theme";

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
  { to: "/admin/jenis-surat", icon: "📋", label: "Kelola Surat" },
  { to: "/kiosk-status", icon: "📡", label: "Status Kiosk" },
  { to: "/admin/desa", icon: "🏘️", label: "Kelola Desa", admin: true },
  { to: "/admin/users", icon: "👤", label: "Kelola User", admin: true },
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
  const { theme, toggleTheme } = useTheme();

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
                  padding: "10px 16px 4px 16px",
                  fontSize: "11px",
                  fontWeight: "bold",
                  letterSpacing: "0.5px",
                  textTransform: "uppercase",
                  color: "var(--text-muted)",
                  borderTop: "1px solid var(--border-color)",
                  marginTop: "6px",
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

      <main className="main-content">
        <div className="top-bar">
          <h2>{title}</h2>
          <button
            className="theme-toggle"
            onClick={toggleTheme}
            title={theme === "dark" ? "Mode Terang" : "Mode Gelap"}
          >
            {theme === "dark" ? <Sun size={18} /> : <Moon size={18} />}
          </button>
        </div>
        {children}
      </main>
    </div>
  );
}
