import { Navigate, Outlet } from "react-router-dom";
import { useAppSelector } from "../../app/hooks";
import { AccountSidebar } from "../AccountSidebar/AccountSidebar";
import { Footer } from "../Footer/Footer";
import { Header } from "../Header/Header";
import styles from "./ProfileLayout.module.css";

export function ProfileLayout() {
  const user = useAppSelector((state) => state.auth.user);

  if (!user) {
    return null;
  }

  return (
    <div className="page">
      <Header />
      <main className={`container ${styles.main}`}>
        <div className={styles.layout}>
          <AccountSidebar />
          <div className={styles.content}>
            <Outlet />
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}

export function ProfileIndexRedirect() {
  return <Navigate to="/profile/orders" replace />;
}
