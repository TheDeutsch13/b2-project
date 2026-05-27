import { useEffect, useRef, useState, type FormEvent } from "react";
import { Link, NavLink, useLocation, useNavigate } from "react-router-dom";
import { Menu, Search, ShoppingCart, User, X } from "lucide-react";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { CartDrawer } from "../CartDrawer/CartDrawer";
import { toggleCart } from "../../features/cart/cartSlice";
import styles from "./Header.module.css";

const AUTH_PATHS = ["/login", "/register"];

export function Header() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAppSelector((state) => state.auth);
  const cartItems = useAppSelector((state) => state.cart.items);
  const notifications = useAppSelector((state) => state.notifications.items);
  const [searchQuery, setSearchQuery] = useState("");
  const [navOpen, setNavOpen] = useState(false);
  const navRef = useRef<HTMLDivElement>(null);

  const isAuthPage = AUTH_PATHS.includes(location.pathname);
  const cartCount = cartItems.reduce((sum, item) => sum + item.quantity, 0);

  useEffect(() => {
    if (location.pathname === "/catalog") {
      setSearchQuery(new URLSearchParams(location.search).get("q") ?? "");
    }
  }, [location.pathname, location.search]);

  useEffect(() => {
    setNavOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    if (!navOpen) {
      return;
    }

    const onPointerDown = (event: MouseEvent) => {
      if (navRef.current && !navRef.current.contains(event.target as Node)) {
        setNavOpen(false);
      }
    };

    window.addEventListener("mousedown", onPointerDown);
    return () => window.removeEventListener("mousedown", onPointerDown);
  }, [navOpen]);

  const handleSearchSubmit = (event: FormEvent) => {
    event.preventDefault();
    const query = searchQuery.trim();
    const params = new URLSearchParams(location.search);

    if (query) {
      params.set("q", query);
    } else {
      params.delete("q");
    }

    const suffix = params.toString();
    navigate(suffix ? `/catalog?${suffix}` : "/catalog");
  };

  return (
    <>
      <header className={styles.header}>
        <div className="container">
          <div className={styles.inner}>
            <div className={styles.topRow}>
              <Link className="logo" to="/">
                GAME<span>GEAR</span>
              </Link>

              {!isAuthPage && (
                <div className={styles.quickActions} ref={navRef}>
                  <button
                    type="button"
                    className={styles.iconBtn}
                    title="Корзина"
                    onClick={() => dispatch(toggleCart())}
                  >
                    <ShoppingCart size={20} />
                    {cartCount > 0 && (
                      <span className={styles.badge}>{cartCount}</span>
                    )}
                  </button>

                  {user ? (
                    <NavLink
                      to="/profile/orders"
                      className={styles.iconBtn}
                      title="Профиль"
                    >
                      <User size={20} />
                    </NavLink>
                  ) : (
                    <NavLink to="/login" className={styles.iconBtn} title="Войти">
                      <User size={20} />
                    </NavLink>
                  )}

                  <button
                    type="button"
                    className={styles.iconBtn}
                    title="Меню"
                    aria-expanded={navOpen}
                    aria-haspopup="true"
                    onClick={() => setNavOpen((open) => !open)}
                  >
                    {navOpen ? <X size={20} /> : <Menu size={20} />}
                  </button>

                  {navOpen && (
                    <div className={styles.navMenu} role="menu">
                      <NavLink
                        to="/"
                        className={styles.navMenuLink}
                        role="menuitem"
                        onClick={() => setNavOpen(false)}
                      >
                        Главная
                      </NavLink>
                      <NavLink
                        to="/catalog"
                        className={styles.navMenuLink}
                        role="menuitem"
                        onClick={() => setNavOpen(false)}
                      >
                        Каталог
                      </NavLink>
                      {!user && (
                        <>
                          <NavLink
                            to="/login"
                            className={styles.navMenuLink}
                            role="menuitem"
                            onClick={() => setNavOpen(false)}
                          >
                            Вход
                          </NavLink>
                          <NavLink
                            to="/register"
                            className={`${styles.navMenuLink} ${styles.navMenuAccent}`}
                            role="menuitem"
                            onClick={() => setNavOpen(false)}
                          >
                            Регистрация
                          </NavLink>
                        </>
                      )}
                    </div>
                  )}
                </div>
              )}
            </div>

            {!isAuthPage && (
              <>
                <form className={styles.search} onSubmit={handleSearchSubmit}>
                  <Search size={18} className={styles.searchIcon} />
                  <input
                    type="search"
                    placeholder="Поиск девайсов..."
                    aria-label="Поиск"
                    value={searchQuery}
                    onChange={(event) => setSearchQuery(event.target.value)}
                  />
                </form>

                <div className={styles.desktopActions}>
                  {user && (
                    <>
                      <NavLink to="/" className={styles.navLink}>
                        Главная
                      </NavLink>
                      <NavLink to="/catalog" className={styles.navLink}>
                        Каталог
                      </NavLink>
                    </>
                  )}

                  {!user && (
                    <>
                      <NavLink to="/login" className={styles.navLink}>
                        Вход
                      </NavLink>
                      <NavLink to="/register" className={styles.navLinkAccent}>
                        Регистрация
                      </NavLink>
                    </>
                  )}

                  <button
                    type="button"
                    className={styles.iconBtn}
                    title="Корзина"
                    onClick={() => dispatch(toggleCart())}
                  >
                    <ShoppingCart size={20} />
                    {cartCount > 0 && (
                      <span className={styles.badge}>{cartCount}</span>
                    )}
                  </button>

                  {user ? (
                    <NavLink
                      to="/profile/orders"
                      className={styles.iconBtn}
                      title="Профиль"
                    >
                      <User size={20} />
                    </NavLink>
                  ) : (
                    <NavLink to="/login" className={styles.iconBtn} title="Войти">
                      <User size={20} />
                    </NavLink>
                  )}
                </div>
              </>
            )}
          </div>

          {notifications.length > 0 && !isAuthPage && (
            <div className={styles.notifications}>
              {notifications.slice(0, 3).map((notification) => (
                <div key={notification.id} className={styles.notificationItem}>
                  {notification.message}
                </div>
              ))}
            </div>
          )}
        </div>
      </header>
      <CartDrawer />
    </>
  );
}
