import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { BarChart3, Bell, FileDown, LayoutDashboard, LogOut, Package } from 'lucide-react';
import { useAuth } from '../features/auth/AuthContext';

export function AppLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

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
          <NavLink to="/assets">
            <Package size={18} /> Assets
          </NavLink>
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
              navigate('/login');
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
