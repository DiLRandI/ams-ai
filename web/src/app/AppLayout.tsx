import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import {
  BarChart3,
  Bell,
  FileDown,
  ChevronDown,
  ChevronRight,
  LayoutDashboard,
  LogOut,
  Package,
  FolderCog,
  Boxes,
} from "lucide-react";
import { useAuth } from "../features/auth/AuthContext";

export function AppLayout() {
  const { user, logout } = useAuth();
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const [assetsOpen, setAssetsOpen] = useState(() => {
    const stored = localStorage.getItem("ams_sidebar_assets_open");
    return stored ? stored === "true" : true;
  });
  const assetsSectionOpen =
    assetsOpen || pathname === "/assets" || pathname === "/assets/categories";

  useEffect(() => {
    localStorage.setItem("ams_sidebar_assets_open", String(assetsOpen));
  }, [assetsOpen]);

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="brandMark">A</span>
          <div>
            <strong>AMS</strong>
            <small>Asset Management</small>
          </div>
        </div>
        <nav className="nav">
          <NavLink to="/dashboard">
            <LayoutDashboard size={18} /> Dashboard
          </NavLink>
          <div className="navGroup">
            <button
              className="navGroupToggle"
              type="button"
              aria-expanded={assetsSectionOpen}
              onClick={() => setAssetsOpen((value) => !value)}
            >
              <span className="navGroupLabel">
                <Package size={18} /> Assets
              </span>
              {assetsSectionOpen ? (
                <ChevronDown size={16} />
              ) : (
                <ChevronRight size={16} />
              )}
            </button>
            {assetsSectionOpen && (
              <div className="navSubgroup">
                <NavLink to="/assets" end>
                  <Boxes size={16} /> Asset list
                </NavLink>
                <NavLink to="/assets/categories">
                  <FolderCog size={16} /> Categories
                </NavLink>
              </div>
            )}
          </div>
          <NavLink to="/reminders">
            <Bell size={18} /> Reminders
          </NavLink>
          <NavLink to="/reports">
            <FileDown size={18} /> Reports
          </NavLink>
        </nav>
        <div className="userPanel">
          <div>
            <strong>{user?.fullName}</strong>
            <small>{user?.role}</small>
          </div>
          <button
            className="iconButton"
            type="button"
            title="Log out"
            onClick={() => {
              logout();
              navigate("/login");
            }}
          >
            <LogOut size={18} />
          </button>
        </div>
      </aside>
      <main className="main">
        <Outlet />
      </main>
      <BarChart3 className="mobileWatermark" size={1} aria-hidden />
    </div>
  );
}
