import {
  Heart,
  HelpCircle,
  LogOut,
  Package,
  Settings,
  Shield,
  Star,
  Truck,
} from "lucide-react";
import {
  canAccessAdminPanel,
  getRoleLabel,
} from "../../constants/roles";
import { NavLink, useNavigate } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { logout } from "../../features/auth/authSlice";
import { getProfileDisplayName, userAvatarUrl } from "../../utils/userDisplay";
import { AccountAvatar } from "../AccountAvatar/AccountAvatar";
import styles from "./AccountSidebar.module.css";

const navItems = [
  { to: "/profile/orders", label: "Мои покупки", icon: Package },
  { to: "/profile/reviews", label: "Отзывы", icon: Star },
  { to: "/profile/favorites", label: "Избранное", icon: Heart },
  { to: "/profile/settings", label: "Профиль", icon: Settings },
  { to: "#", label: "FAQ", icon: HelpCircle, disabled: true },
];

export function AccountSidebar() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const user = useAppSelector((state) => state.auth.user);
  const profile = useAppSelector((state) => state.profile);

  if (!user) {
    return null;
  }

  const fullName = getProfileDisplayName(user, profile);

  const handleLogout = () => {
    dispatch(logout());
    navigate("/login");
  };

  return (
    <aside className={styles.sidebar}>
      <div className={styles.profileCard}>
        <AccountAvatar
          name={fullName}
          avatarUrl={userAvatarUrl(profile.avatarUrl)}
          size="lg"
        />
        <strong className={styles.userName}>{fullName}</strong>
        <NavLink to="/profile/settings" className={styles.editBtn}>
          Изменить профиль
        </NavLink>
      </div>

      <nav className={styles.nav}>
        {navItems.map(({ to, label, icon: Icon, disabled }) =>
          disabled ? (
            <span key={label} className={styles.navLinkDisabled}>
              <Icon size={18} />
              {label}
            </span>
          ) : (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                `${styles.navLink} ${isActive ? styles.active : ""}`
              }
            >
              <Icon size={18} />
              {label}
            </NavLink>
          )
        )}

        {user.role === "courier" && (
          <NavLink
            to="/courier"
            className={({ isActive }) =>
              `${styles.navLink} ${isActive ? styles.active : ""}`
            }
          >
            <Truck size={18} />
            Доставка заказов
          </NavLink>
        )}

        {canAccessAdminPanel(user.role) && (
          <NavLink
            to="/admin"
            className={({ isActive }) =>
              `${styles.navLink} ${isActive ? styles.active : ""}`
            }
          >
            <Shield size={18} />
            {user.role === "moderator"
              ? getRoleLabel("moderator")
              : "Админ-панель"}
          </NavLink>
        )}
      </nav>

      <button type="button" className={styles.logout} onClick={handleLogout}>
        <LogOut size={18} />
        Выйти
      </button>
    </aside>
  );
}
