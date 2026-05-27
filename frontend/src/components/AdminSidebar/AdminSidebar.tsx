import { MessageCircle, Package, ShoppingBag, Star, Users } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { useAppSelector } from "../../app/hooks";
import { getRoleLabel, type UserRole } from "../../constants/roles";
import { getProfileDisplayName, userAvatarUrl } from "../../utils/userDisplay";
import { AccountAvatar } from "../AccountAvatar/AccountAvatar";
import styles from "./AdminSidebar.module.css";

const allLinks = [
  { tab: "products", label: "Товары", icon: Package, roles: ["admin"] as UserRole[] },
  { tab: "orders", label: "Заказы", icon: ShoppingBag, roles: ["admin"] as UserRole[] },
  { tab: "users", label: "Пользователи", icon: Users, roles: ["admin", "moderator"] as UserRole[] },
  {
    tab: "support",
    label: "Поддержка",
    icon: MessageCircle,
    roles: ["admin", "moderator"] as UserRole[],
  },
  { tab: "reviews", label: "Отзывы", icon: Star, roles: ["admin"] as UserRole[] },
];

export function AdminSidebar() {
  const user = useAppSelector((state) => state.auth.user);
  const profile = useAppSelector((state) => state.profile);
  const [searchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") ?? "products";

  if (!user) {
    return null;
  }

  const fullName = getProfileDisplayName(user, profile);
  const links = allLinks.filter((link) => link.roles.includes(user.role));
  const roleLabel =
    user.role === "admin" ? "Администратор" : getRoleLabel(user.role);

  return (
    <aside className={styles.sidebar}>
      <Link to="/profile/orders" className={styles.adminProfile}>
        <AccountAvatar
          name={fullName}
          avatarUrl={userAvatarUrl(profile.avatarUrl)}
          size="md"
        />
        <div>
          <strong>{fullName}</strong>
          <span>{roleLabel}</span>
        </div>
      </Link>

      <nav className={styles.nav}>
        {links.map(({ tab, label, icon: Icon }) => (
          <Link
            key={tab}
            to={`/admin?tab=${encodeURIComponent(tab)}`}
            className={`${styles.link} ${
              tab === activeTab ? styles.active : ""
            }`}
          >
            <Icon size={18} />
            {label}
          </Link>
        ))}
      </nav>
    </aside>
  );
}
